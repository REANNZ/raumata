package raumata

import (
	"fmt"
	"os"
	"slices"

	"github.com/REANNZ/raumata/internal"
	"github.com/REANNZ/raumata/internal/f32"
	"github.com/REANNZ/raumata/vec"
)

const (
	// Cap on the number of iterations the search algorithm does
	searchLimit = 8192
	// Cap on the number of iterations the fix-point pass does
	routeIterLimit = 32
	// The weight to apply to the link-crossing penalty.
	// The higher this number, the further a route will go
	// out of it's way to avoid crossing.
	linkPenaltyWeight = 10.0
)

// LinkRouter routes links through a grid.
// The zero value is not usable.
type LinkRouter struct {
	// Avoid other nodes when routing (default true)
	AvoidNodes        bool
	// Attach to multi-cell nodes in cardinal directions (default true)
	AttachMultiCellsCardinal bool
	// Encourage links to space themselves out (default true)
	SpreadLinks       bool
	Orthogonal        bool
	topo              *Topology
	nodes             internal.Grid[NodeId]
	nodeLabels        internal.Grid[bool]
	linkMap           internal.Grid[[]LinkId]
	extentMin         internal.GridPos
	extentMax         internal.GridPos
	linkPenaltyWeight float32
}

func NewLinkRouter(topo *Topology) *LinkRouter {
	router := &LinkRouter{
		AvoidNodes:        true,
		AttachMultiCellsCardinal: true,
		SpreadLinks:       true,
		topo:              topo,
		nodes:             internal.Grid[NodeId]{},
		nodeLabels:        map[internal.GridPos]bool{},
		linkMap:           map[internal.GridPos][]LinkId{},
		linkPenaltyWeight: linkPenaltyWeight,
	}

	setExtents := false
	// Add all the nodes
	for _, node := range topo.Nodes {
		if node != nil && node.Pos != nil {
			pos := internal.GridPos{
				X: node.Pos[0],
				Y: node.Pos[1],
			}

			if !setExtents {
				router.extentMin = pos
				router.extentMax = pos
				setExtents = true
			} else {
				router.extentMin = router.extentMin.Min(pos)
				router.extentMax = router.extentMax.Max(pos)
			}

			router.nodes[pos] = node.Id
			if node.IsMultiCell() {
				w := node.Extents.Width
				h := node.Extents.Height

				if w > 0 && h > 0 {
					minVec, maxVec := node.GetExtents()

					minX := int16(f32.Ceil(minVec.X))
					minY := int16(f32.Ceil(minVec.Y))
					maxX := int16(f32.Ceil(maxVec.X))
					maxY := int16(f32.Ceil(maxVec.Y))

					for x := minX; x < maxX; x++ {
						for y := minY; y < maxY; y++ {
							p := internal.GridPos{
								X: x,
								Y: y,
							}

							router.nodes[p] = node.Id
						}
					}

					router.extentMin = router.extentMin.Min(internal.GridPos{
						X: minX,
						Y: minY,
					})
					router.extentMax = router.extentMax.Max(internal.GridPos{
						X: maxX,
						Y: maxY,
					})
				}
			}

			labelAt := pos
			switch node.LabelAt {
			case "n":
				labelAt.Y -= 1
			case "ne":
				labelAt.X += 1
				labelAt.Y -= 1
			case "e":
				labelAt.X += 1
			case "se":
				labelAt.X += 1
				labelAt.Y += 1
			case "s":
				labelAt.Y += 1
			case "sw":
				labelAt.X -= 1
				labelAt.Y += 1
			case "w":
				labelAt.X -= 1
			case "nw":
				labelAt.X -= 1
				labelAt.Y -= 1
			}

			if labelAt != pos {
				router.nodeLabels[labelAt] = true

				router.extentMin = router.extentMin.Min(labelAt)
				router.extentMax = router.extentMax.Max(labelAt)
			}
		}
	}

	// Add the links at the start, end and via points
	for id, link := range topo.Links {
		if link == nil {
			continue
		}

		// If the link already has a route, add it
		if len(link.Route) > 0 {
			router.addRoute(id, link.Route)
			continue
		}

		// Adding link at the via points helps to nudge
		// routes away from those locations during initial
		// routing
		for _, via := range link.Via {
			pos := internal.GridPos{
				X: via[0],
				Y: via[1],
			}

			router.addLink(pos, id)
		}

		from := topo.GetNode(link.From)
		if from != nil && from.Pos != nil {
			pos := internal.GridPos{
				X: from.Pos[0],
				Y: from.Pos[1],
			}

			router.addLink(pos, id)
		}

		to := topo.GetNode(link.To)
		if to != nil && to.Pos != nil {
			pos := internal.GridPos{
				X: to.Pos[0],
				Y: to.Pos[1],
			}

			router.addLink(pos, id)
		}
	}

	return router
}

