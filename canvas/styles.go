package canvas

import (
	"encoding/json"

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
		sp := &s
		*sp = NewStyle()
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
