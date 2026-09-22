package ansi

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// svgSniffLen is the number of leading bytes inspected to tell an SVG
// document (XML) apart from a raster image.
const svgSniffLen = 512

// sniffHead returns the leading bytes of r without consuming them, for
// telling SVGs apart from raster images. Fewer bytes are returned for
// shorter input.
func sniffHead(r *bufio.Reader) []byte {
	head, _ := r.Peek(svgSniffLen)
	return head
}

// svgFormat is the format name reported for SVG images in imageConfig.
const svgFormat = "svg"

// maxSVGRasterPixels caps the pixel count of a rasterized SVG. A vector
// image can be declared at any size, so rasterizing it is bounded to keep
// huge documents from exhausting memory; the terminal scales the result to
// the display box.
const maxSVGRasterPixels = 8 << 20 // 8 MP

// isSVGURL reports whether the image URL points at an SVG document by its
// file extension or data URI media type. SVG documents named otherwise are
// detected by isSVGData.
func isSVGURL(url string) bool {
	l := strings.ToLower(url)
	return strings.HasSuffix(l, ".svg") || strings.HasPrefix(l, "data:image/svg")
}

// isSVGData reports whether the data is an SVG document, i.e. XML with an
// <svg> root element, skipping a BOM, the XML declaration, comments and a
// DOCTYPE.
func isSVGData(data []byte) bool {
	rest := bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	for {
		rest = bytes.TrimSpace(rest)
		switch {
		case bytes.HasPrefix(rest, []byte("<?xml")):
			i := bytes.Index(rest, []byte("?>"))
			if i < 0 {
				return false
			}
			rest = rest[i+2:]
		case bytes.HasPrefix(rest, []byte("<!--")):
			i := bytes.Index(rest, []byte("-->"))
			if i < 0 {
				return false
			}
			rest = rest[i+3:]
		case bytes.HasPrefix(rest, []byte("<!")):
			i := bytes.Index(rest, []byte(">"))
			if i < 0 {
				return false
			}
			rest = rest[i+1:]
		default:
			return bytes.HasPrefix(rest, []byte("<svg"))
		}
	}
}

// parseSVG parses an SVG document and returns it with its intrinsic size.
func parseSVG(r io.Reader) (*oksvg.SvgIcon, error) {
	icon, err := oksvg.ReadIconStream(r)
	if err != nil {
		return nil, fmt.Errorf("glamour: error parsing SVG image: %w", err)
	}
	if icon.ViewBox.W <= 0 || icon.ViewBox.H <= 0 {
		return nil, fmt.Errorf("glamour: SVG image has no intrinsic size")
	}
	return icon, nil
}

// readSVGConfig parses the SVG document and returns its intrinsic size as
// the image config.
func readSVGConfig(r io.Reader) (imageConfig, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return imageConfig{}, fmt.Errorf("glamour: error reading SVG image: %w", err)
	}
	icon, err := parseSVG(bytes.NewReader(normalizeSVGPaths(data)))
	if err != nil {
		return imageConfig{}, err
	}
	return imageConfig{
		config: image.Config{
			Width:  int(math.Round(icon.ViewBox.W)),
			Height: int(math.Round(icon.ViewBox.H)),
		},
		format: svgFormat,
	}, nil
}

// svgDisplaySize computes the display size of an SVG image in terminal cells
// from its intrinsic size in CSS pixels. An SVG's declared size is the size
// its author chose for display, so the image is drawn at exactly that size:
// one cell per cellPixelWidth x cellPixelHeight pixels, constrained to the
// available width and the configured limits.
func svgDisplaySize(width, height int, ctx RenderContext) (int, int) {
	cols := max(1, int(math.Ceil(float64(width)/cellPixelWidth)))
	rows := max(1, int(math.Ceil(float64(height)/cellPixelHeight)))

	maxCols := int(ctx.blockStack.Width(ctx))
	if maxCols <= 0 {
		maxCols = cols
	}
	if ctx.options.MaxImageColumns > 0 {
		maxCols = min(maxCols, ctx.options.MaxImageColumns)
	}
	if cols > maxCols {
		rows = max(1, int(math.Round(float64(rows)*float64(maxCols)/float64(cols))))
		cols = maxCols
	}
	if maxRows := ctx.options.MaxImageRows; maxRows > 0 && rows > maxRows {
		cols = max(1, int(math.Round(float64(cols)*float64(maxRows)/float64(rows))))
		rows = maxRows
	}
	return cols, rows
}

