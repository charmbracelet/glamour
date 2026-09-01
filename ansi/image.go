package ansi

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"image"

	// Register image formats for decoding.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/charmbracelet/x/ansi/sixel"
)

const (
	// approxTerminalCellAspect is the approximate height of a terminal cell
	// relative to its width. Used to calculate the display height of an image
	// when its width is constrained to the available width.
	approxTerminalCellAspect = 0.5

	// httpClientTimeout is the timeout used for fetching remote images. It is
	// kept low because fetching happens synchronously while rendering.
	httpClientTimeout = 10 * time.Second

	// kittyCellPixelWidth and kittyCellPixelHeight are the assumed physical
	// pixel dimensions of a terminal cell. Images are transmitted at up to
	// kittyHiDPIScale times the size of their on-screen box so they stay
	// sharp on hi-dpi screens, but never with more pixels than the terminal
	// can actually display.
	kittyCellPixelWidth  = 10
	kittyCellPixelHeight = 20
	kittyHiDPIScale      = 2
)

var httpClient = &http.Client{
	Timeout: httpClientTimeout,
}

// imageConfig holds an image's dimensions and encoded format, as read from
// its header without decoding the pixels.
type imageConfig struct {
	config image.Config
	format string
}

// imageCache is an in-memory cache of decoded images keyed by URL. It avoids
// re-fetching and re-decoding remote images on every render, which is
// important for interactive applications (e.g. a TUI pager) that re-render
// the same document repeatedly.
var imageCache = struct {
	sync.Mutex
	m map[string]image.Image
}{m: make(map[string]image.Image)}

// imageConfigCache is an in-memory cache of image headers keyed by URL, so
// sizing an image doesn't require reading it repeatedly.
var imageConfigCache = struct {
	sync.Mutex
	m map[string]imageConfig
}{m: make(map[string]imageConfig)}

// remoteImageCache is an in-memory cache of the raw bytes of remote images,
// so configuring and decoding an image shares a single fetch.
var remoteImageCache = struct {
	sync.Mutex
	m map[string][]byte
}{m: make(map[string][]byte)}

// sequenceCacheKey uniquely identifies an encoded graphics sequence.
type sequenceCacheKey struct {
	url      string
	protocol ImageProtocol
	cols     int
	rows     int
}

// graphicsSequences holds the graphics protocol sequences produced for a
// single image.
type graphicsSequences struct {
	// inline is embedded in the rendered document at the position of the
	// image, reserving the space the image occupies.
	inline string
	// commands are out-of-band sequences that must be written to the
	// terminal before displaying the rendered document. They are exposed
	// via [ANSIRenderer.GraphicsCommands].
	commands []string
}

// sequenceCache is an in-memory cache of encoded graphics protocol sequences.
// Encoding an image (especially PNG-encoding for Kitty) is expensive, so
// caching the resulting sequences avoids re-encoding on every render in
// interactive applications that re-render the same document repeatedly.
var sequenceCache = struct {
	sync.Mutex
	m map[sequenceCacheKey]graphicsSequences
}{m: make(map[sequenceCacheKey]graphicsSequences)}

// htmlImgRegex matches <img> tags and captures the src attribute.
var htmlImgRegex = regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)

// An ImageElement is used to render images elements.
type ImageElement struct {
	Text     string
	BaseURL  string
	URL      string
	Child    ElementRenderer
	TextOnly bool
}

// Render renders an ImageElement.
func (e *ImageElement) Render(w io.Writer, ctx RenderContext) error {
	// Make OSC 8 hyperlink token.
	hyperlink, resetHyperlink, _ := makeHyperlink(e.URL)

	style := ctx.options.Styles.ImageText
	if e.TextOnly {
		style.Format = strings.TrimSuffix(style.Format, " →")
	}

	if len(e.Text) > 0 {
		token := hyperlink + e.Text + resetHyperlink
		el := &BaseElement{
			Token: token,
			Style: style,
		}
		err := el.Render(w, ctx)
		if err != nil {
			return err
		}
	}

	if e.TextOnly || len(e.URL) == 0 {
		return nil
	}

	url := resolveRelativeURL(e.BaseURL, e.URL)
	token := hyperlink + url + resetHyperlink
	el := &BaseElement{
		Token:  token,
		Prefix: " ",
		Style:  ctx.options.Styles.Image,
	}
	err := el.Render(w, ctx)
	if err != nil {
		return err
	}

	if ctx.options.ImageProtocol != ImageProtocolNone {
		seqs, err := e.graphicsSequence(ctx, url)
		if err != nil {
			// Silently skip images that fail to load or encode. The
			// alt text and link are already rendered above.
			return nil //nolint:nilerr
		}
		if seqs.inline != "" {
			*ctx.pendingImages = append(*ctx.pendingImages, seqs.inline)
		}
		if len(seqs.commands) > 0 {
			*ctx.graphicsCommands = append(*ctx.graphicsCommands, seqs.commands...)
		}
	}

	return nil
}