// Set the minimum and maximum extents of the grid
//
// These are otherwise determined by the positions of nodes and
// vias in the topology.
//
// Setting the extents such that nodes lie outside the grid will
// cause links to fail to route
func (r *LinkRouter) SetExtents(minX, minY, maxX, maxY int) {
	min := internal.GridPos{
		X: int16(minX),
		Y: int16(minY),
	}
	max := internal.GridPos{
		X: int16(maxX),
		Y: int16(maxY),
	}
	r.extentMin = min.Min(max)
	r.extentMax = min.Max(max)
}

func (r *LinkRouter) GetExtents() (min, max vec.Vec2) {
	return r.extentMin.ToVec(), r.extentMax.ToVec()
}

// Route all the links in the topology and update the
// links.
func (r *LinkRouter) RouteLinks() {
	routes := []*route{}
	links := r.topo.Links

	// Routing the links happens in three passes.
	//
	// First, all the links are routed independently, that
	// is, each path is routed without // considering other
	// paths (and thus without considering crossings).
	//
	// Then each route is recorded and links re-routed with
	// knowledge of the other routes. The recorded route is
	// moved as it is re-routed to avoid extra collisions due
	// to multiple routes re-routing in the same direction.
	//
	// Finally, the links are iteratively re-routed and updated
	// whenever a lower-cost route is found until no paths are
	// updated. This fixes up any lingering issues from the
	// previous pass where re-routing a later link allows a better
	// path for an earlier link.

	// Find the initial routes
	for id, link := range links {
		if len(link.Route) > 0 {
			// Don't re-route links that have already been routed
			continue
		}
		route := r.routeLink(id)
		if route != nil {
			routes = append(routes, route)
			link.Route = route.path
		}
	}

	// Add the links to the grid cells
	for _, route := range routes {
		r.addRoute(route.id, route.path)
	}

	// Sort the routes by their weight. Since the results of the
	// next pass is dependent on the order we route the links,
	// sorting them makes the output consistent between invocations.
	slices.SortStableFunc(routes, func(a, b *route) int {
		d := a.weight - b.weight
		if d < 0 {
			return -1
		} else if d > 0 {
			return 1
		} else {
			return 0
		}
	})

	newRoutes := []*route{}
	for _, initRoute := range routes {
		route := r.routeLink(initRoute.id)
		if route != nil {
			r.moveRoute(route.id, initRoute.path, route.path)

			// Set the route on the link
			link := r.topo.GetLink(route.id)
			if link != nil {
				link.Route = route.path
				newRoutes = append(newRoutes, route)
			}
		}
	}

	// Sort again, this favours improving short links
	// over long ones, which works because short links
	// tend to have less flexibility in possible routes
	slices.SortStableFunc(newRoutes, func(a, b *route) int {
		aWeightRatio := float32(a.path.Length()) / float32(a.weight)
		bWeightRatio := float32(b.path.Length()) / float32(b.weight)
		d := aWeightRatio - bWeightRatio
		if d < 0 {
			return -1
		} else if d > 0 {
			return 1
		} else {
			return 0
		}
	})

	// Iterate until a fix-point or we reach the iteration limit.
	// In practise this loop only tends to run once or twice.
	for i := 0; i < routeIterLimit; i++ {
		updated := false
		for i, rt := range newRoutes {
			route := r.routeLink(rt.id)
			if route != nil {
				if route.weight < rt.weight {
					link := r.topo.GetLink(route.id)
					if link != nil {
						r.moveRoute(route.id, rt.path, route.path)
						link.Route = route.path
						newRoutes[i] = route
						updated = true
					}
				}
			}
		}

		if !updated {
			break
		}
	}
}

func (r *LinkRouter) addLink(pos internal.GridPos, id LinkId) {
	curLinks := r.linkMap[pos]
	// Check that it's not already in the list
	for _, lid := range curLinks {
		if lid == id {
			return
		}
	}
	curLinks = append(curLinks, id)
	r.linkMap[pos] = curLinks

	r.extentMin = r.extentMin.Min(pos)
	r.extentMax = r.extentMax.Max(pos)
}

