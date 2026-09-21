package ansi

import (
	"image/color"
	"math/rand"

	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/gif"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const kittyPlaceholder = '\U0010EEEE'

// escapeSequencePattern matches ANSI SGR, OSC 8 hyperlink, and graphics
// protocol sequences.
var escapeSequencePattern = regexp.MustCompile(`\x1b\]8;[^\x07\x9c]*(\x07|\x9c)|\x1b\[[0-9;:]*[a-zA-Z]|\x1b_G[^\x1b]*\x1b\\|\x1bP[^\x1b]*\x1b\\`)

// visibleText returns the text a terminal would display for the given
// rendered output, i.e. with all escape sequences removed.
func visibleText(s string) string {
	return escapeSequencePattern.ReplaceAllString(s, "")
}

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
	// File transmission depends on the terminal being on the same machine,
	// which the environment the tests run in must not influence.
	t.Setenv("SSH_TTY", "")
	t.Setenv("SSH_CONNECTION", "")
	ctx := NewRenderContext(Options{})

	// A PNG that doesn't need resizing must be transmitted by path, without
	// being decoded or re-encoded.
	path := filepath.Join(t.TempDir(), "small.png")
	small := image.NewRGBA(image.Rect(0, 0, 100, 50))
	writePNG(t, path, small)
	config, err := loadImageConfig(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := &kitty.Options{Action: kitty.Transmit, Format: kitty.PNG, Transmission: kitty.Direct, ID: 1, Quite: 2}
	if err := writeKittyImage(ctx, &buf, path, config, 10, 5, opts); err != nil {
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
	config2, err := loadImageConfig(ctx, path2)
	if err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	opts2 := &kitty.Options{Action: kitty.Transmit, Format: kitty.PNG, Transmission: kitty.Direct, ID: 2, Quite: 2}
	if err := writeKittyImage(ctx, &buf, path2, config2, 10, 5, opts2); err != nil {
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

func TestWriteKittyImageTransmissionOverSSH(t *testing.T) {
	// Over SSH the terminal cannot read local file paths, so images must
	// be transmitted inline instead, even when a file would be cheaper.
	t.Setenv("SSH_TTY", "/dev/pts/0")
	ctx := NewRenderContext(Options{})

	path := filepath.Join(t.TempDir(), "small.png")
	writePNG(t, path, image.NewRGBA(image.Rect(0, 0, 100, 50)))
	config, err := loadImageConfig(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := &kitty.Options{Action: kitty.Transmit, Format: kitty.PNG, Transmission: kitty.Direct, ID: 1, Quite: 2}
	if err := writeKittyImage(ctx, &buf, path, config, 10, 5, opts); err != nil {
		t.Fatal(err)
	}
	seq := buf.String()
	if strings.Contains(seq, "t=f") {
		t.Errorf("expected inline transmission over SSH, got: %q", seq[:min(80, len(seq))])
	}
	if !strings.Contains(seq, ";iVBORw") {
		// Direct transmission sends the image inline, so the sequence must
		// carry the PNG data itself rather than a file path.
		t.Errorf("expected inline image data over SSH, got: %q", seq[:min(80, len(seq))])
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

// renderImage renders in as markdown with the given options and returns the
// rendered document and its renderer.
func renderImage(t *testing.T, options Options, in string) (string, *ANSIRenderer) {
	t.Helper()

	b, err := os.ReadFile("../styles/dark.json")
	if err != nil {
		t.Fatal(err)
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

	var buf bytes.Buffer
	if err := md.Convert([]byte(in), &buf); err != nil {
		t.Fatal(err)
	}
	return buf.String(), ar
}

func TestCheckImageSize(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		maxPixels int
		wantOK    bool
	}{
		{name: "small image", width: 100, height: 50, wantOK: true},
		{name: "at default limit", width: 10000, height: 10000, wantOK: true},
		{name: "over default limit", width: 10001, height: 10000, wantOK: false},
		{name: "custom limit", width: 100, height: 100, maxPixels: 9999, wantOK: false},
		{name: "custom limit met", width: 100, height: 100, maxPixels: 10000, wantOK: true},
		{name: "no limit", width: 100000, height: 100000, maxPixels: -1, wantOK: true},
		{name: "invalid dimensions", width: 0, height: 100, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkImageSize(image.Config{Width: tc.width, Height: tc.height}, tc.maxPixels)
			if tc.wantOK && err != nil {
				t.Errorf("expected image to be accepted, got error: %v", err)
			}
			if !tc.wantOK && err == nil {
				t.Error("expected image to be rejected")
			}
		})
	}
}

// TestImageTooLargeIsSkipped ensures that images whose header declares huge
// dimensions are never decoded. Without the size check, decoding a GIF like
// this allocates a full canvas of the declared size, exhausting memory.
func TestImageTooLargeIsSkipped(t *testing.T) {
	// A GIF whose logical screen declares a 30000x30000 canvas, i.e. 900
	// megapixels, but that contains only a single 1x1 pixel frame. Reading
	// its header reports the huge dimensions without any data to match.
	var b bytes.Buffer
	if err := gif.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil); err != nil {
		t.Fatal(err)
	}
	data := b.Bytes()
	binary.LittleEndian.PutUint16(data[6:8], 30000)
	binary.LittleEndian.PutUint16(data[8:10], 30000)
	path := filepath.Join(t.TempDir(), "huge.gif")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	out, _ := renderImage(t, Options{ImageProtocol: ImageProtocolKitty}, "![]("+path+")")
	if strings.Contains(out, "\x1b_G") {
		t.Errorf("expected oversized image to be skipped, got: %q", out)
	}
}

func TestClassifyImageURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    imageSourceKind
		wantErr bool
	}{
		{name: "plain path", url: "/tmp/x/img.png", want: imageSourceLocal},
		{name: "file url", url: "file:///tmp/x/img.png", want: imageSourceLocal},
		{name: "windows drive", url: "C:/Users/x/img.png", want: imageSourceLocal},
		{name: "unc path", url: "//server/share/img.png", want: imageSourceLocal},
		{name: "http", url: "http://example.com/img.png", want: imageSourceRemote},
		{name: "https", url: "https://example.com/img.png", want: imageSourceRemote},
		{name: "data", url: "data:image/png;base64,AAAA", want: imageSourceData},
		{name: "ftp", url: "ftp://example.com/img.png", wantErr: true},
		{name: "mailto", url: "mailto:x@example.com", wantErr: true},
		{name: "javascript", url: "javascript:alert(1)", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := classifyImageURL(tc.url)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got kind %v", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("classifyImageURL(%q) = %v, want %v", tc.url, got, tc.want)
			}
		})
	}
}

func TestDecodeDataURL(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

	tests := []struct {
		name    string
		url     string
		want    []byte
		wantErr bool
	}{
		{
			name: "base64",
			url:  "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
			want: png,
		},
		{
			name: "percent encoded",
			url:  "data:image/png," + url.QueryEscape(string(png)),
			want: png,
		},
		{name: "no payload", url: "data:image/png", wantErr: true},
		{name: "empty payload", url: "data:image/png,", wantErr: true},
		{name: "invalid base64", url: "data:image/png;base64,****", wantErr: true},
		{name: "not a data url", url: "http://example.com/img.png", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeDataURL(tc.url)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Errorf("decodeDataURL(%q) = %v, want %v", tc.url, got, tc.want)
			}
		})
	}
}

