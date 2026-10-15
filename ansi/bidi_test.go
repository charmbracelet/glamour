package ansi

import "testing"

func TestReorderBidi(t *testing.T) {
	flow := func(s string) string { return bidiOpen + s + bidiClose }

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no flows", "مرحبا بالعالم", "مرحبا بالعالم"},
		{"left to right", flow("hello world"), "hello world"},
		{"right to left", flow("مرحبا بالعالم"), "ملاعلاب ابحرم"},
		{"hebrew", flow("שלום עולם"), "םלוע םולש"},
		{"punctuation", flow("مرحبا بالعالم."), ".ملاعلاب ابحرم"},
		{"latin in arabic", flow("مرحبا Glow عالم"), "ملاع Glow ابحرم"},
		{"arabic in latin", flow("Hello عالم world"), "Hello ملاع world"},
		{"number in arabic", flow("عام 2024"), "2024 ماع"},
		{"number after arabic in latin", flow("Hello عام 2024 world"), "Hello 2024 ماع world"},
		{"brackets", flow("(مرحبا)"), "(ابحرم)"},
		{"brackets in latin", flow("Hello عالم (برنامج) world"), "Hello (جمانرب) ملاع world"},
		{
			// عَلَيْكُم: the vowel marks stay on their letters.
			"combining marks",
			flow("عَلَيْكُم"),
			"مكُيْلَعَ",
		},
		{"surrounding whitespace", flow("  مرحبا بالعالم  "), "  ملاعلاب ابحرم  "},
		{"outside of flow", "• " + flow("مرحبا") + "\n• " + flow("hello"), "• ابحرم\n• hello"},
		{
			// The second line starts with a Latin word, but belongs to a
			// right-to-left paragraph.
			"wrapped flow",
			flow("مرحبا\nGlow هنا"),
			"ابحرم\nانه Glow",
		},
		{"styles", flow("\x1b[1mمرحبا\x1b[m بالعالم"), "ملاعلاب \x1b[1mابحرم\x1b[m"},
		{
			"hyperlinks",
			flow("\x1b]8;;https://charm.land\aرابط\x1b]8;;\a نص"),
			"صن \x1b]8;;https://charm.land\aطبار\x1b]8;;\a",
		},
		{"unsupported escape sequence", flow("\x1b[2Kمرحبا"), "\x1b[2Kمرحبا"},
		{"empty flow", "a" + flow("") + "b", "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reorderBidi(tt.in); got != tt.want {
				t.Errorf("reorderBidi(%q)\n got: %q\nwant: %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCloseBidiFlow(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"# " + bidiOpen + "title", "# " + bidiOpen + "title" + bidiClose},
		{"# " + bidiOpen + " ", "#  "},
		{"no flow", "no flow"},
	}
	for _, tt := range tests {
		if got := closeBidiFlow(tt.in); got != tt.want {
			t.Errorf("closeBidiFlow(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