// rasterizeSVGReader parses and rasterizes the SVG document, see
// [rasterizeSVG].
func rasterizeSVGReader(r io.Reader, maxW, maxH int) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("glamour: error reading SVG image: %w", err)
	}
	data = normalizeSVGPaths(data)
	icon, err := parseSVG(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return rasterizeSVG(icon, data, maxW, maxH)
}

// rasterizeSVG rasterizes the SVG document at up to maxW x maxH pixels,
// preserving the aspect ratio; vectors scale cleanly, so the image is
// rasterized at the largest size that fits the box. A non-positive bound
// rasterizes at hiDPIScale times the intrinsic size. The result is capped
// to maxSVGRasterPixels to bound memory use.
func rasterizeSVG(icon *oksvg.SvgIcon, data []byte, maxW, maxH int) (image.Image, error) {
	w, h := icon.ViewBox.W, icon.ViewBox.H
	scale := float64(hiDPIScale)
	if maxW > 0 && maxH > 0 {
		scale = math.Min(float64(maxW)/w, float64(maxH)/h)
	}
	if w*h*scale*scale > maxSVGRasterPixels {
		scale = math.Sqrt(maxSVGRasterPixels / (w * h))
	}
	width := max(1, int(math.Round(w*scale)))
	height := max(1, int(math.Round(h*scale)))

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	scanner := rasterx.NewScannerGV(width, height, img, img.Bounds())
	dasher := rasterx.NewDasher(width, height, scanner)
	icon.SetTarget(0, 0, float64(width), float64(height))
	icon.Draw(dasher, 1.0)

	// oksvg doesn't render text, so the document's text runs are drawn over
	// the rasterized shapes here.
	if runs := svgTextRuns(data); len(runs) > 0 {
		drawSVGTexts(img, icon.ViewBox.X, icon.ViewBox.Y,
			float64(width)/icon.ViewBox.W, float64(height)/icon.ViewBox.H, runs)
	}
	return img, nil
}

// svgTextRun is a single run of SVG text with its position and style,
// resolved to viewBox coordinates.
type svgTextRun struct {
	text   string
	x, y   float64 // the baseline position, in viewBox units
	size   float64 // the font size, in viewBox units
	fill   color.RGBA
	anchor string // start, middle or end
	bold   bool
}

// svgStyle holds the presentation attributes that SVG text elements inherit
// from their ancestors.
type svgStyle struct {
	fill     color.RGBA
	opacity  float64
	fontSize float64
	anchor   string
	bold     bool
	noFill   bool // the element is explicitly not filled
	filtered bool // the element is filtered, e.g. a blurred text shadow
	m        svgMatrix
}

