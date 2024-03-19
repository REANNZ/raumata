package internal

import "github.com/REANNZ/raumata/vec"

// A simple abstraction of an infinite(ish) using a map
// to store values
type Grid[T any] map[GridPos]T

// Type representing positions in a grid
type GridPos struct {
	X, Y int16
}

// Returns a [vec.Vec2] with the same values as the
// grid position
func (g GridPos) ToVec() vec.Vec2 {
	return vec.Vec2{
		X: float32(g.X),
		Y: float32(g.Y),
	}
}

func (g GridPos) Min(p GridPos) GridPos {
	minPos := GridPos{}
	if g.X < p.X {
		minPos.X = g.X
	} else {
		minPos.X = p.X
	}
	if g.Y < p.Y {
		minPos.Y = g.Y
	} else {
		minPos.Y = p.Y
	}

	return minPos
}

func (g GridPos) Max(p GridPos) GridPos {
	maxPos := GridPos{}
	if g.X > p.X {
		maxPos.X = g.X
	} else {
		maxPos.X = p.X
	}
	if g.Y > p.Y {
		maxPos.Y = g.Y
	} else {
		maxPos.Y = p.Y
	}

	return maxPos
}

// Returns the Chebyshev distance between two points
//
//	d = max(abs(a.X-b.X), abs(a.Y-b.Y))
func (a GridPos) ChebyshevDistance(b GridPos) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y

	if dx < 0 {
		dx *= -1
	}
	if dy < 0 {
		dy *= -1
	}

	if dx > dy {
		return float32(dx)
	} else {
		return float32(dy)
	}
}
