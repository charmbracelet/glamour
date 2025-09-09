// Package glamour lets you render markdown documents & templates on ANSI
// compatible terminals. You can create your own stylesheet or simply use one of
// the stylish defaults
package glamour

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/muesli/termenv"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"golang.org/x/term"

	"github.com/charmbracelet/glamour/ansi"
	styles "github.com/charmbracelet/glamour/styles"
)

const (
	defaultWidth = 80
	highPriority = 1000
)

// A TermRendererOption sets an option on a TermRenderer.
type TermRendererOption func(*TermRenderer) error

// TermRenderer can be used to render markdown content, posing a depth of
// customization and styles to fit your needs.
type TermRenderer struct {
	md          goldmark.Markdown
	ansiOptions ansi.Options
	buf         bytes.Buffer
	renderBuf   bytes.Buffer
}

// Render initializes a new TermRenderer and renders a markdown with a specific
// style.
func Render(in string, stylePath string) (string, error) {
	b, err := RenderBytes([]byte(in), stylePath)
	return string(b), err
}

// RenderWithEnvironmentConfig initializes a new TermRenderer and renders a
// markdown with a specific style defined by the GLAMOUR_STYLE environment variable.
func RenderWithEnvironmentConfig(in string) (string, error) {
	b, err := RenderBytes([]byte(in), getEnvironmentStyle())
	return string(b), err
}

// RenderBytes initializes a new TermRenderer and renders a markdown with a
// specific style.
func RenderBytes(in []byte, stylePath string) ([]byte, error) {
	r, err := NewTermRenderer(
		WithStylePath(stylePath),
	)
	if err != nil {
		return nil, err
	}
	return r.RenderBytes(in)
}

// NewTermRenderer returns a new TermRenderer the given options.
func NewTermRenderer(options ...TermRendererOption) (*TermRenderer, error) {
	tr := &TermRenderer{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.DefinitionList,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		),
		ansiOptions: ansi.Options{
			WordWrap:      defaultWidth,
			ColorProfile:  termenv.TrueColor,
			LinkFormatter: ansi.DefaultFormatter, // Ensure consistent rendering path
		},
	}
	for _, o := range options {
		if err := o(tr); err != nil {
			return nil, err
		}
	}
	ar := ansi.NewRenderer(tr.ansiOptions)
	tr.md.SetRenderer(
		renderer.NewRenderer(
			renderer.WithNodeRenderers(
				util.Prioritized(ar, highPriority),
			),
		),
	)
	return tr, nil
}

// WithBaseURL sets a TermRenderer's base URL.
func WithBaseURL(baseURL string) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.BaseURL = baseURL
		return nil
	}
}

// WithColorProfile sets the TermRenderer's color profile
// (TrueColor / ANSI256 / ANSI).
func WithColorProfile(profile termenv.Profile) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.ColorProfile = profile
		return nil
	}
}

// WithStandardStyle sets a TermRenderer's styles with a standard (builtin)
// style.
func WithStandardStyle(style string) TermRendererOption {
	return func(tr *TermRenderer) error {
		styles, err := getDefaultStyle(style)
		if err != nil {
			return err
		}
		tr.ansiOptions.Styles = *styles
		return nil
	}
}

// WithAutoStyle sets a TermRenderer's styles with either the standard dark
// or light style, depending on the terminal's background color at run-time.
func WithAutoStyle() TermRendererOption {
	return WithStandardStyle(styles.AutoStyle)
}

// WithEnvironmentConfig sets a TermRenderer's styles based on the
// GLAMOUR_STYLE environment variable.
func WithEnvironmentConfig() TermRendererOption {
	return WithStylePath(getEnvironmentStyle())
}

// WithStylePath sets a TermRenderer's style from stylePath. stylePath is first
// interpreted as a filename. If no such file exists, it is re-interpreted as a
// standard style.
func WithStylePath(stylePath string) TermRendererOption {
	return func(tr *TermRenderer) error {
		styles, err := getDefaultStyle(stylePath)
		if err != nil {
			jsonBytes, err := os.ReadFile(stylePath)
			if err != nil {
				return fmt.Errorf("glamour: error reading file: %w", err)
			}

			return json.Unmarshal(jsonBytes, &tr.ansiOptions.Styles)
		}
		tr.ansiOptions.Styles = *styles
		return nil
	}
}

