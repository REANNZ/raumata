package canvas_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/REANNZ/raumata/canvas"
	. "github.com/REANNZ/raumata/canvas"
)

func TestColorHSLToRGB(t *testing.T) {
	check := func(hsl *HSLColor, rgb *RGBColor) {
		t.Helper()
		conv := hsl.ToRGB()

		if *conv != *rgb {
			t.Errorf("Bad conversion of %s, expected %s, got %s",
				hsl, rgb, conv)
		}
	}

	check(HSL(0, 0.0, 0.0), RGB(0.0, 0.0, 0.0))
	check(HSL(0, 0.0, 1.0), RGB(1.0, 1.0, 1.0))
	check(HSL(0, 1.0, 0.5), RGB(1.0, 0.0, 0.0))
	check(HSL(120, 1.0, 0.5), RGB(0.0, 1.0, 0.0))
	check(HSL(240, 1.0, 0.5), RGB(0.0, 0.0, 1.0))
	check(HSL(60, 1.0, 0.5), RGB(1.0, 1.0, 0.0))
	check(HSL(180, 1.0, 0.5), RGB(0.0, 1.0, 1.0))
	check(HSL(300, 1.0, 0.5), RGB(1.0, 0.0, 1.0))
	check(HSL(210, 0.8, 0.4), RGB(0.080, 0.4, 0.72))
}

func TestColorRGBToHSL(t *testing.T) {
	check := func(rgb *RGBColor, hsl *HSLColor) {
		t.Helper()
		conv := rgb.ToHSL()

		if *conv != *hsl {
			t.Errorf("Bad conversion of %s, expected %s, got %s",
				rgb, hsl, conv)
		}
	}

	check(RGB(0.0, 0.0, 0.0), HSL(0, 0.0, 0.0))
	check(RGB(1.0, 1.0, 1.0), HSL(0, 0.0, 1.0))
	check(RGB(1.0, 0.0, 0.0), HSL(0, 1.0, 0.5))
	check(RGB(0.0, 1.0, 0.0), HSL(120, 1.0, 0.5))
	check(RGB(0.0, 0.0, 1.0), HSL(240, 1.0, 0.5))
	check(RGB(1.0, 1.0, 0.0), HSL(60, 1.0, 0.5))
	check(RGB(0.0, 1.0, 1.0), HSL(180, 1.0, 0.5))
	check(RGB(1.0, 0.0, 1.0), HSL(300, 1.0, 0.5))
	check(RGB(0.078, 0.4, 0.722), HSL(210, 0.805, 0.4))

}

func TestColorHSLInterpolate(t *testing.T) {
	check := func(expected, actual *HSLColor) {
		t.Helper()

		if *expected != *actual {
			t.Errorf("Expected %s, got %s", expected, actual)
		}
	}

	a := HSL(0, 1, 0)
	b := HSL(60, 0, 1)

	check(a.Interpolate(b, 0.0), a)
	check(a.Interpolate(b, 1.0), b)
	check(a.Interpolate(b, 0.5), HSL(30, 0.5, 0.5))

	a = HSL(300, 0, 0)
	b = HSL(60, 1, 1)

	check(a.Interpolate(b, 0.0), a)
	check(a.Interpolate(b, 1.0), b)
	check(a.Interpolate(b, 0.5), HSL(0, 0.5, 0.5))
}

func TestColorEqual(t *testing.T) {
	a := RGB(0, 0, 0)
	b := HSL(0, 0, 0)

	if !ColorEqual(a, a) {
		t.Errorf("ColorEqual returned false for %s == %s", a, a)
	}
	if !ColorEqual(b, b) {
		t.Errorf("ColorEqual returned false for %s == %s", b, b)
	}
	if !ColorEqual(a, b) {
		t.Errorf("ColorEqual returned false for %s == %s", a, b)
	}
	if !ColorEqual(b, a) {
		t.Errorf("ColorEqual returned false for %s == %s", b, a)
	}

	c := HSL(60, 0, 0)
	if !ColorEqual(b, c) {
		t.Errorf("ColorEqual returned false for %s == %s", b, c)
	}

	d := RGB(0, 0.5, 0.5)
	if ColorEqual(a, d) {
		t.Errorf("ColorEqual returned true for %s == %s", a, d)
	}
}

