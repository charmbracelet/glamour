package ansi

import (
	"image/color"
	"math/rand"

	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const kittyPlaceholder = '\U0010EEEE'

func TestImageProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol ImageProtocol
	}{
		{"kitty", ImageProtocolKitty},
		{"sixel", ImageProtocolSixel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			imgPath, err := filepath.Abs(filepath.Join("testdata", "TestImageProtocol", "test.png"))
			if err != nil {
				t.Fatal(err)
			}

			b, err := os.ReadFile("../styles/dark.json")
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap:      80,
				ImageProtocol: tc.protocol,
			}
			if err := json.Unmarshal(b, &options.Styles); err != nil {
				t.Fatal(err)
			}

			md := goldmark.New(
				goldmark.WithExtensions(
					extension.GFM,
					extension.DefinitionList,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
			)

			ar := NewRenderer(options)
			md.SetRenderer(
				renderer.NewRenderer(
					renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

			in := "![Test Image](" + imgPath + ")"
			var buf bytes.Buffer
			if err := md.Convert([]byte(in), &buf); err != nil {
				t.Fatal(err)
			}

			out := buf.String()
			if tc.protocol == ImageProtocolKitty && !bytes.Contains(buf.Bytes(), []byte("\x1b_G")) {
				t.Errorf("expected kitty graphics sequence, got: %q", out)
			}
			if tc.protocol == ImageProtocolSixel && !bytes.Contains(buf.Bytes(), []byte("\x1bP")) {
				t.Errorf("expected sixel graphics sequence, got: %q", out)
			}
		})
	}
}

func TestImageProtocolKittyPlaceholders(t *testing.T) {
	imgPath, err := filepath.Abs(filepath.Join("testdata", "TestImageProtocol", "test.png"))
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile("../styles/dark.json")
	if err != nil {
		t.Fatal(err)
	}

	options := Options{
		WordWrap:      80,
		ImageProtocol: ImageProtocolKittyPlaceholders,
	}
	if err := json.Unmarshal(b, &options.Styles); err != nil {
		t.Fatal(err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	ar := NewRenderer(options)
	md.SetRenderer(
		renderer.NewRenderer(
			renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

	in := "![Test Image](" + imgPath + ")"
	var buf bytes.Buffer
	if err := md.Convert([]byte(in), &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	// The document itself must only contain the placeholder grid, not the
	// image payload; cell-based TUI renderers would drop raw APC sequences.
	if bytes.Contains(buf.Bytes(), []byte("\x1b_G")) {
		t.Errorf("expected no inline kitty graphics sequence, got: %q", out)
	}
	if !strings.ContainsRune(out, 0x10EEEE) {
		t.Errorf("expected unicode placeholders in document, got: %q", out)
	}

	// The transmission and virtual placement commands must be exposed for
	// out-of-band output.
	cmds := ar.GraphicsCommands()
	if len(cmds) != 2 {
		t.Fatalf("expected transmit and place commands, got %d: %q", len(cmds), cmds)
	}
	if !strings.Contains(cmds[0], "f=100") || !strings.Contains(cmds[0], "i="+strconv.Itoa(imageID(imgPath))) {
		t.Errorf("unexpected transmit command: %q", cmds[0])
	}
	if !strings.Contains(cmds[1], "a=p") || !strings.Contains(cmds[1], "U=1") {
		t.Errorf("unexpected place command: %q", cmds[1])
	}

	// The grid must be sized to the width available inside the surrounding
	// block (the document has a margin of 2, i.e. 4 cells), so that it is
	// never wrapped by the block, which would chop up the image.
	if !strings.Contains(cmds[1], "c=76") {
		t.Errorf("expected image constrained to the block width, got: %q", cmds[1])
	}

	// Only the first cell of each row carries the row diacritic; the terminal
	// infers the columns of the remaining cells, which keeps the document
	// small. Each row must repeat the image ID in its foreground color, as
	// style handling in wrapping writers can otherwise recolor the
	// placeholders.
	lines := strings.Split(strings.TrimPrefix(out, "\n"), "\n")
	var gridLines []string
	for _, l := range lines {
		if strings.ContainsRune(l, 0x10EEEE) {
			gridLines = append(gridLines, l)
		}
	}
	if len(gridLines) == 0 {
		t.Fatalf("expected placeholder grid in document, got: %q", out)
	}
	for _, l := range gridLines {
		ph := string(kittyPlaceholder)
		want := "\x1b[38;2;" +
			strconv.Itoa(imageID(imgPath)>>16&0xff) + ";" +
			strconv.Itoa(imageID(imgPath)>>8&0xff) + ";" +
			strconv.Itoa(imageID(imgPath)&0xff) + "m" + ph
		if !strings.HasPrefix(strings.TrimLeft(l, " "), want) {
			t.Errorf("expected grid row to start with the image id color and a placeholder, got: %q", l[:min(80, len(l))])
		}
		if strings.Count(l, ph) != 76 {
			t.Errorf("expected 76 placeholder cells, got %d", strings.Count(l, ph))
		}
		if strings.Count(l, "\u0305")+strings.Count(l, "\u030D") > 2 {
			t.Errorf("expected at most one row diacritic on the first cells, got: %q", l[:min(80, len(l))])
		}
	}
}

func TestImageDisplaySize(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		opts          Options
		wantCols      int
		wantRows      int
	}{
		{
			name:  "narrow image keeps natural size",
			width: 40, height: 20,
			opts:     Options{WordWrap: 80},
			wantCols: 40, wantRows: 10,
		},
		{
			name:  "wide image is constrained to the available width",
			width: 160, height: 80,
			opts:     Options{WordWrap: 80},
			wantCols: 80, wantRows: 20,
		},
		{
			name:  "max columns",
			width: 160, height: 80,
			opts:     Options{WordWrap: 80, MaxImageColumns: 40},
			wantCols: 40, wantRows: 10,
		},
		{
			name:  "max rows preserves the aspect ratio",
			width: 160, height: 160,
			opts:     Options{WordWrap: 80, MaxImageRows: 10},
			wantCols: 20, wantRows: 10,
		},
		{
			name:  "zero word wrap falls back to the natural size",
			width: 120, height: 40,
			opts:     Options{},
			wantCols: 120, wantRows: 20,
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewRenderContext(tc.opts)
			cols, rows := imageDisplaySize(tc.width, tc.height, ctx)
			if cols != tc.wantCols || rows != tc.wantRows {
				t.Errorf("expected %dx%d cells, got %dx%d", tc.wantCols, tc.wantRows, cols, rows)
			}
		})
	}
}

func TestDownscaleImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1000, 500))

	scaled := downscaleImage(img, 200, 200)
	if b := scaled.Bounds(); b.Dx() != 200 || b.Dy() != 100 {
		t.Errorf("expected 200x100, got %dx%d", b.Dx(), b.Dy())
	}

	// Images that fit are returned unchanged.
	if got := downscaleImage(img, 2000, 2000); got != img {
		t.Error("expected image to be returned unchanged")
	}
}

