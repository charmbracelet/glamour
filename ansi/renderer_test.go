package ansi

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const (
	examplesDir = "../styles/examples/"
	issuesDir   = "../testdata/issues/"
)

func TestRenderer(t *testing.T) {
	files, err := filepath.Glob(examplesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(bn, func(t *testing.T) {
			sn := filepath.Join(examplesDir, bn+".style")

			in, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(sn)
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap:     80,
				ColorProfile: termenv.TrueColor,
			}
			err = json.Unmarshal(b, &options.Styles)
			if err != nil {
				t.Fatal(err)
			}

			switch bn {
			case "table_wrap":
				tableWrap := true
				options.TableWrap = &tableWrap
			case "table_truncate":
				tableWrap := false
				options.TableWrap = &tableWrap
			case "table_with_inline_links":
				options.InlineTableLinks = true
			case "table_with_footer_links", "table_with_footer_links_no_color":
				options.InlineTableLinks = false
			}

			md := goldmark.New(
				goldmark.WithExtensions(
					extension.GFM,
					extension.DefinitionList,
					emoji.Emoji,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
			)

			ar := NewRenderer(options)
			md.SetRenderer(
				renderer.NewRenderer(
					renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

			var buf bytes.Buffer
			if err := md.Convert(in, &buf); err != nil {
				t.Error(err)
			}

			golden.RequireEqual(t, buf.Bytes())
		})
	}
}

func TestRendererIssues(t *testing.T) {
	files, err := filepath.Glob(issuesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(bn, func(t *testing.T) {
			in, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile("../styles/dark.json")
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap:     80,
				ColorProfile: termenv.TrueColor,
			}
			err = json.Unmarshal(b, &options.Styles)
			if err != nil {
				t.Fatal(err)
			}
			if bn == "493" {
				tableWrap := false
				options.TableWrap = &tableWrap
			}

			md := goldmark.New(
				goldmark.WithExtensions(
					extension.GFM,
					extension.DefinitionList,
					emoji.Emoji,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
			)

			ar := NewRenderer(options)
			md.SetRenderer(
				renderer.NewRenderer(
					renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

			var buf bytes.Buffer
			if err := md.Convert(in, &buf); err != nil {
				t.Error(err)
			}

			golden.RequireEqual(t, buf.Bytes())
		})
	}
}

func TestRenderContextImageStorage(t *testing.T) {
	ctx := NewRenderContext(Options{})

	// Initially no images
	if ctx.HasImages() {
		t.Error("new context should have no images")
	}

	// Store an image
	imageData := "\x1b_Gtest\x1b\\"
	placeholder := ctx.StoreImage(imageData)

	// Should now have images
	if !ctx.HasImages() {
		t.Error("context should have images after StoreImage")
	}

	// Placeholder should have expected format
	if !strings.HasPrefix(placeholder, imagePlaceholderPrefix) {
		t.Errorf("placeholder should start with prefix, got %q", placeholder)
	}
	if !strings.HasSuffix(placeholder, imagePlaceholderSuffix) {
		t.Errorf("placeholder should end with suffix, got %q", placeholder)
	}

	// Store another image
	imageData2 := "\x1b_Gtest2\x1b\\"
	placeholder2 := ctx.StoreImage(imageData2)

	// Placeholders should be unique
	if placeholder == placeholder2 {
		t.Error("each placeholder should be unique")
	}
}

func TestRenderContextReplaceImagePlaceholders(t *testing.T) {
	ctx := NewRenderContext(Options{})

	imageData := "IMAGE_DATA_HERE"
	placeholder := ctx.StoreImage(imageData)

	// Test replacement
	content := "before " + placeholder + " after"
	result := ctx.ReplaceImagePlaceholders(content)

	expected := "before " + imageData + " after"
	if result != expected {
		t.Errorf("ReplaceImagePlaceholders() = %q, want %q", result, expected)
	}

	// Test with unknown placeholder (should remain unchanged)
	unknownContent := "text with " + imagePlaceholderPrefix + "9999" + imagePlaceholderSuffix + " unknown"
	unknownResult := ctx.ReplaceImagePlaceholders(unknownContent)
	// Unknown placeholders are not replaced (not in data map)
	if !strings.Contains(unknownResult, imagePlaceholderPrefix+"9999"+imagePlaceholderSuffix) {
		t.Error("unknown placeholders should remain unchanged")
	}

	// Test with no placeholders
	noPlaceholders := "just regular text"
	noResult := ctx.ReplaceImagePlaceholders(noPlaceholders)
	if noResult != noPlaceholders {
		t.Error("text without placeholders should be unchanged")
	}
}

func TestRenderContextWriteWithImageReplacement(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		marginPrefix string
		wantRaw      string
		wantMargin   string
	}{
		{
			name:         "no placeholders",
			content:      "just text",
			marginPrefix: "",
			wantRaw:      "",
			wantMargin:   "just text",
		},
		{
			name:         "text with margin prefix (no images)",
			content:      "text content",
			marginPrefix: "  ",
			wantRaw:      "",
			wantMargin:   "text content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewRenderContext(Options{})

			var rawBuf, marginBuf bytes.Buffer
			err := ctx.WriteWithImageReplacement(&rawBuf, &marginBuf, tt.content, tt.marginPrefix)
			if err != nil {
				t.Fatalf("WriteWithImageReplacement() error = %v", err)
			}

			if rawBuf.String() != tt.wantRaw {
				t.Errorf("raw output = %q, want %q", rawBuf.String(), tt.wantRaw)
			}
			if marginBuf.String() != tt.wantMargin {
				t.Errorf("margin output = %q, want %q", marginBuf.String(), tt.wantMargin)
			}
		})
	}
}

func TestRenderContextWriteWithImageReplacementWithImages(t *testing.T) {
	ctx := NewRenderContext(Options{})

	imageData := "IMG_ESCAPE_SEQ"
	placeholder := ctx.StoreImage(imageData)

	var rawBuf, marginBuf bytes.Buffer
	content := "before " + placeholder + " after"
	marginPrefix := "> "

	err := ctx.WriteWithImageReplacement(&rawBuf, &marginBuf, content, marginPrefix)
	if err != nil {
		t.Fatalf("WriteWithImageReplacement() error = %v", err)
	}

	// Image data should go to raw output with margin prefix
	if !strings.Contains(rawBuf.String(), marginPrefix+imageData) {
		t.Errorf("raw output should contain margin prefix + image data, got %q", rawBuf.String())
	}

	// Text before/after should go to margin writer
	if !strings.Contains(marginBuf.String(), "before ") {
		t.Errorf("margin output should contain 'before ', got %q", marginBuf.String())
	}
	if !strings.Contains(marginBuf.String(), " after") {
		t.Errorf("margin output should contain ' after', got %q", marginBuf.String())
	}
}

func TestRenderContextWriteWithImageReplacementMalformed(t *testing.T) {
	ctx := NewRenderContext(Options{})

	var rawBuf, marginBuf bytes.Buffer

	// Malformed placeholder (prefix without suffix)
	malformed := "text " + imagePlaceholderPrefix + "no suffix here"

	err := ctx.WriteWithImageReplacement(&rawBuf, &marginBuf, malformed, "")
	if err != nil {
		t.Fatalf("WriteWithImageReplacement() error = %v", err)
	}

	// Should handle gracefully - write remaining text
	combined := rawBuf.String() + marginBuf.String()
	if !strings.Contains(combined, "text ") {
		t.Error("should still output text before malformed placeholder")
	}
}

func TestProtocolToImages(t *testing.T) {
	tests := []struct {
		input    ImageProtocol
		expected string
	}{
		{ImageProtocolAuto, "auto"},
		{ImageProtocolKitty, "kitty"},
		{ImageProtocolSixel, "sixel"},
		{ImageProtocolITerm, "iterm"},
		{ImageProtocolNone, "none"},
		{"unknown", "none"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := protocolToImages(tt.input)
			if string(result) != tt.expected {
				t.Errorf("protocolToImages(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestImageProtocolConstants(t *testing.T) {
	// Verify constants have expected values
	if ImageProtocolNone != "none" {
		t.Errorf("ImageProtocolNone = %q, want %q", ImageProtocolNone, "none")
	}
	if ImageProtocolAuto != "auto" {
		t.Errorf("ImageProtocolAuto = %q, want %q", ImageProtocolAuto, "auto")
	}
	if ImageProtocolKitty != "kitty" {
		t.Errorf("ImageProtocolKitty = %q, want %q", ImageProtocolKitty, "kitty")
	}
	if ImageProtocolSixel != "sixel" {
		t.Errorf("ImageProtocolSixel = %q, want %q", ImageProtocolSixel, "sixel")
	}
	if ImageProtocolITerm != "iterm" {
		t.Errorf("ImageProtocolITerm = %q, want %q", ImageProtocolITerm, "iterm")
	}
}