func (r *LinkRouter) removeLink(pos internal.GridPos, id LinkId) {
	curLinks, ok := r.linkMap[pos]
	if !ok {
		return
	}
	newList := make([]LinkId, 0, len(curLinks))
	for _, lid := range curLinks {
		if lid != id {
			newList = append(newList, lid)
		}
	}
	if len(newList) > 0 {
		r.linkMap[pos] = newList
	} else {
		delete(r.linkMap, pos)
	}
}

func (r *LinkRouter) addRoute(id LinkId, path vec.Polyline) {
	for _, point := range path {
		pos := internal.GridPos{
			X: int16(point.X),
			Y: int16(point.Y),
		}

		r.addLink(pos, id)
	}
}

func (r *LinkRouter) removeRoute(id LinkId, path vec.Polyline) {
	for _, point := range path {
		pos := internal.GridPos{
			X: int16(point.X),
			Y: int16(point.Y),
		}

		r.removeLink(pos, id)
	}
}

func (r *LinkRouter) moveRoute(id LinkId, oldPath, newPath vec.Polyline) {
	r.removeRoute(id, oldPath)
	r.addRoute(id, newPath)
}

func (r *LinkRouter) routeLink(id LinkId) *route {
	link := r.topo.GetLink(id)
	if link == nil {
		return nil
	}

	start := r.topo.GetNode(link.From)
	if start == nil || start.Pos == nil {
		return nil
	}
	goal := r.topo.GetNode(link.To)
	if goal == nil || goal.Pos == nil {
		return nil
	}

	startNode := link.From
	goalNode := link.To

	finder := routeFinder{
		startNode: startNode,
		goalNode:  goalNode,
		goalIsMulti: goal.IsMultiCell(),
		linkId:    id,
		router:    r,
	}

	vias := make([]internal.GridPos, len(link.Via))

	for i, via := range link.Via {
		vias[i] = internal.GridPos{
			X: via[0],
			Y: via[1],
		}

	}

	// Use a set of start positions instead of a single
	// position to allow for multi-cell to multi-cell links
	startPositions := make([]internal.GridPos, 0, 1)
	if start.IsMultiCell() {
		node := r.topo.GetNode(startNode)
		minExtent, maxExtent := node.GetExtents()

		minX := int16(f32.Ceil(minExtent.X))
		minY := int16(f32.Ceil(minExtent.Y))
		maxX := int16(f32.Ceil(maxExtent.X))
		maxY := int16(f32.Ceil(maxExtent.Y))

		for x := minX; x < maxX; x += 1 {
			pos := internal.GridPos{
				X: x,
				Y: minY,
			}
			startPositions = append(startPositions, pos)

			pos.Y = maxY - 1
			startPositions = append(startPositions, pos)
		}

		for y := minY + 1; y < maxY; y += 1 {
			pos := internal.GridPos{
				X: minX,
				Y: y,
			}
			startPositions = append(startPositions, pos)

			pos.X = maxX - 1
			startPositions = append(startPositions, pos)
		}
	} else {
		startPositions = append(startPositions, internal.GridPos{
			X: start.Pos[0],
			Y: start.Pos[1],
		})
	}

	goalPos := internal.GridPos{
		X: goal.Pos[0],
		Y: goal.Pos[1],
	}

	route := finder.run(startPositions, goalPos, vias)
	return route
}

type route struct {
	id     LinkId
	path   vec.Polyline
	weight float32
}

// Useful for debugging
func (r *route) dump() {
	if r == nil {
		fmt.Fprintf(os.Stderr, "nil\n")
		return
	}
	if len(r.path) > 0 {
		for i, p := range r.path {
			if i == 0 {
				fmt.Fprintf(os.Stderr, "[%d] %s", i, p)
			} else {
				fmt.Fprintf(os.Stderr, " -> [%d] %s", i, p)
			}
		}
		fmt.Fprintf(os.Stderr, " %f\n", r.weight)
	}
}

type routeFinder struct {
	startNode, goalNode NodeId
	goal                gridNode
	goalIsMulti         bool
	vias                []internal.GridPos
	linkId              LinkId
	router              *LinkRouter
	cameFrom            map[gridNode]gridNode
}

// Represents a node in the implicit graph we are traversing
type gridNode struct {
	gridPos    internal.GridPos // The grid positions
	dirX, dirY int16            // The current direction
	via        int              // Which via point we need to head to next
}

