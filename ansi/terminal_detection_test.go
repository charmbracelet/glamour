package ansi

import (
	"os"
	"testing"
)

// TestTerminalDetectionComprehensive tests supportsHyperlinks with comprehensive environment combinations
func TestTerminalDetectionComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
		category string
	}{
		// TERM_PROGRAM based detection
		{
			name:     "iTerm2",
			envVars:  map[string]string{"TERM_PROGRAM": "iTerm.app"},
			expected: true,
			category: "TERM_PROGRAM",
		},
		{
			name:     "VS Code integrated terminal",
			envVars:  map[string]string{"TERM_PROGRAM": "vscode"},
			expected: true,
			category: "TERM_PROGRAM",
		},
		{
			name:     "Windows Terminal",
			envVars:  map[string]string{"TERM_PROGRAM": "Windows Terminal"},
			expected: true,
			category: "TERM_PROGRAM",
		},
		{
			name:     "WezTerm",
			envVars:  map[string]string{"TERM_PROGRAM": "WezTerm"},
			expected: true,
			category: "TERM_PROGRAM",
		},
		{
			name:     "Hyper terminal",
			envVars:  map[string]string{"TERM_PROGRAM": "Hyper"},
			expected: true,
			category: "TERM_PROGRAM",
		},
		{
			name:     "Unknown TERM_PROGRAM",
			envVars:  map[string]string{"TERM_PROGRAM": "unknown-terminal"},
			expected: false,
			category: "TERM_PROGRAM",
		},

		// TERM variable based detection
		{
			name:     "xterm-256color",
			envVars:  map[string]string{"TERM": "xterm-256color"},
			expected: true,
			category: "TERM",
		},
		{
			name:     "screen-256color",
			envVars:  map[string]string{"TERM": "screen-256color"},
			expected: true,
			category: "TERM",
		},
		{
			name:     "tmux-256color",
			envVars:  map[string]string{"TERM": "tmux-256color"},
			expected: true,
			category: "TERM",
		},
		{
			name:     "alacritty",
			envVars:  map[string]string{"TERM": "alacritty"},
			expected: true,
			category: "TERM",
		},
		{
			name:     "xterm-kitty",
			envVars:  map[string]string{"TERM": "xterm-kitty"},
			expected: true,
			category: "TERM",
		},
		{
			name:     "basic xterm",
			envVars:  map[string]string{"TERM": "xterm"},
			expected: false,
			category: "TERM",
		},
		{
			name:     "dumb terminal",
			envVars:  map[string]string{"TERM": "dumb"},
			expected: false,
			category: "TERM",
		},

		// Special environment variables
		{
			name:     "Kitty terminal by KITTY_WINDOW_ID",
			envVars:  map[string]string{"KITTY_WINDOW_ID": "1"},
			expected: true,
			category: "SPECIAL_ENV",
		},
		{
			name:     "Alacritty by ALACRITTY_LOG",
			envVars:  map[string]string{"ALACRITTY_LOG": "/tmp/alacritty.log"},
			expected: true,
			category: "SPECIAL_ENV",
		},
		{
			name:     "Alacritty by ALACRITTY_SOCKET",
			envVars:  map[string]string{"ALACRITTY_SOCKET": "/tmp/alacritty.sock"},
			expected: true,
			category: "SPECIAL_ENV",
		},

		// Priority testing - TERM_PROGRAM should take precedence
		{
			name: "iTerm2 overrides basic TERM",
			envVars: map[string]string{
				"TERM_PROGRAM": "iTerm.app",
				"TERM":         "xterm",
			},
			expected: true,
			category: "PRIORITY",
		},
		{
			name: "Unknown TERM_PROGRAM with supported TERM",
			envVars: map[string]string{
				"TERM_PROGRAM": "unknown",
				"TERM":         "xterm-256color",
			},
			expected: true,
			category: "PRIORITY",
		},

		// Edge cases
		{
			name:     "Empty environment",
			envVars:  map[string]string{},
			expected: false,
			category: "EDGE_CASE",
		},
		{
			name: "Multiple indicators - all support",
			envVars: map[string]string{
				"TERM_PROGRAM":    "vscode",
				"TERM":            "xterm-256color",
				"KITTY_WINDOW_ID": "1",
			},
			expected: true,
			category: "EDGE_CASE",
		},
		{
			name: "Mixed support indicators",
			envVars: map[string]string{
				"TERM_PROGRAM":    "unknown",
				"TERM":            "dumb",
				"KITTY_WINDOW_ID": "1",
			},
			expected: true, // KITTY_WINDOW_ID should trigger support
			category: "EDGE_CASE",
		},

		// Case sensitivity tests
		{
			name:     "Case sensitive TERM_PROGRAM - correct case",
			envVars:  map[string]string{"TERM_PROGRAM": "iTerm.app"},
			expected: true,
			category: "CASE_SENSITIVITY",
		},
		{
			name:     "Case sensitive TERM_PROGRAM - wrong case",
			envVars:  map[string]string{"TERM_PROGRAM": "iterm.app"},
			expected: false,
			category: "CASE_SENSITIVITY",
		},

		// Real-world scenarios
		{
			name: "macOS iTerm2 typical setup",
			envVars: map[string]string{
				"TERM_PROGRAM":         "iTerm.app",
				"TERM_PROGRAM_VERSION": "3.4.16",
				"TERM":                 "xterm-256color",
			},
			expected: true,
			category: "REAL_WORLD",
		},
		{
			name: "VS Code integrated terminal typical setup",
			envVars: map[string]string{
				"TERM_PROGRAM":         "vscode",
				"TERM_PROGRAM_VERSION": "1.74.0",
				"TERM":                 "xterm-256color",
			},
			expected: true,
			category: "REAL_WORLD",
		},
		{
			name: "SSH session with screen",
			envVars: map[string]string{
				"TERM":           "screen-256color",
				"SSH_CONNECTION": "192.168.1.100 12345 192.168.1.1 22",
			},
			expected: true,
			category: "REAL_WORLD",
		},
		{
			name: "WSL with Windows Terminal",
			envVars: map[string]string{
				"TERM_PROGRAM":    "Windows Terminal",
				"TERM":            "xterm-256color",
				"WSL_DISTRO_NAME": "Ubuntu",
			},
			expected: true,
			category: "REAL_WORLD",
		},
		{
			name: "Docker container with basic terminal",
			envVars: map[string]string{
				"TERM":      "xterm",
				"container": "docker",
			},
			expected: false,
			category: "REAL_WORLD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant environment variables first
			clearTerminalEnvVars(t)

			// Set the test environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Test the function
			ctx := RenderContext{} // Empty context is fine for this test
			result := supportsHyperlinks(ctx)

			if result != tt.expected {
				t.Errorf("supportsHyperlinks() = %v, expected %v for test case %q (category: %s)",
					result, tt.expected, tt.name, tt.category)
				t.Logf("Environment variables set: %+v", tt.envVars)
			}
		})
	}
}