// svgTextRuns extracts the SVG document's text runs. oksvg, which parses the
// shapes, doesn't render text, so it is parsed separately here. Malformed
// documents return whatever was parsed so far.
func svgTextRuns(data []byte) []svgTextRun {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var stack []svgStyle
	cur := svgStyle{fill: color.RGBA{A: 0xff}, opacity: 1, fontSize: 11, m: svgIdentity()}

	var runs []svgTextRun
	var penX, penY float64
	textDepth := 0

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, cur)
			cur = svgApplyAttributes(cur, t)
			if t.Name.Local == "text" || t.Name.Local == "tspan" {
				textDepth++
				if x, ok := svgAttrFloat(t, "x"); ok {
					penX = x
				}
				if y, ok := svgAttrFloat(t, "y"); ok {
					penY = y
				}
			}
		case xml.CharData:
			text := strings.Join(strings.Fields(string(t)), " ")
			if textDepth == 0 || cur.filtered || cur.noFill || text == "" {
				continue
			}
			x, y := cur.m.point(penX, penY)
			size := cur.fontSize * cur.m.scale()
			fill := cur.fill
			fill.A = uint8(math.Round(math.Min(float64(fill.A), 0xff*cur.opacity)))
			runs = append(runs, svgTextRun{
				text: text, x: x, y: y, size: size,
				fill: fill, anchor: cur.anchor, bold: cur.bold,
			})
			// Advance the pen for any character data that follows, e.g. a
			// second tspan without its own x.
			if s := cur.m.scale(); s > 0 {
				penX += svgMeasure(text, size, cur.bold) / s
			}
		case xml.EndElement:
			if t.Name.Local == "text" || t.Name.Local == "tspan" {
				textDepth--
			}
			cur, stack = stack[len(stack)-1], stack[:len(stack)-1]
		}
	}
	return runs
}

// svgAttributes returns the element's attributes, including the ones packed
// into the style attribute.
func svgAttributes(el xml.StartElement) map[string]string {
	attrs := make(map[string]string, len(el.Attr))
	for _, a := range el.Attr {
		attrs[strings.ToLower(a.Name.Local)] = a.Value
	}
	if v := attrs["style"]; v != "" {
		for _, decl := range strings.Split(v, ";") {
			name, value, ok := strings.Cut(decl, ":")
			if !ok {
				continue
			}
			attrs[strings.ToLower(strings.TrimSpace(name))] = strings.TrimSpace(value)
		}
	}
	return attrs
}

// svgApplyAttributes returns the style for an element, inheriting from the
// given style.
func svgApplyAttributes(s svgStyle, el xml.StartElement) svgStyle {
	attrs := svgAttributes(el)

	if attrs["filter"] != "" {
		// Filters are usually blur effects, e.g. the softened shadow below
		// a text. They aren't supported; the text is skipped instead of
		// drawn sharp.
		s.filtered = true
	}
	if v := attrs["fill"]; v != "" && v != "inherit" {
		if v == "none" {
			s.noFill = true
		} else if c, err := parseSVGColor(v); err == nil {
			s.fill = c
			s.noFill = false
		}
	}
	for _, key := range []string{"fill-opacity", "opacity"} {
		if v := attrs[key]; v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				s.opacity *= f
			}
		}
	}
	if v := attrs["font-size"]; v != "" {
		if f, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(v), "px"), 64); err == nil && f > 0 {
			s.fontSize = f
		}
	}
	if v := attrs["font-weight"]; v != "" {
		if v == "bold" || v == "bolder" {
			s.bold = true
		} else if f, err := strconv.Atoi(v); err == nil {
			s.bold = f >= 600
		}
	}
	if v := attrs["text-anchor"]; v != "" {
		s.anchor = v
	}
	if v := attrs["transform"]; v != "" {
		s.m = s.m.times(svgParseTransform(v))
	}
	return s
}

// svgAttrFloat returns the element's attribute parsed as a float.
func svgAttrFloat(el xml.StartElement, name string) (float64, bool) {
	for _, a := range el.Attr {
		if a.Name.Local == name {
			f, err := strconv.ParseFloat(strings.TrimSpace(a.Value), 64)
			return f, err == nil
		}
	}
	return 0, false
}

// svgColorNames maps the named SVG colors used in badges and icons. Hex and
// rgb() colors are parsed by parseSVGColor.
var svgColorNames = map[string]color.RGBA{
	"black":  {0x00, 0x00, 0x00, 0xff},
	"blue":   {0x00, 0x00, 0xff, 0xff},
	"gray":   {0x80, 0x80, 0x80, 0xff},
	"green":  {0x00, 0x80, 0x00, 0xff},
	"grey":   {0x80, 0x80, 0x80, 0xff},
	"orange": {0xff, 0xa5, 0x00, 0xff},
	"purple": {0x80, 0x00, 0x80, 0xff},
	"red":    {0xff, 0x00, 0x00, 0xff},
	"white":  {0xff, 0xff, 0xff, 0xff},
	"yellow": {0xff, 0xff, 0x00, 0xff},
}

