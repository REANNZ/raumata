package canvas

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/REANNZ/raumata/internal"
	"github.com/REANNZ/raumata/internal/f32"
	"github.com/REANNZ/raumata/vec"
)

// Controls the way styles are rendered
type SVGStyleMode int

const (
	// Don't use stylesheets at all
	SVGStyleNone SVGStyleMode = iota
	// Render an embedded stylesheet into the SVG document
	SVGStyleInternal
	// Include element styles, but don't embed a stylesheet
	// into the document, relying instead on an external
	// stylesheet
	SVGStyleExternal
)

// Renders a canvas to a SVG format
//
// The size of the image is determined by the width and height
// of the canvas and the Width and Height fields.
type SVGRenderer struct {
	Indent         int            // Controls the size of the indent
	Width          int            // The width of the image, <= 0 means automatic
	Height         int            // The height of the image, <= 0 means automatic
	IncludeHeader  bool           // Include an XML header, set to false if embedding the file in another document
	StyleMode      SVGStyleMode   // Mode to use for rendering styles, defaults to SVGStyleNone
	Precision      int            // Controls the precision used for printing floats
	RootAttributes map[string]any // Attributes for the root svg element
	f              io.Writer
	level          int
	currentStyle   *Style
	canvas         *Canvas
}

// NewSVGRenderer returns a new renderer that writes an SVG to f
func NewSVGRenderer(f io.Writer) *SVGRenderer {
	return &SVGRenderer{
		f:            f,
		level:        0,
		currentStyle: NewStyle(),

		IncludeHeader:  true,
		Precision:      2,
		RootAttributes: make(map[string]any),
	}
}

func (r *SVGRenderer) RenderCanvas(canvas *Canvas) error {
	if r.IncludeHeader {
		_, err := io.WriteString(r.f, `<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">`)
		if err != nil {
			return err
		}
	}

	r.canvas = canvas

	attrs := r.convertAttributeMap(r.RootAttributes)

	attrs["xmlns"] = "http://www.w3.org/2000/svg"

	aabb := canvas.GetAABB()

	min, max := aabb.Bounds()

	// Calculate the viewbox, which is the coordinate system
	// used by the elements in the the image
	min = min.Sub(canvas.Margin)
	max = max.Add(canvas.Margin)
	size := max.Sub(min)

	viewBox := fmt.Sprintf("%s %s %s %s",
		r.formatFloat32(min.X),
		r.formatFloat32(min.Y),
		r.formatFloat32(size.X),
		r.formatFloat32(size.Y))

	// Calculate the image's width and height
	var width, height int
	if r.Width <= 0 && r.Height > 0 {
		h := float32(r.Height)
		w := (h / size.Y) * size.X
		height = r.Height
		width = int(f32.Round(w))
	}
	if r.Height <= 0 && r.Width > 0 {
		w := float32(r.Width)
		h := (w / size.X) * size.Y
		width = r.Width
		height = int(f32.Round(h))
	}

	if r.Width > 0 && r.Height > 0 {
		width = r.Width
		height = r.Height
	}

	if width == 0 && height == 0 {
		width = int(f32.Round(size.X))
		height = int(f32.Round(size.Y))
	}

	attrs["width"] = fmt.Sprintf("%dpx", width)
	attrs["height"] = fmt.Sprintf("%dpx", height)
	attrs["viewBox"] = viewBox

	// Start rendering
	if r.StyleMode != SVGStyleInternal || !canvas.Stylesheet.HasRules() {
		return r.writeElement("svg", attrs, canvas.Children, nil)
	} else {
		err := r.writeOpenElement("svg", attrs, false)
		if err != nil {
			return err
		}

		r.level += 1
		err = r.writeStylesheet(canvas.Stylesheet)
		if err != nil {
			return err
		}

		RenderChildren(r, canvas.Children)

		r.level -= 1
		_, err = fmt.Fprintf(r.f, "</svg>")
		return err
	}
}

// RenderGroup renders a [Group] object to a `<g>` element
func (r *SVGRenderer) RenderGroup(group *Group) error {

	attrs := r.convertAttributes(&group.Attributes)

	// Try to handle the transform nicely, if there is one.
	// While the matrix form will always work, using the translate/rotate
	// forms makes the markup more understandable
	if group.Transform != nil && !group.Transform.IsIdentity() {
		t := group.Transform

		transformStr := ""

		trans, ok := t.GetTranslation()
		if ok {
			xStr := r.formatFloat32(trans.X)
			yStr := r.formatFloat32(trans.Y)
			transformStr = fmt.Sprintf("translate(%s, %s)", xStr, yStr)
		}
		rot, ok := t.GetRotation()
		if ok {
			transformStr = fmt.Sprintf("rotate(%s)", r.formatFloat32(rot))
		}

		if transformStr == "" {
			// Fallback to the matrix form
			transformStr = fmt.Sprintf("matrix(%s,%s,%s,%s,%s,%s)",
				r.formatFloat32(t.A),
				r.formatFloat32(t.B),
				r.formatFloat32(t.C),
				r.formatFloat32(t.D),
				r.formatFloat32(t.E),
				r.formatFloat32(t.F))
		}

		attrs["transform"] = transformStr
	}

	return r.writeElement("g", attrs, group.Children, group.Attributes.Style)
}

