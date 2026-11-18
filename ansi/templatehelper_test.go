package ansi

import "testing"

func TestTemplateHelperBounds(t *testing.T) {
	call := func(name string, values ...interface{}) string {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("%s%v panicked: %v", name, values, r)
			}
		}()
		return TemplateFuncMap[name].(func(...interface{}) string)(values...)
	}

	tests := []struct {
		name   string
		values []interface{}
		want   string
	}{
		{"Left", []interface{}{"glamour", 3}, "gla"},
		{"Left", []interface{}{"glamour", 7}, "glamour"},
		{"Left", []interface{}{"glamour", 8}, "glamour"},
		{"Left", []interface{}{"glamour", 0}, ""},
		{"Left", []interface{}{"glamour", -1}, ""},

		{"Mid", []interface{}{"glamour", 2}, "amour"},
		{"Mid", []interface{}{"glamour", 7}, ""},
		{"Mid", []interface{}{"glamour", 8}, ""},
		{"Mid", []interface{}{"glamour", -1}, "glamour"},
		{"Mid", []interface{}{"glamour", 2, 4}, "am"},
		{"Mid", []interface{}{"glamour", 2, 7}, "amour"},
		{"Mid", []interface{}{"glamour", 2, 8}, "amour"},
		{"Mid", []interface{}{"glamour", 4, 2}, ""},
		{"Mid", []interface{}{"glamour", 3, 3}, ""},

		{"Right", []interface{}{"glamour", 4}, "mour"},
		{"Right", []interface{}{"glamour", 7}, "glamour"},
		{"Right", []interface{}{"glamour", 8}, "glamour"},
		{"Right", []interface{}{"glamour", 0}, ""},
		{"Right", []interface{}{"glamour", -1}, ""},

		{"Last", []interface{}{[]string{"a", "b", "c"}}, "c"},
		{"Last", []interface{}{[]string{"a"}}, "a"},
		{"Last", []interface{}{[]string{}}, ""},
	}

	for _, tt := range tests {
		if got := call(tt.name, tt.values...); got != tt.want {
			t.Errorf("%s%v = %q, want %q", tt.name, tt.values, got, tt.want)
		}
	}
}

func TestTemplateHelpersValidSlices(t *testing.T) {
	t.Parallel()
	left := TemplateFuncMap["Left"].(func(...interface{}) string)
	right := TemplateFuncMap["Right"].(func(...interface{}) string)
	if left("glamour", 3) != "gla" || right("glamour", 4) != "mour" {
		t.Fatal("valid slices changed")
	}
}

func TestTemplateHelpersDoNotExpandEmptyRanges(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"Left", "Right"} {
		fn := TemplateFuncMap[name].(func(...interface{}) string)
		if got := fn("glamour", 0); got != "" {
			t.Fatalf("%s zero count = %q", name, got)
		}
	}
	mid := TemplateFuncMap["Mid"].(func(...interface{}) string)
	if got := mid("glamour", 3, 3); got != "" {
		t.Fatalf("empty range = %q", got)
	}
}

func TestTemplateRightMinimumInteger(t *testing.T) {
	t.Parallel()
	right := TemplateFuncMap["Right"].(func(...interface{}) string)
	minInt := -int(^uint(0)>>1) - 1
	if got := right("glamour", minInt); got != "" {
		t.Fatalf("Right with minimum integer = %q", got)
	}
}