// Test the reflection-based unmarshalling code

func TestColorUnmarshal(t *testing.T) {
	type testObj struct {
		C      Color
		Cp     *Color
		Cpp    **Color
		Str    string
		Nested struct {
			NC Color
		}
		NestedP *struct {
			NC Color
		}
		Colors   []Color
		ColorMap map[string]Color
	}

	checkField := func(name string, actual, expected any) {
		t.Helper()
		if actual != expected {
			t.Errorf("Value for field '%s' is '%s', expected '%s'",
				name, actual, expected)
		}
	}

	// Check that the basic idea works
	jsonBlob := []byte(`{}`)

	var obj testObj
	err := UnmarshalColorStruct(jsonBlob, &obj)
	if err != nil {
		t.Errorf("Error parsing json: %s", err)
	}

	// Check unmarshalling with all fields set

	jsonBlob = []byte(`{
  "C": "#000000",
  "Cp": "#ffffff",
  "Cpp": "#808080",
  "Str": "abc",
  "Nested": {
    "NC": "#101010"
  },
  "NestedP": {
    "NC": "#202020"
  },
  "Colors": [
    "#000000",
    "#111111",
    "#222222"
  ],
  "ColorMap": {
    "red": "#ff0000",
    "green": "#00ff00",
    "blue": "#0000ff"
  }
}`)

	obj = testObj{}

	err = UnmarshalColorStruct(jsonBlob, &obj)
	if err != nil {
		t.Errorf("Error parsing json: %s", err)
	}

	checkField("C", obj.C.ToRGB().ToHex(), "#000000")
	checkField("Cp", (*obj.Cp).ToRGB().ToHex(), "#ffffff")
	checkField("Cpp", (**obj.Cpp).ToRGB().ToHex(), "#808080")
	checkField("Str", obj.Str, "abc")
	checkField("Nested.NC", obj.Nested.NC.ToRGB().ToHex(), "#101010")
	checkField("NestedP.NC", obj.NestedP.NC.ToRGB().ToHex(), "#202020")

	if len(obj.Colors) != 3 {
		t.Errorf("Length of field colors is %d, expected 3", len(obj.Colors))
	}
	checkField("Colors[0]", obj.Colors[0].ToRGB().ToHex(), "#000000")
	checkField("Colors[1]", obj.Colors[1].ToRGB().ToHex(), "#111111")
	checkField("Colors[2]", obj.Colors[2].ToRGB().ToHex(), "#222222")

	checkField("ColorMap[\"red\"]", obj.ColorMap["red"].ToRGB().ToHex(), "#ff0000")
	checkField("ColorMap[\"green\"]", obj.ColorMap["green"].ToRGB().ToHex(), "#00ff00")
	checkField("ColorMap[\"blue\"]", obj.ColorMap["blue"].ToRGB().ToHex(), "#0000ff")

	// Check unmarshalling into an existing value

	jsonBlob = []byte(`{
  "C": "#ffffff",
  "Cp": "#0000aa",
  "NestedP": {
    "NC": null
  }
}`)

	err = UnmarshalColorStruct(jsonBlob, &obj)
	if err != nil {
		t.Errorf("Error parsing json: %s", err)
	}

	checkField("C", obj.C.ToRGB().ToHex(), "#ffffff")
	checkField("Cp", (*obj.Cp).ToRGB().ToHex(), "#0000aa")
	checkField("Cpp", (**obj.Cpp).ToRGB().ToHex(), "#808080")
	checkField("Str", obj.Str, "abc")
	checkField("Nested.NC", obj.Nested.NC.ToRGB().ToHex(), "#101010")
	checkField("NestedP.NC", obj.NestedP.NC, nil)

	if len(obj.Colors) != 3 {
		t.Errorf("Length of field colors is %d, expected 3", len(obj.Colors))
	}
	checkField("Colors[0]", obj.Colors[0].ToRGB().ToHex(), "#000000")
	checkField("Colors[1]", obj.Colors[1].ToRGB().ToHex(), "#111111")
	checkField("Colors[2]", obj.Colors[2].ToRGB().ToHex(), "#222222")

	checkField("ColorMap[\"red\"]", obj.ColorMap["red"].ToRGB().ToHex(), "#ff0000")
	checkField("ColorMap[\"green\"]", obj.ColorMap["green"].ToRGB().ToHex(), "#00ff00")
	checkField("ColorMap[\"blue\"]", obj.ColorMap["blue"].ToRGB().ToHex(), "#0000ff")
}