// graphicsSequence loads the image and returns the encoded graphics protocol
// sequences. Encoded sequences are cached so repeated renders of the same
// image don't re-encode it.
func (e *ImageElement) graphicsSequence(ctx RenderContext, url string) (graphicsSequences, error) {
	config, err := loadImageConfig(url)
	if err != nil {
		return graphicsSequences{}, fmt.Errorf("glamour: error loading image: %w", err)
	}

	cols, rows := imageDisplaySize(config.config.Width, config.config.Height, ctx)
	key := sequenceCacheKey{url: url, protocol: ctx.options.ImageProtocol, cols: cols, rows: rows}

	sequenceCache.Lock()
	if seqs, ok := sequenceCache.m[key]; ok {
		sequenceCache.Unlock()
		return seqs, nil
	}
	sequenceCache.Unlock()

	seqs, err := e.encodeGraphics(ctx, url, config, cols, rows)
	if err != nil {
		return graphicsSequences{}, err
	}

	sequenceCache.Lock()
	sequenceCache.m[key] = seqs
	sequenceCache.Unlock()
	return seqs, nil
}

// encodeGraphics encodes the image at url into graphics protocol sequences
// sized to cols x rows terminal cells, along with the newlines needed to
// reserve vertical space for the image in the text output.
func (e *ImageElement) encodeGraphics(ctx RenderContext, url string, config imageConfig, cols, rows int) (graphicsSequences, error) {
	switch ctx.options.ImageProtocol {
	case ImageProtocolKitty:
		opts := kitty.Options{
			Action:          kitty.TransmitAndPut,
			Format:          kitty.PNG,
			Transmission:    kitty.Direct,
			ID:              imageID(url),
			Columns:         cols,
			Rows:            rows,
			DoNotMoveCursor: true,
			Quite:           2,
		}
		var buf bytes.Buffer
		if err := writeKittyImage(&buf, url, config, cols, rows, &opts); err != nil {
			return graphicsSequences{}, err
		}
		return graphicsSequences{inline: reserveRows(buf.String(), rows)}, nil
	case ImageProtocolSixel:
		// Sixel draws at pixel resolution, so scale the image down to the
		// target cell dimensions to match the reserved space.
		img, err := loadImage(url)
		if err != nil {
			return graphicsSequences{}, fmt.Errorf("glamour: error loading image: %w", err)
		}
		scaled := scaleToCells(img, cols, rows)
		var sb bytes.Buffer
		enc := sixel.Encoder{}
		if err := enc.Encode(&sb, scaled); err != nil {
			return graphicsSequences{}, fmt.Errorf("glamour: error encoding sixel image: %w", err)
		}
		seq := "\x1bPq" + sb.String() + "\x1b\\"
		return graphicsSequences{inline: reserveRows(seq, rows)}, nil
	case ImageProtocolKittyPlaceholders:
		return encodeKittyPlaceholders(url, config, cols, rows)
	case ImageProtocolNone:
		return graphicsSequences{}, nil
	default:
		return graphicsSequences{}, nil
	}
}

