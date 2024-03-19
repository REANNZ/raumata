package canvas

import "github.com/REANNZ/raumata/vec"

// An axis-aligned bounding box
//
// The zero value is a zero-sized bounding
// box around the origin
type AABB struct {
	min vec.Vec2
	max vec.Vec2
}

// Constructs a new axis-aligned bounding box
// from the two points.
func NewAABB(min, max vec.Vec2) *AABB {
	return &AABB{
		min: min.Min(max),
		max: min.Max(max),
	}
}

func (a *AABB) Bounds() (min, max vec.Vec2) {
	return a.min, a.max
}

func (a *AABB) Size() vec.Vec2 {
	return a.max.Sub(a.min)
}

// Union returns the union of the two AABBs
func (a *AABB) Union(b *AABB) *AABB {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return &AABB{
		min: a.min.Min(b.min),
		max: a.max.Max(b.max),
	}
}

// Transform returns the AABB of a transformed by t
func (a *AABB) Transform(t *vec.Transform) *AABB {
	// Construct the four corners of the box, we
	// can't just transform the min and max points as
	// that won't produce an AABB for the whole box.
	p0 := a.min
	p1 := vec.Vec2{X: a.min.X, Y: a.max.Y}
	p2 := a.max
	p3 := vec.Vec2{X: a.max.X, Y: a.min.Y}

	// Transform all the points
	p0 = t.Apply(p0)
	p1 = t.Apply(p1)
	p2 = t.Apply(p2)
	p3 = t.Apply(p3)

	// Find the minimum and maximum of the
	// points
	min := p0.Min(p1.Min(p2.Min(p3)))
	max := p0.Max(p1.Max(p2.Max(p3)))

	return NewAABB(min, max)
}
