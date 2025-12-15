// Package images provides terminal image rendering support for glamour.
package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BourgeoisBear/rasterm"
	"golang.org/x/term"
)

// Protocol represents the terminal image protocol to use.
type Protocol string

const (
	// ProtocolNone disables image rendering.
	ProtocolNone Protocol = "none"
	// ProtocolAuto automatically detects the best protocol.
	ProtocolAuto Protocol = "auto"
	// ProtocolKitty uses the Kitty graphics protocol.
	ProtocolKitty Protocol = "kitty"
	// ProtocolSixel uses the SIXEL protocol.
	ProtocolSixel Protocol = "sixel"
	// ProtocolITerm uses the iTerm2 inline image protocol.
	ProtocolITerm Protocol = "iterm"
)

// RenderOptions configures image rendering behavior.
type RenderOptions struct {
	Protocol    Protocol
	BaseURL     string
	FetchRemote bool // Allow fetching remote images via HTTP (default: false)
}

// detectedProtocol caches the result of protocol detection.
var (
	detectedProtocol     Protocol
	detectedProtocolOnce sync.Once
)

// detectProtocol detects the best available terminal image protocol.
// Uses environment-based detection to avoid terminal queries that can hang.
func detectProtocol() Protocol {
	detectedProtocolOnce.Do(func() {
		// Graphics protocols require stdout to be a TTY
		if !isTerminal() {
			detectedProtocol = ProtocolNone
			return
		}

		// Detect protocol from environment variables
		detectedProtocol = detectProtocolFromEnvironment()
	})
	return detectedProtocol
}

// detectProtocolFromEnvironment detects the terminal image protocol using
// rasterm's detection helpers and environment variables.
func detectProtocolFromEnvironment() Protocol {
	// Use rasterm's detection helpers
	if rasterm.IsKittyCapable() {
		return ProtocolKitty
	}

	if rasterm.IsItermCapable() {
		return ProtocolITerm
	}

	// Fallback: check for Ghostty which also supports Kitty protocol
	// (rasterm might not detect it yet)
	if os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return ProtocolKitty
	}

	// Could add Sixel detection here for terminals like mlterm, foot, etc.
	// rasterm.IsSixelCapable() does terminal queries which can hang,
	// so we skip it for now.
	return ProtocolNone
}

// isTerminal checks if stdout is connected to a terminal.
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// CanRender returns true if image rendering is available with the given options.
func CanRender(opts RenderOptions) bool {
	if opts.Protocol == ProtocolNone || opts.Protocol == "" {
		return false
	}

	// For auto mode, check if any protocol is detected
	if opts.Protocol == ProtocolAuto {
		detected := detectProtocol()
		return detected != ProtocolNone
	}

	// For specific protocols, they're always "available" if requested
	return true
}

// LoadAndRender loads an image from the given URL and renders it for terminal display.
// It supports:
// - Remote URLs (http://, https://) - only if FetchRemote is true
// - Local file paths
// - Data URIs (data:image/png;base64,...)
func LoadAndRender(imageURL string, opts RenderOptions) (string, error) {
	if !CanRender(opts) {
		return "", fmt.Errorf("image rendering not available")
	}

	// Load the image data
	img, err := loadImage(imageURL, opts)
	if err != nil {
		return "", fmt.Errorf("failed to load image: %w", err)
	}

	// Render the image
	return renderImage(img, opts)
}

