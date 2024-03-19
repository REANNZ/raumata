package canvas_test

import (
	"testing"

	. "github.com/REANNZ/raumata/canvas"
)

func checkStyleEq(t *testing.T, expected, actual *Style) {
	t.Helper()

	if !ColorEqual(expected.FillColor, actual.FillColor) {
		t.Errorf("FillColor not correct, expected %s, got %s",
			expected.FillColor, actual.FillColor)
	}
	if !ColorEqual(expected.StrokeColor, actual.StrokeColor) {
		t.Errorf("StrokeColor not correct, expected %s, got %s",
			expected.StrokeColor, actual.StrokeColor)
	}

	if actual.Opacity != expected.Opacity {
		t.Errorf("Opacity not correct, expected %s, got %s",
			&expected.Opacity, &actual.Opacity)
	}

	if actual.FillOpacity != expected.FillOpacity {
		t.Errorf("FillOpacity not correct, expected %s, got %s",
			&expected.FillOpacity, &actual.FillOpacity)
	}

	if actual.StrokeOpacity != expected.StrokeOpacity {
		t.Errorf("StrokeOpacity not correct, expected %s, got %s",
			&expected.StrokeOpacity, &actual.StrokeOpacity)
	}
}

func TestStyleChanged(t *testing.T) {
	blank := NewStyle()

	s := NewStyle()
	s.FillColor = RGB(0, 0, 0)
	s.StrokeColor = RGB(0, 0, 0)
	s.Opacity.Set(1)
	s.StrokeWidth.Set(1)
	s.StrokeOpacity.Set(1)
	s.FillOpacity.Set(1)

	changed := blank.Changed(s)

	checkStyleEq(t, s, changed)

	s2 := NewStyle()
	s2.FillColor = RGB(1, 0, 1)
	s2.StrokeColor = RGB(0, 0, 0)
	s2.StrokeWidth.Set(0)
	s2.StrokeOpacity.Set(1)

	expected := NewStyle()
	expected.FillColor = RGB(1, 0, 1)
	expected.StrokeColor = nil
	expected.StrokeWidth.Set(0)

	changed = s.Changed(s2)

	checkStyleEq(t, expected, changed)
}