// WithStyles sets a TermRenderer's styles.
func WithStyles(styles ansi.StyleConfig) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.Styles = styles
		return nil
	}
}

// WithStylesFromJSONBytes sets a TermRenderer's styles by parsing styles from
// jsonBytes.
func WithStylesFromJSONBytes(jsonBytes []byte) TermRendererOption {
	return func(tr *TermRenderer) error {
		return json.Unmarshal(jsonBytes, &tr.ansiOptions.Styles)
	}
}

// WithStylesFromJSONFile sets a TermRenderer's styles from a JSON file.
func WithStylesFromJSONFile(filename string) TermRendererOption {
	return func(tr *TermRenderer) error {
		jsonBytes, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("glamour: error reading file: %w", err)
		}
		return json.Unmarshal(jsonBytes, &tr.ansiOptions.Styles)
	}
}

// WithWordWrap sets a TermRenderer's word wrap.
func WithWordWrap(wordWrap int) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.WordWrap = wordWrap
		return nil
	}
}

// WithTableWrap controls whether table content will wrap if too long.
// This is true by default. If false, table content will be truncated with an
// ellipsis if too long to fit.
func WithTableWrap(tableWrap bool) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.TableWrap = &tableWrap
		return nil
	}
}

// WithInlineTableLinks forces tables to render links inline. By default,links
// are rendered as a list of links at the bottom of the table.
func WithInlineTableLinks(inlineTableLinks bool) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.InlineTableLinks = inlineTableLinks
		return nil
	}
}

// WithPreservedNewLines preserves newlines from being replaced.
func WithPreservedNewLines() TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.PreserveNewLines = true
		return nil
	}
}

// WithEmoji sets a TermRenderer's emoji rendering.
func WithEmoji() TermRendererOption {
	return func(tr *TermRenderer) error {
		emoji.New().Extend(tr.md)
		return nil
	}
}

// WithChromaFormatter sets a TermRenderer's chroma formatter used for code blocks.
func WithChromaFormatter(formatter string) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.ChromaFormatter = formatter
		return nil
	}
}