// parseSVGColor parses an SVG color: a #rgb or #rrggbb hex value, an
// rgb()/rgba() function, or a named color.
func parseSVGColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	if c, ok := svgColorNames[strings.ToLower(s)]; ok {
		return c, nil
	}

	if hex, ok := strings.CutPrefix(s, "#"); ok {
		if len(hex) == 3 || len(hex) == 4 {
			var b []byte
			for i := 0; i < len(hex); i++ {
				b = append(b, hex[i], hex[i])
			}
			hex = string(b)
		}
		if len(hex) == 6 || len(hex) == 8 {
			v, err := strconv.ParseUint(hex, 16, 32)
			if err != nil {
				return color.RGBA{}, fmt.Errorf("glamour: invalid SVG color %q", s)
			}
			c := color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 0xff}
			if len(hex) == 8 {
				c = color.RGBA{uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), uint8(v)}
			}
			return c, nil
		}
		return color.RGBA{}, fmt.Errorf("glamour: invalid SVG color %q", s)
	}

	if strings.HasPrefix(s, "rgb(") || strings.HasPrefix(s, "rgba(") {
		inner := s[strings.IndexByte(s, '(')+1 : len(s)-1]
		parts := strings.FieldsFunc(inner, func(r rune) bool {
			return r == ',' || r == ' ' || r == '/'
		})
		if len(parts) >= 3 {
			var c color.RGBA
			c.A = 0xff
			for i := 0; i < 3; i++ {
				v, err := strconv.ParseUint(parts[i], 10, 8)
				if err != nil {
					return color.RGBA{}, fmt.Errorf("glamour: invalid SVG color %q", s)
				}
				switch i {
				case 0:
					c.R = uint8(v)
				case 1:
					c.G = uint8(v)
				case 2:
					c.B = uint8(v)
				}
			}
			return c, nil
		}
	}
	return color.RGBA{}, fmt.Errorf("glamour: unknown SVG color %q", s)
}

// svgMatrix is a 2D affine transformation mapping (x, y) to
// (a*x + c*y + e, b*x + d*y + f).
type svgMatrix struct{ a, b, c, d, e, f float64 }

func svgIdentity() svgMatrix { return svgMatrix{a: 1, d: 1} }

func svgTranslate(x, y float64) svgMatrix { return svgMatrix{a: 1, d: 1, e: x, f: y} }

func svgScale(x, y float64) svgMatrix { return svgMatrix{a: x, d: y} }

// times returns m applied after o.
func (m svgMatrix) times(o svgMatrix) svgMatrix {
	return svgMatrix{
		a: m.a*o.a + m.c*o.b,
		b: m.b*o.a + m.d*o.b,
		c: m.a*o.c + m.c*o.d,
		d: m.b*o.c + m.d*o.d,
		e: m.a*o.e + m.c*o.f + m.e,
		f: m.b*o.e + m.d*o.f + m.f,
	}
}

// point applies the transformation to (x, y).
func (m svgMatrix) point(x, y float64) (float64, float64) {
	return m.a*x + m.c*y + m.e, m.b*x + m.d*y + m.f
}

// scale returns the absolute scale of the transformation, used to scale the
// font size.
func (m svgMatrix) scale() float64 {
	return math.Sqrt(math.Abs(m.a*m.d - m.b*m.c))
}

// svgTransformPattern matches a transform function, e.g. scale(.1).
var svgTransformPattern = regexp.MustCompile(`([a-zA-Z]+)\s*\(([^)]*)\)`)

