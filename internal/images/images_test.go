package images

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// createTestPNG creates a simple 2x2 PNG image for testing.
func createTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})
	img.Set(1, 1, color.RGBA{255, 255, 255, 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test PNG: %v", err)
	}
	return buf.Bytes()
}

// createTestImage creates a simple 2x2 RGBA image for testing.
func createTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})
	img.Set(1, 1, color.RGBA{255, 255, 255, 255})
	return img
}

func TestProtocolConstants(t *testing.T) {
	// Verify protocol constants have expected values
	tests := []struct {
		protocol Protocol
		expected string
	}{
		{ProtocolNone, "none"},
		{ProtocolAuto, "auto"},
		{ProtocolKitty, "kitty"},
		{ProtocolSixel, "sixel"},
		{ProtocolITerm, "iterm"},
	}

	for _, tt := range tests {
		if string(tt.protocol) != tt.expected {
			t.Errorf("Protocol %v: got %q, want %q", tt.protocol, string(tt.protocol), tt.expected)
		}
	}
}

func TestDetectProtocolFromEnvironment(t *testing.T) {
	// Save original environment
	origKittyWindowID := os.Getenv("KITTY_WINDOW_ID")
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origGhosttyResources := os.Getenv("GHOSTTY_RESOURCES_DIR")
	origLCTerminal := os.Getenv("LC_TERMINAL")

	// Restore environment after test
	defer func() {
		os.Setenv("KITTY_WINDOW_ID", origKittyWindowID)
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("GHOSTTY_RESOURCES_DIR", origGhosttyResources)
		os.Setenv("LC_TERMINAL", origLCTerminal)
	}()

	tests := []struct {
		name     string
		envSetup func()
		expected Protocol
	}{
		{
			name: "kitty detected via KITTY_WINDOW_ID",
			envSetup: func() {
				os.Setenv("KITTY_WINDOW_ID", "1")
				os.Unsetenv("TERM_PROGRAM")
				os.Unsetenv("GHOSTTY_RESOURCES_DIR")
				os.Unsetenv("LC_TERMINAL")
			},
			expected: ProtocolKitty,
		},
		{
			name: "wezterm detected via TERM_PROGRAM",
			envSetup: func() {
				os.Unsetenv("KITTY_WINDOW_ID")
				os.Setenv("TERM_PROGRAM", "WezTerm")
				os.Unsetenv("GHOSTTY_RESOURCES_DIR")
				os.Unsetenv("LC_TERMINAL")
			},
			expected: ProtocolKitty, // WezTerm uses Kitty protocol
		},
		{
			name: "ghostty detected via GHOSTTY_RESOURCES_DIR",
			envSetup: func() {
				os.Unsetenv("KITTY_WINDOW_ID")
				os.Unsetenv("TERM_PROGRAM")
				os.Setenv("GHOSTTY_RESOURCES_DIR", "/usr/share/ghostty")
				os.Unsetenv("LC_TERMINAL")
			},
			expected: ProtocolKitty, // Ghostty uses Kitty protocol
		},
		{
			name: "iterm detected via LC_TERMINAL",
			envSetup: func() {
				os.Unsetenv("KITTY_WINDOW_ID")
				os.Unsetenv("TERM_PROGRAM")
				os.Unsetenv("GHOSTTY_RESOURCES_DIR")
				os.Setenv("LC_TERMINAL", "iTerm2")
			},
			expected: ProtocolITerm,
		},
		{
			name: "no protocol detected",
			envSetup: func() {
				os.Unsetenv("KITTY_WINDOW_ID")
				os.Unsetenv("TERM_PROGRAM")
				os.Unsetenv("GHOSTTY_RESOURCES_DIR")
				os.Unsetenv("LC_TERMINAL")
			},
			expected: ProtocolNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envSetup()
			got := detectProtocolFromEnvironment()
			if got != tt.expected {
				t.Errorf("detectProtocolFromEnvironment() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCanRender(t *testing.T) {
	tests := []struct {
		name     string
		opts     RenderOptions
		expected bool
	}{
		{
			name:     "none protocol returns false",
			opts:     RenderOptions{Protocol: ProtocolNone},
			expected: false,
		},
		{
			name:     "empty protocol returns false",
			opts:     RenderOptions{Protocol: ""},
			expected: false,
		},
		{
			name:     "kitty protocol returns true",
			opts:     RenderOptions{Protocol: ProtocolKitty},
			expected: true,
		},
		{
			name:     "sixel protocol returns true",
			opts:     RenderOptions{Protocol: ProtocolSixel},
			expected: true,
		},
		{
			name:     "iterm protocol returns true",
			opts:     RenderOptions{Protocol: ProtocolITerm},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanRender(tt.opts)
			if got != tt.expected {
				t.Errorf("CanRender(%v) = %v, want %v", tt.opts, got, tt.expected)
			}
		})
	}
}

func TestLoadDataURI(t *testing.T) {
	pngData := createTestPNG(t)
	base64Data := base64.StdEncoding.EncodeToString(pngData)

	tests := []struct {
		name    string
		dataURI string
		wantErr bool
	}{
		{
			name:    "valid base64 PNG data URI",
			dataURI: "data:image/png;base64," + base64Data,
			wantErr: false,
		},
		{
			name:    "invalid - not a data URI",
			dataURI: "http://example.com/image.png",
			wantErr: true,
		},
		{
			name:    "invalid - missing comma",
			dataURI: "data:image/png;base64" + base64Data,
			wantErr: true,
		},
		{
			name:    "invalid - bad base64",
			dataURI: "data:image/png;base64,not-valid-base64!!!",
			wantErr: true,
		},
		{
			name:    "invalid - not an image",
			dataURI: "data:text/plain;base64," + base64.StdEncoding.EncodeToString([]byte("hello")),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := loadDataURI(tt.dataURI)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadDataURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && img == nil {
				t.Error("loadDataURI() returned nil image without error")
			}
		})
	}
}

func TestLoadLocalFile(t *testing.T) {
	// Create a temporary directory with a test image
	tmpDir := t.TempDir()
	pngData := createTestPNG(t)

	// Write test PNG file
	pngPath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(pngPath, pngData, 0644); err != nil {
		t.Fatalf("failed to write test PNG: %v", err)
	}

	// Write test file in subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	subPngPath := filepath.Join(subDir, "nested.png")
	if err := os.WriteFile(subPngPath, pngData, 0644); err != nil {
		t.Fatalf("failed to write nested PNG: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "absolute path",
			path:    pngPath,
			baseURL: "",
			wantErr: false,
		},
		{
			name:    "relative path with base URL",
			path:    "test.png",
			baseURL: tmpDir + "/",
			wantErr: false,
		},
		{
			name:    "relative path with file:// base URL",
			path:    "test.png",
			baseURL: "file://" + tmpDir + "/index.md",
			wantErr: false,
		},
		{
			name:    "nested relative path",
			path:    "subdir/nested.png",
			baseURL: tmpDir + "/",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    filepath.Join(tmpDir, "nonexistent.png"),
			baseURL: "",
			wantErr: true,
		},
		{
			name:    "invalid image file",
			path:    filepath.Join(tmpDir, "invalid.png"),
			baseURL: "",
			wantErr: true,
		},
	}

	// Create an invalid image file
	invalidPath := filepath.Join(tmpDir, "invalid.png")
	if err := os.WriteFile(invalidPath, []byte("not an image"), 0644); err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := loadLocalFile(tt.path, tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadLocalFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && img == nil {
				t.Error("loadLocalFile() returned nil image without error")
			}
		})
	}
}

