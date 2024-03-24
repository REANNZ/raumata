package raumata_test

import (
	"os"

	"github.com/REANNZ/raumata"
	"github.com/REANNZ/raumata/canvas"
)

// This is a bare-minimum example of usage
func Example() {
	// This would normally come from an outside source,
	// e.g. a JSON file or database
	networkTopology := &raumata.Topology{
		Nodes: map[raumata.NodeId]*raumata.Node{
			"ruru": {
				Id: "ruru",
				Pos: &[2]int16{0, 0},
			},
			"kea": {
				Id:  "kea",
				Pos: &[2]int16{10, 5},
			},
			"kaka": {
				Id:  "kaka",
				Pos: &[2]int16{0, 10},
			},
		},
		Links: map[raumata.LinkId]*raumata.Link{
			"ruru-kea": {
				Id:   "ruru-kea",
				From: "ruru",
				To:   "kea",
			},
			"kea-kaka": {
				Id:   "kea-kaka",
				From: "kea",
				To:   "kaka",
			},
			"kaka-ruru": {
				Id:   "kaka-ruru",
				From: "kaka",
				To:   "ruru",
			},
		},
	}

	router := raumata.NewLinkRouter(networkTopology)
	router.RouteLinks()

	raumata.PlaceLabels(networkTopology)

	renderer := raumata.NewRenderer()

	drawCanvas := canvas.NewCanvas()

	renderer.RenderTopologyToCanvas(networkTopology, drawCanvas)

	svgRenderer := canvas.NewSVGRenderer(os.Stdout)
	drawCanvas.Render(svgRenderer)
}
