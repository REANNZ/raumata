package raumata_test

import (
	"testing"

	. "github.com/REANNZ/raumata"
	"github.com/REANNZ/raumata/vec"
)

func TestLinkRouter1(t *testing.T) {
	topo := Topology{
		Nodes: map[NodeId]*Node{
			"A": {
				Id:      "A",
				Pos:     &[2]int16{0, 0},
				Label:   "A",
				LabelAt: "n",
			},
			"B": {
				Id:      "B",
				Pos:     &[2]int16{0, 5},
				Label:   "B",
				LabelAt: "w",
			},
			"C": {
				Id:      "C",
				Pos:     &[2]int16{0, 10},
				Label:   "C",
				LabelAt: "s",
			},
			"D": {
				Id:      "D",
				Pos:     &[2]int16{8, 5},
				Label:   "D",
				LabelAt: "s",
			},
			"E": {
				Id:      "E",
				Pos:     &[2]int16{10, 10},
				Label:   "E",
				LabelAt: "e",
			},
		},
		Links: map[LinkId]*Link{
			"A-B": {
				Id:   "A-B",
				From: "A",
				To:   "B",
			},
			"B-C": {
				Id:   "B-C",
				From: "B",
				To:   "C",
			},
			"A-D": {
				Id:   "A-D",
				From: "A",
				To:   "D",
			},
			"B-D": {
				Id:   "B-D",
				From: "B",
				To:   "D",
			},
			"C-D": {
				Id:   "C-D",
				From: "C",
				To:   "D",
			},
			"D-E": {
				Id:   "D-E",
				From: "D",
				To:   "E",
			},
		},
	}

	linkRouter := NewLinkRouter(&topo)

	linkRouter.RouteLinks()

	// Just check some simple properties
	for id, link := range topo.Links {
		if len(link.Route) == 0 {
			t.Errorf("No route for link %s", id)
		}
		if len(link.Route) == 1 {
			t.Errorf("Route for link %s only has one position", id)
		}

		from := topo.Nodes[link.From]
		to := topo.Nodes[link.To]

		fromPos := vec.Vec2{X: float32(from.Pos[0]), Y: float32(from.Pos[1])}
		toPos := vec.Vec2{X: float32(to.Pos[0]), Y: float32(to.Pos[1])}

		if link.Route[0] != fromPos {
			t.Errorf("Route for link %s does not start at 'from' node (%s != %s)",
				id, link.Route[0], fromPos)
		}
		if link.Route[len(link.Route)-1] != toPos {
			t.Errorf("Route for link %s does not end at 'to' node (%s != %s)",
				id, link.Route[len(link.Route)-1], toPos)
		}
	}
}

func TestLinkRouterMulti(t *testing.T) {
	topo := Topology{
		Nodes: map[NodeId]*Node{
			"A": {
				Id:      "A",
				Pos:     &[2]int16{0, 0},
				Label:   "A",
				LabelAt: "c",
				Extents: &NodeExtents{
					Width:  3,
					Height: 10,
				},
			},
			"B": {
				Id:      "B",
				Pos:     &[2]int16{0, 10},
				Label:   "B",
				LabelAt: "c",
				Extents: &NodeExtents{
					Width:  3,
					Height: 10,
				},
			},
		},
		Links: map[LinkId]*Link{
			"A-B-1": {
				Id:   "A-B-1",
				From: "A",
				To:   "B",
			},
			"A-B-2": {
				Id:   "A-B-2",
				From: "A",
				To:   "B",
			},
			"A-B-3": {
				Id:   "A-B-3",
				From: "A",
				To:   "B",
			},
			"A-B-4": {
				Id:   "A-B-4",
				From: "A",
				To:   "B",
			},
			"A-B-5": {
				Id:   "A-B-5",
				From: "A",
				To:   "B",
			},
		},
	}

	linkRouter := NewLinkRouter(&topo)

	linkRouter.RouteLinks()

	// Just check some simple properties
	for id, link := range topo.Links {
		if len(link.Route) == 0 {
			t.Errorf("No route for link %s", id)
		}
		if len(link.Route) == 1 {
			t.Errorf("Route for link %s only has one position", id)
		}
	}
}

func BenchmarkLinkRouter(b *testing.B) {
	topo := Topology{
		Nodes: map[NodeId]*Node{
			"A": {
				Id:  "A",
				Pos: &[2]int16{0, 0},
			},
			"B": {
				Id:  "B",
				Pos: &[2]int16{10, 10},
			},
		},
		Links: map[LinkId]*Link{
			"A-B": {
				Id:   "A-B",
				From: "A",
				To:   "B",
				Via: [][2]int16{
					{0, 2},
					{2, 2},
				},
			},
		},
	}

	linkRouter := NewLinkRouter(&topo)

	for i := 0; i < b.N; i++ {
		linkRouter.RouteLinks()
	}
}
