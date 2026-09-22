package ansi

import (
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testBadgeSVG is a badge in the shape of the ones served by shields.io: a
// rounded background split in two, and text drawn in scaled groups.
const testBadgeSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="94" height="20" role="img" aria-label="release: v3.0.0">` +
	`<title>release: v3.0.0</title>` +
	`<filter id="blur"><feGaussianBlur stdDeviation="16"/></filter>` +
	`<clipPath id="r"><rect width="94" height="20" rx="3"/></clipPath>` +
	`<g clip-path="url(#r)"><rect width="49" height="20" fill="#555"/>` +
	`<rect x="49" width="45" height="20" fill="#007ec6"/></g>` +
	`<g fill="#fff" text-anchor="middle" font-family="Verdana,sans-serif" font-size="110">` +
	`<g transform="scale(.1)"><g aria-hidden="true" fill="#010101">` +
	`<text x="255" y="150" fill-opacity=".8" filter="url(#blur)" textLength="390">release</text>` +
	`<text x="255" y="150" fill-opacity=".3" textLength="390">release</text></g>` +
	`<text x="255" y="140" textLength="390">release</text></g>` +
	`<g transform="scale(.1)"><g aria-hidden="true" fill="#010101">` +
	`<text x="705" y="150" fill-opacity=".8" filter="url(#blur)" textLength="350">v3.0.0</text>` +
	`<text x="705" y="150" fill-opacity=".3" textLength="350">v3.0.0</text></g>` +
	`<text x="705" y="140" textLength="350">v3.0.0</text></g></g></svg>`

// TestSVGDisplaySize checks that SVG images are sized by their intrinsic CSS
// pixel size, i.e. one cell per cellPixelWidth x cellPixelHeight pixels.
func TestSVGDisplaySize(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		opts          Options
		wantCols      int
		wantRows      int
	}{
		{
			name:  "badge is drawn at its own size",
			width: 94, height: 20,
			opts:     Options{WordWrap: 80},
			wantCols: 10, wantRows: 1,
		},
		{
			name:  "large image is constrained to the available width",
			width: 900, height: 200,
			opts:     Options{WordWrap: 80},
			wantCols: 80, wantRows: 9,
		},
		{
			name:  "max columns",
			width: 900, height: 200,
			opts:     Options{WordWrap: 80, MaxImageColumns: 45},
			wantCols: 45, wantRows: 5,
		},
		{
			name:  "max rows preserves the aspect ratio",
			width: 400, height: 400,
			opts:     Options{WordWrap: 80, MaxImageRows: 4},
			wantCols: 8, wantRows: 4,
		},
		{
			name:  "tiny image is at least one cell",
			width: 4, height: 4,
			opts:     Options{WordWrap: 80},
			wantCols: 1, wantRows: 1,
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewRenderContext(tc.opts)
			cols, rows := svgDisplaySize(tc.width, tc.height, ctx)
			if cols != tc.wantCols || rows != tc.wantRows {
				t.Errorf("expected %dx%d cells, got %dx%d", tc.wantCols, tc.wantRows, cols, rows)
			}
		})
	}
}

// TestSVGBadgeRendering checks that a badge is displayed at its own size and
// that its text is drawn, as oksvg alone would leave the badge without one.
func TestSVGBadgeRendering(t *testing.T) {
	// The badge's group carries the text attributes and the transform; the
	// runs must be resolved to the document's coordinates.
	runs := svgTextRuns([]byte(testBadgeSVG))
	if len(runs) != 4 {
		t.Fatalf("expected 4 text runs, got %d: %+v", len(runs), runs)
	}
	// The blurred copy of each label is skipped; the sharp one and the
	// crisper shadow remain.
	labels := map[string]bool{}
	for _, r := range runs {
		labels[r.text] = true
		if r.size < 10 || r.size > 12 {
			t.Errorf("expected the font size to be scaled by the group's transform, got %v", r.size)
		}
		if r.anchor != "middle" {
			t.Errorf("expected the text anchor to be inherited, got %q", r.anchor)
		}
	}
	if !labels["release"] || !labels["v3.0.0"] {
		t.Errorf("expected both labels, got %v", labels)
	}

	// Labels are centered on their half of the badge: the first one around
	// x=25.5 and the second around x=70.5, at the text's baseline.
	for _, r := range runs {
		want := 25.5
		if r.text == "v3.0.0" {
			want = 70.5
		}
		if math.Abs(r.x-want) > 0.01 {
			t.Errorf("expected %q to be positioned at x=%v, got x=%v", r.text, want, r.x)
		}
		if math.Abs(r.y-14) > 0.01 && math.Abs(r.y-15) > 0.01 {
			t.Errorf("expected a baseline around y=14, got y=%v", r.y)
		}
		if r.fill.A == 0xff && (r.fill.R != 0xff || r.fill.G != 0xff || r.fill.B != 0xff) {
			t.Errorf("expected white text, got %v", r.fill)
		}
	}

	// The rasterized badge must contain the drawn text: white pixels on the
	// white label background and dark ones on the colored background.
	img, err := rasterizeSVGReader(strings.NewReader(testBadgeSVG), 940, 200)
	if err != nil {
		t.Fatal(err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 940 || bounds.Dy() != 200 {
		t.Errorf("expected a 940x200 raster, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	var leftWhite, rightWhite int
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r > 0xe000 && g > 0xe000 && b > 0xe000 {
				if x < bounds.Max.X/2 {
					leftWhite++
				} else {
					rightWhite++
				}
			}
		}
	}
	// Both labels are drawn in white, on their half of the badge.
	if leftWhite == 0 {
		t.Error("expected white text pixels on the left half of the badge")
	}
	if rightWhite == 0 {
		t.Error("expected white text pixels on the right half of the badge")
	}
}

// TestSVGTextRuns checks the attributes the text parser resolves.
func TestSVGTextRuns(t *testing.T) {
	runs := svgTextRuns([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="40">
		<g fill="#fff" font-size="20" text-anchor="end">
			<text x="10" y="20" transform="translate(5,5) scale(2)">hello <tspan fill="rgb(1,2,3)">world</tspan></text>
		</g>
		<text x="10" y="20" fill="none">hidden</text>
		<text x="10" y="20" fill="red" font-weight="600" style="font-size: 8px">bold</text>
	</svg>`))

	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d: %+v", len(runs), runs)
	}

	first := runs[0]
	if first.text != "hello" {
		t.Errorf("expected the text to be trimmed, got %q", first.text)
	}
	// translate(5,5) then scale(2): the point is scaled first, then moved.
	if math.Abs(first.x-25) > 0.01 || math.Abs(first.y-45) > 0.01 {
		t.Errorf("expected the transform to be applied, got x=%v y=%v", first.x, first.y)
	}
	if math.Abs(first.size-40) > 0.01 {
		t.Errorf("expected the font size to be scaled to 40, got %v", first.size)
	}
	if first.anchor != "end" {
		t.Errorf("expected the anchor to be inherited, got %q", first.anchor)
	}

	// A tspan without a position continues after the text before it.
	second := runs[1]
	if second.text != "world" {
		t.Errorf("expected the tspan text, got %q", second.text)
	}
	if second.x <= first.x {
		t.Errorf("expected the tspan to continue after the text, got x=%v", second.x)
	}
	if second.fill != (color.RGBA{R: 1, G: 2, B: 3, A: 0xff}) {
		t.Errorf("expected the tspan color, got %v", second.fill)
	}

	// fill="none" text is not drawn; a style attribute overrides the
	// inherited font size, and a bold weight switches font.
	third := runs[2]
	if third.text != "bold" {
		t.Errorf("expected the bold run, got %q", third.text)
	}
	if !third.bold {
		t.Error("expected the run to be bold")
	}
	if math.Abs(third.size-8) > 0.01 {
		t.Errorf("expected the font size from the style attribute, got %v", third.size)
	}
	if third.fill != (color.RGBA{R: 0xff, A: 0xff}) {
		t.Errorf("expected the red fill, got %v", third.fill)
	}
}

