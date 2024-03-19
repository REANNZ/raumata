package canvas

import "github.com/REANNZ/raumata/vec"

// A Canvas represents an abstract surface to draw to
type Canvas struct {
	Element
	Margin vec.Vec2 // Specifies the margin around the image
	Styles map[string]*Style
}

// NewCanvas returns a new Canvas to draw to
func NewCanvas() *Canvas {
	return &Canvas{
		Styles: map[string]*Style{},
	}
}

// Returns the axis aligned bounding box of the image
func (c *Canvas) GetAABB() *AABB {
	if c == nil {
		return nil
	}
	return GetCombinedAABB(c.Children)
}

// Render the canvas using the given renderer
func (c *Canvas) Render(renderer Renderer) error {
	return renderer.RenderCanvas(c)
}

// Object is the interface implemented by Canvas objects
type Object interface {
	GetAABB() *AABB
	GetAttributes() *Attributes
	Render(Renderer) error
}

type Container interface {
	Object
	AppendChild(Object)
}

// Element holds common fields for [Object]s
type Element struct {
	Attributes Attributes
	Children   []Object
}

func (e *Element) AppendChild(obj Object) {
	e.Children = append(e.Children, obj)
}

func (e *Element) GetAttributes() *Attributes {
	return &e.Attributes
}

// Renderer is an interface for Canvas renderers.
// It implements the Visitor pattern
type Renderer interface {
	RenderCanvas(*Canvas) error
	RenderGroup(*Group) error
	RenderRect(*Rect) error
	RenderEllipse(*Ellipse) error
	RenderLine(*Line) error
	RenderPolygon(*Polygon) error
	RenderPath(*Path) error
	RenderText(*Text) error
}

// Helper function for rendering children
func RenderChildren(renderer Renderer, children []Object) error {
	for _, obj := range children {
		if obj == nil {
			continue
		}
		err := obj.Render(renderer)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper function for calculating the union of the
// AABBs of a set of objects
func GetCombinedAABB(objs []Object) *AABB {
	var unionAabb *AABB = nil

	for _, obj := range objs {
		if obj == nil {
			continue
		}
		aabb := obj.GetAABB()
		if aabb == nil {
			continue
		}

		g, ok := obj.(*Group)
		if ok && g.Transform != nil {
			aabb = aabb.Transform(g.Transform)
		}

		unionAabb = unionAabb.Union(aabb)
	}

	return unionAabb
}