// encodeKittyPlaceholders encodes the image for display via Kitty graphics
// Unicode placeholders. The returned sequences carry the out-of-band transmit
// and virtual placement commands in commands, and the placeholder grid, which
// anchors the image to the text grid, in inline. This allows full-screen
// TUI applications to display images that move with the text while scrolling
// without re-transmitting the image data.
// See https://sw.kovidgoyal.net/kitty/graphics-protocol/#unicode-placeholders
func encodeKittyPlaceholders(url string, config imageConfig, cols, rows int) (graphicsSequences, error) {
	id := imageID(url)

	// Transmit the image without displaying it.
	opts := kitty.Options{
		Action:       kitty.Transmit,
		Format:       kitty.PNG,
		Transmission: kitty.Direct,
		ID:           id,
		Quite:        2,
	}
	var transmit bytes.Buffer
	if err := writeKittyImage(&transmit, url, config, cols, rows, &opts); err != nil {
		return graphicsSequences{}, err
	}

	// Create a virtual placement of the image, sized to cols x rows cells.
	// The image will be displayed wherever the corresponding Unicode
	// placeholders appear in the text, and will move with them as the text
	// scrolls, without any further graphics protocol output.
	//
	// Rows and columns are capped to the number of values representable by
	// the row/column diacritics (360, see rowcolumn-diacritics.txt in the
	// kitty docs). The terminal fits the image into the placeholder box,
	// preserving its aspect ratio.
	rows = min(rows, maxDiacritics)
	cols = min(cols, maxDiacritics)
	place := kitty.Options{
		Action:           kitty.Put,
		VirtualPlacement: true,
		ID:               id,
		Columns:          cols,
		Rows:             rows,
		Quite:            2,
	}

	return graphicsSequences{
		inline:   "\n" + placeholderGrid(id, cols, rows),
		commands: []string{transmit.String(), ansi.KittyGraphics(nil, place.Options()...)},
	}, nil
}

// writeKittyImage writes the kitty graphics sequence transmitting the image
// at url with the given options, for display in a cols x rows cell box. It
// picks the cheapest transmission medium: local PNG files that don't need
// resizing are transmitted by path so the terminal reads them directly,
// without any decoding or encoding on this side, and everything else is
// decoded, downscaled to at most twice the size of the on-screen box, and
// re-encoded as PNG.
func writeKittyImage(w io.Writer, url string, config imageConfig, cols, rows int, opts *kitty.Options) error {
	if isFileURL(url) && config.format == "png" && fitsDisplayBox(config.config, cols, rows) {
		opts.Transmission = kitty.File
		opts.File = localImagePath(url)
		if err := kitty.EncodeGraphics(w, nil, opts); err != nil {
			return fmt.Errorf("glamour: error encoding kitty image: %w", err)
		}
		return nil
	}

	img, err := loadImage(url)
	if err != nil {
		return fmt.Errorf("glamour: error loading image: %w", err)
	}
	maxW, maxH := maxSourcePixels(cols, rows)
	img = downscaleImage(img, maxW, maxH)
	opts.Chunk = true
	if err := kitty.EncodeGraphics(w, img, opts); err != nil {
		return fmt.Errorf("glamour: error encoding kitty image: %w", err)
	}
	return nil
}

// maxSourcePixels returns the maximum number of source pixels an image
// displayed in cols x rows cells can use without visible loss of quality.
func maxSourcePixels(cols, rows int) (int, int) {
	return max(cols, 0) * kittyCellPixelWidth * kittyHiDPIScale,
		max(rows, 0) * kittyCellPixelHeight * kittyHiDPIScale
}

// fitsDisplayBox reports whether an image of the given dimensions can be
// displayed in cols x rows cells without downscaling.
func fitsDisplayBox(config image.Config, cols, rows int) bool {
	maxW, maxH := maxSourcePixels(cols, rows)
	return config.Width > 0 && config.Height > 0 &&
		config.Width <= maxW && config.Height <= maxH
}

// placeholderGrid returns a cols x rows grid of Kitty graphics Unicode
// placeholders for the image with the given ID. Each line of the grid carries
// the image ID in its foreground color, and its first cell encodes the row
// with a combining diacritic; the terminal infers the column of the remaining
// cells from the preceding ones. This requires a terminal with 24-bit color
// support.
func placeholderGrid(id, cols, rows int) string {
	var b strings.Builder
	for row := 0; row < rows; row++ {
		if row > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("\x1b[38;2;")
		b.WriteString(strconv.Itoa((id >> 16) & 0xff))
		b.WriteByte(';')
		b.WriteString(strconv.Itoa((id >> 8) & 0xff))
		b.WriteByte(';')
		b.WriteString(strconv.Itoa(id & 0xff))
		b.WriteByte('m')
		b.WriteRune(kitty.Placeholder)
		b.WriteRune(kitty.Diacritic(row))
		for col := 1; col < cols; col++ {
			b.WriteRune(kitty.Placeholder)
		}
		b.WriteString("\x1b[39m")
	}
	return b.String()
}