// svgParseTransform parses a transform attribute. translate, scale and
// matrix are supported; other functions are ignored.
func svgParseTransform(s string) svgMatrix {
	m := svgIdentity()
	for _, match := range svgTransformPattern.FindAllStringSubmatch(s, -1) {
		args := svgFloats(match[2])
		var t svgMatrix
		switch strings.ToLower(match[1]) {
		case "matrix":
			if len(args) != 6 {
				continue
			}
			t = svgMatrix{args[0], args[1], args[2], args[3], args[4], args[5]}
		case "translate":
			if len(args) == 0 {
				continue
			}
			t = svgTranslate(args[0], 0)
			if len(args) > 1 {
				t = svgTranslate(args[0], args[1])
			}
		case "scale":
			if len(args) == 0 {
				continue
			}
			t = svgScale(args[0], args[0])
			if len(args) > 1 {
				t = svgScale(args[0], args[1])
			}
		default:
			continue
		}
		m = m.times(t)
	}
	return m
}

// svgFloats parses a list of space- or comma-separated floats.
func svgFloats(s string) []float64 {
	var floats []float64
	for _, v := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' }) {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}
		floats = append(floats, f)
	}
	return floats
}

// The fonts used to render SVG text, parsed lazily. SVG documents, e.g.
// badges, are usually designed for common sans-serif fonts, and the Go fonts
// are metrically close to them.
var (
	svgFontsOnce sync.Once
	svgRegular   *opentype.Font
	svgBold      *opentype.Font
	svgFontsErr  error
)

func svgFonts() (*opentype.Font, *opentype.Font, error) {
	svgFontsOnce.Do(func() {
		svgRegular, svgFontsErr = opentype.Parse(goregular.TTF)
		if svgFontsErr != nil {
			return
		}
		svgBold, svgFontsErr = opentype.Parse(gobold.TTF)
	})
	return svgRegular, svgBold, svgFontsErr
}

// svgFace returns a font face of the given size in pixels.
func svgFace(size float64, bold bool) (font.Face, error) {
	regular, boldFont, err := svgFonts()
	if err != nil {
		return nil, fmt.Errorf("glamour: error parsing the builtin SVG font: %w", err)
	}
	f := regular
	if bold {
		f = boldFont
	}
	return opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

// svgMeasure returns the width of the text at the given size, in pixels.
func svgMeasure(text string, size float64, bold bool) float64 {
	face, err := svgFace(size, bold)
	if err != nil {
		return size * 0.6 * float64(len(text))
	}
	d := font.Drawer{Face: face}
	return float64(d.MeasureString(text).Ceil())
}

// drawSVGTexts draws the SVG's text runs over the rasterized image. originX
// and originY are the viewBox's origin and sx and sy the scale from viewBox
// units to pixels.
func drawSVGTexts(img *image.RGBA, originX, originY, sx, sy float64, runs []svgTextRun) {
	for _, r := range runs {
		face, err := svgFace(r.size*sy, r.bold)
		if err != nil {
			continue
		}
		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(r.fill),
			Face: face,
		}
		x := (r.x - originX) * sx
		y := (r.y - originY) * sy
		w := float64(d.MeasureString(r.text).Ceil())
		switch r.anchor {
		case "middle":
			x -= w / 2
		case "end":
			x -= w
		}
		d.Dot = fixed.P(int(math.Round(x)), int(math.Round(y)))
		d.DrawString(r.text)
	}
}

// svgCommandArgs returns the number of arguments an SVG path command takes.
// Commands may be followed by more than one set of arguments.
var svgCommandArgs = map[byte]int{
	'M': 2, 'm': 2, 'L': 2, 'l': 2, 'H': 1, 'h': 1, 'V': 1, 'v': 1,
	'C': 6, 'c': 6, 'S': 4, 's': 4, 'Q': 4, 'q': 4, 'T': 2, 't': 2,
	'A': 7, 'a': 7, 'Z': 0, 'z': 0,
}

