package raumata

import (
	"math"
	"slices"

	"github.com/REANNZ/raumata/canvas"
	"github.com/REANNZ/raumata/internal/f32"
	"github.com/REANNZ/raumata/option"
	"github.com/REANNZ/raumata/vec"
)

// Stores style information for nodes
type NodeStyle struct {
	// Size of the node
	Size float32 `json:"size"`
	*canvas.Style
}

// Stores style information for links
type LinkStyle struct {
	Size float32 `json:"size"`
	// Bend radius for the drawn line
	Radius option.Float32 `json:"radius"`
	*canvas.Style
}

// Style information for node and link labels
type LabelStyle struct {
	Size         float32      `json:"size"`                       // Font size
	Color        canvas.Color `json:"color"`                      // Text color
	FontFamily   string       `json:"font-family"`                // Font family
	Background   canvas.Color `json:"background-color,omitempty"` // Background color - Link only
	Border       canvas.Color `json:"border-color,omitempty"`     // Border color - Link only
	BorderRadius float32      `json:"border-radius,omityempty"`   // Border radius - Link only
	Width        float32      `json:"width,omitempty"`            // Label width - Link only
	Opacity      float32      `json:"opacity,omitempty"`          // Label background opacity - Link only
}

// Configuration values for the renderer
//
// The zero value is not usable, instead it is better to
// create one with [DefaultRenderConfig] and modify it.
type RenderConfig struct {
	MinNodeSep       float32              `json:"min-node-sep"`
	DefaultNodeStyle NodeStyle            `json:"node-style"`
	NodeStyles       map[string]NodeStyle `json:"node-styles,omitempty"`
	DefaultLinkStyle LinkStyle            `json:"link-style"`
	LinkStyles       map[string]LinkStyle `json:"link-styles,omitempty"`
	NodeLabelStyle   LabelStyle           `json:"node-label-style"`
	LinkLabelStyle   LabelStyle           `json:"link-label-style"`
	LinkColorScale   *canvas.ColorScale   `json:"link-color-scale"`
}

func DefaultRenderConfig() *RenderConfig {

	config := &RenderConfig{
		MinNodeSep: 5,
		DefaultNodeStyle: NodeStyle{
			Size: 20,
			Style: &canvas.Style{
				StrokeWidth: option.Float32{},
				StrokeColor: canvas.NewStyleColor(canvas.RGB(0, 0, 0)),
				FillColor:   canvas.NewStyleColor(canvas.RGB(1, 1, 1)),
			},
		},
		DefaultLinkStyle: LinkStyle{
			Size:   10,
			Radius: option.Float32{},
			Style: &canvas.Style{
				StrokeWidth: option.Float32{},
				FillColor:   canvas.NewStyleColor(canvas.RGB(0.5, 0.5, 0.5)),
			},
		},
		LinkColorScale: canvas.HeatColorScale(),
		NodeStyles:     map[string]NodeStyle{},
		LinkStyles:     map[string]LinkStyle{},
		NodeLabelStyle: LabelStyle{
			Size:       16,
			FontFamily: "sans-serif",
			Color:      canvas.RGB(0, 0, 0),
		},
		LinkLabelStyle: LabelStyle{
			Size:         8,
			FontFamily:   "monospace",
			Color:        canvas.RGB(0, 0, 0),
			Background:   canvas.RGB(1, 1, 1),
			Border:       canvas.RGB(0, 0, 0),
			Opacity:      0.9,
			BorderRadius: 3,
			Width:        28,
		},
	}

	config.DefaultNodeStyle.StrokeWidth.Set(4)
	config.DefaultLinkStyle.StrokeWidth.Set(0)
	config.DefaultLinkStyle.Radius.Set(10)

	return config
}

type Renderer struct {
	Config *RenderConfig
	scale  float32
	nodeSizes map[NodeId]float32
}

func NewRenderer() *Renderer {
	return &Renderer{
		Config: DefaultRenderConfig(),
	}
}

func NewRendererWithConfig(config *RenderConfig) *Renderer {
	return &Renderer{
		Config: config,
	}
}

