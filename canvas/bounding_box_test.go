package canvas_test

import (
	"math"
	"testing"

	. "github.com/REANNZ/raumata/canvas"
	"github.com/REANNZ/raumata/vec"
)

func checkVec(t *testing.T, actual, expected vec.Vec2) {
	t.Helper()

	if !expected.ApproxEq(actual, 1e-6) {
		t.Errorf("Vec mismatch, expected %s, got %s", expected, actual)
	}
}

func TestAABB(t *testing.T) {
	p0 := vec.Vec2{X: 0, Y: 0}
	p1 := vec.Vec2{X: 1, Y: 1}

	aabb := NewAABB(p0, p1)

	min, max := aabb.Bounds()

	checkVec(t, min, p0)
	checkVec(t, max, p1)

	size := aabb.Size()
	checkVec(t, size, vec.Vec2{X: 1, Y: 1})

	aabb = NewAABB(p1, p0)
	min, max = aabb.Bounds()

	checkVec(t, min, p0)
	checkVec(t, max, p1)

	size = aabb.Size()
	checkVec(t, size, vec.Vec2{X: 1, Y: 1})
}

func TestAABBUnion(t *testing.T) {
	p0 := vec.Vec2{X: -1, Y: -1}
	p1 := vec.Vec2{X: 5, Y: 5}
	p2 := vec.Vec2{X: 2, Y: 2}
	p3 := vec.Vec2{X: 9, Y: 9}

	aabb0 := NewAABB(p0, p1)
	aabb1 := NewAABB(p2, p3)

	aabbUnion := aabb0.Union(aabb1)

	min, max := aabbUnion.Bounds()
	checkVec(t, min, p0)
	checkVec(t, max, p3)

	size := aabbUnion.Size()
	checkVec(t, size, vec.Vec2{X: 10, Y: 10})
}

func TestAABBTransform(t *testing.T) {
	p0 := vec.Vec2{X: 0, Y: 0}
	p1 := vec.Vec2{X: 5, Y: 5}

	aabb := NewAABB(p0, p1)

	translate := vec.NewTranslate(vec.Vec2{X: 1, Y: 1})

	aabb1 := aabb.Transform(translate)

	checkVec(t, aabb.Size(), aabb1.Size())

	min, max := aabb1.Bounds()
	checkVec(t, min, vec.Vec2{X: 1, Y: 1})
	checkVec(t, max, vec.Vec2{X: 6, Y: 6})

	// Note that this a rotation around the origin
	rotate := vec.NewRotate(math.Pi / 2)
	aabb2 := aabb.Transform(rotate)

	min, max = aabb2.Bounds()
	checkVec(t, min, vec.Vec2{X: -5, Y: 0})
	checkVec(t, max, vec.Vec2{X: 0, Y: 5})

	// This is a 45deg rotation, which would produce the wrong
	// answers if we didn't take all 4 corners of the box into consideration
	rotate = vec.NewRotate(math.Pi / 4)
	aabb3 := aabb.Transform(rotate)

	min, max = aabb3.Bounds()
	// The corners are (-5 / sqrt(2), 0) and (5 / sqrt(2), sqrt(50)
	// sqrt(50) == 10 / sqrt(2)
	checkVec(t, min, vec.Vec2{X: -5.0 / math.Sqrt2, Y: 0})
	checkVec(t, max, vec.Vec2{X: 5.0 / math.Sqrt2, Y: 10.0 / math.Sqrt2})
}
