package ansi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
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

	// maxRemoteImageBytes is the maximum size of a remote image download.
	// Larger downloads are rejected, so a huge image cannot exhaust memory
	// or block rendering for minutes, and neither can a server that lies
	// about its content type.
	maxRemoteImageBytes = 32 << 20 // 32 MiB

	// cellPixelWidth and cellPixelHeight are the assumed physical pixel
	// dimensions of a terminal cell, used to convert terminal cells to
	// pixels. Images are displayed at up to 1:1 in sixel, and transmitted
	// at up to hiDPIScale times the size of their on-screen box for kitty,
	// so they stay sharp on hi-dpi screens, but never with more pixels than
	// the terminal can actually display.
	cellPixelWidth  = 10
	cellPixelHeight = 20
	hiDPIScale      = 2
)

// DefaultMaxImagePixels is the default value of [Options.MaxImagePixels]:
// the maximum number of pixels of images that are decoded and displayed.
// Decoding larger images can exhaust memory, so they are skipped instead.
const DefaultMaxImagePixels = 100_000_000 // 100 megapixels

var httpClient = &http.Client{
	Timeout: httpClientTimeout,
}

// errRemoteImagesDisabled is wrapped into the error returned when an image
// is not loaded because loading remote images is disabled (see
// [Options.LoadRemoteImages]). Callers can match it with errors.Is, e.g. to
// render a hint of their own.
var errRemoteImagesDisabled = errors.New("remote images are disabled")

// defaultRemoteImageNotLoadedNote is appended after the URL of a remote
// image that was not loaded because loading remote images is disabled. It
// starts with a space because it directly follows the URL.
const defaultRemoteImageNotLoadedNote = " (not loaded: remote image loading is disabled)"

// imageConfig holds an image's dimensions and encoded format, as read from
// its header without decoding the pixels.
type imageConfig struct {
	config image.Config
	format string
}

// imageCaches holds a renderer's image caches: image headers, decoded
// images, downloaded remote images, and encoded graphics sequences, all
// keyed by URL. They avoid re-fetching, re-decoding, and re-encoding images
// on every render, which is important for interactive applications (e.g. a
// TUI pager) that re-render the same document repeatedly.
//
// Each renderer has its own caches, so applications that create a new
// renderer per document release the cached images when the renderer is
// garbage collected. All methods are safe to call on a nil *imageCaches;
// such calls simply skip the cache.
type imageCaches struct {
	sync.Mutex
	configs   map[string]imageConfig
	images    map[string]image.Image
	data      map[string][]byte
	sequences map[sequenceCacheKey]graphicsSequences
}

// newImageCaches returns an empty set of image caches.
func newImageCaches() *imageCaches {
	return &imageCaches{
		configs:   make(map[string]imageConfig),
		images:    make(map[string]image.Image),
		data:      make(map[string][]byte),
		sequences: make(map[sequenceCacheKey]graphicsSequences),
	}
}

func (c *imageCaches) config(url string) (imageConfig, bool) {
	if c == nil {
		return imageConfig{}, false
	}
	c.Lock()
	defer c.Unlock()
	config, ok := c.configs[url]
	return config, ok
}

func (c *imageCaches) setConfig(url string, config imageConfig) {
	if c == nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	c.configs[url] = config
}

func (c *imageCaches) image(url string) (image.Image, bool) {
	if c == nil {
		return nil, false
	}
	c.Lock()
	defer c.Unlock()
	img, ok := c.images[url]
	return img, ok
}

func (c *imageCaches) setImage(url string, img image.Image) {
	if c == nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	c.images[url] = img
}

func (c *imageCaches) remoteData(url string) ([]byte, bool) {
	if c == nil {
		return nil, false
	}
	c.Lock()
	defer c.Unlock()
	data, ok := c.data[url]
	return data, ok
}

func (c *imageCaches) setRemoteData(url string, data []byte) {
	if c == nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	c.data[url] = data
}