// TestParseSVGColor checks the color formats found in SVG documents.
func TestParseSVGColor(t *testing.T) {
	tests := []struct {
		in   string
		want color.RGBA
	}{
		{"#fff", color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}},
		{"#007ec6", color.RGBA{R: 0x00, G: 0x7e, B: 0xc6, A: 0xff}},
		{"#01010180", color.RGBA{R: 0x01, G: 0x01, B: 0x01, A: 0x80}},
		{"rgb(1,2,3)", color.RGBA{R: 1, G: 2, B: 3, A: 0xff}},
		{"rgba(1, 2, 3)", color.RGBA{R: 1, G: 2, B: 3, A: 0xff}},
		{"white", color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}},
		{"Black", color.RGBA{A: 0xff}},
	}
	for _, tc := range tests {
		got, err := parseSVGColor(tc.in)
		if err != nil {
			t.Errorf("parseSVGColor(%q) returned an error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseSVGColor(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
	if _, err := parseSVGColor("chartreuse"); err == nil {
		t.Error("expected an error for an unknown color")
	}
}

// TestSVGTextRunsMalformed checks that a broken document doesn't panic and
// keeps the text parsed up to the point where it breaks.
func TestSVGTextRunsMalformed(t *testing.T) {
	runs := svgTextRuns([]byte(`<svg width="10" height="10"><text x="1" y="2">ok</text><g`))
	if len(runs) != 1 || runs[0].text != "ok" {
		t.Errorf("expected the parsed text, got %+v", runs)
	}
}

// TestSVGImageConfig checks that an SVG's intrinsic size is read from its
// document, and that a document without one is rejected.
func TestSVGImageConfig(t *testing.T) {
	cfg, err := readSVGConfig(strings.NewReader(testBadgeSVG))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.format != svgFormat {
		t.Errorf("expected the svg format, got %q", cfg.format)
	}
	if cfg.config.Width != 94 || cfg.config.Height != 20 {
		t.Errorf("expected the badge's size, got %dx%d", cfg.config.Width, cfg.config.Height)
	}
	if _, err := readSVGConfig(strings.NewReader(`<svg><rect width="1" height="1"/></svg>`)); err == nil {
		t.Error("expected an error for an SVG without an intrinsic size")
	}
}

// TestSVGRasterScale checks that scaling an SVG keeps its aspect ratio and
// that the text stays on the image.
func TestSVGRasterScale(t *testing.T) {
	img, err := rasterizeSVGReader(strings.NewReader(testBadgeSVG), 188, 40)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 188 || b.Dy() != 40 {
		t.Errorf("expected a 188x40 raster, got %dx%d", b.Dx(), b.Dy())
	}

	// The blue half of the badge holds white text. Sample the label's
	// rectangle and count the white pixels there.
	var white int
	for x := 100; x < 188; x++ {
		for y := 0; y < 40; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r > 0xe000 && g > 0xe000 && b > 0xe000 {
				white++
			}
		}
	}
	if white < 40 {
		t.Errorf("expected the text to be drawn at the scaled size, got %d white pixels", white)
	}
}

// TestSVGImageRenders checks that an SVG image is rendered in a document,
// sized to its intrinsic size in cells.
func TestSVGImageRenders(t *testing.T) {
	svgPath := writeTestSVG(t, testBadgeSVG)

	out, ar := renderImage(t, Options{WordWrap: 80, ImageProtocol: ImageProtocolKittyPlaceholders}, "![]("+svgPath+")")
	if !strings.ContainsRune(out, kittyPlaceholder) {
		t.Fatalf("expected unicode placeholders in document, got: %q", out)
	}

	// A 94x20 badge is 10x1 cells: one row of ten placeholders.
	cmds := ar.GraphicsCommands()
	if len(cmds) != 2 {
		t.Fatalf("expected transmit and place commands, got %d", len(cmds))
	}
	if !strings.Contains(cmds[1], "c=10,r=1") {
		t.Errorf("expected the badge to be placed in a 10x1 cell box, got: %q", cmds[1])
	}
	lines := strings.Split(strings.TrimPrefix(out, "\n"), "\n")
	var gridLines []string
	for _, l := range lines {
		if strings.ContainsRune(l, kittyPlaceholder) {
			gridLines = append(gridLines, l)
		}
	}
	if len(gridLines) != 1 {
		t.Fatalf("expected a single grid row, got %d: %q", len(gridLines), out)
	}
	if got := strings.Count(gridLines[0], string(kittyPlaceholder)); got != 10 {
		t.Errorf("expected 10 placeholder cells, got %d", got)
	}
}

// TestSVGImageRaster is a sanity check that rasterizing an SVG yields an
// image with the badge's colors.
func TestSVGImageRaster(t *testing.T) {
	img, err := rasterizeSVGReader(strings.NewReader(testBadgeSVG), 94, 20)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 94 || b.Dy() != 20 {
		t.Fatalf("expected a 94x20 raster, got %dx%d", b.Dx(), b.Dy())
	}
	// The left half is #555 and the right one #007ec6.
	if r, g, b, _ := img.At(5, 10).RGBA(); r>>8 != 0x55 || g>>8 != 0x55 || b>>8 != 0x55 {
		t.Errorf("expected the label background to be #555, got %04x%04x%04x", r, g, b)
	}
	right := img.Bounds().Max.X
	if r, g, b, _ := img.At(right-5, 10).RGBA(); b>>8 != 0xc6 || g>>8 != 0x7e {
		t.Errorf("expected the value background to be #007ec6, got %04x%04x%04x", r, g, b)
	}
}

// writeTestSVG writes an SVG document to a temporary file and returns its
// path.
func writeTestSVG(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.svg")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestNormalizeSVGPath checks that path data is rewritten into a form parsers
// read reliably.
func TestNormalizeSVGPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "arc flags glued to the following number",
			in:   "M5 15a.977.977 0 00-.275-.082",
			want: "M5,15a0.977,0.977,0,0,0,-0.275,-0.082",
		},
		{
			name: "arc flags on their own",
			in:   "a8 8 0 0 1 16 0",
			want: "a8,8,0,0,1,16,0",
		},
		{
			name: "arc flags glued to each other",
			in:   "a8 8 0 0116 0",
			want: "a8,8,0,0,1,16,0",
		},
		{
			name: "implicit decimals",
			in:   "l1.5.5",
			want: "l1.5,0.5",
		},
		{
			name: "repeated commands",
			in:   "c1 2 3 4 5 6 7 8 9 10 11 12",
			want: "c1,2,3,4,5,6c7,8,9,10,11,12",
		},
		{
			name: "moveto pairs become line-tos",
			in:   "M1 2 3 4",
			want: "M1,2L3,4",
		},
		{
			name: "closepath and relative moveto",
			in:   "M1 1h2v2H1z m2 0h2",
			want: "M1,1h2v2H1zm2,0h2",
		},
		{
			name: "exponents are kept",
			in:   "l1e-3 2",
			want: "l0.001,2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeSVGPath(tc.in); got != tc.want {
				t.Errorf("normalizeSVGPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestNormalizeSVGPaths checks that the path data of a document is rewritten
// without touching other attributes.
func TestNormalizeSVGPaths(t *testing.T) {
	in := `<svg width="10" height="10" stroke-dasharray="1,2"><path d="M1 1a.5.5 0 001 1" stroke="red"/></svg>`
	want := `<svg width="10" height="10" stroke-dasharray="1,2"><path d="M1,1a0.5,0.5,0,0,0,1,1" stroke="red"/></svg>`
	if got := string(normalizeSVGPaths([]byte(in))); got != want {
		t.Errorf("normalizeSVGPaths() = %q, want %q", got, want)
	}
}

// TestSVGArcsRender checks that paths with arcs are drawn, and that the
// document's shapes are all rendered.
func TestSVGArcsRender(t *testing.T) {
	// A rounded badge shape drawn with arcs, closed by hand, plus text.
	svg := `<svg width="40" height="20" xmlns="http://www.w3.org/2000/svg">` +
		`<path d="M2 0h36a2 2 0 012 2v16a2 2 0 01-2 2H2a2 2 0 01-2-2V2a2 2 0 012-2z" fill="#5C5C5C"/>` +
		`<path d="M20 5a5 5 0 100 10 5 5 0 000-10z" fill="#fff"/>` +
		`</svg>`

	img, err := rasterizeSVGReader(strings.NewReader(svg), 400, 200)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 400 || b.Dy() != 200 {
		t.Fatalf("expected a 400x200 raster, got %dx%d", b.Dx(), b.Dy())
	}

	// The rounded rectangle fills the badge, and the circle is drawn in
	// white on top of it.
	var gray, white int
	for x := 0; x < 400; x++ {
		for y := 0; y < 200; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			switch {
			case r>>8 == 0x5c && g>>8 == 0x5c && b>>8 == 0x5c:
				gray++
			case r > 0xe000 && g > 0xe000 && b > 0xe000:
				white++
			}
		}
	}
	if gray < 400*200/4 {
		t.Errorf("expected the badge shape to be filled, got %d gray pixels", gray)
	}
	if white < 1000 {
		t.Errorf("expected the circle to be drawn, got %d white pixels", white)
	}
}

// TestInlineBadges checks that images that are one cell row tall, e.g. badges,
// are drawn in the text flow, so images next to each other stay on one line.
func TestInlineBadges(t *testing.T) {
	dir := t.TempDir()
	badge := func(name string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(testBadgeSVG), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	one, two := badge("one.svg"), badge("two.svg")

	md := "<p>\n" +
		"  <a href=\"https://example.com\"><img src=\"" + one + "\"></a>\n" +
		"  <a href=\"https://example.com\"><img src=\"" + two + "\"></a>\n" +
		"</p>\n"

	out, ar := renderImage(t, Options{
		WordWrap:      80,
		ImageProtocol: ImageProtocolKittyPlaceholders,
	}, md)

	// Both badges are placed in the same line of text, with a space between
	// them, as a browser would render them.
	var gridLines []string
	for _, l := range strings.Split(strings.TrimPrefix(out, "\n"), "\n") {
		if strings.ContainsRune(l, kittyPlaceholder) {
			gridLines = append(gridLines, l)
		}
	}
	if len(gridLines) != 1 {
		t.Fatalf("expected both badges on one line, got %d lines: %q", len(gridLines), out)
	}
	if got := strings.Count(gridLines[0], string(kittyPlaceholder)); got != 20 {
		t.Errorf("expected 20 placeholder cells, got %d: %q", got, gridLines[0])
	}
	// The badges are separated by a space, and neither starts a line.
	grid := visibleText(gridLines[0])
	grid = strings.TrimLeft(grid, " ")
	if len([]rune(grid)) != 0 && strings.HasPrefix(grid, " ") {
		t.Errorf("expected the line to start with a badge, got: %q", grid)
	}
	if !strings.Contains(grid, string(kittyPlaceholder)+" ") {
		t.Errorf("expected a space between the badges, got: %q", grid)
	}
	if cmds := ar.GraphicsCommands(); len(cmds) != 4 {
		t.Errorf("expected transmit and place commands for both badges, got %d", len(cmds))
	}
}

// TestImageOrderInHTMLBlocks checks that images are drawn in the order they
// appear in the document: HTML blocks are not wrapped, so their images can be
// written in place, and a banner ahead of a row of badges stays ahead of it.
func TestImageOrderInHTMLBlocks(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"banner.svg": `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="60"><rect width="200" height="60" fill="#555"/></svg>`,
		"badge.svg":  testBadgeSVG,
		"demo.svg":   `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="120"><rect width="300" height="120" fill="#333"/></svg>`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Use absolute paths: base URLs with a drive letter don't survive the
	// URL resolution on Windows.
	banner, badge, demo := filepath.Join(dir, "banner.svg"), filepath.Join(dir, "badge.svg"), filepath.Join(dir, "demo.svg")

	doc := "# Title\n\n" +
		"<p align=\"center\">\n" +
		"    <img src=\"" + banner + "\" alt=\"Banner\">\n" +
		"    <a href=\"https://example.com\"><img src=\"" + badge + "\" alt=\"Release\"></a>\n" +
		"</p>\n\n" +
		"<p align=\"center\">\n" +
		"    <img src=\"" + demo + "\" alt=\"Demo\">\n" +
		"</p>\n\n" +
		"## What is it?\n"

	out, _ := renderImage(t, Options{
		WordWrap: 100, ImageProtocol: ImageProtocolKittyPlaceholders,
	}, doc)

	// Collect the grid lines, in order: the banner's three rows of 20 cells,
	// the badge row of 10 cells, then the demo's rows of 30 cells.
	var cells []int
	textBetween := -1
	for i, line := range strings.Split(out, "\n") {
		vis := visibleText(line)
		if strings.ContainsRune(vis, kittyPlaceholder) {
			if textBetween >= 0 {
				t.Errorf("expected the images to come before %q, got one at line %d", strings.TrimSpace(visibleText(strings.Split(out, "\n")[textBetween])), i)
			}
			cells = append(cells, strings.Count(vis, string(kittyPlaceholder)))
			continue
		}
		// Headings and paragraphs come last.
		if trimmed := strings.TrimSpace(vis); trimmed == "## What is it?" && textBetween < 0 {
			textBetween = i
		}
	}

	want := []int{20, 20, 20, 10, 30, 30, 30, 30, 30, 30}
	if len(cells) != len(want) {
		t.Fatalf("expected %d image rows, got %d: %v", len(want), len(cells), cells)
	}
	for i := range want {
		if cells[i] != want[i] {
			t.Errorf("image row %d: expected %d cells, got %d (rows: %v)", i, want[i], cells[i], cells)
		}
	}
}