// TestTerminalDetectionEdgeCases tests edge cases and error conditions
func TestTerminalDetectionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
		desc     string
	}{
		{
			name: "Empty string values",
			envVars: map[string]string{
				"TERM_PROGRAM": "",
				"TERM":         "",
			},
			expected: false,
			desc:     "Empty environment variable values should not trigger support",
		},
		{
			name: "Whitespace in values",
			envVars: map[string]string{
				"TERM_PROGRAM": " iTerm.app ",
			},
			expected: false, // Current implementation doesn't trim whitespace
			desc:     "Whitespace around values should not match",
		},
		{
			name: "Partial matches in TERM",
			envVars: map[string]string{
				"TERM": "my-xterm-256color-custom",
			},
			expected: true, // Uses strings.Contains()
			desc:     "Partial matches should work for TERM variable",
		},
		{
			name: "Multiple TERM patterns",
			envVars: map[string]string{
				"TERM": "screen-256color-tmux",
			},
			expected: true, // Should match both screen-256color and tmux patterns
			desc:     "TERM with multiple supported patterns should match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTerminalEnvVars(t)

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			result := supportsHyperlinks(RenderContext{})
			if result != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.desc, result, tt.expected)
			}
		})
	}
}

// TestTerminalDetectionPerformance tests that detection is fast
func TestTerminalDetectionPerformance(t *testing.T) {
	// Set up a typical environment
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("TERM", "xterm-256color")

	ctx := RenderContext{}

	// Run detection many times to check for performance issues
	for i := 0; i < 1000; i++ {
		result := supportsHyperlinks(ctx)
		if !result {
			t.Errorf("Expected hyperlink support in iteration %d", i)
		}
	}
}

// TestTerminalDetectionWithContext tests that RenderContext doesn't affect detection
func TestTerminalDetectionWithContext(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "vscode")

	// Test with different context configurations
	contexts := []RenderContext{
		{}, // Empty context
		{options: Options{WordWrap: 80}},
		{options: Options{Styles: StyleConfig{}}},
	}

	for i, ctx := range contexts {
		result := supportsHyperlinks(ctx)
		if !result {
			t.Errorf("Context %d should not affect hyperlink detection", i)
		}
	}
}

// clearTerminalEnvVars clears all terminal-related environment variables for clean testing
func clearTerminalEnvVars(t *testing.T) {
	t.Helper()

	terminalEnvVars := []string{
		"TERM_PROGRAM",
		"TERM_PROGRAM_VERSION",
		"TERM",
		"TERMINAL_EMULATOR",
		"KITTY_WINDOW_ID",
		"ALACRITTY_LOG",
		"ALACRITTY_SOCKET",
		"COLORTERM",
		"WT_SESSION",
		"SSH_CONNECTION",
		"WSL_DISTRO_NAME",
		"container",
	}

	for _, envVar := range terminalEnvVars {
		t.Setenv(envVar, "")
	}
}

// BenchmarkTerminalDetection benchmarks the hyperlink detection function
func BenchmarkTerminalDetection(b *testing.B) {
	scenarios := map[string]map[string]string{
		"iTerm2":         {"TERM_PROGRAM": "iTerm.app"},
		"VSCode":         {"TERM_PROGRAM": "vscode"},
		"xterm-256color": {"TERM": "xterm-256color"},
		"kitty":          {"KITTY_WINDOW_ID": "1"},
		"unsupported":    {"TERM": "dumb"},
		"empty":          {},
	}

	for name, envVars := range scenarios {
		b.Run(name, func(b *testing.B) {
			// Set up environment
			for key, value := range envVars {
				os.Setenv(key, value)
			}

			ctx := RenderContext{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				supportsHyperlinks(ctx)
			}
		})
	}
}
