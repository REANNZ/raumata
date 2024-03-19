package canvas

import "github.com/REANNZ/raumata/vec"

// A group of objects
// Can also have a transformation applied to
// it
type Group struct {
	Element
	Transform *vec.Transform
}

func NewGroup() *Group {
	return &Group{}
}

func (g *Group) GetAABB() *AABB {
	if g == nil {
		return nil
	}

	return GetCombinedAABB(g.Children)
}

func (g *Group) Render(r Renderer) error {
	return r.RenderGroup(g)
}