// RenderRect renders a [Rect] object to a `<rect>` element
func (r *SVGRenderer) RenderRect(rect *Rect) error {

	attrs := r.convertAttributes(&rect.Attributes)

	attrs["x"] = r.formatFloat32(rect.Pos.X)
	attrs["y"] = r.formatFloat32(rect.Pos.Y)
	attrs["width"] = r.formatFloat32(rect.Width)
	attrs["height"] = r.formatFloat32(rect.Height)
	if rect.Rx > 0 {
		attrs["rx"] = r.formatFloat32(rect.Rx)
	}
	if rect.Ry > 0 {
		attrs["ry"] = r.formatFloat32(rect.Ry)
	}
	return r.writeElement("rect", attrs, rect.Children, rect.Attributes.Style)
}

// RenderEllipse renders an [Ellipse] object to either an
// `<ellipse>` elements or a `<circle>` element
func (r *SVGRenderer) RenderEllipse(ellipse *Ellipse) error {

	attrs := r.convertAttributes(&ellipse.Attributes)

	name := "ellipse"
	attrs["cx"] = r.formatFloat32(ellipse.Center.X)
	attrs["cy"] = r.formatFloat32(ellipse.Center.Y)

	if ellipse.Rx == ellipse.Ry {
		name = "circle"
		attrs["r"] = r.formatFloat32(ellipse.Rx)
	} else {
		attrs["rx"] = r.formatFloat32(ellipse.Rx)
		attrs["ry"] = r.formatFloat32(ellipse.Ry)
	}
	return r.writeElement(name, attrs, ellipse.Children, ellipse.Attributes.Style)
}

// RenderLine renders a [Line] object to a `<line>` element
func (r *SVGRenderer) RenderLine(line *Line) error {

	attrs := r.convertAttributes(&line.Attributes)

	attrs["x1"] = r.formatFloat32(line.Start.X)
	attrs["y1"] = r.formatFloat32(line.Start.Y)
	attrs["x2"] = r.formatFloat32(line.End.X)
	attrs["y2"] = r.formatFloat32(line.End.Y)

	return r.writeElement("line", attrs, line.Children, line.Attributes.Style)
}

// RenderPolygon renders a [Polygon] object to a `<polygon>` element
func (r *SVGRenderer) RenderPolygon(polygon *Polygon) error {

	attrs := r.convertAttributes(&polygon.Attributes)

	points := ""
	for _, p := range polygon.Points {
		xStr := r.formatFloat32(p.X)
		yStr := r.formatFloat32(p.Y)
		points += fmt.Sprintf("%s, %s ", xStr, yStr)
	}

	attrs["points"] = points

	return r.writeElement("polygon", attrs, polygon.Children, polygon.Attributes.Style)
}

// RenderPath renders a [Path] object to a `<path>` object
func (r *SVGRenderer) RenderPath(path *Path) error {

	eps := f32.Pow(10, -(float32(r.Precision + 1)))

	attrs := r.convertAttributes(&path.Attributes)

	data := ""

	prevPos := vec.Vec2{}
	prevCmdCode := ""
	for _, cmd := range path.Data {
		switch cmd.Type {
		case CommandClosePath:
			data += "Z"
			prevCmdCode = "Z"
		case CommandMoveTo:
			data += fmt.Sprintf("M%s,%s ", r.formatFloat32(cmd.Args[0]), r.formatFloat32(cmd.Args[1]))
			prevCmdCode = "M"
		case CommandLineTo:
			if prevPos.ApproxEq(cmd.Pos, eps) {
				continue
			}
			if prevPos.X == cmd.Pos.X {
				data += fmt.Sprintf("V%s ", r.formatFloat32(cmd.Args[1]))
				prevCmdCode = "V"
			} else if prevPos.Y == cmd.Pos.Y {
				data += fmt.Sprintf("H%s ", r.formatFloat32(cmd.Args[0]))
				prevCmdCode = "H"
			} else {
				if prevCmdCode != "L" && prevCmdCode != "M" {
					data += "L"
					prevCmdCode = "L"
				}
				data += fmt.Sprintf("%s,%s ", r.formatFloat32(cmd.Args[0]), r.formatFloat32(cmd.Args[1]))
			}
		case CommandArcTo:
			start := vec.Vec2{X: cmd.Args[0], Y: cmd.Args[1]}
			end := vec.Vec2{X: cmd.Args[2], Y: cmd.Args[3]}
			radius := cmd.Args[4]
			sweep := int(cmd.Args[5])

			dir := end.Sub(start)
			dist := dir.Length()

			if radius < (dist / 2) {
				radius = (dist / 2)
			}

			radStr := r.formatFloat32(radius)
			data += fmt.Sprintf("A%s,%s 0 0,%d %s,%s ",
				radStr, radStr, sweep, r.formatFloat32(end.X), r.formatFloat32(end.Y))
			prevCmdCode = "A"
		}
		prevPos = cmd.Pos
	}

	attrs["d"] = data

	return r.writeElement("path", attrs, path.Children, path.Attributes.Style)

}

