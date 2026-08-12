package ansi

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	east "github.com/yuin/goldmark-emoji/ast"
	"github.com/yuin/goldmark/ast"
	astext "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// HyperlinkMode controls how hyperlinks are rendered.
type HyperlinkMode int

const (
	// HyperlinkModeAuto renders link text followed by the URL in parentheses,
	// with an OSC 8 hyperlink sequence applied to the URL.
	HyperlinkModeAuto HyperlinkMode = iota
	// HyperlinkModeInline renders only the link text, underlined, with an
	// OSC 8 hyperlink sequence applied to it. The URL is hidden.
	// Use this when the terminal is known to support OSC 8 hyperlinks.
	HyperlinkModeInline
)

// ImageProtocol is the graphics protocol used to render images inline.
type ImageProtocol int

const (
	// ImageProtocolNone renders images as styled text and links only.
	ImageProtocolNone ImageProtocol = iota
	// ImageProtocolKitty renders images using the Kitty graphics protocol.
	ImageProtocolKitty
	// ImageProtocolSixel renders images using the Sixel graphics format.
	ImageProtocolSixel
	// ImageProtocolKittyPlaceholders renders images using the Kitty
	// graphics protocol with Unicode placeholders. The image is transmitted
	// out-of-band (see [ANSIRenderer.GraphicsCommands]) and displayed via
	// Unicode placeholder characters that anchor the image to the text
	// grid. This is meant for full-screen TUI applications whose
	// cell-based renderers would drop raw graphics escape sequences, and
	// makes images move naturally with the text when scrolling, without
	// re-transmitting the image data.
	ImageProtocolKittyPlaceholders
)

// Options is used to configure an ANSIRenderer.
type Options struct {
	BaseURL          string
	WordWrap         int
	TableWrap        *bool
	InlineTableLinks bool
	PreserveNewLines bool
	Styles           StyleConfig
	ChromaFormatter  string
	HyperlinkMode    HyperlinkMode
	ImageProtocol    ImageProtocol

	// MaxImageColumns and MaxImageRows limit the number of terminal cells an
	// image may occupy, in addition to the constraints of the surrounding
	// blocks. Zero means no limit.
	MaxImageColumns int
	MaxImageRows    int
}

// ANSIRenderer renders markdown content as ANSI escaped sequences.
type ANSIRenderer struct { //nolint: revive
	context RenderContext
}

// NewRenderer returns a new ANSIRenderer with style and options set.
func NewRenderer(options Options) *ANSIRenderer {
	return &ANSIRenderer{
		context: NewRenderContext(options),
	}
}

// GraphicsCommands returns the out-of-band graphics protocol sequences
// (image transmission and placement commands) collected during the last
// render. Callers using [ImageProtocolKittyPlaceholders] must write these
// sequences to the terminal before displaying the rendered document, since
// the document itself only contains Unicode placeholders referencing them.
func (r *ANSIRenderer) GraphicsCommands() []string {
	return *r.context.graphicsCommands
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *ANSIRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks
	reg.Register(ast.KindDocument, r.renderNode)
	reg.Register(ast.KindHeading, r.renderNode)
	reg.Register(ast.KindBlockquote, r.renderNode)
	reg.Register(ast.KindCodeBlock, r.renderNode)
	reg.Register(ast.KindFencedCodeBlock, r.renderNode)
	reg.Register(ast.KindHTMLBlock, r.renderNode)
	reg.Register(ast.KindList, r.renderNode)
	reg.Register(ast.KindListItem, r.renderNode)
	reg.Register(ast.KindParagraph, r.renderNode)
	reg.Register(ast.KindTextBlock, r.renderNode)
	reg.Register(ast.KindThematicBreak, r.renderNode)

	// inlines
	reg.Register(ast.KindAutoLink, r.renderNode)
	reg.Register(ast.KindCodeSpan, r.renderNode)
	reg.Register(ast.KindEmphasis, r.renderNode)
	reg.Register(ast.KindImage, r.renderNode)
	reg.Register(ast.KindLink, r.renderNode)
	reg.Register(ast.KindRawHTML, r.renderNode)
	reg.Register(ast.KindText, r.renderNode)
	reg.Register(ast.KindString, r.renderNode)

	// tables
	reg.Register(astext.KindTable, r.renderNode)
	reg.Register(astext.KindTableHeader, r.renderNode)
	reg.Register(astext.KindTableRow, r.renderNode)
	reg.Register(astext.KindTableCell, r.renderNode)

	// definitions
	reg.Register(astext.KindDefinitionList, r.renderNode)
	reg.Register(astext.KindDefinitionTerm, r.renderNode)
	reg.Register(astext.KindDefinitionDescription, r.renderNode)

	// footnotes
	reg.Register(astext.KindFootnote, r.renderNode)
	reg.Register(astext.KindFootnoteList, r.renderNode)
	reg.Register(astext.KindFootnoteLink, r.renderNode)
	reg.Register(astext.KindFootnoteBacklink, r.renderNode)

	// checkboxes
	reg.Register(astext.KindTaskCheckBox, r.renderNode)

	// strikethrough
	reg.Register(astext.KindStrikethrough, r.renderNode)

	// emoji
	reg.Register(east.KindEmoji, r.renderNode)
}

func (r *ANSIRenderer) renderNode(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	writeTo := io.Writer(w)
	bs := r.context.blockStack

	// children get rendered by their parent
	if isChild(node) {
		return ast.WalkContinue, nil
	}

	e := r.NewElement(node, source)
	if entering { //nolint: nestif
		// everything below the Document element gets rendered into a block buffer
		if bs.Len() > 0 {
			writeTo = io.Writer(bs.Current().Block)
		}

		_, _ = io.WriteString(writeTo, e.Entering)
		if e.Renderer != nil {
			err := e.Renderer.Render(writeTo, r.context)
			if err != nil {
				return ast.WalkStop, fmt.Errorf("glamour: error rendering: %w", err)
			}
		}
	} else {
		// everything below the Document element gets rendered into a block buffer
		if bs.Len() > 0 {
			writeTo = io.Writer(bs.Parent().Block)
		}

		// if we're finished rendering the entire document,
		// flush to the real writer
		if node.Type() == ast.TypeDocument {
			writeTo = w
		}

		if e.Finisher != nil {
			err := e.Finisher.Finish(writeTo, r.context)
			if err != nil {
				return ast.WalkStop, fmt.Errorf("glamour: error finishing render: %w", err)
			}
		}

		// flush any remaining image sequences at the end of the document
		if node.Type() == ast.TypeDocument {
			if err := r.context.flushPendingImages(w); err != nil {
				return ast.WalkStop, err
			}
		}

		_, _ = io.WriteString(bs.Current().Block, e.Exiting)
	}

	return ast.WalkContinue, nil
}

func isChild(node ast.Node) bool {
	for n := node.Parent(); n != nil; n = n.Parent() {
		// These types are already rendered by their parent
		switch n.Kind() {
		case ast.KindCodeSpan, ast.KindAutoLink, ast.KindLink, ast.KindImage, ast.KindEmphasis, astext.KindStrikethrough, astext.KindTableCell:
			return true
		}
	}

	return false
}

func resolveRelativeURL(baseURL string, rel string) string {
	u, err := url.Parse(rel)
	if err != nil {
		return rel
	}
	if u.IsAbs() {
		return rel
	}
	u.Path = strings.TrimPrefix(u.Path, "/")

	base, err := url.Parse(baseURL)
	if err != nil {
		return rel
	}
	return base.ResolveReference(u).String()
}