func TestWriteKittyImageTransmission(t *testing.T) {
	// A PNG that doesn't need resizing must be transmitted by path, without
	// being decoded or re-encoded.
	path := filepath.Join(t.TempDir(), "small.png")
	small := image.NewRGBA(image.Rect(0, 0, 100, 50))
	writePNG(t, path, small)
	config, err := loadImageConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := &kitty.Options{Action: kitty.Transmit, Format: kitty.PNG, Transmission: kitty.Direct, ID: 1, Quite: 2}
	if err := writeKittyImage(&buf, path, config, 10, 5, opts); err != nil {
		t.Fatal(err)
	}
	seq := buf.String()
	if !strings.Contains(seq, "t=f") || !strings.Contains(seq, "f=100") {
		t.Errorf("expected file transmission, got: %q", seq)
	}

	// A PNG larger than its display box must be downscaled and transmitted
	// directly, in chunks.
	path2 := filepath.Join(t.TempDir(), "large.png")
	writePNG(t, path2, noiseImage(512, 512))
	config2, err := loadImageConfig(path2)
	if err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	opts2 := &kitty.Options{Action: kitty.Transmit, Format: kitty.PNG, Transmission: kitty.Direct, ID: 2, Quite: 2}
	if err := writeKittyImage(&buf, path2, config2, 10, 5, opts2); err != nil {
		t.Fatal(err)
	}
	seq2 := buf.String()
	if strings.Contains(seq2, "t=f") {
		t.Errorf("expected direct transmission of downscaled image, got: %q", seq2[:min(80, len(seq2))])
	}
	if !strings.Contains(seq2, "m=1;") || !strings.Contains(seq2, "m=0;") {
		t.Errorf("expected chunked transmission, got: %q", seq2[:min(80, len(seq2))])
	}
}

func TestLocalImagePath(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"plain path", "/tmp/x/test.png", "/tmp/x/test.png"},
		{"file url", "file:///tmp/x/test.png", "/tmp/x/test.png"},
		{"windows drive", "file:///C:/Users/x/test.png", "C:/Users/x/test.png"},
		{"file url with spaces", "file:///tmp/my%20docs/test.png", "/tmp/my docs/test.png"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := filepath.ToSlash(localImagePath(tc.url)); got != tc.want {
				t.Errorf("localImagePath(%q) = %q, want %q", tc.url, got, tc.want)
			}
		})
	}
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() //nolint:errcheck
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

// noiseImage returns a random noise image, which doesn't compress well, so
// its encoded size is dominated by its dimensions.
func noiseImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rnd := rand.New(rand.NewSource(1)) //nolint:gosec
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(rnd.Intn(256)), G: uint8(rnd.Intn(256)), B: uint8(rnd.Intn(256)), A: 255})
		}
	}
	return img
}