// RenderText renders a [Text] object to a `<text>` element
func (r *SVGRenderer) RenderText(text *Text) error {
	attrs := r.convertAttributes(&text.Attributes)

	attrs["x"] = r.formatFloat32(text.Pos.X)
	attrs["y"] = r.formatFloat32(text.Pos.Y)
	if text.Size > 0 {
		attrs["font-size"] = r.formatFloat32(text.Size)
	}

	anchor := text.Anchor.String()
	if anchor != "" {
		attrs["text-anchor"] = anchor
	}

	if err := r.writeOpenElement("text", attrs, false); err != nil {
		return err
	}

	if _, err := io.WriteString(r.f, text.Text); err != nil {
		return err
	}

	_, err := io.WriteString(r.f, "</text>")
	return err

}

func (r *SVGRenderer) writeStylesheet(stylesheet Stylesheet) error {
	if err := r.writeOpenElement("defs", nil, false); err != nil {
		return err
	}

	r.level += 1
	if err := r.writeOpenElement("style", map[string]string{"type": "text/css"}, false); err != nil {
		return err
	}

	if _, err := io.WriteString(r.f, "<![CDATA[\n"); err != nil {
		return err
	}

	ssRules := stylesheet.GetAllRules()
	rules := make([]Rule, len(ssRules))

	copy(rules, ssRules)

	slices.Reverse(rules)

	for _, rule := range rules {
		selector := strings.Join(rule.Selector, ".")
		if _, err := fmt.Fprintf(r.f, ".%s {\n", selector); err != nil {
			return err
		}

		if _, err := io.WriteString(r.f, rule.Style.toCSS(r.Indent)); err != nil {
			return err
		}

		if _, err := io.WriteString(r.f, "}\n"); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(r.f, "]]>"); err != nil {
		return err
	}

	if _, err := io.WriteString(r.f, "</style>"); err != nil {
		return err
	}

	r.level -= 1
	if err := r.newline(); err != nil {
		return err
	}
	_, err := io.WriteString(r.f, "</defs>")
	return err
}

func (r *SVGRenderer) writeOpenElement(name string, attrs map[string]string, selfClose bool) error {
	if err := r.newline(); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(r.f, "<%s", name); err != nil {
		return err
	}

	// Sort the attributes by key to make the output consistent and
	// more diff-friendly
	type attrPair struct {
		key string
		val string
	}

	var attrPairs []attrPair

	for key, val := range attrs {
		attrPairs = append(attrPairs, attrPair{
			key: key, val: val,
		})
	}

	slices.SortFunc(attrPairs, func(a, b attrPair) int {
		if a.key < b.key {
			return -1
		} else if a.key == b.key {
			return 0
		} else {
			return 1
		}
	})

	for _, pair := range attrPairs {
		if _, err := fmt.Fprintf(r.f, " %s=\"%s\"", pair.key, pair.val); err != nil {
			return err
		}
	}

	var err error
	if selfClose {
		_, err = io.WriteString(r.f, "/>")
	} else {
		_, err = io.WriteString(r.f, ">")
	}
	return err
}

func (r *SVGRenderer) writeElement(name string, attrs map[string]string, children []Object, style *Style) error {
	if err := r.writeOpenElement(name, attrs, len(children) == 0); err != nil {
		return err
	}
	if len(children) > 0 {
		prevStyle := *r.currentStyle
		if style != nil {
			*r.currentStyle = *style
			r.currentStyle.Merge(&prevStyle)
		}

		r.level += 1
		if err := RenderChildren(r, children); err != nil {
			return err
		}
		r.level -= 1

		*r.currentStyle = prevStyle

		if err := r.newline(); err != nil {
			return err
		}
		_, err := fmt.Fprintf(r.f, "</%s>", name)
		return err
	}

	return nil
}

func (r *SVGRenderer) newline() error {
	if r.Indent == 0 {
		return nil
	}

	buf := [64]byte{}
	buf[0] = '\n'

	numSpaces := r.Indent * r.level
	if numSpaces >= len(buf) {
		numSpaces = len(buf) - 1
	}
	for i := 1; i <= numSpaces; i++ {
		buf[i] = ' '
	}

	_, err := r.f.Write(buf[:numSpaces+1])
	return err
}

func (r *SVGRenderer) formatFloat32(f float32) string {
	return internal.FormatFloat32(f, r.Precision)
}

func (r *SVGRenderer) convertAttributeMap(attrs map[string]any) map[string]string {
	out := map[string]string{}

	if attrs == nil {
		return out
	}

	for attr, v := range attrs {
		switch val := v.(type) {
		case nil:
			// Do nothing
		case int:
			out[attr] = strconv.FormatInt(int64(val), 10)
		case bool:
			if val {
				out[attr] = "true"
			} else {
				out[attr] = "false"
			}
		case float32:
			out[attr] = strconv.FormatFloat(float64(val), 'f', r.Precision, 64)
		case float64:
			out[attr] = strconv.FormatFloat(val, 'f', r.Precision, 64)
		case string:
			out[attr] = val
		case []string:
			list := ""
			for i, s := range val {
				if i != 0 {
					list += " "
				}
				list += s
			}
			out[attr] = list
		case Color:
			out[attr] = val.ToRGB().ToHex()
		case fmt.Stringer:
			out[attr] = val.String()
		}
	}

	return out
}

// Converts attributes into a map[string]string.
func (r *SVGRenderer) convertAttributes(attrs *Attributes) map[string]string {
	// Convert the `Extra` field first
	out := r.convertAttributeMap(attrs.Extra)

	if attrs.Id != "" {
		out["id"] = attrs.Id
	}

	// Handle converting the styles

	// Create a new blank style
	style := NewStyle()

	if attrs.Style != nil {
		// If there is an element style, use it
		style.Merge(attrs.Style)
	}

	if r.StyleMode == SVGStyleNone {
		// We aren't using stylesheets, so we need to include the
		// styles from classes
		classStyle := r.canvas.Stylesheet.GetStyle(attrs.Classes)
		style.Merge(classStyle)

		// Only emit attributes for changed style values
		style = r.currentStyle.Changed(style)

		// Lower styles to element attributes
		if style.Opacity.Valid {
			out["opacity"] = r.formatFloat32(style.Opacity.Value)
		}
		if style.FillColor != nil {
			color := style.FillColor.ToRGB().ToHex()
			out["fill"] = color
		}
		if style.StrokeOpacity.Valid {
			out["stroke-opacity"] = r.formatFloat32(style.StrokeOpacity.Value)
		}
		if style.StrokeColor != nil {
			color := style.StrokeColor.ToRGB().ToHex()
			out["stroke"] = color
		}
		if style.StrokeOpacity.Valid {
			out["stroke-opacity"] = r.formatFloat32(style.StrokeOpacity.Value)
		}
		if style.StrokeWidth.Valid {
			out["stroke-width"] = r.formatFloat32(style.StrokeWidth.Value)
		}
		if style.FontFamily != "" {
			out["font-family"] = style.FontFamily
		}
	} else {
		// Only emit style values that have changed
		style = r.currentStyle.Changed(style)
		css := style.toCSS(0)
		if css != "" {
			out["style"] = css
		}
	}

	if len(attrs.Classes) > 0 {
		out["class"] = strings.Join(attrs.Classes, " ")
	}

	return out
}

func (s *Style) toCSS(indent int) string {
	if s == nil {
		return ""
	}
	css := ""

	indentStr := make([]byte, indent)
	for i := 0; i < indent; i++ {
		indentStr[i] = ' '
	}

	appendStyle := func(style, value string) {
		if indent > 0 {
			css += string(indentStr)
		}

		css += fmt.Sprintf("%s: %s;", style, value)
		if indent > 0 {
			css += "\n"
		}
	}

	appendColor := func(style string, color Color) {
		if color == nil {
			return
		}
		if color.Space() == ColorSpaceHSL {
			appendStyle(style, color.ToHSL().String())
		} else {
			appendStyle(style, color.ToRGB().ToHex())
		}
	}

	if s.Opacity.Valid {
		appendStyle("opacity", s.Opacity.String())
	}

	appendColor("fill", s.FillColor)
	if s.FillOpacity.Valid {
		appendStyle("fill-opacity", s.FillOpacity.String())
	}

	appendColor("stroke", s.StrokeColor)

	if s.StrokeOpacity.Valid {
		appendStyle("stroke-opacity", s.StrokeOpacity.String())
	}
	if s.StrokeWidth.Valid {
		appendStyle("stroke-width", s.StrokeWidth.String())
	}
	if s.FontFamily != "" {
		appendStyle("font-family", s.FontFamily)
	}

	return css
}
