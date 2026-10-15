package ansi

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// AlertKind is the type of a GitHub-style alert.
type AlertKind string

// The alert types GitHub recognizes. See
// https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#alerts
const (
	AlertNote      AlertKind = "note"
	AlertTip       AlertKind = "tip"
	AlertImportant AlertKind = "important"
	AlertWarning   AlertKind = "warning"
	AlertCaution   AlertKind = "caution"
)

// Title returns the label GitHub displays for the alert, e.g. "Note".
func (k AlertKind) Title() string {
	switch k {
	case AlertNote:
		return "Note"
	case AlertTip:
		return "Tip"
	case AlertImportant:
		return "Important"
	case AlertWarning:
		return "Warning"
	case AlertCaution:
		return "Caution"
	}
	return string(k)
}

// alertAttribute marks a blockquote as an alert. Its value is an AlertKind.
const alertAttribute = "glamour-alert"

var (
	// alertMatcher matches the first line of an alert blockquote, e.g. [!NOTE].
	// Like GitHub, the marker must be the only thing on the line, and the type
	// must be uppercase.
	alertMatcher      = regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]$`)
	alertMarkerPrefix = []byte("[!")
)

// AlertKindFromNode returns the kind of alert the given node was tagged with,
// if any.
func AlertKindFromNode(node ast.Node) (AlertKind, bool) {
	v, ok := node.AttributeString(alertAttribute)
	if !ok {
		return "", false
	}
	kind, ok := v.(AlertKind)
	return kind, ok
}

// An AlertTransformer is a [parser.ParagraphTransformer] that detects GitHub
// alerts: blockquotes whose first line is a marker such as [!NOTE].
//
// Detected markers are removed from the document, and the enclosing blockquote
// is tagged with the alert's [AlertKind], which the ANSI renderer turns into a
// callout. It runs before inline parsing, so the marker never becomes part of
// the rendered output.
type AlertTransformer struct{}

// NewAlertTransformer returns a new AlertTransformer.
func NewAlertTransformer() *AlertTransformer { return &AlertTransformer{} }

// Transform implements [parser.ParagraphTransformer].
func (*AlertTransformer) Transform(node *ast.Paragraph, reader text.Reader, _ parser.Context) {
	blockquote, ok := node.Parent().(*ast.Blockquote)
	if !ok {
		return
	}
	// Only the first line of the blockquote can open an alert.
	if node.PreviousSibling() != nil {
		return
	}
	if _, ok := blockquote.AttributeString(alertAttribute); ok {
		return
	}

	lines := node.Lines()
	if lines.Len() == 0 {
		return
	}

	// Blockquote markers have already been stripped by the block parser, so
	// this is the line as the author wrote it, minus the ">".
	first := lines.At(0)
	line := bytes.TrimSpace(first.Value(reader.Source()))
	if !bytes.HasPrefix(line, alertMarkerPrefix) {
		return
	}

	match := alertMatcher.FindSubmatch(line)
	if match == nil {
		return
	}

	blockquote.SetAttributeString(alertAttribute, AlertKind(strings.ToLower(string(match[1]))))

	if lines.Len() == 1 {
		// The marker was the paragraph's only content, so the paragraph has
		// nothing left to render. Removing it here is supported: the parser
		// skips closing paragraphs that a transformer has detached.
		blockquote.RemoveChild(blockquote, node)
		return
	}

	// Drop the marker line and keep the rest of the paragraph.
	rest := lines.Sliced(1, lines.Len())
	lines.SetSliced(0, 0)
	lines.AppendAll(rest)
	node.SetLines(lines)
}

// An AlertElement renders a GitHub-style alert: a styled title line followed by
// the alert's contents.
type AlertElement struct {
	BlockElement
	Kind  AlertKind
	Title StylePrimitive
}

// Render renders an AlertElement.
func (e *AlertElement) Render(w io.Writer, ctx RenderContext) error {
	if err := e.BlockElement.Render(w, ctx); err != nil {
		return err
	}

	bs := ctx.blockStack
	title := BaseElement{Token: e.Kind.Title(), Style: e.Title}
	if err := title.Render(bs.Current().Block, ctx); err != nil {
		return err
	}

	if _, err := io.WriteString(bs.Current().Block, "\n"); err != nil {
		return fmt.Errorf("glamour: error writing to writer: %w", err)
	}
	return nil
}

// alertStyle returns the style of the given alert kind, falling back to the
// blockquote style for anything the alert style leaves unset.
func alertStyle(ctx RenderContext, kind AlertKind) (StyleBlock, StylePrimitive) {
	parent := ctx.blockStack.Current().Style
	quote := cascadeStyle(parent, ctx.options.Styles.BlockQuote, false)
	alert := ctx.options.Styles.Alerts.For(kind)

	style := quote
	style.StylePrimitive = cascadeStylePrimitive(quote.StylePrimitive, alert.StylePrimitive, false)
	if alert.Indent != nil {
		style.Indent = alert.Indent
	}
	if alert.IndentToken != nil {
		style.IndentToken = alert.IndentToken
	}
	if alert.Margin != nil {
		style.Margin = alert.Margin
	}
	if alert.IndentTokenStyle != nil {
		style.IndentTokenStyle = alert.IndentTokenStyle
	}

	return style, cascadeStylePrimitive(style.StylePrimitive, alert.Title, false)
}