// This is the start of the route finding algorithm.
//
// The algorithm works by finding a path through an implicit graph defined
// by [gridNode] above using A* search.
//
// In order to find a good path that also runs through the via points,
// the grid is essentially duplicated for each via point, connected at the
// via position. The start node is then placed on the highest grid and
// the goal node placed on the lowest grid, forcing the path to traverse
// the via points by construction.
func (f *routeFinder) run(startPositions []internal.GridPos, goal internal.GridPos, vias []internal.GridPos) *route {
	f.goal = gridNode{gridPos: goal, via: 0}
	f.vias = vias

	if len(startPositions) == 0 {
		return nil
	}

	start := startPositions[0]

	// Used to estimate the initial size of the datastructures used
	// in path finding
	minDist := int(start.ChebyshevDistance(goal))

	// Create datastructures for path finding
	f.cameFrom = make(map[gridNode]gridNode, minDist*2)
	openSet := internal.PriorityQueue[gridNode]{}
	weights := make(map[gridNode]float32, minDist*2)

	// Add all the start nodes to the initial open set
	// This has a similar effect to starting with a virtual node
	// that has the start positions as zero-distance neighbours
	for _, pos := range startPositions {
		node := gridNode{
			gridPos: pos,
			via: len(vias),
		}

		openSet.Push(node, 0)
		weights[node] = 0
	}

	iterNum := 0
	for !openSet.Empty() && iterNum < searchLimit {

		curP, _ := openSet.Pop()
		current := *curP

		curWeight := weights[current]

		currentId, _ := f.router.nodes[current.gridPos]
		// We've reached the destination. Due to the way the graph is defined,
		// we have to ignore the direction values, which means there are up to
		// 8 valid goal nodes (one for each approaching direction), fortunately
		// the algorithm will find the closest one anyway.
		if current.via == f.goal.via && (current.gridPos == f.goal.gridPos || currentId == f.goalNode) {
			return f.buildRoute(current, curWeight)
		}

		f.neighbours(current, func(n gridNode) {
			newWeight := curWeight + f.weight(current, n)

			neighbourWeight, ok := weights[n]

			if !ok || newWeight < neighbourWeight {
				f.cameFrom[n] = current
				weights[n] = newWeight

				// The distance by itself is an admissable/consistent heuristic.
				// Adding the "via distance" causes the algorithm to favour exploring
				// paths that have already been through a via point at the cost of
				// potentially finding sub-optimal routes.
				h := f.goalDistance(n) + float32(n.via)

				// Multiply the priority by 100 to keep some of the precision from the
				// weight calculation
				priority := int((newWeight + h) * 100)

				openSet.Push(n, priority)
			}
		})

		iterNum += 1
	}

	return nil
}

func (f *routeFinder) buildRoute(pos gridNode, weight float32) *route {
	path := []internal.GridPos{pos.gridPos}

	c, ok := f.cameFrom[pos]
	if !ok {
		return nil
	}

	// Limit the number of iterations the route reconstruction
	// can do to avoid infinite loops
	maxIter := len(f.cameFrom) + 1
	i := 0
	for i < maxIter && ok {
		path = append(path, c.gridPos)
		prev := c
		c, ok = f.cameFrom[c]
		if ok && c == prev {
			// This is very simplistic loop detection
			panic(fmt.Errorf("Loop in path! (%d, %d)", c.gridPos.X, c.gridPos.Y))
		}

		i += 1
	}

	// If ok == true, then we didn't reach the end of the route
	if ok {
		panic("buildRoute could not build route!")
	}

	// Reverse the path of grid positions and turn it into
	// a vec.Polyline
	line := vec.Polyline(make([]vec.Vec2, 0, len(path)))
	for i := len(path) - 1; i >= 0; i-- {
		line = append(line, path[i].ToVec())
	}

	// Remove duplicated values
	line = line.Fix()

	return &route{
		id:     f.linkId,
		path:   line,
		weight: weight,
	}
}

func (f *routeFinder) getVia(n int) (internal.GridPos, bool) {
	if n == 0 || n > len(f.vias) {
		return internal.GridPos{}, false
	} else {
		return f.vias[len(f.vias)-n], true
	}
}

