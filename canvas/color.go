package canvas

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/REANNZ/raumata/internal/f32"
)

const (
	componentPrec = 3
)

// Represents an abstract color.
type Color interface {
	Space() ColorSpace
	ToRGB() *RGBColor
	ToHSL() *HSLColor
}

// Compare two colors for equality. This will convert the colors
// to the same space if necessary to do the comparison
//
//	a = RGB(0, 0, 0)
//	b = HSL(0, 0, 0)
//	ColorEqual(a, b) // true
func ColorEqual(a, b Color) bool {
	if a == b {
		return true
	}
	// If one of the colors is nil, then they
	// can't be equal
	if a == nil || b == nil {
		return false
	}

	// Both a and b are non-nil now

	// Do the comparison in the original space if
	// the match
	if a.Space() == b.Space() {
		if a.Space() == ColorSpaceHSL {
			return a.ToHSL().Equal(b.ToHSL())
		}
	}

	// Convert both to RGB and compare
	return *a.ToRGB() == *b.ToRGB()
}

type ColorSpace int

const (
	ColorSpaceRGB ColorSpace = iota
	ColorSpaceHSL
)

func (sp ColorSpace) String() string {
	switch sp {
	case ColorSpaceHSL:
		return "hsl"
	default:
		return "rgb"
	}
}

func (sp *ColorSpace) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "hsl" {
		*sp = ColorSpaceHSL
	} else {
		*sp = ColorSpaceRGB
	}

	return nil
}

func (sp ColorSpace) MarshalJSON() ([]byte, error) {
	return json.Marshal(sp.String())
}

// Represents a color in RGB space using three
// values from the interval [0, 1]
type RGBColor struct {
	R, G, B float32
}

// Constructs an RGBColor value, with the given values.
// values are expected to be between 0 and 1. Values
// outside that range are clamped to within 0 and 1
func RGB(r, g, b float32) *RGBColor {
	r = f32.Max(0, f32.Min(r, 1))
	g = f32.Max(0, f32.Min(g, 1))
	b = f32.Max(0, f32.Min(b, 1))

	r = roundTo(r, componentPrec)
	g = roundTo(g, componentPrec)
	b = roundTo(b, componentPrec)

	return &RGBColor{
		R: r,
		G: g,
		B: b,
	}
}

// Constructs an RGBColor from 3 integer component
// values. This is equivalent to calling [RGB] as:
//
//     RGB(r/255, g/255, b/255)
func RGBInt(r, g, b int) *RGBColor {
	rf := float32(r) / 255
	gf := float32(g) / 255
	bf := float32(b) / 255

	return RGB(rf, gf, bf)
}

type ColorParseError struct {
	Input string
	Err   error
}

func (e *ColorParseError) Error() string {
	return fmt.Sprintf("Error parsing '%s': %s", e.Input, e.Err)
}

func (e *ColorParseError) Unwrap() error {
	return e.Err
}

// Parse the given string into a [Color].
//
// Currently only hex-strings starting with '#' are supported
func ParseColor(s string) (Color, error) {
	if s[0] == '#' {
		return ParseHexColor(s)
	}

	return nil, &ColorParseError{
		Input: s,
		Err:   errors.New("Invalid color format"),
	}
}

// Parse a hex-encoded string, with an optional leading '#', into an RGBColor.
//
// The string must use two bytes per value
func ParseHexColor(s string) (*RGBColor, error) {

	input := s
	makeError := func(e error) error {
		err := &ColorParseError{
			Input: input,
			Err:   e,
		}

		if numErr, ok := e.(*strconv.NumError); ok {
			err.Err = fmt.Errorf("'%s' %w", numErr.Num, numErr.Err)
		}

		return err
	}

	if s[0] == '#' {
		s = s[1:]
	}

	if len(s) < 6 {
		return nil, makeError(errors.New("too short"))
	}

	var redPart, greenPart, bluePart string

	redPart = s[0:2]
	greenPart = s[2:4]
	bluePart = s[4:6]

	red, err := strconv.ParseInt(redPart, 16, 16)
	if err != nil {
		return nil, makeError(err)
	}
	green, err := strconv.ParseInt(greenPart, 16, 16)
	if err != nil {
		return nil, makeError(err)
	}
	blue, err := strconv.ParseInt(bluePart, 16, 16)
	if err != nil {
		return nil, makeError(err)
	}

	return RGBInt(int(red), int(green), int(blue)), nil
}