// GetScale returns the scale factor used for converting
// positions from the topology grid into canvas positions
//
// By default it is calculated so the largest node size (from
// configured styles) is approximately the same as one unit in the grid.
//
// Use [Renderer.SetScale] to override the scale
func (r *Renderer) GetScale() float32 {
	if r.scale > 0 {
		return r.scale
	}

	maxNodeSize := r.Config.DefaultNodeStyle.Size
	maxNodeStrokeWidth := r.Config.DefaultNodeStyle.StrokeWidth.Value
	for _, style := range r.Config.NodeStyles {
		if style.Size > maxNodeSize {
			maxNodeSize = style.Size
		}
		if style.StrokeWidth.Valid && style.StrokeWidth.Value > maxNodeStrokeWidth {
			maxNodeStrokeWidth = style.StrokeWidth.Value
		}
	}

	r.scale = r.Config.MinNodeSep + maxNodeSize + maxNodeStrokeWidth

	return r.scale
}

// Explicitly set the scale, s must be greater than 0
func (r *Renderer) SetScale(s float32) {
	r.scale = s
}

// RenderTopologyToCanvas renders the given Topology to the top level of the given
// This also adds the styles to the canvas.
func (r *Renderer) RenderTopologyToCanvas(topo *Topology, c *canvas.Canvas) error {
	g, err := r.RenderTopology(topo)
	if err != nil {
		return err
	}

	c.AppendChild(g)
	r.SetStyles(c)

	return nil
}

// RenderTopology renders the given Topology and returns a [canvas.Object] that
// can be added to a canvas or other object
func (r *Renderer) RenderTopology(topo *Topology) (canvas.Object, error) {
	links := make([]*Link, 0, len(topo.Links))
	nodes := make([]*Node, 0, len(topo.Nodes))

	r.nodeSizes = map[NodeId]float32{}

	// Collect and sort the links and nodes, this keeps the output
	// consistent between runs
	for _, l := range topo.Links {
		// Filter out un-routed links
		if l != nil && len(l.Route) >= 2 {
			links = append(links, l)
		}
	}
	for _, n := range topo.Nodes {
		// Filter out nodes without a position
		if n != nil && n.Pos != nil {
			nodes = append(nodes, n)
			style := r.getNodeStyle(n)
			r.nodeSizes[n.Id] = style.Size
		}
	}

	slices.SortFunc(links, func(a, b *Link) int {
		if a.Id < b.Id {
			return -1
		} else if a.Id > b.Id {
			return 1
		} else {
			return 0
		}
	})

	slices.SortFunc(nodes, func(a, b *Node) int {
		if a.Id < b.Id {
			return -1
		} else if a.Id > b.Id {
			return 1
		} else {
			return 0
		}
	})

	group := canvas.NewGroup()
	group.Attributes.Id = "topology"

	linkGroup, err := r.RenderLinks(links)
	if err != nil {
		return nil, err
	}

	nodeGroup, err := r.RenderNodes(nodes)
	if err != nil {
		return nil, err
	}

	group.AppendChild(linkGroup)
	group.AppendChild(nodeGroup)

	return group, nil
}

// RenderNodes renders a list of nodes and returns a [canvas.Object]
func (r *Renderer) RenderNodes(nodes []*Node) (canvas.Object, error) {
	group := canvas.NewGroup()
	group.Attributes.Id = "nodes"

	for _, node := range nodes {
		obj, err := r.RenderNode(node)
		if err != nil {
			return nil, err
		}
		if obj != nil {
			group.AppendChild(obj)
		}
	}

	return group, nil
}

// RenderLinks renders a list of links and returns a [canvas.Object]
func (r *Renderer) RenderLinks(links []*Link) (canvas.Object, error) {
	group := canvas.NewGroup()
	group.Attributes.Id = "links"

	for _, link := range links {
		obj, err := r.RenderLink(link)
		if err != nil {
			return nil, err
		}
		if obj != nil {
			group.AppendChild(obj)
		}
	}

	return group, nil
}

