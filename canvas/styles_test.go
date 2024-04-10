package canvas_test

import (
	"testing"

	. "github.com/REANNZ/raumata/canvas"
)

func checkStyleEq(t *testing.T, expected, actual *Style) {
	t.Helper()

	styleColorEqual := func(a, b StyleColor) bool {
		if a == b {
			return true
		}

		if a.IsNone() != b.IsNone() {
			return false
		} else if a.IsNone() && b.IsNone() {
			return true
		}

		return ColorEqual(a.Color(), b.Color())
	}

	if !styleColorEqual(expected.FillColor, actual.FillColor) {
		t.Errorf("FillColor not correct, expected %s, got %s",
			&expected.FillColor, &actual.FillColor)
	}
	if !styleColorEqual(expected.StrokeColor, actual.StrokeColor) {
		t.Errorf("StrokeColor not correct, expected %s, got %s",
			&expected.StrokeColor, &actual.StrokeColor)
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

	if actual.FontFamily != expected.FontFamily {
		t.Errorf("FontFamily not correct, expected %s, got %s",
			expected.FontFamily, actual.FontFamily)
	}
}

func TestStyleChanged(t *testing.T) {
	blank := NewStyle()

	s := NewStyle()
	s.FillColor.SetColor(RGB(0, 0, 0))
	s.StrokeColor.SetColor(RGB(0, 0, 0))
	s.Opacity.Set(1)
	s.StrokeWidth.Set(1)
	s.StrokeOpacity.Set(1)
	s.FillOpacity.Set(1)

	changed := blank.Changed(s)

	checkStyleEq(t, s, changed)

	s2 := NewStyle()
	s2.FillColor.SetColor(RGB(1, 0, 1))
	s2.StrokeColor.SetColor(RGB(0, 0, 0))
	s2.StrokeWidth.Set(0)
	s2.StrokeOpacity.Set(1)

	expected := NewStyle()
	expected.FillColor.SetColor(RGB(1, 0, 1))
	expected.StrokeWidth.Set(0)

	changed = s.Changed(s2)

	checkStyleEq(t, expected, changed)
}

func TestSelectorMatches(t *testing.T) {
	selector := Selector{}

	if !selector.Matches([]string{"test"}) {
		t.Errorf("Empty selector should match any classes")
	}

	if !selector.Matches([]string{"foo", "bar", "baz"}) {
		t.Errorf("Empty selector should match any classes")
	}

	selector = Selector{"foo"}

	if !selector.Matches([]string{"foo"}) {
		t.Errorf("'foo' selector did not match 'foo'")
	}
	if selector.Matches([]string{"bar"}) {
		t.Errorf("'foo' selector should not match 'bar'")
	}
	if !selector.Matches([]string{"foo", "bar", "baz"}) {
		t.Errorf("'foo' selector did not match 'foo', 'bar', 'baz'")
	}

	selector = Selector{"foo", "bar", "baz"}
	if selector.Matches([]string{"foo"}) {
		t.Errorf("'foo', 'bar', 'baz' selector should not match 'foo'")
	}
	if !selector.Matches([]string{"foo", "bar", "baz"}) {
		t.Errorf("'foo', 'bar', 'baz' selector did not match 'foo', 'bar', 'baz'")
	}
}

func TestStylesheet(t *testing.T) {
	stylesheet := Stylesheet{}

	a := NewStyle()
	a.FillColor = NewStyleColor(RGB(1, 0, 0))
	stylesheet.AddRule(Selector{"a"}, a)

	b := NewStyle()
	b.StrokeColor = NewStyleColor(RGB(0, 1, 0))
	stylesheet.AddRule(Selector{"b"}, b)

	c := NewStyle()
	c.Opacity.Set(0.5)
	stylesheet.AddRule(Selector{"c"}, c)

	rules := stylesheet.GetRules([]string{"a"})
	if len(rules) != 1 {
		t.Errorf("Expected one rule to match 'a', got %d", len(rules))
	}

	rule := rules[0]
	if len(rule.Selector) != 1 && rule.Selector[0] != "a" {
		t.Errorf("Incorrect selector: %v", rule.Selector)
	}
	checkStyleEq(t, a, rule.Style)

	rules = stylesheet.GetRules([]string{"b"})
	if len(rules) != 1 {
		t.Errorf("Expected one rule to match 'b', got %d", len(rules))
	}

	rule = rules[0]
	if len(rule.Selector) != 1 && rule.Selector[0] != "b" {
		t.Errorf("Incorrect selector: %v", rule.Selector)
	}
	checkStyleEq(t, b, rule.Style)

	rules = stylesheet.GetRules([]string{"c"})
	if len(rules) != 1 {
		t.Errorf("Expected one rule to match 'c', got %d", len(rules))
	}

	rule = rules[0]
	if len(rule.Selector) != 1 && rule.Selector[0] != "c" {
		t.Errorf("Incorrect selector: %v", rule.Selector)
	}
	checkStyleEq(t, c, rule.Style)

	rules = stylesheet.GetRules([]string{"a", "b", "c"})
	if len(rules) != 3 {
		t.Errorf("Expected three rules to match 'a', 'b', 'c', got %d", len(rules))
	}

	expectedStyle := NewStyle()
	expectedStyle.FillColor.SetColor(RGB(1, 0, 0))
	expectedStyle.StrokeColor.SetColor(RGB(0, 1, 0))
	expectedStyle.Opacity.Set(0.5)
	style := stylesheet.GetStyle([]string{"a", "b", "c"})

	checkStyleEq(t, expectedStyle, style)
}