// Produces the set of neighbours of the given node
func (f *routeFinder) neighbours(pos gridNode, fn func(gridNode)) {
	extMin := f.router.extentMin
	extMax := f.router.extentMax

	// Helper function to prune the graph a little
	produce := func(g gridNode) {
		// the current node isn't it's own neighbour
		if g == pos {
			return
		}
		// don't consider the node we just came from
		prev, ok := f.cameFrom[pos]
		if ok && prev == g {
			return
		}

		via, ok := f.getVia(pos.via)
		if ok && g.gridPos == via {
			g.via -= 1
		}

		nodeId := f.router.nodes[g.gridPos]
		if g.gridPos == f.goal.gridPos || nodeId == f.goalNode {
			if f.goalIsMulti && f.router.AttachMultiCellsCardinal {
				if g.dirX == 0 || g.dirY == 0 {
					fn(g)
				}
			} else {
				fn(g)
			}
		} else {
			// Check that neighbour is in-bounds
			gridPos := g.gridPos
			inBounds := gridPos.X >= extMin.X && gridPos.X <= extMax.X &&
				gridPos.Y >= extMin.Y && gridPos.Y <= extMax.Y

			// Skip over neighbours that have nodes in them
			// (The target node is handled by the check above)
			_, isNode := f.router.nodes[gridPos]

			isNode = f.router.AvoidNodes && isNode

			// Skip over neighbours that have node labels in them
			_, isLabel := f.router.nodeLabels[gridPos]

			if inBounds && !isNode && !isLabel {
				fn(g)
			}
		}
	}

	// Produce the next grid pos in the current direction
	if pos.dirX != 0 || pos.dirY != 0 {
		// TODO: implement some basic jump point search techniques
		// to make searching straight-line paths faster.
		// https://en.wikipedia.org/wiki/Jump_point_search
		n := pos
		n.gridPos.X += pos.dirX
		n.gridPos.Y += pos.dirY

		produce(n)
	} else {
		// Handle the special case where dirX == 0 and dirY == 0
		// Produce the 8 neighbours directly

		// Produce cardinal directions first in order to
		// create a slight preference for paths that leave the node
		// in a cardinal direction.
		// This produces better results when routing to multi-cell
		// nodes.
		for dx := int16(-1); dx <= 1; dx++ {
			for dy := int16(-1); dy <= 1; dy++ {
				// Skip null direction
				if dx == 0 && dy == 0 {
					continue
				}
				// Skip diagonal directions
				if dx != 0 && dy != 0 {
					continue
				}
				n := pos
				n.dirX = dx
				n.dirY = dy
				n.gridPos.X = pos.gridPos.X + dx
				n.gridPos.Y = pos.gridPos.Y + dy
				produce(n)
			}
		}

		if !f.router.Orthogonal {
			// Now produce the diagonals
			for dx := int16(-1); dx <= 1; dx++ {
				for dy := int16(-1); dy <= 1; dy++ {
					// Skip cardinal directions
					// (also skips null direction)
					if dx == 0 || dy == 0 {
						continue
					}
					n := pos
					n.dirX = dx
					n.dirY = dy
					n.gridPos.X = pos.gridPos.X + dx
					n.gridPos.Y = pos.gridPos.Y + dy
					produce(n)
				}
			}
		}
		return
	}

	if f.router.Orthogonal {
		if pos.dirX == 0 {
			n := pos
			n.dirY = 0
			n.dirX = pos.dirY
			produce(n)
			n.dirX = -pos.dirY
			produce(n)
		} else {
			n := pos
			n.dirX = 0
			n.dirY = pos.dirX
			produce(n)
			n.dirY = -pos.dirX
			produce(n)
		}
	} else {
		// Produce the two 45deg turns from the current direction

		if pos.dirX == 0 {
			n := pos
			n.dirX = 1
			produce(n)
			n.dirX = -1
			produce(n)
		} else if pos.dirY != 0 {
			n := pos
			n.dirX = 0
			produce(n)
		}

		if pos.dirY == 0 {
			n := pos
			n.dirY = 1
			produce(n)
			n.dirY = -1
			produce(n)
		} else if pos.dirX != 0 {
			n := pos
			n.dirY = 0
			produce(n)
		}
	}
}