// WithOptions sets multiple TermRenderer options within a single TermRendererOption.
func WithOptions(options ...TermRendererOption) TermRendererOption {
	return func(tr *TermRenderer) error {
		for _, o := range options {
			if err := o(tr); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithLinkFormatter sets a TermRenderer's custom link formatter.
//
// Link formatters control how markdown links are rendered in the terminal output.
// When set to nil (default), the standard Glamour link formatting is used.
//
// Example:
//
//	// Use a custom formatter
//	customFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
//		return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
//	})
//	renderer, err := glamour.NewTermRenderer(
//		glamour.WithLinkFormatter(customFormatter),
//	)
//
// See the ansi package documentation for available built-in formatters and
// examples of creating custom formatters.
func WithLinkFormatter(formatter ansi.LinkFormatter) TermRendererOption {
	return func(tr *TermRenderer) error {
		tr.ansiOptions.LinkFormatter = formatter
		return nil
	}
}

// WithTextOnlyLinks configures the TermRenderer to show only link text.
//
// This formatter renders links showing only their text content, hiding URLs.
// In terminals that support OSC 8 hyperlinks, the text becomes clickable.
// In other terminals, only the styled text is displayed.
//
// This is useful when you want a cleaner appearance without visible URLs,
// especially in terminals that support hyperlink functionality.
//
// Example:
//
//	renderer, err := glamour.NewTermRenderer(
//		glamour.WithTextOnlyLinks(),
//	)
//
// Terminal compatibility:
// - Terminals with OSC 8 support: Clickable text links
// - Other terminals: Text only, URLs hidden
func WithTextOnlyLinks() TermRendererOption {
	return WithLinkFormatter(ansi.TextOnlyFormatter)
}

// WithURLOnlyLinks configures the TermRenderer to show only URLs.
//
// This formatter renders links showing only the URL, hiding link text.
// This is useful when space is limited or when the URL itself is more
// important than descriptive text.
//
// Fragment-only URLs (like "#section") are not displayed as they typically
// refer to document anchors rather than external resources.
//
// Example:
//
//	renderer, err := glamour.NewTermRenderer(
//		glamour.WithURLOnlyLinks(),
//	)
//
// Terminal compatibility: Works in all terminals
func WithURLOnlyLinks() TermRendererOption {
	return WithLinkFormatter(ansi.URLOnlyFormatter)
}

// WithHyperlinks configures the TermRenderer to use OSC 8 hyperlinks.
//
// This formatter renders links as OSC 8 hyperlinks, making the link text
// clickable in supporting terminals while keeping URLs hidden from view.
// This provides the cleanest visual appearance when terminal support is available.
//
// WARNING: This formatter does not provide fallback for terminals without
// OSC 8 support, which may result in escape sequences being displayed.
// Use WithSmartHyperlinks() for automatic fallback behavior.
//
// Example:
//
//	renderer, err := glamour.NewTermRenderer(
//		glamour.WithHyperlinks(),
//	)
//
// Terminal compatibility:
// - Modern terminals (iTerm2, Windows Terminal, etc.): Clickable hyperlinks
// - Legacy terminals: May display escape sequences
func WithHyperlinks() TermRendererOption {
	return WithLinkFormatter(ansi.HyperlinkFormatter)
}

// WithSmartHyperlinks configures the TermRenderer with intelligent hyperlink handling.
//
// This formatter automatically detects terminal hyperlink support:
// - In terminals that support OSC 8: Shows clickable hyperlinks (text only)
// - In other terminals: Falls back to standard "text url" format
//
// This provides the best user experience across different terminal environments
// by combining modern hyperlink capabilities with universal fallback support.
//
// Example:
//
//	renderer, err := glamour.NewTermRenderer(
//		glamour.WithSmartHyperlinks(),
//	)
//
// Terminal compatibility:
// - Modern terminals: Clickable hyperlinks with hidden URLs
// - Legacy terminals: Standard "text url" format
// - All terminals: Graceful degradation
func WithSmartHyperlinks() TermRendererOption {
	return WithLinkFormatter(ansi.SmartHyperlinkFormatter)
}

func (tr *TermRenderer) Read(b []byte) (int, error) {
	n, err := tr.renderBuf.Read(b)
	if err == io.EOF {
		return n, io.EOF
	}
	if err != nil {
		return 0, fmt.Errorf("glamour: error reading from buffer: %w", err)
	}
	return n, nil
}

func (tr *TermRenderer) Write(b []byte) (int, error) {
	n, err := tr.buf.Write(b)
	if err != nil {
		return 0, fmt.Errorf("glamour: error writing bytes: %w", err)
	}
	return n, nil
}

// Close must be called after writing to TermRenderer. You can then retrieve
// the rendered markdown by calling Read.
func (tr *TermRenderer) Close() error {
	err := tr.md.Convert(tr.buf.Bytes(), &tr.renderBuf)
	if err != nil {
		return fmt.Errorf("glamour: error converting markdown: %w", err)
	}

	tr.buf.Reset()
	return nil
}

// Render returns the markdown rendered into a string.
func (tr *TermRenderer) Render(in string) (string, error) {
	b, err := tr.RenderBytes([]byte(in))
	return string(b), err
}

// RenderBytes returns the markdown rendered into a byte slice.
func (tr *TermRenderer) RenderBytes(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := tr.md.Convert(in, &buf)
	return buf.Bytes(), err
}

func getEnvironmentStyle() string {
	glamourStyle := os.Getenv("GLAMOUR_STYLE")
	if len(glamourStyle) == 0 {
		glamourStyle = styles.AutoStyle
	}

	return glamourStyle
}

func getDefaultStyle(style string) (*ansi.StyleConfig, error) {
	if style == styles.AutoStyle {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			return &styles.NoTTYStyleConfig, nil
		}
		if termenv.HasDarkBackground() {
			return &styles.DarkStyleConfig, nil
		}
		return &styles.LightStyleConfig, nil
	}

	styles, ok := styles.DefaultStyles[style]
	if !ok {
		return nil, fmt.Errorf("%s: style not found", style)
	}
	return styles, nil
}
