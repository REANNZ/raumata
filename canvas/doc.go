// The canvas package provides an API for drawing.
//
// It is designed with SVG generation in mind, so the way it describes
// most operations are derived from the way SVG describes those same
// operations.
//
// The basic usage looks like:
//
//	c := canvas.NewCanvas()
//	// Add objects to the canvas
//	r := canvas.NewSVGRenderer(os.Stdout)
//	c.Render(r)
//
// See the documentation for specific types for more information.
package canvas