// maxDiacritics is the number of values the row/column diacritics of the
// Kitty graphics Unicode placeholders can encode. From
// https://sw.kovidgoyal.net/kitty/graphics-protocol/#unicode-placeholders
const maxDiacritics = 360

// reserveRows appends one newline per row the image occupies so the terminal
// reserves vertical space for it. Without this, the terminal draws the image
// over the text that follows it.
func reserveRows(seq string, rows int) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(seq)
	for i := 0; i < rows; i++ {
		sb.WriteString("\n")
	}
	return sb.String()
}

// isFileURL reports whether url points to a local file rather than a remote
// resource.
func isFileURL(url string) bool {
	return !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")
}

// imageID returns a stable, positive identifier for an image URL, fitting in
// 24 bits so it can be encoded in the foreground color of a Unicode
// placeholder. Stable IDs let terminals replace a previous placement of the
// same image instead of piling up copies when a document is re-rendered.
func imageID(url string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(url))
	id := int(h.Sum32() & 0xffffff)
	if id == 0 {
		id = 1
	}
	return id
}

// imageDisplaySize calculates the display size of an image in terminal cells.
// It constrains the image to the width available in the block the image is
// rendered into (accounting for margins and indentation), so the image never
// gets wrapped by surrounding block elements, and computes the height
// preserving the aspect ratio. The size can be further limited via the
// [Options.MaxImageColumns] and [Options.MaxImageRows] options.
func imageDisplaySize(width, height int, ctx RenderContext) (int, int) {
	maxWidth := int(ctx.blockStack.Width(ctx))
	if maxWidth <= 0 {
		maxWidth = width
	}
	if ctx.options.MaxImageColumns > 0 {
		maxWidth = min(maxWidth, ctx.options.MaxImageColumns)
	}
	if width > maxWidth && width > 0 {
		height = height * maxWidth / width
		width = maxWidth
	}

	rows := int(math.Ceil(float64(height) * approxTerminalCellAspect))
	if rows < 1 {
		rows = 1
	}
	if maxRows := ctx.options.MaxImageRows; maxRows > 0 && rows > maxRows {
		// Shrink the width proportionally so the aspect ratio is preserved
		// by the terminal when it fits the image into the display box.
		width = max(1, width*maxRows/rows)
		rows = maxRows
	}
	return width, rows
}

// scaleToCells scales img to approximately cols x rows terminal cells using
// nearest-neighbor sampling. Terminal cells are assumed to be twice as tall
// as they are wide (see approxTerminalCellAspect), so one column maps to one
// pixel and one row maps to two pixels.
func scaleToCells(img image.Image, cols, rows int) image.Image {
	if cols <= 0 || rows <= 0 {
		return img
	}
	return downscaleImage(img, cols, rows*2)
}