// loadImage loads an image from various sources.
func loadImage(imageURL string, opts RenderOptions) (image.Image, error) {
	// Check for data URI
	if strings.HasPrefix(imageURL, "data:") {
		return loadDataURI(imageURL)
	}

	// Parse the URL
	u, err := url.Parse(imageURL)
	if err != nil {
		// Treat as local file path
		return loadLocalFile(imageURL, opts.BaseURL)
	}

	// Handle based on scheme
	switch u.Scheme {
	case "http", "https":
		if !opts.FetchRemote {
			return nil, fmt.Errorf("remote image fetching disabled (use WithImageFetchRemote to enable)")
		}
		return loadRemoteURL(imageURL)
	case "file":
		return loadLocalFile(u.Path, opts.BaseURL)
	case "":
		// No scheme - treat as local file or relative path
		return loadLocalFile(imageURL, opts.BaseURL)
	default:
		return nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
}

// loadDataURI loads an image from a data URI.
func loadDataURI(dataURI string) (image.Image, error) {
	// Parse data URI: data:[<mediatype>][;base64],<data>
	if !strings.HasPrefix(dataURI, "data:") {
		return nil, fmt.Errorf("invalid data URI")
	}

	// Find the comma that separates metadata from data
	commaIdx := strings.Index(dataURI, ",")
	if commaIdx == -1 {
		return nil, fmt.Errorf("invalid data URI: missing comma")
	}

	metadata := dataURI[5:commaIdx] // Skip "data:"
	data := dataURI[commaIdx+1:]

	// Check if base64 encoded
	isBase64 := strings.Contains(metadata, ";base64")

	var imgData []byte
	var err error

	if isBase64 {
		imgData, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 data: %w", err)
		}
	} else {
		// URL-encoded data
		imgData = []byte(data)
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// loadLocalFile loads an image from a local file.
func loadLocalFile(path, baseURL string) (image.Image, error) {
	// Resolve relative paths
	if !filepath.IsAbs(path) && baseURL != "" {
		// Try to resolve against baseURL
		if u, err := url.Parse(baseURL); err == nil {
			if u.Scheme == "file" || u.Scheme == "" {
				// Local base path
				basePath := u.Path
				if basePath == "" {
					basePath = baseURL
				}
				path = filepath.Join(filepath.Dir(basePath), path)
			}
		}
	}

	// Open and decode the file
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// loadRemoteURL loads an image from a remote URL.
func loadRemoteURL(imageURL string) (image.Image, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: HTTP %d", resp.StatusCode)
	}

	// Limit the size we'll read (10MB max)
	limitedReader := io.LimitReader(resp.Body, 10*1024*1024)

	img, _, err := image.Decode(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// renderImage renders an image using the appropriate terminal protocol.
func renderImage(img image.Image, opts RenderOptions) (string, error) {
	proto := opts.Protocol

	// If auto, resolve to a concrete protocol
	if proto == ProtocolAuto {
		proto = detectProtocol()
		if proto == ProtocolNone {
			return "", fmt.Errorf("no image protocol available")
		}
	}

	// Render to a buffer
	var buf bytes.Buffer

	switch proto {
	case ProtocolKitty:
		// Let the terminal handle native sizing - don't set DstCols/DstRows
		// unless we need to constrain to a maximum. This preserves small
		// image dimensions while capping large images.
		kittyOpts := rasterm.KittyImgOpts{}
		// Note: Not setting DstCols/DstRows means the terminal displays
		// the image at its native pixel size, scaled to terminal cells.
		if err := rasterm.KittyWriteImage(&buf, img, kittyOpts); err != nil {
			return "", fmt.Errorf("failed to render image with Kitty protocol: %w", err)
		}

	case ProtocolITerm:
		if err := rasterm.ItermWriteImage(&buf, img); err != nil {
			return "", fmt.Errorf("failed to render image with iTerm protocol: %w", err)
		}

	case ProtocolSixel:
		// Sixel requires a paletted image
		palettedImg := convertToPaletted(img)
		if err := rasterm.SixelWriteImage(&buf, palettedImg); err != nil {
			return "", fmt.Errorf("failed to render image with Sixel protocol: %w", err)
		}

	default:
		return "", fmt.Errorf("unsupported protocol: %s", proto)
	}

	return buf.String(), nil
}

// convertToPaletted converts an image to a paletted image for Sixel rendering.
func convertToPaletted(img image.Image) *image.Paletted {
	// Check if already paletted
	if p, ok := img.(*image.Paletted); ok {
		return p
	}

	// Create a 256-color palette using a 6x6x6 color cube + grayscale
	bounds := img.Bounds()
	palette := make(color.Palette, 0, 256)

	// Generate a simple 6x6x6 color cube (216 colors)
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				palette = append(palette, simpleColor{uint8(r * 51), uint8(g * 51), uint8(b * 51)})
			}
		}
	}
	// Add grayscale ramp (40 colors to reach 256)
	for i := 0; i < 40; i++ {
		gray := uint8(i * 255 / 39)
		palette = append(palette, simpleColor{gray, gray, gray})
	}

	// Use standard library's draw package for conversion
	palettedImg := image.NewPaletted(bounds, palette)
	draw.Draw(palettedImg, bounds, img, bounds.Min, draw.Src)

	return palettedImg
}

// simpleColor implements color.Color for palette generation.
type simpleColor struct {
	r, g, b uint8
}

func (c simpleColor) RGBA() (r, g, b, a uint32) {
	return uint32(c.r) * 257, uint32(c.g) * 257, uint32(c.b) * 257, 0xffff
}