// RenderNode renders the given Node and returns a [canvas.Object]
func (r *Renderer) RenderNode(node *Node) (canvas.Object, error) {
	if node == nil || node.Pos == nil {
		return nil, nil
	}
	scale := r.GetScale()
	// pos is the center of the node shape
	pos := vec.Vec2{X: float32(node.Pos[0]), Y: float32(node.Pos[1])}
	pos = pos.Mul(scale)

	style := r.getNodeStyle(node)

	// Create a group for the node
	nodeGroup := canvas.NewGroup()
	nodeGroup.Attributes.Id = string("N-" + node.Id)
	nodeGroup.Attributes.SetExtra("data-node", string(node.Id))

	// NOTE: this is where you'd branch off for different node styles
	var nodeShape canvas.Object = canvas.NewCircle(pos, style.Size/2)

	if node.IsMultiCell() {
		radius := style.Size / 2;
		nodeMin, nodeMax := node.GetExtents()
		nodeShape = r.RenderShape(radius, vec.Polyline{
			{ X: nodeMin.X, Y: nodeMin.Y },
			{ X: nodeMax.X, Y: nodeMin.Y },
			{ X: nodeMax.X, Y: nodeMax.Y },
			{ X: nodeMin.X, Y: nodeMax.Y },
		})
	}

	attrs := nodeShape.GetAttributes()
	attrs.AddClass("node")
	if node.Class != "" {
		attrs.AddClass(node.Class)
	}

	if node.Style != nil {
		// Copy the node style over to the node shape
		attrs.Style = node.Style.Style
	}

	nodeGroup.AppendChild(nodeShape)

	if node.IsMultiCell() || node.LabelAt != "" {
		label, err := r.RenderNodeLabel(node)
		if err != nil {
			return nil, err
		}
		if label != nil {
			nodeGroup.AppendChild(label)
		}
	}

	return nodeGroup, nil
}

// RenderLink renders the given Link and returns a [canvas.Object]
func (r *Renderer) RenderLink(link *Link) (canvas.Object, error) {
	if link == nil || link.Route == nil {
		return nil, nil
	}

	route := link.Route.Simplify()

	style := r.getLinkStyle(link)
	scale := r.GetScale()

	linkGroup := canvas.NewGroup()
	linkGroup.Attributes.Id = string("L-" + link.Id)
	linkGroup.Attributes.AddClass("link")
	if link.Class != "" {
		linkGroup.Attributes.AddClass(link.Class)
	}

	// The node sizes are used to adjust lengths along links
	fromSize := r.getNodeSize(link.From)
	toSize := r.getNodeSize(link.To)

	// NOTE: This is where you'd branch off for different link styles
	//       (e.g. double line instead of opposing arrows)

	var splitAt float32
	if link.SplitAt != nil {
		splitAt = *link.SplitAt
	} else if fromSize == toSize {
		// Optimisation for common case
		splitAt = 0.5
	} else {
		// Calculate the split point halfway along the visual length of the
		// link.
		routeLen := route.Length()
		// Scale the to/from sizes to grid space
		fromSizeGrid := fromSize / scale
		toSizeGrid := toSize / scale

		// This calculates a split point that has been moved further along
		// the path proportional to fromSize and pulled back along the path
		// proportional to toSize
		splitAt = 1 + (fromSizeGrid - toSizeGrid) / routeLen
		splitAt = splitAt / 2
	}

	// Clamp splitAt to 0 < x < 1
	splitAt = f32.Max(f32.Min(splitAt, 0.99), 0.01)

	splitTolerance := style.Size / scale
	routeA, routeB := findSplit(route, splitAt, splitTolerance)
	routeA = routeA.Mul(scale)
	routeB = routeB.Mul(scale)

	// TODO: handle state-dependent link-coloring (e.g. grey for down)

	// Helper function for rendering the individual link parts
	renderLinkSegment := func(route vec.Polyline, data *LinkData, from, to string) (canvas.Object, error) {
		var color canvas.StyleColor = style.FillColor
		if data != nil && data.Value.Valid {
			color.SetColor(r.Config.LinkColorScale.GetColor(data.Value.Value))
		}
		path := renderArrow(route, style.Size, style.Radius.Value)
		if path == nil {
			return nil, nil
		}

		if !color.IsZero() {
			path.Attributes.EnsureStyle()
			path.Attributes.Style.FillColor = color
		}

		linkSeg := canvas.NewGroup()
		linkSeg.Attributes.AddClass("link-segment")
		linkSeg.Attributes.SetExtra("data-from", from)
		linkSeg.Attributes.SetExtra("data-to", to)

		linkSeg.AppendChild(path)

		if data != nil && data.Label != "" {
			// Calculate the adjustment to the centre point
			// due to the node and the arrow head
			adjustment := r.getNodeSize(NodeId(from))
			adjustment -= style.Size
			// Calculate the offset 0.5 along the path as seen
			t := 1 + (adjustment / (route.Length()))
			t = t / 2
			labelPos := route.Interpolate(t)
			label, err := r.RenderLinkLabel(labelPos, data.Label)
			if err != nil {
				return nil, err
			}
			linkSeg.AppendChild(label)
		}

		return linkSeg, nil
	}

	linkSegA, err := renderLinkSegment(routeA, link.FromData, string(link.From), string(link.To))
	if err != nil {
		return nil, err
	}
	linkSegB, err := renderLinkSegment(routeB, link.ToData, string(link.To), string(link.From))
	if err != nil {
		return nil, err
	}

	if link.Class != "" {
		linkSegA.GetAttributes().AddClass(link.Class)
		linkSegB.GetAttributes().AddClass(link.Class)
	}

	linkGroup.AppendChild(linkSegA)
	linkGroup.AppendChild(linkSegB)

	// TODO: State handling

	return linkGroup, nil
}