// downscaleImage scales img down to at most maxW x maxH pixels using
// nearest-neighbor sampling, preserving the aspect ratio. Images that already
// fit are returned unchanged.
func downscaleImage(img image.Image, maxW, maxH int) image.Image {
	src := img.Bounds()
	srcW, srcH := src.Dx(), src.Dy()
	if srcW == 0 || srcH == 0 || maxW <= 0 || maxH <= 0 ||
		(srcW <= maxW && srcH <= maxH) {
		return img
	}

	// Preserve the aspect ratio while fitting into the box.
	scale := min(float64(maxW)/float64(srcW), float64(maxH)/float64(srcH))
	dstW := max(1, int(float64(srcW)*scale))
	dstH := max(1, int(float64(srcH)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		sy := src.Min.Y + y*srcH/dstH
		for x := 0; x < dstW; x++ {
			sx := src.Min.X + x*srcW/dstW
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	return dst
}

// loadImageConfig reads an image's dimensions and format from its header,
// without decoding its pixels. Results are cached per URL.
func loadImageConfig(url string) (imageConfig, error) {
	imageConfigCache.Lock()
	if config, ok := imageConfigCache.m[url]; ok {
		imageConfigCache.Unlock()
		return config, nil
	}
	imageConfigCache.Unlock()

	config, err := readImageConfig(url)
	if err != nil {
		return imageConfig{}, err
	}

	imageConfigCache.Lock()
	imageConfigCache.m[url] = config
	imageConfigCache.Unlock()
	return config, nil
}

func readImageConfig(url string) (imageConfig, error) {
	if isFileURL(url) {
		path := localImagePath(url)
		f, err := os.Open(path)
		if err != nil {
			return imageConfig{}, fmt.Errorf("glamour: error opening image file: %w", err)
		}
		defer f.Close() //nolint:errcheck
		cfg, format, err := image.DecodeConfig(f)
		if err != nil {
			return imageConfig{}, fmt.Errorf("glamour: error decoding image config: %w", err)
		}
		return imageConfig{config: cfg, format: format}, nil
	}

	buf, err := fetchRemoteImage(url)
	if err != nil {
		return imageConfig{}, err
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return imageConfig{}, fmt.Errorf("glamour: error decoding image config: %w", err)
	}
	return imageConfig{config: cfg, format: format}, nil
}

// loadImage loads and decodes an image from a local file path or remote URL.
// Decoded images are cached in memory so repeated renders don't re-fetch and
// re-decode them.
func loadImage(url string) (image.Image, error) {
	if !isFileURL(url) {
		imageCache.Lock()
		if img, ok := imageCache.m[url]; ok {
			imageCache.Unlock()
			return img, nil
		}
		imageCache.Unlock()

		buf, err := fetchRemoteImage(url)
		if err != nil {
			return nil, err
		}
		img, _, err := image.Decode(bytes.NewReader(buf))
		if err != nil {
			return nil, fmt.Errorf("glamour: error decoding image: %w", err)
		}

		imageCache.Lock()
		imageCache.m[url] = img
		imageCache.Unlock()
		return img, nil
	}

	path := localImagePath(url)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("glamour: error opening image file: %w", err)
	}
	defer f.Close() //nolint:errcheck

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("glamour: error decoding image: %w", err)
	}
	return img, nil
}

// fetchRemoteImage fetches the bytes of a remote image, checking for
// unsupported content types. Responses are cached in memory so repeated
// renders don't re-fetch them.
func fetchRemoteImage(url string) ([]byte, error) {
	remoteImageCache.Lock()
	if buf, ok := remoteImageCache.m[url]; ok {
		remoteImageCache.Unlock()
		return buf, nil
	}
	remoteImageCache.Unlock()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("glamour: error creating request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("glamour: error fetching image: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("glamour: unexpected status code %d fetching image", resp.StatusCode)
	}

	// Read the body into a buffer so we can inspect the content type.
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("glamour: error reading image body: %w", err)
	}

	// Check for SVG which Go's image.Decode cannot handle.
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "image/svg") || bytes.HasPrefix(bytes.TrimSpace(buf), []byte("<svg")) {
		return nil, fmt.Errorf("glamour: SVG images are not supported")
	}

	remoteImageCache.Lock()
	remoteImageCache.m[url] = buf
	remoteImageCache.Unlock()
	return buf, nil
}

// localImagePath returns the local file path for the given URL, which can be
// a plain path or a file:// URL. Windows drive letters are handled, i.e.
// file:///C:/dir/img.png resolves to C:\dir\img.png on Windows.
func localImagePath(s string) string {
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "file" || u.Path == "" {
		return strings.TrimPrefix(s, "file://")
	}
	p := u.Path
	// A Windows drive letter is encoded as a leading path segment, e.g.
	// /C:/dir/img.png.
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' &&
		(p[1] >= 'a' && p[1] <= 'z' || p[1] >= 'A' && p[1] <= 'Z') {
		p = p[1:]
	}
	return filepath.FromSlash(p)
}

// parseHTMLImages extracts <img> tags from HTML and returns an ImageElement
// for each one found. Returns nil if no images are found.
func parseHTMLImages(ctx RenderContext, html string) ElementRenderer {
	matches := htmlImgRegex.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil
	}

	var elements []ElementRenderer
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		src := m[1]
		elements = append(elements, &ImageElement{
			Text:    "",
			BaseURL: ctx.options.BaseURL,
			URL:     src,
		})
	}

	if len(elements) == 0 {
		return nil
	}

	return &CompoundElement{Elements: elements}
}

// A CompoundElement renders multiple elements sequentially.
type CompoundElement struct {
	Elements []ElementRenderer
}

// Render renders all child elements.
func (e *CompoundElement) Render(w io.Writer, ctx RenderContext) error {
	for _, el := range e.Elements {
		if err := el.Render(w, ctx); err != nil {
			return fmt.Errorf("glamour: error rendering element: %w", err)
		}
	}
	return nil
}