// TestDataURLImages ensures inline data: images render, since they involve
// no network access.
func TestDataURLImages(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	out, _ := renderImage(t, Options{ImageProtocol: ImageProtocolKitty}, "![]("+dataURL+")")
	if !strings.Contains(out, "\x1b_G") {
		t.Errorf("expected data URL image to render, got: %q", out)
	}
}

func TestRemoteImages(t *testing.T) {
	png, err := os.ReadFile(filepath.Join("testdata", "TestImageProtocol", "test.png"))
	if err != nil {
		t.Fatal(err)
	}

	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(png)
	}))
	defer srv.Close()

	// Remote images are disabled by default: no request is ever made.
	md := "![](img.png)"
	out, _ := renderImage(t,
		Options{ImageProtocol: ImageProtocolKitty, BaseURL: srv.URL + "/"}, md)
	if requests.Load() != 0 {
		t.Errorf("expected no requests with remote images disabled, got %d", requests.Load())
	}
	if strings.Contains(out, "\x1b_G") {
		t.Errorf("expected remote image to be skipped, got: %q", out)
	}

	// When enabled, remote images are fetched and rendered.
	out, _ = renderImage(t,
		Options{ImageProtocol: ImageProtocolKitty, LoadRemoteImages: true, BaseURL: srv.URL + "/"}, md)
	if requests.Load() == 0 {
		t.Error("expected a request with remote images enabled")
	}
	if !strings.Contains(out, "\x1b_G") {
		t.Errorf("expected remote image to render, got: %q", out)
	}
}