// RenderNodeLabel renders the label for the given Node and returns a [canvas.Object]
func (r *Renderer) RenderNodeLabel(node *Node) (canvas.Object, error) {
	scale := r.GetScale()

	style := r.getNodeStyle(node)

	pos := vec.Vec2{X: float32(node.Pos[0]), Y: float32(node.Pos[1])}
	if node.IsMultiCell() {
		minPos, maxPos := node.GetExtents()
		pos = minPos.Add(maxPos).Div(2)
	}
	labelPos := pos.Mul(scale)
	anchor := canvas.TextAnchorNone
	offsetDist := (style.Size / 2) + style.StrokeWidth.Value

	textSize := r.Config.NodeLabelStyle.Size

	// Calculate the offset from the node position
	// by rotating a vector to the appropriate position

	offsetVec := vec.Vec2{X: offsetDist, Y: 0}
	textAdjust := vec.Vec2{}

	// Don't place diagonal labels at the 45deg rotation,
	// instead rotate them so they're closer to the vertical.
	// This makes the association with the nodes slighly clearer.
	// The angle 3π/8 is 67.5deg
	var diagAngle float32 = (3 * math.Pi) / 8
	switch node.LabelAt {
	case "n":
		offsetVec = offsetVec.Rotate(-math.Pi / 2)
		anchor = canvas.TextAnchorMiddle
	case "ne":
		offsetVec = offsetVec.Rotate(-diagAngle)
		anchor = canvas.TextAnchorStart
	case "e":
		textAdjust.Y = textSize / 2
		anchor = canvas.TextAnchorStart
	case "se":
		offsetVec = offsetVec.Rotate(diagAngle)
		textAdjust.Y = textSize
		anchor = canvas.TextAnchorStart
	case "s":
		offsetVec = offsetVec.Rotate(math.Pi / 2)
		textAdjust.Y = textSize
		anchor = canvas.TextAnchorMiddle
	case "sw":
		offsetVec = offsetVec.Rotate(math.Pi - diagAngle)
		textAdjust.Y = textSize
		anchor = canvas.TextAnchorEnd
	case "w":
		offsetVec = offsetVec.Rotate(math.Pi)
		textAdjust.Y = textSize / 2
		anchor = canvas.TextAnchorEnd
	case "nw":
		offsetVec = offsetVec.Rotate(math.Pi + diagAngle)
		anchor = canvas.TextAnchorEnd
	case "c":
		if node.IsMultiCell() {
			offsetVec = vec.Vec2{}
			anchor = canvas.TextAnchorMiddle
			textAdjust.Y = textSize / 2
		}
	}

	if anchor != canvas.TextAnchorNone {
		labelPos = labelPos.Add(offsetVec).Add(textAdjust)
		labelText := string(node.Id)
		if node.Label != "" {
			labelText = node.Label
		}
		label := canvas.NewText(labelPos, labelText)
		label.Anchor = anchor
		label.Size = textSize
		label.Attributes.AddClass("node-label-text")

		return label, nil
	}

	return nil, nil
}

// RenderLinkLabel renders a link label at pos and returns a [canvas.Object]
func (r *Renderer) RenderLinkLabel(pos vec.Vec2, text string) (canvas.Object, error) {

	size := r.Config.LinkLabelStyle.Size
	radius := r.Config.LinkLabelStyle.BorderRadius

	textPos := vec.Vec2{X: 0, Y: size / 2}

	textObj := canvas.NewText(textPos, text)
	textObj.Anchor = canvas.TextAnchorMiddle
	textObj.Size = size
	textObj.Attributes.AddClass("link-label-text")

	width := r.Config.LinkLabelStyle.Width
	height := size + 5
	border := canvas.NewRect(vec.Vec2{X: -width / 2, Y: -height / 2}, width, height)
	if radius > 0 {
		radius = f32.Min(radius, height/2)
		border.Rx = radius
		border.Ry = radius
	}
	border.Attributes.AddClass("link-label-box")

	transform := vec.NewTranslate(pos)
	labelGroup := canvas.NewGroup()
	labelGroup.Transform = transform
	labelGroup.Attributes.AddClass("link-label")
	labelGroup.AppendChild(border)
	labelGroup.AppendChild(textObj)

	return labelGroup, nil
}