// svgPathDataPattern matches the path data attributes of an SVG document,
// i.e. d="..." and d='...', including the whitespace before them.
var svgPathDataPattern = regexp.MustCompile(`(?is)(\s)d(\s*=\s*)("([^"]*)"|'([^']*)')`)

// normalizeSVGPaths rewrites the path data of an SVG document, see
// [normalizeSVGPath].
func normalizeSVGPaths(data []byte) []byte {
	return svgPathDataPattern.ReplaceAllFunc(data, func(m []byte) []byte {
		sub := svgPathDataPattern.FindSubmatch(m)
		if sub == nil {
			return m
		}
		value, quote := sub[4], `"`
		if len(sub[5]) > 0 {
			value, quote = sub[5], `'`
		}
		var b []byte
		b = append(b, sub[1]...)
		b = append(b, 'd')
		b = append(b, sub[2]...)
		b = append(b, quote...)
		b = append(b, normalizeSVGPath(string(value))...)
		b = append(b, quote...)
		return b
	})
}

// normalizeSVGPath rewrites SVG path data into a form that parsers read
// reliably: every command letter is followed by exactly one set of
// arguments, separated by commas. SVG allows arguments to be packed tightly
// together, e.g. the flags of an arc can be glued to the numbers after them
// (a common badge idiom is "a.977.977 0 00-.275-.082", where "00" are the
// two flags), and parsers commonly read those as a single number, which
// makes the whole path fail to parse.
func normalizeSVGPath(d string) string {
	var b strings.Builder
	for _, tok := range svgPathTokens(d) {
		arity := svgCommandArgs[tok.cmd]
		if arity == 0 {
			b.WriteByte(tok.cmd)
			continue
		}
		// A moveto's additional argument pairs are line-to commands.
		cmd := tok.cmd
		for i := 0; i+arity <= len(tok.args); i += arity {
			b.WriteByte(cmd)
			for j := range arity {
				if j > 0 {
					b.WriteByte(',')
				}
				b.WriteString(strconv.FormatFloat(tok.args[i+j], 'f', -1, 64))
			}
			switch cmd {
			case 'm':
				cmd = 'l'
			case 'M':
				cmd = 'L'
			}
		}
	}
	return b.String()
}

// svgPathToken is a command of an SVG path with its arguments.
type svgPathToken struct {
	cmd  byte
	args []float64
}

// svgPathTokens splits SVG path data into its commands and arguments.
func svgPathTokens(d string) []svgPathToken {
	var (
		tokens []svgPathToken
		cur    *svgPathToken
		num    strings.Builder
	)

	flush := func() {
		s := num.String()
		num.Reset()
		if s == "" || cur == nil {
			return
		}
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			cur.args = append(cur.args, v)
		}
	}

	for i := 0; i < len(d); i++ {
		c := d[i]
		switch {
		case c == ' ' || c == ',' || c == '\t' || c == '\n' || c == '\r':
			flush()
		case svgCommandArgs[c] != 0 || c == 'z' || c == 'Z':
			flush()
			tokens = append(tokens, svgPathToken{cmd: c})
			cur = &tokens[len(tokens)-1]
		case c == '-' || c == '+':
			// A sign starts a new number, unless it is an exponent's sign.
			if s := num.String(); s == "" || s[len(s)-1] != 'e' && s[len(s)-1] != 'E' {
				flush()
			}
			num.WriteByte(c)
		case c == '.':
			// A second dot starts a new number, e.g. the implicit "1.5.5".
			if strings.Contains(num.String(), ".") {
				flush()
			}
			num.WriteByte(c)
		case c >= '0' && c <= '9', c == 'e', c == 'E':
			num.WriteByte(c)
			// The flags of an arc command are single digits, and can be
			// glued to the number after them, e.g. "00-1".
			if cur != nil && (cur.cmd == 'a' || cur.cmd == 'A') &&
				(len(cur.args)%7 == 3 || len(cur.args)%7 == 4) {
				flush()
			}
		default:
			flush()
		}
	}
	flush()
	return tokens
}