// TestImageURLHiddenWhenDisplayed checks that an image that is displayed is
// rendered without its URL, which would only be redundant next to it, and
// that the URL is kept when the image isn't displayed.
func TestImageURLHiddenWhenDisplayed(t *testing.T) {
	imgPath, err := filepath.Abs(filepath.Join("testdata", "TestImageProtocol", "test.png"))
	if err != nil {
		t.Fatal(err)
	}
	md := "![alt text](" + imgPath + ")"

	// Displayed: the image replaces the URL, in both the protocol that draws
	// the image itself and the one that anchors it to the text grid.
	for _, protocol := range []ImageProtocol{ImageProtocolKitty, ImageProtocolKittyPlaceholders} {
		out, _ := renderImage(t, Options{ImageProtocol: protocol}, md)
		displayMarkers := 0
		switch protocol {
		case ImageProtocolKitty:
			displayMarkers = strings.Count(out, "\x1b_G")
		case ImageProtocolKittyPlaceholders:
			displayMarkers = strings.Count(out, string(kittyPlaceholder))
		}
		if displayMarkers == 0 {
			t.Errorf("protocol %v: expected the image to be displayed, got: %q", protocol, out)
		}
		if got := visibleText(out); strings.Contains(got, imgPath) {
			t.Errorf("protocol %v: expected no URL next to the displayed image, got: %q", protocol, got)
		}
	}

	// Not displayed: the URL is how the reader gets to the image.
	out, _ := renderImage(t, Options{ImageProtocol: ImageProtocolNone}, md)
	if got := visibleText(out); !strings.Contains(got, imgPath) {
		t.Errorf("expected the image URL, got: %q", got)
	}
}

