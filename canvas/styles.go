package canvas

import (
	"encoding/json"
	"slices"

	"github.com/REANNZ/raumata/option"
)

// Stores style information for an element
type Style struct {
	// The overall opacity of the object
	Opacity option.Float32 `json:"opacity,omitempty"`

	// The color used to fill the object
	FillColor Color `json:"fill,omitempty"`
	// The opacity of the fill
	FillOpacity option.Float32 `json:"fill-opacity,omitempty"`
	// The color used to paint the stroke/outline of the object
	StrokeColor Color `json:"stroke,omitempty"`
	// The opacity of the stroke/outline
	StrokeOpacity option.Float32 `json:"stroke-opacity,omitempty"`
	// The width of the stroke/outline
	StrokeWidth option.Float32 `json:"stroke-width,omitempty"`

	// The font family used for text
	FontFamily string `json:"font-family,omitempty"`
}

func NewStyle() *Style {
	return &Style{}
}

func (s *Style) Merge(other *Style) {
	if s == nil || other == nil {
		return
	}

	if !s.Opacity.Valid {
		s.Opacity = other.Opacity
	}
	if s.FillColor == nil {
		s.FillColor = other.FillColor
	}
	if !s.FillOpacity.Valid {
		s.FillOpacity = other.FillOpacity
	}
	if s.StrokeColor == nil {
		s.StrokeColor = other.StrokeColor
	}
	if !s.StrokeOpacity.Valid {
		s.StrokeOpacity = other.StrokeOpacity
	}
	if !s.StrokeWidth.Valid {
		s.StrokeWidth = other.StrokeWidth
	}
	if s.FontFamily == "" {
		s.FontFamily = other.FontFamily
	}
}

// Return a style with only the values that have changed from
// s to other
func (s *Style) Changed(other *Style) *Style {
	newStyle := NewStyle()

	if s.Opacity != other.Opacity {
		newStyle.Opacity = other.Opacity
	}
	if s.FillColor == nil && other.FillColor != nil {
		newStyle.FillColor = other.FillColor
	} else if other.FillColor != nil {
		color := s.FillColor
		otherColor := other.FillColor

		if !ColorEqual(color, otherColor) {
			newStyle.FillColor = other.FillColor
		}
	}
	if s.FillOpacity != other.FillOpacity {
		newStyle.FillOpacity = other.FillOpacity
	}
	if s.StrokeColor == nil && other.StrokeColor != nil {
		newStyle.StrokeColor = other.StrokeColor
	} else if other.StrokeColor != nil {
		color := s.StrokeColor
		otherColor := other.StrokeColor

		if !ColorEqual(color, otherColor) {
			newStyle.StrokeColor = other.StrokeColor
		}
	}
	if s.StrokeOpacity != other.StrokeOpacity {
		newStyle.StrokeOpacity = other.StrokeOpacity
	}
	if s.StrokeWidth != other.StrokeWidth {
		newStyle.StrokeWidth = other.StrokeWidth
	}

	if s.FontFamily != other.FontFamily {
		newStyle.FontFamily = other.FontFamily
	}

	return newStyle
}

func (s *Style) UnmarshalJSON(data []byte) error {
	if s == nil {
		return nil
	}

	newStyle := Style{}
	if err := UnmarshalColorStruct(data, &newStyle); err != nil {
		return err
	}

	newStyle.Merge(s)

	*s = newStyle

	return nil
}

func (s *Style) MarshalJSON() ([]byte, error) {
	// `omitempty` doesn't work on struct types, meaning it includes
	// the option.Float32 values in the output as nulls. This isn't
	// desirable, so we need to filter them out ourselves
	obj := map[string]json.RawMessage{}

	marshal := func(key string, val any) error {
		d, err := json.Marshal(val)
		if err != nil {
			return err
		}
		if string(d) != "null" {
			obj[key] = d
		}
		return nil
	}

	if err := marshal("opacity", &s.Opacity); err != nil {
		return nil, err
	}
	if err := marshal("fill", s.FillColor); err != nil {
		return nil, err
	}
	if err := marshal("fill-opacity", &s.FillOpacity); err != nil {
		return nil, err
	}
	if err := marshal("stroke", s.StrokeColor); err != nil {
		return nil, err
	}
	if err := marshal("stroke-opacity", &s.StrokeOpacity); err != nil {
		return nil, err
	}
	if err := marshal("stroke-width", &s.StrokeWidth); err != nil {
		return nil, err
	}
	if s.FontFamily != "" {
		if err := marshal("font-family", s.FontFamily); err != nil {
			return nil, err
		}
	}

	return json.Marshal(obj)
}

// Stylesheet represents a set of reusable styles that
// allow for style information to be defined separately from
// individual elements.
//
// It is loosely modeled on a simplified version of CSS, basically
// only supporting classes.
type Stylesheet struct {
	rules []Rule
}

// An individual rule in a stylesheet
type Rule struct {
	Selector Selector
	Style    *Style
}

// The selection rule that matches classes to styles.
type Selector []string

// GetAllRules returns all the rules in the stylesheet
func (ss *Stylesheet) GetAllRules() []Rule {
	return ss.rules
}

// HasRule returns if the stylesheet has any rules defined
func (ss *Stylesheet) HasRules() bool {
	return len(ss.rules) > 0
}

// AddRule adds a new rule to the stylesheet
func (ss *Stylesheet) AddRule(sel Selector, style *Style) {
	if ss == nil || style == nil {
		return
	}
	r := Rule{
		Selector: sel,
		Style:    style,
	}

	ss.rules = append(ss.rules, r)

	// Ensure the rules stay sorted as `GetStyle` relies on
	// this property
	slices.SortStableFunc(ss.rules, func(a, b Rule) int {
		aLen := len(a.Selector)
		bLen := len(b.Selector)

		if aLen < bLen {
			return 1
		} else if aLen > bLen {
			return -1
		} else {
			return 0
		}
	})
}

// GetRules returns all the rules matching the given classes
func (ss *Stylesheet) GetRules(classes []string) []Rule {
	if ss == nil {
		return nil
	}

	rules := []Rule{}
	for _, rule := range ss.rules {
		if rule.Selector.Matches(classes) {
			rules = append(rules, rule)
		}
	}

	return rules
}

// GetStyle returns the combined style of all styles that match
// the given classes
func (ss *Stylesheet) GetStyle(classes []string) *Style {
	if ss == nil {
		return nil
	}

	newStyle := NewStyle()

	// This relies on the styles being sorted from most specific
	// to least specific
	for _, r := range ss.GetRules(classes) {
		newStyle.Merge(r.Style)
	}

	return newStyle
}

// Matches returns true if this selector matches the given
// classes
func (s Selector) Matches(classes []string) bool {
	for _, selClass := range s {
		hasClass := false
		for _, cls := range classes {
			if selClass == cls {
				hasClass = true
				break
			}
		}
		if !hasClass {
			return false
		}
	}

	return true
}
