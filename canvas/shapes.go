package canvas

import (
	"math"

	"github.com/REANNZ/raumata/vec"
)

// Rect is a rectangle with optionally
// rounded corners
type Rect struct {
	Element
	Pos    vec.Vec2
	Width  float32
	Height float32
	Rx     float32
	Ry     float32
}

func NewRect(pos vec.Vec2, width, height float32) *Rect {
	return &Rect{
		Pos:    pos,
		Width:  width,
		Height: height,
	}
}

func NewSquare(pos vec.Vec2, width float32) *Rect {
	return NewRect(pos, width, width)
}

func (r *Rect) GetAABB() *AABB {
	if r == nil {
		return nil
	}

	a := r.Pos
	b := r.Pos.Add(vec.Vec2{X: r.Width, Y: r.Height})

	return NewAABB(a, b)
}

func (rect *Rect) Render(r Renderer) error {
	return r.RenderRect(rect)
}

// Ellipse is an ellipse centered at Center
// with x and y radiuses of Rx and Ry.
type Ellipse struct {
	Element
	Center vec.Vec2
	Rx     float32
	Ry     float32
}

func NewEllipse(center vec.Vec2, rx, ry float32) *Ellipse {
	return &Ellipse{
		Center: center,
		Rx:     rx,
		Ry:     ry,
	}
}

func NewCircle(center vec.Vec2, radius float32) *Ellipse {
	return NewEllipse(center, radius, radius)
}

func (e *Ellipse) GetAABB() *AABB {
	if e == nil {
		return nil
	}

	offset := vec.Vec2{X: e.Rx, Y: e.Ry}

	a := e.Center.Add(offset)
	b := e.Center.Sub(offset)

	return NewAABB(a, b)
}

func (ellipse *Ellipse) Render(r Renderer) error {
	return r.RenderEllipse(ellipse)
}

// Line is a straight line segment from
// Start to End
type Line struct {
	Element
	Start vec.Vec2
	End   vec.Vec2
}

func NewLine(start, end vec.Vec2) *Line {
	return &Line{
		Start: start,
		End:   end,
	}
}

func (l *Line) GetAABB() *AABB {
	if l == nil {
		return nil
	}

	return NewAABB(l.Start, l.End)
}

func (line *Line) Render(r Renderer) error {
	return r.RenderLine(line)
}

// Polygon is a closed shape with only straight sides
type Polygon struct {
	Element
	Points []vec.Vec2
}

func NewPolygon(points []vec.Vec2) *Polygon {
	return &Polygon{
		Points: points,
	}
}

// Makes a regular polygon with the given number of sides.
//
// The polygon has vertices the given radius away from the center.
// Setting `sideTop` to true rotates the polygon such that it has a side at the top,
// setting it to false means that there is always a point at the top of the polygon
//
// Panics if numSides is less than 3
func NewRegularPolygon(center vec.Vec2, radius float32, numSides int, sideTop bool) *Polygon {
	if numSides < 3 {
		panic("Cannot make a polygon with less than three sides!")
	}
	points := make([]vec.Vec2, numSides)

	internalAngle := (2 * math.Pi) / float32(numSides)
	offsetStart := vec.Vec2{X: 0, Y: -radius}
	if sideTop {
		offsetStart = offsetStart.Rotate(internalAngle / 2)
	}
	for i := 0; i < numSides; i++ {
		angle := internalAngle * float32(i)

		offset := offsetStart.Rotate(angle)
		points[i] = center.Add(offset)
	}

	return NewPolygon(points)
}

func (p *Polygon) GetAABB() *AABB {
	if p == nil {
		return nil
	}

	if len(p.Points) < 3 {
		return nil
	}

	min := p.Points[0]
	max := p.Points[0]

	for _, pt := range p.Points {
		min = min.Min(pt)
		max = max.Max(pt)
	}

	return NewAABB(min, max)
}

func (Polygon *Polygon) Render(r Renderer) error {
	return r.RenderPolygon(Polygon)
}