// Sets the styles configured in the Renderer to the canvas
//
// The following classes are created in the canvas:
//
//   - "node" - Styles that apply to all nodes
//   - "link-segment" - Styles that apply to all link segments
//   - "node-label-text" - Styles that apply to all node labels
//   - "link-label-text" - Styles that apply to all link labels
//   - "link-label-box" - Styles that apply to all link labels
func (r *Renderer) SetStyles(c *canvas.Canvas) {
	c.Stylesheet.AddRule(canvas.Selector{"node"}, r.Config.DefaultNodeStyle.Style)
	for cls, style := range r.Config.NodeStyles {
		sel := canvas.Selector{"node", cls}
		c.Stylesheet.AddRule(sel, style.Style)
	}
	c.Stylesheet.AddRule(canvas.Selector{"link-segment"}, r.Config.DefaultLinkStyle.Style)
	for cls, style := range r.Config.LinkStyles {
		sel := canvas.Selector{"link-segment", cls}
		c.Stylesheet.AddRule(sel, style.Style)
	}

	nodeLabelStyle := canvas.NewStyle()
	nodeLabelStyle.FillColor.SetColor(r.Config.NodeLabelStyle.Color)
	nodeLabelStyle.FontFamily = r.Config.NodeLabelStyle.FontFamily
	c.Stylesheet.AddRule(canvas.Selector{"node-label-text"}, nodeLabelStyle)

	linkLabelTextStyle := canvas.NewStyle()
	linkLabelTextStyle.FillColor.SetColor(r.Config.LinkLabelStyle.Color)
	linkLabelTextStyle.FontFamily = r.Config.LinkLabelStyle.FontFamily
	c.Stylesheet.AddRule(canvas.Selector{"link-label-text"}, linkLabelTextStyle)

	linkLabelBoxStyle := canvas.NewStyle()
	linkLabelBoxStyle.FillColor.SetColor(r.Config.LinkLabelStyle.Background)
	linkLabelBoxStyle.StrokeColor.SetColor(r.Config.LinkLabelStyle.Border)
	linkLabelBoxStyle.Opacity.Set(r.Config.LinkLabelStyle.Opacity)
	linkLabelBoxStyle.StrokeWidth.Set(1)
	c.Stylesheet.AddRule(canvas.Selector{"link-label-box"}, linkLabelBoxStyle)
}

// Helper function for rendering shapes in grid-space at the appropriate scale.
// Paths is a set of paths that define the shape, the shape is always closed, corners
// are radiused if radius > 0
func (r *Renderer) RenderShape(radius float32, paths ...vec.Polyline) canvas.Object {
	pathObj := canvas.NewPath()

	scale := r.GetScale()

	for _, path := range paths {
		path = path.Mul(scale).Simplify()

		if radius <= 0 {
			// Handle the simple case where it's just a polygon
			for i, p := range path {
				if i == 0 {
					pathObj.MoveTo(p)
				} else {
					pathObj.LineTo(p)
				}
			}
			pathObj.ClosePath()
		} else {
			// Handling radiused corners is more challenging
			// As the final path may never actually pass through any of
			// the points in the given path
			for i := range path {
				var prevPoint, curPoint, nextPoint vec.Vec2
				curPoint = path[i]
				if i == 0 {
					prevPoint = path[len(path)-1]
				} else {
					prevPoint = path[i-1]
				}
				if i == len(path)-1 {
					nextPoint = path[0]
				} else {
					nextPoint = path[i+1]
				}

				prevPoint = prevPoint.Add(curPoint).Div(2)
				nextPoint = curPoint.Add(nextPoint).Div(2)

				if i == 0 {
					pathObj.MoveTo(prevPoint)
				}

				pathObj.RoundCorner(radius, prevPoint, curPoint, nextPoint)
			}
		}
	}

	return pathObj
}