// Calculate the weight of the edge from `fromNode` to `toNode`.
func (f *routeFinder) weight(fromNode, toNode gridNode) float32 {
	from := fromNode.gridPos
	to := toNode.gridPos

	toNodeId := f.router.nodes[to]

	// This currently always returns 1, but if JPS is implemented,
	// the nodes won't be adjacent cells
	dist := from.ChebyshevDistance(to)
	var linkPenalty float32 = 0

	// If the grid positions are the same, it's a turn
	if from == to {
		// Penalize turns more than single steps
		dist = 2
		cur := fromNode
		prevNode, ok := f.cameFrom[cur]
		// If the previous step was also a turn, then
		// increase the penalty, this encourages two 45deg turns
		// spaced apart (a total weight of 4) over a single 90deg turn
		// (a total weight of 6)
		if ok && prevNode.gridPos == cur.gridPos {
			dist = 4
		}
	} else if to != f.goal.gridPos && toNodeId != f.goalNode {
		// Add a penalty to cells that contain links, this is
		// primarily to avoid having multiple paths take the
		// same route when other optimal paths exist.
		links := f.router.linkMap[to]
		var n float32 = 1
		for _, l := range links {
			if l != f.linkId {
				// Apply a penalty for each link, but make
				// the penalty smaller for each successive link.
				linkPenalty += 1 / n
				n *= 2
			}
		}

		// Handle diagonal crossings:
		//
		// +--+--+--+--+
		// |a |  |  |b |
		// +--+--+--+--+
		// |  |a0|x |  |
		// +--+--+--+--+
		// |  |* |a1|  |
		// +--+--+--+--+
		// |  |  |  |a |
		// +--+--+--+--+
		//
		// If we're on `x` and evaluating the move to `*`
		// then move crosses `a` even though the destination
		// node is empty. We need to check a0 and a1 in that
		// case for links that are both locations.
		if fromNode.dirX != 0 && fromNode.dirY != 0 {
			n1 := from
			n1.X += fromNode.dirX
			n2 := from
			n2.Y += fromNode.dirY

			links1 := f.router.linkMap[n1]
			links2 := f.router.linkMap[n2]

			// Get all the links that are in both of the two relevant positions
			linksIntersection := []LinkId{}
			for _, l1 := range links1 {
				if l1 == f.linkId {
					continue
				}
				for _, l2 := range links2 {
					if l1 == l2 {
						linksIntersection = append(linksIntersection, l1)
						break
					}
				}
			}

			for _, l := range linksIntersection {
				if l != f.linkId {
					linkPenalty += 1 / n
					n *= 2
				}
			}
		}

		// Apply a penalty for being adjacent to other links,
		// this is to try and spread out links radially at the
		// start and end nodes since otherwise they can bunch
		// up weird ways.
		addPenalty := func(at internal.GridPos) {
			if !f.router.SpreadLinks {
				return
			}
			links := f.router.linkMap[at]
			// Start the penalty fairly low, since we really
			// just want to pick between otherwise-equal paths
			var n float32 = 16
			for _, l := range links {
				if l != f.linkId {
					linkPenalty += 1 / n
					n *= 2
				}
			}
		}

		// Check to the left and right of the "to" node,
		// (relative to the direction the route is heading)
		// for neighbours.
		if toNode.dirX == 0 {
			n := to
			n.Y += toNode.dirY
			n.X += 1
			addPenalty(n)
			n = to
			n.Y += toNode.dirY
			n.X -= 1
			addPenalty(n)
		} else if toNode.dirY != 0 {
			n := to
			n.Y += toNode.dirY
			addPenalty(n)
		}
		if toNode.dirY == 0 {
			n := to
			n.X += toNode.dirX
			n.Y += 1
			addPenalty(n)
			n = to
			n.X += toNode.dirX
			n.Y -= 1
			addPenalty(n)
		} else if toNode.dirX != 0 {
			n := to
			n.X += toNode.dirX
			addPenalty(n)
		}
	}

	weight := dist + (linkPenalty * f.router.linkPenaltyWeight)

	return weight
}

func (f *routeFinder) goalDistance(fromNode gridNode) float32 {
	from := fromNode.gridPos

	if f.goalIsMulti {
		goalNode := f.router.topo.GetNode(f.goalNode)

		minVec, maxVec := goalNode.GetExtents()

		minX := int16(f32.Ceil(minVec.X))
		minY := int16(f32.Ceil(minVec.Y))
		maxX := int16(f32.Ceil(maxVec.X))
		maxY := int16(f32.Ceil(maxVec.Y))

		dist := float32(-1)

		for x := minX; x < maxX; x++ {
			for y := minY; y < maxY; y++ {
				pos := internal.GridPos{
					X: x,
					Y: y,
				}

				d := from.ChebyshevDistance(pos)

				if dist < 0 || d < dist {
					dist = d
				}
			}
		}

		return dist
	} else {
		return from.ChebyshevDistance(f.goal.gridPos)
	}
}
