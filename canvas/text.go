package canvas

import "github.com/REANNZ/raumata/vec"

type TextAnchor int

const (
	TextAnchorNone TextAnchor = iota
	TextAnchorStart
	TextAnchorMiddle
	TextAnchorEnd
)

// Text is some text drawn to the canvas
type Text struct {
	Attributes Attributes
	Pos        vec.Vec2
	Text       string
	Size       float32
	Anchor     TextAnchor
}

func NewText(pos vec.Vec2, text string) *Text {
	return &Text{
		Pos:  pos,
		Text: text,
		Size: 10,
	}
}

func (t *Text) GetAABB() *AABB {
	if t == nil {
		return nil
	}
	// TODO: use actual font-based calcuations to derive the bounding-box
	// instead of these arbitrary heuristics
	// golang.org/x/image/font would be the most useful.
	ascender := t.Size * 0.85
	advance := t.Size * 0.65

	min := t.Pos.Sub(vec.Vec2{X: 0, Y: ascender})

	width := advance * float32(len(t.Text))

	switch t.Anchor {
	case TextAnchorMiddle:
		min.X -= width / 2
	case TextAnchorEnd:
		min.X -= width
	}

	max := min.Add(vec.Vec2{X: width, Y: t.Size})

	return NewAABB(min, max)
}

func (t *Text) Render(r Renderer) error {
	return r.RenderText(t)
}

func (t *Text) GetAttributes() *Attributes {
	return &t.Attributes
}

func (a TextAnchor) String() string {
	switch a {
	case TextAnchorStart:
		return "start"
	case TextAnchorMiddle:
		return "middle"
	case TextAnchorEnd:
		return "end"
	default:
		return ""
	}
}