func TestRemoteImageNotLoadedNote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("no request must be made while remote images are disabled")
		w.Header().Set("Content-Type", "image/png")
	}))
	defer srv.Close()

	url := srv.URL + "/img.png"
	md := "![alt text](" + url + ")"
	options := Options{ImageProtocol: ImageProtocolKitty, BaseURL: srv.URL + "/"}

	// By default the URL is followed by a note saying the image wasn't
	// loaded, so a skipped remote image isn't mistaken for a broken one.
	out, _ := renderImage(t, options, md)
	if !strings.Contains(out, url) {
		t.Errorf("expected the image URL to render, got: %q", out)
	}
	u, n := strings.Index(out, url), strings.Index(out, "not loaded: remote image loading is disabled")
	if n < 0 {
		t.Errorf("expected the default not-loaded note, got: %q", out)
	} else if n < u {
		t.Errorf("expected the note after the URL, got: %q", out)
	}
	if strings.Contains(out, "\x1b_G") {
		t.Errorf("expected no image data, got: %q", out)
	}

	// A custom note replaces the default one, so applications can point at
	// their own setting.
	note := " (not loaded: set loadRemoteImages to true)"
	out, _ = renderImage(t, Options{
		ImageProtocol:            ImageProtocolKitty,
		BaseURL:                  srv.URL + "/",
		RemoteImageNotLoadedNote: &note,
	}, md)
	if !strings.Contains(out, note) {
		t.Errorf("expected the custom note, got: %q", out)
	}
	if strings.Contains(out, "remote image loading is disabled") {
		t.Errorf("expected the default note to be replaced, got: %q", out)
	}

	// An empty note suppresses the note, for renderers that load the
	// images in a second pass.
	empty := ""
	out, _ = renderImage(t, Options{
		ImageProtocol:            ImageProtocolKitty,
		BaseURL:                  srv.URL + "/",
		RemoteImageNotLoadedNote: &empty,
	}, md)
	if strings.Contains(out, "not loaded") {
		t.Errorf("expected no note, got: %q", out)
	}
	if !strings.Contains(out, url) {
		t.Errorf("expected the image URL to render, got: %q", out)
	}
}

func TestRemoteImageSizeLimit(t *testing.T) {
	t.Run("declared content length", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Announce more bytes than the limit allows; the download must
			// be rejected without reading the body.
			w.Header().Set("Content-Length", strconv.Itoa(maxRemoteImageBytes+1))
			_, _ = io.WriteString(w, "x")
		}))
		defer srv.Close()

		out, _ := renderImage(t,
			Options{ImageProtocol: ImageProtocolKitty, LoadRemoteImages: true, BaseURL: srv.URL + "/"}, "![](img.png)")
		if strings.Contains(out, "\x1b_G") {
			t.Errorf("expected oversized remote image to be skipped, got: %q", out)
		}
	})

	t.Run("streamed body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Stream more bytes than the limit allows without declaring the
			// size; the download must be cut off at the limit.
			_, _ = io.Copy(w, bytes.NewReader(make([]byte, maxRemoteImageBytes+1024)))
		}))
		defer srv.Close()

		out, _ := renderImage(t,
			Options{ImageProtocol: ImageProtocolKitty, LoadRemoteImages: true, BaseURL: srv.URL + "/"}, "![](img.png)")
		if strings.Contains(out, "\x1b_G") {
			t.Errorf("expected oversized remote image to be skipped, got: %q", out)
		}
	})
}

func TestPerRendererImageCaches(t *testing.T) {
	imgPath, err := filepath.Abs(filepath.Join("testdata", "TestImageProtocol", "test.png"))
	if err != nil {
		t.Fatal(err)
	}
	md := "![](" + imgPath + ")"
	options := Options{ImageProtocol: ImageProtocolKitty}

	out1, r1 := renderImage(t, options, md)
	out2, r2 := renderImage(t, options, md)

	// Every renderer gets its own caches, so switching documents releases
	// the cached images along with the old renderer.
	if r1.context.options.caches == r2.context.options.caches {
		t.Error("expected renderers to have separate image caches")
	}
	if out1 != out2 {
		t.Errorf("expected identical output, got %q and %q", out1, out2)
	}

	// A nil cache must be safe to use, so hand-made render contexts don't
	// crash: all methods simply skip the cache.
	var caches *imageCaches
	caches.setImage("x", nil)
	if img, ok := caches.image("x"); ok {
		t.Errorf("expected no cached image, got %v", img)
	}
	caches.setConfig("x", imageConfig{})
	if _, ok := caches.config("x"); ok {
		t.Error("expected no cached config")
	}
	caches.setRemoteData("x", nil)
	if _, ok := caches.remoteData("x"); ok {
		t.Error("expected no cached data")
	}
	caches.setSequence(sequenceCacheKey{}, graphicsSequences{})
	if _, ok := caches.sequence(sequenceCacheKey{}); ok {
		t.Error("expected no cached sequence")
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