func (r *Renderer) RenderGrid(bounds *canvas.AABB) canvas.Object {
	gridGroup := canvas.NewGroup()
	attrs := &gridGroup.Attributes
	attrs.EnsureStyle()
	attrs.Style.StrokeColor.SetColor(canvas.HSL(0, 0, 0.5))

	scale := r.GetScale()

	minPos, maxPos := bounds.Bounds()

	minPos = minPos.Div(scale).Floor().Mul(scale)
	maxPos = maxPos.Div(scale).Floor().Mul(scale)

	minPos.X -= scale / 2
	minPos.Y -= scale / 2

	for x := minPos.X; x <= maxPos.X; x += scale {
		start := vec.Vec2{ X: x, Y: minPos.Y }
		end := vec.Vec2{ X: x, Y: maxPos.Y }
		line := canvas.NewLine(start, end)
		gridGroup.AppendChild(line)
	}

	for y := minPos.Y; y <= maxPos.Y; y += scale {
		start := vec.Vec2{ X: minPos.X, Y: y }
		end := vec.Vec2{ X: maxPos.X, Y: y }
		line := canvas.NewLine(start, end)
		gridGroup.AppendChild(line)
	}

	return gridGroup
}

func (r *Renderer) getLinkStyle(link *Link) *LinkStyle {
	style := &LinkStyle{
		Style: canvas.NewStyle(),
	}

	if link.Style != nil {
		style.merge(link.Style)
	}

	if link.Class != "" {
		classStyle, ok := r.Config.LinkStyles[link.Class]
		if ok {
			style.merge(&classStyle)
		}
	}

	style.merge(&r.Config.DefaultLinkStyle)

	return style
}

func (r *Renderer) getNodeStyle(node *Node) *NodeStyle {
	style := &NodeStyle{
		Style: canvas.NewStyle(),
	}

	if node.Style != nil {
		*style = *node.Style
	}

	if node.Class != "" {
		classStyle, ok := r.Config.NodeStyles[node.Class]
		if ok {
			style.merge(&classStyle)
		}
	}

	style.merge(&r.Config.DefaultNodeStyle)

	return style
}

func (r *Renderer) getNodeSize(nodeId NodeId) float32 {
	if r.nodeSizes == nil {
		return r.Config.DefaultNodeStyle.Size
	}
	size, ok := r.nodeSizes[nodeId]
	if !ok {
		return r.Config.DefaultNodeStyle.Size
	}

	return size
}

func (s *NodeStyle) merge(other *NodeStyle) {
	if s.Style == nil {
		s.Style = canvas.NewStyle()
	}
	s.Style.Merge(other.Style)
	if s.Size == 0 {
		s.Size = other.Size
	}
}

func (s *LinkStyle) merge(other *LinkStyle) {
	if s.Style == nil {
		s.Style = canvas.NewStyle()
	}
	s.Style.Merge(other.Style)
	if s.Size == 0 {
		s.Size = other.Size
	}
	if !s.Radius.Valid {
		s.Radius = other.Radius
	}
}