func (c *imageCaches) sequence(key sequenceCacheKey) (graphicsSequences, bool) {
	if c == nil {
		return graphicsSequences{}, false
	}
	c.Lock()
	defer c.Unlock()
	seqs, ok := c.sequences[key]
	return seqs, ok
}

func (c *imageCaches) setSequence(key sequenceCacheKey, seqs graphicsSequences) {
	if c == nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	c.sequences[key] = seqs
}

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

// htmlImgRegex matches <img> tags and captures the src attribute.
var htmlImgRegex = regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)

// An ImageElement is used to render images elements.
type ImageElement struct {
	Text     string
	BaseURL  string
	URL      string
	Child    ElementRenderer
	TextOnly bool

	// Inline writes the image's graphics sequences where the image appears
	// in the document instead of queueing them for after the block. Used for
	// images parsed from HTML blocks, which aren't wrapped, so their
	// sequences can be written in place without breaking the text flow.
	Inline bool
}

// Render renders an ImageElement.
func (e *ImageElement) Render(w io.Writer, ctx RenderContext) error {
	// Make OSC 8 hyperlink token.
	hyperlink, resetHyperlink, _ := makeHyperlink(e.URL)

	// Queue the image's graphics sequence first: when the image displays,
	// its URL would only be redundant. The sequence is written after the
	// paragraph has been rendered (see flushPendingImages), so this doesn't
	// affect the order of the text rendered below.
	url := resolveRelativeURL(e.BaseURL, e.URL)
	imageDisplayed, imageErr := e.displayImage(ctx, w, url)

	style := ctx.options.Styles.ImageText
	if e.TextOnly || imageDisplayed {
		// No URL follows the text, so don't render its arrow.
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

	if e.TextOnly || len(e.URL) == 0 || imageDisplayed {
		return nil
	}

	// The image doesn't display, so its URL is the reader's only access to it.
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

	// A remote image that wasn't loaded because loading remote images is
	// disabled gets a note saying so, so it isn't mistaken for a broken one.
	// Other failures are skipped silently, and the link remains.
	if note := ctx.remoteImageNotLoadedNote(imageErr); note != "" {
		el := &BaseElement{
			Token: note,
			Style: ctx.options.Styles.Image,
		}
		if err := el.Render(w, ctx); err != nil {
			return err
		}
	}

	return nil
}

// displayImage queues the graphics protocol sequences that display the image
// at url, if this element references an image and the configured protocol
// can render it. It reports whether the image is displayed, along with the
// error that kept it from being displayed. w is the writer of the block the
// image is rendered into.
func (e *ImageElement) displayImage(ctx RenderContext, w io.Writer, url string) (bool, error) {
	if e.TextOnly || len(e.URL) == 0 || ctx.options.ImageProtocol == ImageProtocolNone {
		return false, nil
	}

	seqs, inline, err := e.graphicsSequence(ctx, url)
	if err != nil {
		return false, err
	}

	if seqs.inline != "" {
		switch {
		case e.Inline:
			// Images in HTML blocks are written where they appear in the
			// document: HTML blocks aren't wrapped, so even multi-row
			// sequences can be written in place.
			if _, err := io.WriteString(w, seqs.inline); err != nil {
				return false, fmt.Errorf("glamour: error writing image: %w", err)
			}
		case inline:
			// Single-cell-row images, e.g. badges, are anchored to the text
			// they appear in, so images next to each other end up on one
			// line, the way they do in a browser.
			if _, err := io.WriteString(w, seqs.inline); err != nil {
				return false, fmt.Errorf("glamour: error writing image: %w", err)
			}
		default:
			*ctx.pendingImages = append(*ctx.pendingImages, seqs.inline)
		}
	}
	if len(seqs.commands) > 0 {
		*ctx.graphicsCommands = append(*ctx.graphicsCommands, seqs.commands...)
	}
	return true, nil
}

// remoteImageNotLoadedNote returns the note to render after the URL of a
// remote image that was not loaded because loading remote images is disabled
// (see [Options.RemoteImageNotLoadedNote]), or an empty string when there is
// no such note to render.
func (ctx RenderContext) remoteImageNotLoadedNote(err error) string {
	if !errors.Is(err, errRemoteImagesDisabled) {
		return ""
	}
	if ctx.options.RemoteImageNotLoadedNote != nil {
		return *ctx.options.RemoteImageNotLoadedNote
	}
	return defaultRemoteImageNotLoadedNote
}

// graphicsSequence loads the image and returns the encoded graphics protocol
// sequences. Encoded sequences are cached so repeated renders of the same
// image don't re-encode it.
func (e *ImageElement) graphicsSequence(ctx RenderContext, url string) (graphicsSequences, bool, error) {
	config, err := loadImageConfig(ctx, url)
	if err != nil {
		return graphicsSequences{}, false, fmt.Errorf("glamour: error loading image: %w", err)
	}

	// SVGs are vectors whose intrinsic size only determines the aspect
	// ratio, so the pixel limit doesn't apply to them: the rasterized image
	// is sized to the display box, not to the declared dimensions.
	if config.format != svgFormat {
		if err := checkImageSize(config.config, ctx.options.MaxImagePixels); err != nil {
			return graphicsSequences{}, false, err
		}
	}

	// An SVG's declared size is a CSS pixel size chosen for display, so it
	// is drawn at exactly that size. Raster images use pixels as cells,
	// constrained by the block width.
	var cols, rows int
	if config.format == svgFormat {
		cols, rows = svgDisplaySize(config.config.Width, config.config.Height, ctx)
	} else {
		cols, rows = imageDisplaySize(config.config.Width, config.config.Height, ctx)
	}
	// Images displayed via unicode placeholders take part in the text flow
	// when they are one cell row tall.
	inline := ctx.options.ImageProtocol == ImageProtocolKittyPlaceholders && rows == 1
	key := sequenceCacheKey{url: url, protocol: ctx.options.ImageProtocol, cols: cols, rows: rows}

	if seqs, ok := ctx.imageCachesOf().sequence(key); ok {
		return seqs, inline, nil
	}

	seqs, err := e.encodeGraphics(ctx, url, config, cols, rows)
	if err != nil {
		return graphicsSequences{}, false, err
	}

	ctx.imageCachesOf().setSequence(key, seqs)
	return seqs, inline, nil
}

// checkImageSize reports whether an image with the given header dimensions
// may be decoded, under the given pixel limit. Zero applies
// [DefaultMaxImagePixels], a negative value means no limit. Checking the
// header dimensions before decoding keeps oversized images from exhausting
// memory, no matter how small they are on disk.
func checkImageSize(config image.Config, maxPixels int) error {
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("glamour: invalid image dimensions %dx%d", config.Width, config.Height)
	}
	if maxPixels < 0 {
		return nil
	}
	if maxPixels == 0 {
		maxPixels = DefaultMaxImagePixels
	}
	if w, h := int64(config.Width), int64(config.Height); w*h > int64(maxPixels) {
		return fmt.Errorf("glamour: image is too large: %dx%d pixels, limit is %d", config.Width, config.Height, maxPixels)
	}
	return nil
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
		if err := writeKittyImage(ctx, &buf, url, config, cols, rows, &opts); err != nil {
			return graphicsSequences{}, err
		}
		return graphicsSequences{inline: reserveRows(buf.String(), rows)}, nil
	case ImageProtocolSixel:
		// Sixel draws at pixel resolution, so the image is loaded at the
		// target cell dimensions (SVGs are rasterized at that size) and
		// scaled down to match the reserved space.
		img, err := loadImage(ctx, url, cols*cellPixelWidth, rows*cellPixelHeight)
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
		return encodeKittyPlaceholders(ctx, url, config, cols, rows)
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
func encodeKittyPlaceholders(ctx RenderContext, url string, config imageConfig, cols, rows int) (graphicsSequences, error) {
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
	if err := writeKittyImage(ctx, &transmit, url, config, cols, rows, &opts); err != nil {
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

	// Single-row images are anchored to the text flow, so they are placed
	// where they appear; taller ones take up whole lines of their own, before
	// and after the text that follows them.
	inline := placeholderGrid(id, cols, rows)
	if rows > 1 {
		inline = "\n" + inline + "\n"
	}

	return graphicsSequences{
		inline:   inline,
		commands: []string{transmit.String(), ansi.KittyGraphics(nil, place.Options()...)},
	}, nil
}

// writeKittyImage writes the kitty graphics sequence transmitting the image
// at url with the given options, for display in a cols x rows cell box. It
// picks the cheapest transmission medium: local PNG files that don't need
// resizing are transmitted by path so the terminal reads them directly,
// without any decoding or encoding on this side, and everything else is
// decoded, downscaled to at most twice the size of the on-screen box, and
// re-encoded as PNG. Paths are only transmitted when the terminal can
// read them, i.e. outside of SSH sessions; over SSH the image data is
// transmitted inline instead.
func writeKittyImage(ctx RenderContext, w io.Writer, url string, config imageConfig, cols, rows int, opts *kitty.Options) error {
	if isLocalImage(url) && config.format == "png" && fitsDisplayBox(config.config, cols, rows) && !inSSHSession() {
		opts.Transmission = kitty.File
		opts.File = localImagePath(url)
		if err := kitty.EncodeGraphics(w, nil, opts); err != nil {
			return fmt.Errorf("glamour: error encoding kitty image: %w", err)
		}
		return nil
	}

	maxW, maxH := maxSourcePixels(cols, rows)
	// SVGs are rasterized at the display size; raster images are decoded at
	// their own size and scaled down here.
	img, err := loadImage(ctx, url, maxW, maxH)
	if err != nil {
		return fmt.Errorf("glamour: error loading image: %w", err)
	}
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
	return max(cols, 0) * cellPixelWidth * hiDPIScale,
		max(rows, 0) * cellPixelHeight * hiDPIScale
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

// imageSourceKind classifies image URLs by where the image data comes from.
type imageSourceKind int

const (
	imageSourceLocal imageSourceKind = iota
	imageSourceRemote
	imageSourceData
)

// classifyImageURL reports where the image at the given URL comes from.
// Anything that isn't a local path, an http(s) URL, or a data: URI is an
// error: URLs like ftp://host/img.png would otherwise be treated as local
// file names. Single-letter schemes are Windows drive letters, like
// C:/dir/img.png, and count as local paths.
func classifyImageURL(s string) (imageSourceKind, error) {
	u, err := url.Parse(s)
	if err != nil {
		return 0, fmt.Errorf("glamour: error parsing image URL: %w", err)
	}
	switch u.Scheme {
	case "", "file":
		return imageSourceLocal, nil
	case "http", "https":
		return imageSourceRemote, nil
	case "data":
		return imageSourceData, nil
	}
	if len(u.Scheme) == 1 && (u.Scheme[0] >= 'a' && u.Scheme[0] <= 'z' || u.Scheme[0] >= 'A' && u.Scheme[0] <= 'Z') {
		return imageSourceLocal, nil
	}
	return 0, fmt.Errorf("glamour: unsupported image URL scheme %q", u.Scheme)
}

// isLocalImage reports whether url points to a local file rather than a
// remote resource or inline data.
func isLocalImage(url string) bool {
	kind, err := classifyImageURL(url)
	return err == nil && kind == imageSourceLocal
}

// inSSHSession reports whether this process runs inside an SSH session, i.e.
// whether the terminal runs on another machine. The terminal cannot read
// local file paths there, so images must be transmitted inline.
func inSSHSession() bool {
	return os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CONNECTION") != ""
}

// imageID returns a stable, positive identifier for an image URL, fitting in
// 24 bits so it can be encoded in the foreground color of a Unicode
// placeholder. Stable IDs let terminals replace a previous placement of the
// same image instead of piling up copies when a document is re-rendered.
func imageID(url string) int {
	h := fnv.New32a()
	_, _ = io.WriteString(h, url)
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

// scaleToCells scales img down to the pixel size of a cols x rows terminal
// cell box using nearest-neighbor sampling, preserving the aspect ratio. A
// cell is cellPixelWidth pixels wide and cellPixelHeight pixels tall, so
// e.g. an image filling 40x10 cells must be 400x200 pixels: sixel draws at
// pixel resolution, and an image scaled to its box in cells alone would
// come out a fraction of the reserved size. Images that already fit are
// returned unchanged.
func scaleToCells(img image.Image, cols, rows int) image.Image {
	if cols <= 0 || rows <= 0 {
		return img
	}
	return downscaleImage(img, cols*cellPixelWidth, rows*cellPixelHeight)
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
func loadImageConfig(ctx RenderContext, url string) (imageConfig, error) {
	if config, ok := ctx.imageCachesOf().config(url); ok {
		return config, nil
	}

	config, err := readImageConfig(ctx, url)
	if err != nil {
		return imageConfig{}, err
	}

	ctx.imageCachesOf().setConfig(url, config)
	return config, nil
}

func readImageConfig(ctx RenderContext, url string) (imageConfig, error) {
	kind, err := classifyImageURL(url)
	if err != nil {
		return imageConfig{}, err
	}

	switch kind {
	case imageSourceLocal:
		path := localImagePath(url)
		f, err := os.Open(path)
		if err != nil {
			return imageConfig{}, fmt.Errorf("glamour: error opening image file: %w", err)
		}
		defer f.Close() //nolint:errcheck

		br := bufio.NewReader(f)
		if isSVG(url, sniffHead(br)) {
			return readSVGConfig(br)
		}
		cfg, format, err := image.DecodeConfig(br)
		if err != nil {
			return imageConfig{}, fmt.Errorf("glamour: error decoding image config: %w", err)
		}
		return imageConfig{config: cfg, format: format}, nil
	}

	var buf []byte
	switch kind {
	case imageSourceData:
		buf, err = imageData(ctx, url)
		if err != nil {
			return imageConfig{}, err
		}
	case imageSourceRemote:
		if !ctx.options.LoadRemoteImages {
			return imageConfig{}, fmt.Errorf("glamour: %w", errRemoteImagesDisabled)
		}
		buf, err = fetchRemoteImage(ctx, url)
		if err != nil {
			return imageConfig{}, err
		}
	}

	if isSVG(url, buf) {
		return readSVGConfig(bytes.NewReader(buf))
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return imageConfig{}, fmt.Errorf("glamour: error decoding image config: %w", err)
	}
	return imageConfig{config: cfg, format: format}, nil
}

// isSVG reports whether the image at url, with the given data, is an SVG
// document. SVG isn't a raster image format, so it is parsed and rasterized
// separately (see svg.go); both the URL and the image's own bytes are
// checked, as SVGs are stored and served under a variety of names and
// content types.
func isSVG(url string, data []byte) bool {
	return isSVGURL(url) || isSVGData(data)
}

// loadImage loads and decodes an image from a local file, a data: URI, or a
// remote URL. SVGs are rasterized to fit the maxW x maxH pixel box, keeping
// their aspect ratio; for raster images the limits are ignored, as they are
// scaled to the display size later on. Decoded images are cached in memory
// so repeated renders don't re-fetch and re-decode them.
func loadImage(ctx RenderContext, url string, maxW, maxH int) (image.Image, error) {
	kind, err := classifyImageURL(url)
	if err != nil {
		return nil, err
	}

	if kind != imageSourceLocal {
		if img, ok := ctx.imageCachesOf().image(url); ok {
			return img, nil
		}
	}

	var buf []byte
	switch kind {
	case imageSourceLocal:
		path := localImagePath(url)
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("glamour: error opening image file: %w", err)
		}
		defer f.Close() //nolint:errcheck

		br := bufio.NewReader(f)
		if isSVG(url, sniffHead(br)) {
			return rasterizeSVGReader(br, maxW, maxH)
		}
		img, _, err := image.Decode(br)
		if err != nil {
			return nil, fmt.Errorf("glamour: error decoding image: %w", err)
		}
		return img, nil
	case imageSourceData:
		buf, err = imageData(ctx, url)
		if err != nil {
			return nil, err
		}
	case imageSourceRemote:
		if !ctx.options.LoadRemoteImages {
			return nil, fmt.Errorf("glamour: %w", errRemoteImagesDisabled)
		}
		buf, err = fetchRemoteImage(ctx, url)
		if err != nil {
			return nil, err
		}
	}

	if isSVG(url, buf) {
		return rasterizeSVGReader(bytes.NewReader(buf), maxW, maxH)
	}
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("glamour: error decoding image: %w", err)
	}

	if kind != imageSourceLocal {
		ctx.imageCachesOf().setImage(url, img)
	}
	return img, nil
}

// imageData returns the bytes of the image embedded in the given data: URI,
// e.g. data:image/png;base64,<base64 data>. Both base64 and percent-encoded
// payloads are supported; the decoded bytes are cached per URL.
func imageData(ctx RenderContext, url string) ([]byte, error) {
	if data, ok := ctx.imageCachesOf().remoteData(url); ok {
		return data, nil
	}

	data, err := decodeDataURL(url)
	if err != nil {
		return nil, err
	}

	ctx.imageCachesOf().setRemoteData(url, data)
	return data, nil
}

// decodeDataURL decodes the payload of a data: URI. Both base64 payloads,
// e.g. data:image/png;base64,<base64 data>, and percent-encoded ones are
// supported; anything else is an error.
func decodeDataURL(s string) ([]byte, error) {
	const prefix = "data:"
	if !strings.HasPrefix(s, prefix) {
		return nil, fmt.Errorf("glamour: not a data URL: %q", s)
	}
	mediaType, payload, found := strings.Cut(s[len(prefix):], ",")
	if !found || payload == "" {
		return nil, fmt.Errorf("glamour: data URL without payload")
	}

	for _, param := range strings.Split(mediaType, ";") {
		if strings.EqualFold(strings.TrimSpace(param), "base64") {
			data, err := base64.StdEncoding.DecodeString(payload)
			if err != nil {
				return nil, fmt.Errorf("glamour: error decoding base64 data URL: %w", err)
			}
			return data, nil
		}
	}

	data, err := url.PathUnescape(payload)
	if err != nil {
		return nil, fmt.Errorf("glamour: error decoding data URL: %w", err)
	}
	return []byte(data), nil
}

// fetchRemoteImage fetches the bytes of a remote image, checking for
// unsupported content types and rejecting oversized downloads. Responses
// are cached in memory so repeated renders don't re-fetch them.
func fetchRemoteImage(ctx RenderContext, url string) ([]byte, error) {
	if buf, ok := ctx.imageCachesOf().remoteData(url); ok {
		return buf, nil
	}

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
	if resp.ContentLength > maxRemoteImageBytes {
		return nil, fmt.Errorf("glamour: remote image is too large (%d bytes, limit is %d)", resp.ContentLength, maxRemoteImageBytes)
	}

	// Read the body into a buffer so we can inspect the content type.
	// The limit keeps a huge image, or a server that never stops
	// sending, from exhausting memory.
	buf, err := io.ReadAll(io.LimitReader(resp.Body, maxRemoteImageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("glamour: error reading image body: %w", err)
	}
	if len(buf) > maxRemoteImageBytes {
		return nil, fmt.Errorf("glamour: remote image exceeds %d bytes", maxRemoteImageBytes)
	}

	ctx.imageCachesOf().setRemoteData(url, buf)
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
		if len(elements) > 0 {
			// Browsers render images that are on separate lines of the
			// markup next to each other, with a space between them.
			elements = append(elements, spaceElement{})
		}
		src := m[1]
		elements = append(elements, &ImageElement{
			Text:    "",
			BaseURL: ctx.options.BaseURL,
			URL:     src,
			Inline:  true,
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

// spaceElement renders a single space.
type spaceElement struct{}

// Render writes a space.
func (spaceElement) Render(w io.Writer, _ RenderContext) error {
	_, err := io.WriteString(w, " ")
	if err != nil {
		return fmt.Errorf("glamour: error writing space: %w", err)
	}
	return nil
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