func TestLoadRemoteURL(t *testing.T) {
	pngData := createTestPNG(t)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/image.png":
			w.Header().Set("Content-Type", "image/png")
			w.Write(pngData)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/invalid":
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("not an image"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid PNG URL",
			url:     server.URL + "/image.png",
			wantErr: false,
		},
		{
			name:    "404 response",
			url:     server.URL + "/notfound",
			wantErr: true,
		},
		{
			name:    "invalid image data",
			url:     server.URL + "/invalid",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     "http://invalid.invalid.invalid/image.png",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := loadRemoteURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && img == nil {
				t.Error("loadRemoteURL() returned nil image without error")
			}
		})
	}
}

func TestLoadImage(t *testing.T) {
	pngData := createTestPNG(t)
	base64Data := base64.StdEncoding.EncodeToString(pngData)

	// Create temp file
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(pngPath, pngData, 0644); err != nil {
		t.Fatalf("failed to write test PNG: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(pngData)
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		opts    RenderOptions
		wantErr bool
	}{
		{
			name:    "data URI",
			url:     "data:image/png;base64," + base64Data,
			opts:    RenderOptions{},
			wantErr: false,
		},
		{
			name:    "local file path",
			url:     pngPath,
			opts:    RenderOptions{},
			wantErr: false,
		},
		{
			name:    "file:// URL",
			url:     "file://" + pngPath,
			opts:    RenderOptions{},
			wantErr: false,
		},
		{
			name:    "remote URL with FetchRemote enabled",
			url:     server.URL + "/image.png",
			opts:    RenderOptions{FetchRemote: true},
			wantErr: false,
		},
		{
			name:    "remote URL with FetchRemote disabled",
			url:     server.URL + "/image.png",
			opts:    RenderOptions{FetchRemote: false},
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			url:     "ftp://example.com/image.png",
			opts:    RenderOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := loadImage(tt.url, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && img == nil {
				t.Error("loadImage() returned nil image without error")
			}
		})
	}
}

func TestConvertToPaletted(t *testing.T) {
	// Test with RGBA image
	t.Run("RGBA image", func(t *testing.T) {
		img := createTestImage()
		paletted := convertToPaletted(img)

		if paletted == nil {
			t.Fatal("convertToPaletted() returned nil")
		}

		// Check bounds match
		if paletted.Bounds() != img.Bounds() {
			t.Errorf("bounds mismatch: got %v, want %v", paletted.Bounds(), img.Bounds())
		}

		// Check palette has colors
		if len(paletted.Palette) == 0 {
			t.Error("palette is empty")
		}

		// Palette should have 256 colors (216 color cube + 40 grayscale)
		if len(paletted.Palette) != 256 {
			t.Errorf("palette size: got %d, want 256", len(paletted.Palette))
		}
	})

	// Test with already paletted image
	t.Run("already paletted image", func(t *testing.T) {
		palette := color.Palette{
			color.RGBA{0, 0, 0, 255},
			color.RGBA{255, 255, 255, 255},
		}
		original := image.NewPaletted(image.Rect(0, 0, 2, 2), palette)

		result := convertToPaletted(original)
		if result != original {
			t.Error("convertToPaletted() should return same image for already paletted input")
		}
	})
}

func TestSimpleColorRGBA(t *testing.T) {
	c := simpleColor{128, 64, 32}
	r, g, b, a := c.RGBA()

	// Values should be scaled to 16-bit (multiplied by 257)
	expectedR := uint32(128) * 257
	expectedG := uint32(64) * 257
	expectedB := uint32(32) * 257
	expectedA := uint32(0xffff)

	if r != expectedR || g != expectedG || b != expectedB || a != expectedA {
		t.Errorf("RGBA() = (%d, %d, %d, %d), want (%d, %d, %d, %d)",
			r, g, b, a, expectedR, expectedG, expectedB, expectedA)
	}
}

func TestRenderImage(t *testing.T) {
	img := createTestImage()

	tests := []struct {
		name     string
		protocol Protocol
		wantErr  bool
	}{
		{
			name:     "kitty protocol",
			protocol: ProtocolKitty,
			wantErr:  false,
		},
		{
			name:     "iterm protocol",
			protocol: ProtocolITerm,
			wantErr:  false,
		},
		{
			name:     "sixel protocol",
			protocol: ProtocolSixel,
			wantErr:  false,
		},
		{
			name:     "none protocol",
			protocol: ProtocolNone,
			wantErr:  true,
		},
		{
			name:     "unknown protocol",
			protocol: Protocol("unknown"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := RenderOptions{Protocol: tt.protocol}
			result, err := renderImage(img, opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("renderImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == "" {
					t.Error("renderImage() returned empty string without error")
				}

				// Verify the output contains expected escape sequences
				switch tt.protocol {
				case ProtocolKitty:
					if !bytes.Contains([]byte(result), []byte("\x1b_G")) {
						t.Error("Kitty output missing expected escape sequence")
					}
				case ProtocolITerm:
					if !bytes.Contains([]byte(result), []byte("\x1b]1337;")) {
						t.Error("iTerm output missing expected escape sequence")
					}
				case ProtocolSixel:
					if !bytes.Contains([]byte(result), []byte("\x1bP")) {
						t.Error("Sixel output missing expected escape sequence")
					}
				}
			}
		})
	}
}

func TestLoadAndRender(t *testing.T) {
	pngData := createTestPNG(t)
	base64Data := base64.StdEncoding.EncodeToString(pngData)

	tests := []struct {
		name    string
		url     string
		opts    RenderOptions
		wantErr bool
	}{
		{
			name:    "render data URI with kitty",
			url:     "data:image/png;base64," + base64Data,
			opts:    RenderOptions{Protocol: ProtocolKitty},
			wantErr: false,
		},
		{
			name:    "render with none protocol",
			url:     "data:image/png;base64," + base64Data,
			opts:    RenderOptions{Protocol: ProtocolNone},
			wantErr: true,
		},
		{
			name:    "render with empty protocol",
			url:     "data:image/png;base64," + base64Data,
			opts:    RenderOptions{Protocol: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := LoadAndRender(tt.url, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAndRender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("LoadAndRender() returned empty string without error")
			}
		})
	}
}

func TestRenderOptionsDefaults(t *testing.T) {
	opts := RenderOptions{}

	if opts.Protocol != "" {
		t.Errorf("default Protocol should be empty, got %q", opts.Protocol)
	}
	if opts.BaseURL != "" {
		t.Errorf("default BaseURL should be empty, got %q", opts.BaseURL)
	}
	if opts.FetchRemote != false {
		t.Error("default FetchRemote should be false")
	}
}
