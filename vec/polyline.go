package vec

import "github.com/REANNZ/raumata/internal/f32"

// Polyline is a list of points `{x1, x2, ..., xn}`
// that represents a series of lines:
//
//	{ {x1, x2}, {x2, x3}, ..., {xn-1, xn} }
//
// A polyline with less than 2 points is treated as
// a degenerate case
type Polyline []Vec2

// Returns the result of adding x to all points in pl
func (pl Polyline) Add(x Vec2) Polyline {
	newLine := make([]Vec2, len(pl))

	for i := range pl {
		newLine[i] = pl[i].Add(x)
	}

	return newLine
}

// Returns the result of multipling all the points
// in pl by x
func (pl Polyline) Mul(x float32) Polyline {
	newLine := make([]Vec2, len(pl))

	for i := range pl {
		newLine[i] = pl[i].Mul(x)
	}

	return newLine
}

// Returns the total length of the polyline
//
// Uses the Euclidean Metric L = sqrt(x^2 + y^2)
func (pl Polyline) Length() float32 {
	if len(pl) <= 1 {
		return 0
	}

	// Calculate the total length using pairwise summation
	// to reduce round-off error

	lengths := make([]float32, len(pl)-1)

	for i := 0; i < len(pl)-1; i++ {
		p1 := pl[i]
		p2 := pl[i+1]

		lengths[i] = p2.Sub(p1).Length()
	}

	return f32.Sum(lengths)
}

// Fix returns a new Polyline with invalid or degenerate
// lines removed
//
// Specifically, Fix removes segments with length zero and
// points that have a `NaN` component
func (pl Polyline) Fix() Polyline {
	if len(pl) == 0 {
		return pl
	}
	newLine := make([]Vec2, 0, 2)

	prevPoint := pl[0]

	for i := range pl {
		p := pl[i]
		if i == 0 || p != prevPoint {
			if f32.IsNaN(p.X) || f32.IsNaN(p.Y) {
				continue
			}

			newLine = append(newLine, p)
			prevPoint = p
		}
	}

	return newLine
}

// Threshold for colinearity for Simplify, is not 1
// to allow for the finite precision of floats
const colinearThreshold = 0.99

// Simplifies the polyline by removing intermediate points
// that are colinear
func (pl Polyline) Simplify() Polyline {
	// If we have 2 or fewer points, there's no simplification
	// to do
	if len(pl) <= 2 {
		return pl
	}

	newLine := make([]Vec2, 0, 2)

	// add the first point
	newLine = append(newLine, pl[0])

	for i := 1; i < len(pl)-1; i++ {
		prevPoint := pl[i-1]
		curPoint := pl[i]
		nextPoint := pl[i+1]

		prevDir := curPoint.Sub(prevPoint).Normalized()
		nextDir := nextPoint.Sub(curPoint).Normalized()

		// Returns a number between 1 and -1. If prevDir == nextDir,
		// then prevDir.Dot(nextDir) == 1 with it decreasing as the angle
		// between them decreases
		similarity := prevDir.Dot(nextDir)
		if similarity < colinearThreshold {
			newLine = append(newLine, curPoint)
		}
	}

	return append(newLine, pl[len(pl)-1])
}

// Subdivide returns a polyline with each segment divided into
// count parts
func (pl Polyline) Subdivide(count int) Polyline {
	if count <= 1 {
		return pl
	}

	newLine := make([]Vec2, 0, len(pl)*count)

	for i := range pl[:len(pl)-1] {
		segStart := pl[i]
		segEnd := pl[i+1]

		for j := 0; j < count; j++ {
			t := float32(j) / float32(count)

			newLine = append(newLine, segStart.Lerp(segEnd, t))
		}
	}

	newLine = append(newLine, pl[len(pl)-1])

	return newLine
}

// Interpolate returns the point t*length along the line
//
// t is expected to be in the interval [0, 1]. Values outside of
// that range are clamped.
//
// If len(pl) == 0, it returns (0, 0) (the zero value of Vec)
func (pl Polyline) Interpolate(t float32) Vec2 {
	i, j, t := pl.interpolate(t)

	if i < 0 || j < 0 {
		return Vec2{}
	}

	if i == j {
		return pl[i]
	}

	pi := pl[i]
	pj := pl[j]

	return pi.Lerp(pj, t)
}

// SplitAt returns a pair of Polylines split at a point
// t*length along the polyline.
//
// t is expected to be in the interval [0, 1]. Values outside of
// that range are clamped.
//
// if len(pl) == 0, SplitAt returns (nil, nil), otherwise
// the lines are guaranteed to have at least one point each and
// have the following property:
//
//	line1, line2 := line.SplitAt(t)
//	line1[len(line1)-1] == line2[0]
//
// The order of points in the split lines is the same as the order
// in the original line
func (pl Polyline) SplitAt(t float32) (Polyline, Polyline) {
	i, j, t := pl.interpolate(t)

	if i < 0 || j < 0 {
		return nil, nil
	}

	line1 := make([]Vec2, 0, i+1)
	line2 := make([]Vec2, 0, len(pl)-j)

	// Copy the first part into line1
	for _, p := range pl[:i+1] {
		line1 = append(line1, p)
	}

	// Add the split point if it is between points
	if i != j {
		pi := pl[i]
		pj := pl[j]
		splitP := pi.Lerp(pj, t)

		line1 = append(line1, splitP)
		line2 = append(line2, splitP)
	}

	// Copy the second part into line2
	for _, p := range pl[j:] {
		line2 = append(line2, p)
	}

	return line1, line2
}

// Generic interpolation method, returns the indexes of the two points
// to interpolate between along with a new interpolate variable for the
// line segment
func (pl Polyline) interpolate(t float32) (int, int, float32) {
	// Special cases
	if len(pl) == 0 {
		// A line with no points doesn't make much sense
		return -1, -1, t
	}
	if len(pl) == 1 || t <= 0 {
		// A line with only one point is a zero-length line, all
		// points on the line are that point
		// If t is <= 0, then the point is the first point on the line
		return 0, 0, 0
	}
	if t >= 1 {
		// If t is >= 1, then the point is the last point on the line
		idx := len(pl) - 1
		return idx, idx, 1
	}
	if len(pl) == 2 {
		// If there are only two points, then we're just interpolating
		// along a straight line
		return 0, 1, t
	}

	targetLen := pl.Length() * t

	var curLen float32 = 0

	for i := 0; i < len(pl)-1; i++ {
		segStart := pl[i]
		segEnd := pl[i+1]

		segLen := segEnd.Sub(segStart).Length()
		// Skip over zero-length segments
		if segLen == 0 {
			continue
		}
		// The dance between curLen and nextLen
		// below is required to minimize round-off
		// errors
		nextLen := curLen + segLen
		if nextLen == targetLen {
			return i + 1, i + 1, 0
		}
		if nextLen >= targetLen {
			t := (targetLen - curLen) / segLen

			return i, i + 1, t
		}
		curLen = nextLen
	}

	// We shouldn't ever get here, so return an obviously
	// bad value
	return -1, -1, 0
}