func (rgb *RGBColor) Space() ColorSpace { return ColorSpaceRGB }

// Implement the [Color] interface, returns the receiver
func (rgb *RGBColor) ToRGB() *RGBColor {
	return rgb
}

// Implement the [Color] interface
//
// Returns the color in the HSL color space
func (rgb *RGBColor) ToHSL() *HSLColor {
	cMax := f32.Max(rgb.R, rgb.G, rgb.B)
	cMin := f32.Min(rgb.R, rgb.G, rgb.B)

	delta := cMax - cMin

	var h, s, l float32

	if delta == 0 {
		h = 0
	} else if cMax == rgb.R {
		h = 60 * floatMod((rgb.G-rgb.B)/delta, 6)
	} else if cMax == rgb.G {
		h = 60 * (((rgb.B - rgb.R) / delta) + 2)
	} else {
		h = 60 * (((rgb.R - rgb.G) / delta) + 4)
	}

	l = (cMax + cMin) / 2

	if delta == 0 {
		s = 0
	} else {
		s = delta / (1 - f32.Abs(2*l-1))
	}

	return HSL(h, s, l)
}

// Returns the color as an hex-encoded string with a leading '#'
func (rgb *RGBColor) ToHex() string {
	r := int(f32.Round(rgb.R * 255))
	g := int(f32.Round(rgb.G * 255))
	b := int(f32.Round(rgb.B * 255))

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Implement [encoding/TextUnmarshaler].
// Marshals using [RGBColor.ToHex]
func (rgb *RGBColor) UnmarshalText(text []byte) error {
	c, err := ParseHexColor(string(text))
	if err != nil {
		return err
	}

	*rgb = *c

	return nil
}

// Implement [encoding/TextMarshaler]
func (rgb *RGBColor) MarshalText() ([]byte, error) {
	return []byte(rgb.ToHex()), nil
}

// Returns the result of doing a component-wise interpolation between
// x and y, using the interpolation variable t.
// t is expected to be between 0 and 1, values outside that range are
// clamped
func (x *RGBColor) Interpolate(y *RGBColor, t float32) *RGBColor {
	if t <= 0 {
		return x
	} else if t >= 1 {
		return y
	}

	r := x.R*(1-t) + y.R*t
	g := x.G*(1-t) + y.G*t
	b := x.B*(1-t) + y.B*t

	return RGB(r, g, b)
}

func (rgb *RGBColor) String() string {
	return fmt.Sprintf("rgb(%.3g, %.3g, %.3g)",
		rgb.R, rgb.G, rgb.B)
}

// Represents a color in the HSL color space.
//
// The HSL color space represents colors using
// hue, saturation and lightness values
type HSLColor struct {
	H float32 // Hue as an angle, valid range is [0, 360)
	S float32 // Saturation, valid range is [0, 1]
	L float32 // Lightness, valid range is [0, 1]
}

// Constructs a color in the HSL color space.
//
// Hue values outside [0, 360) are adjusted to fall in the range
// Saturation and lightness values outside of [0, 1] are clamped to that range
func HSL(h, s, l float32) *HSLColor {
	// Adjust h so it falls in [0, 360)
	for h < 0 {
		h += 360
	}
	for h >= 360 {
		h -= 360
	}
	// Just clamp s and l
	s = f32.Max(0, f32.Min(s, 1))
	l = f32.Max(0, f32.Min(l, 1))

	h = roundTo(h, 1)
	s = roundTo(s, componentPrec)
	l = roundTo(l, componentPrec)

	return &HSLColor{
		H: h,
		S: s,
		L: l,
	}
}

func (hsl *HSLColor) Space() ColorSpace { return ColorSpaceHSL }

// Implements the [Color] interface
//
// Returns the equivalent color in RGB color space
func (hsl *HSLColor) ToRGB() *RGBColor {
	c := (1 - f32.Abs(2*hsl.L-1)) * hsl.S
	x := c * (1 - f32.Abs(floatMod(hsl.H/60, 2)-1))
	m := hsl.L - (c / 2)

	var r, g, b float32
	if hsl.H < 60 {
		r, g, b = c, x, 0
	} else if hsl.H < 120 {
		r, g, b = x, c, 0
	} else if hsl.H < 180 {
		r, g, b = 0, c, x
	} else if hsl.H < 240 {
		r, g, b = 0, x, c
	} else if hsl.H < 300 {
		r, g, b = x, 0, c
	} else {
		r, g, b = c, 0, x
	}

	r += m
	g += m
	b += m

	return RGB(r, g, b)
}

// Implements the [Color] interface
//
// Returns the reciever
func (hsl *HSLColor) ToHSL() *HSLColor {
	return hsl
}

// Returns whether two points in HSL color space represent
// the same color.
func (a *HSLColor) Equal(b *HSLColor) bool {
	// This needs to be a little more complicated because
	// multiple points in HSL space map to the same color

	// If all the components are equal, then the colors are equal
	if *a == *b {
		return true
	}

	// If the saturation is 0, then only the lightness component matters
	if a.S == 0 && b.S == 0 {
		return a.L == b.L
	}

	// If the lightness is 1 or 0, then neither hue nor saturation matter
	if a.L == b.L && (a.L == 1 || a.L == 0) {
		return true
	}

	// Otherwise, all components need to be equal for the colors to be
	// equal, which we already checked
	return false
}

// Returns the result of doing a component-wise interpolation between
// x and y, using the interpolation variable t.
// t is expected to be between 0 and 1, values outside that range are
// clamped.
// As the hue represents an angle, there are two lines between any two
// values with different hues. This function will interpolate along the
// shorter of the two lines. If the hues are 180 degrees apart, the
// interpolation will avoid crossing zero.
func (a *HSLColor) Interpolate(b *HSLColor, t float32) *HSLColor {
	if t <= 0 {
		return a
	} else if t >= 1 {
		return b
	}

	var h, s, l float32

	ha := a.H
	hb := b.H

	delta := f32.Abs(ha - hb)
	if delta <= 180 {
		h = ha*(1-t) + hb*t
	} else {
		ha = floatMod(ha+delta, 360)
		hb = floatMod(hb+delta, 360)

		h = ha*(1-t) + hb*t

		h -= delta
	}

	s = a.S*(1-t) + b.S*t
	l = a.L*(1-t) + b.L*t

	return HSL(h, s, l)
}

func (hsl *HSLColor) String() string {
	return fmt.Sprintf("hsl(%.3g, %.3g, %.3g)",
		hsl.H, hsl.S, hsl.L)
}

type colorPoint struct {
	val   float32
	color Color
}

type ColorScale struct {
	Space  ColorSpace
	points []colorPoint
}

func NewColorScale() *ColorScale {
	return &ColorScale{}
}

func ColorScaleFromMap(m map[float32]Color) *ColorScale {
	scale := &ColorScale{}

	for val, color := range m {
		scale.points = append(scale.points, colorPoint{
			val:   val,
			color: color,
		})
	}

	scale.sort()

	return scale
}

func HeatColorScale() *ColorScale {
	colors := map[float32]Color{
		0.0: RGB(0.114, 0.282, 0.467),
		0.1: RGB(0.106, 0.541, 0.353),
		0.5: RGB(0.984, 0.690, 0.123),
		0.7: RGB(0.965, 0.533, 0.220),
		0.9: RGB(0.933, 0.243, 0.196),
	}

	scale := ColorScaleFromMap(colors)
	scale.Space = ColorSpaceHSL

	return scale
}

func (s *ColorScale) AddColor(val float32, color Color) {
	s.points = append(s.points, colorPoint{val: val, color: color})
	s.sort()
}

func (s *ColorScale) getColor(val float32) (i, j int, t float32) {
	for i := 0; i < len(s.points)-1; i++ {
		p1 := s.points[i]
		p2 := s.points[i+1]

		if p1.val >= val {
			return 0, 0, 0
		}

		if p2.val >= val {
			delta := p2.val - p1.val
			t := (val - p1.val) / delta

			return i, i + 1, t
		}
	}

	i = len(s.points) - 1
	j = i
	t = 1
	return
}

func (s *ColorScale) GetColor(val float32) Color {
	if s == nil {
		return nil
	}
	if len(s.points) == 0 {
		return nil
	}
	if len(s.points) == 1 {
		return s.points[0].color
	}

	i, j, t := s.getColor(val)
	p1 := s.points[i]
	p2 := s.points[j]
	switch s.Space {
	case ColorSpaceHSL:
		return p1.color.ToHSL().Interpolate(p2.color.ToHSL(), t)
	default:
		return p1.color.ToRGB().Interpolate(p2.color.ToRGB(), t)
	}
}

func (s *ColorScale) sort() {
	slices.SortStableFunc(s.points, func(a, b colorPoint) int {
		if a.val < b.val {
			return -1
		} else if a.val > b.val {
			return 1
		} else {
			return 0
		}
	})
}

func (s *ColorScale) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	unmarshalSlice := func(slice [][2]json.RawMessage) ([]colorPoint, error) {
		newPoints := make([]colorPoint, len(slice))
		for i := range slice {
			var val float32
			var colorStr string
			rawVal := slice[i][0]
			rawColor := slice[i][1]
			if err := json.Unmarshal(rawVal, &val); err != nil {
				return nil, err
			}
			if err := json.Unmarshal(rawColor, &colorStr); err != nil {
				return nil, err
			}

			color, err := ParseColor(colorStr)
			if err != nil {
				return nil, err
			}

			newPoints[i] = colorPoint{
				val:   val,
				color: color,
			}
		}

		return newPoints, nil
	}

	if data[0] == '[' {
		var array [][2]json.RawMessage
		if err := json.Unmarshal(data, &array); err != nil {
			return err
		}

		newPoints, err := unmarshalSlice(array)
		if err != nil {
			return err
		}

		s.points = newPoints

		s.sort()

		return nil
	} else if data[0] == '{' {
		var object struct {
			Space ColorSpace `json:"space"`
			Colors [][2]json.RawMessage `json:"colors"`
		}

		object.Space = s.Space
		if err := json.Unmarshal(data, &object); err != nil {
			return err
		}

		newPoints, err := unmarshalSlice(object.Colors)
		if err != nil {
			return err
		}

		s.Space = object.Space
		s.points = newPoints

		s.sort()

		return nil
	} else {
		return errors.New("invalid color scale format, must be an array or object")
	}
}

func (s *ColorScale) MarshalJSON() ([]byte, error) {
	if len(s.points) == 0 {
		return []byte("null"), nil
	}
	array := make([][2]json.RawMessage, len(s.points))

	for i := range s.points {
		p := s.points[i]
		valBytes := []byte(strconv.FormatFloat(float64(p.val), 'g', 2, 32))
		colorBytes, err := json.Marshal(p.color)
		if err != nil {
			return nil, err
		}

		array[i][0] = valBytes
		array[i][1] = colorBytes
	}

	var object struct {
		Space ColorSpace `json:"space"`
		Colors [][2]json.RawMessage `json:"colors"`
	}
	object.Space = s.Space
	object.Colors = array

	return json.Marshal(&object)
}

func floatMod(a, b float32) float32 {
	var sign float32
	if a < 0 {
		sign = -1
		a = -a
	} else {
		sign = 1
	}
	for a > b {
		a -= b
	}

	if sign > 0 {
		return a
	} else {
		return b - a
	}
}

func roundTo(x float32, prec int) float32 {
	var factor float32 = 1
	for i := 0; i < prec; i++ {
		factor *= 10
	}

	return f32.Round(x*factor) / factor
}