func renderArrow(route vec.Polyline, width, radius float32) *canvas.Path {
	if len(route) < 2 {
		return nil
	}

	path := canvas.NewPath()

	path.Attributes.Style = canvas.NewStyle()

	halfWidth := width / 2

	// The last point on the line is the point of the arrow
	// We essentially remove that point and replace it with
	// one offset from the end
	arrowPoint := route[len(route)-1]
	prevPoint := route[len(route)-2]

	dir := arrowPoint.Sub(prevPoint)
	dirLen := dir.Length()
	dir = dir.Div(dirLen)

	if dirLen > halfWidth {
		// The common case where we have enough room to simply
		// move the point back without crashing into an existing
		// point
		route[len(route)-1] = arrowPoint.Sub(dir.Mul(halfWidth))
	} else {
		// If we don't have enough room to move the end point back,
		// we need to fallback to the general solution of finding
		// the point to split at and using that instead.
		backOffT := halfWidth / route.Length()

		route, _ = route.SplitAt(1 - backOffT)

	}

	route = route.Simplify()

	if len(route) < 2 {
		return nil
	}

	// Helper function for adding points to the path
	addPoint := func(prevIdx, curIdx, nextIdx int) {
		curPoint := route[curIdx]

		if prevIdx < 0 || prevIdx >= len(route) {
			// curPoint is the first point in the path
			nextPoint := route[nextIdx]
			dir := nextPoint.Sub(curPoint).Normalized()

			curPoint = curPoint.Add(dir.Norm().Mul(halfWidth))
			// LineTo works here because the first LineTo is actually
			// a MoveTo
			path.LineTo(curPoint)
		} else if nextIdx < 0 || nextIdx >= len(route) {
			// curPoint is the last point in the path
			prevPoint := route[prevIdx]
			dir := curPoint.Sub(prevPoint).Normalized()

			curPoint = curPoint.Add(dir.Norm().Mul(halfWidth))
			path.LineTo(curPoint)
		} else {
			// curPoint is in the middle of the path
			prevPoint := route[prevIdx]
			nextPoint := route[nextIdx]

			prevDir := curPoint.Sub(prevPoint).Normalized()
			nextDir := nextPoint.Sub(curPoint).Normalized()

			// Unless the neighbour points are the ends of the
			// route, we need to ensure that the path doesn't
			// double-back on itself. We do this by taking the mid
			// points of the neighbours and curPoint as the maximum
			// extents of the corner.
			if prevIdx > 0 && prevIdx < len(route)-1 {
				prevPoint = prevPoint.Add(curPoint).Div(2)
			}

			if nextIdx > 0 && nextIdx < len(route)-1 {
				nextPoint = curPoint.Add(nextPoint).Div(2)
			}

			prevNorm := prevDir.Norm()
			nextNorm := nextDir.Norm()

			cornerStart := prevPoint.Add(prevNorm.Mul(halfWidth))
			cornerEnd := nextPoint.Add(nextNorm.Mul(halfWidth))

			offsetVec := prevNorm.Add(nextNorm).Normalized()
			cornerOffset := halfWidth / offsetVec.Dot(prevNorm)

			cornerPeak := curPoint.Add(offsetVec.Mul(cornerOffset))

			r := radius
			cornerNorm := cornerEnd.Sub(cornerStart).Norm()
			if cornerNorm.Dot(cornerPeak.Sub(cornerStart)) > 0 {
				r += halfWidth
			} else {
				r -= halfWidth
			}

			path.RoundCorner(r, cornerStart, cornerPeak, cornerEnd)
		}
	}

	// Go around one side of the arrow
	for i := 0; i < len(route); i++ {
		addPoint(i-1, i, i+1)
	}

	// Draw a line to the point of the arrow
	path.LineTo(arrowPoint)

	// Draw the other size of the arrow
	for i := len(route) - 1; i >= 0; i-- {
		addPoint(i+1, i, i-1)
	}

	// Finish
	return path.ClosePath()
}

// Find an appropriate split point along route starting from startPos and
// return the split lines (with the second one reversed).
//
// findSplit will avoid a split point closer than splitTolerance from a
// corner.
func findSplit(route vec.Polyline, startPos float32, splitTolerance float32) (vec.Polyline, vec.Polyline) {
	route = route.Simplify()

	route1, route2 := route.SplitAt(startPos)

	// Check if the split point is itself a corner
	splitP := route2[0]
	for _, p := range route {
		if p == splitP {
			// The split point is a corner, move the split point a tiny amount
			// and split again
			route1, route2 = route.SplitAt(startPos + 0.005)
			break
		}
	}

	route1 = route1.Simplify()
	route2 = route2.Simplify()

	seg1Length := route1[len(route1)-1].Sub(route1[len(route1)-2]).Length()
	seg2Length := route2[0].Sub(route2[1]).Length()

	didAdjust := false
	if seg1Length < splitTolerance {
		adjustment := (splitTolerance - seg1Length) / route1.Length()
		newPos := startPos + adjustment
		if newPos < 1 && newPos > 0 {
			route1, route2 = route.SplitAt(newPos)
			didAdjust = true
		}
	}
	if !didAdjust && seg2Length < splitTolerance {
		adjustment := (splitTolerance - seg2Length) / route2.Length()
		newPos := startPos - adjustment
		if newPos < 1 && newPos > 0 {
			route1, route2 = route.SplitAt(newPos)
			didAdjust = true
		}
	}

	if didAdjust {
		route1 = route1.Simplify()
		route2 = route2.Simplify()
	}

	slices.Reverse(route2)
	return route1, route2
}

func (s *LabelStyle) UnmarshalJSON(data []byte) error {
	return canvas.UnmarshalColorStruct(data, s)
}