type testColorUnmarshalEmbedTop struct {
	A string
	B Color
	TestColorUnmarshalEmbedA
	testColorUnmarshalEmbedB
}
type TestColorUnmarshalEmbedA struct {
	C string
	D string
}
type testColorUnmarshalEmbedB struct {
	E string
	F string
}

func (t *testColorUnmarshalEmbedTop) UnmarshalJSON(data []byte) error {
	return UnmarshalColorStruct(data, t)
}

func TestColorUnmarshalEmbed(t *testing.T) {
	jsonBlob := []byte(`{
  "A": "1",
  "B": "#000000",
  "C": "c",
  "D": "d"
}`)

	var obj testColorUnmarshalEmbedTop

	err := json.Unmarshal(jsonBlob, &obj)
	if err != nil {
		t.Errorf("Failed to parse json: %s", err)
	}

	if obj.A != "1" {
		t.Errorf("Field `obj.A`, expected value 1, got %s", obj.A)
	}

	if obj.B.ToRGB().ToHex() != "#000000" {
		t.Errorf("Field `obj.B`, expected value '#000000', got %s", obj.B.ToRGB().ToHex())
	}

	if obj.C != "c" {
		t.Errorf("Field `obj.C`, expected value \"c\", got \"%s\"", obj.C)
	}
	if obj.D != "d" {
		t.Errorf("Field `obj.D`, expected value \"d\", got \"%s\"", obj.D)
	}
}

type testColorUnmarshalRec struct {
	A string
	B Color
	R *testColorUnmarshalRec
}

func (t *testColorUnmarshalRec) UnmarshalJSON(data []byte) error {
	return UnmarshalColorStruct(data, t)
}
func TestColorUnmarshalRec(t *testing.T) {
	jsonBlob := []byte(`{
  "A": "1",
  "B": "#000000",
  "R": {
    "A": "2",
    "B": "#000000"
  }
}`)

	var obj testColorUnmarshalRec

	err := json.Unmarshal(jsonBlob, &obj)
	if err != nil {
		t.Errorf("Error parsing json: %s", err)
	}

	if obj.R.A != "2" {
		t.Errorf("Field `obj.R.A`, expected value 2, got %s", obj.R.A)
	}
	if obj.R.B.ToRGB().ToHex() != "#000000" {
		t.Errorf("Field `obj.R.B`, expected value '#000000', got %s", obj.R.B.ToRGB().ToHex())
	}
}

func ExampleHSLColor_Interpolate() {
	a := canvas.HSL(60, 0.9, 0.4)
	b := canvas.HSL(120, 0.9, 0.6)

	mid := a.Interpolate(b, 0.5)
	fmt.Println(mid)

	a = canvas.HSL(90, 1.0, 0.5)
	b = canvas.HSL(270, 1.0, 0.5)

	// Returns a hue of 180 instead of 0
	mid = a.Interpolate(b, 0.5)
	fmt.Println(mid)

	// Output:
	// hsl(90, 0.9, 0.5)
	// hsl(180, 1, 0.5)
}
