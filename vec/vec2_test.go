package vec_test

import (
	"math"
	"testing"

	"github.com/REANNZ/raumata/internal/f32"
	"github.com/REANNZ/raumata/vec"
)

func checkVec(t *testing.T, actual, expected vec.Vec2) {
	t.Helper()
	if !actual.ApproxEq(expected, 1e-12) {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestVecLength(t *testing.T) {
	checkLen := func(v vec.Vec2, expected float32) {
		t.Helper()
		actual := v.Length()

		if !f32.ApproxEq(actual, expected, 1e-12) {
			t.Errorf("Test Length of %s, expected '%f', got '%f'",
				v, expected, actual)
		}
	}

	checkLen(vec.Vec2{0, 0}, 0)
	checkLen(vec.Vec2{1, 0}, 1)
	checkLen(vec.Vec2{-1, 0}, 1)
	checkLen(vec.Vec2{0, 1}, 1)
	checkLen(vec.Vec2{0, -1}, 1)
	checkLen(vec.Vec2{1, 1}, math.Sqrt2)
	checkLen(vec.Vec2{-1, -1}, math.Sqrt2)
}

func TestVecNormalize(t *testing.T) {
	checkVecNorm := func(v vec.Vec2) {
		t.Helper()
		vNorm := v.Normalized()
		length := vNorm.Length()

		if !f32.ApproxEq(length, 1.0, 1e-12) {
			t.Errorf("%s normalized to %s, expected length of '1.0', got '%f'",
				v, vNorm, length)
		}
	}

	zeroNorm := (vec.Vec2{0, 0}).Normalized()
	if zeroNorm.X != 0 || zeroNorm.Y != 0 {
		t.Errorf("(0, 0) not normalized to (0, 0), got %s", zeroNorm)
	}

	checkVecNorm(vec.Vec2{1, 0})
	checkVecNorm(vec.Vec2{0, 1})
	checkVecNorm(vec.Vec2{1, 1})
	checkVecNorm(vec.Vec2{-1, -1})
	checkVecNorm(vec.Vec2{0.01, -0.01})
}

func TestVecArithmetic(t *testing.T) {
	z := vec.Vec2{0, 0}
	a := vec.Vec2{1, 1}
	b := vec.Vec2{0, 1}

	// Test that basic things still work
	checkVec(t, a.Add(z), a)
	checkVec(t, a.Sub(z), a)
	checkVec(t, a.Neg(), vec.Vec2{-1, -1})
	checkVec(t, z.Neg(), z)

	checkVec(t, a.Mul(0), z)
	checkVec(t, a.Div(1), a)

	checkVec(t, a.Add(b), vec.Vec2{1, 2})
	checkVec(t, a.Sub(b), vec.Vec2{1, 0})
	checkVec(t, a.Add(b.Neg()), vec.Vec2{1, 0})
	checkVec(t, a.Sub(b.Neg()), vec.Vec2{1, 2})
}

func TextVecLerp(t *testing.T) {

	a := vec.Vec2{0, 0}
	b := vec.Vec2{1, 0}

	checkVec(t, a.Lerp(b, -1), a)
	checkVec(t, a.Lerp(b, 0), a)
	checkVec(t, a.Lerp(b, 1), b)
	checkVec(t, a.Lerp(b, 2), b)

	checkVec(t, a.Lerp(b, 0.1), vec.Vec2{0.1, 0})
	checkVec(t, a.Lerp(b, 0.25), vec.Vec2{0.25, 0})
	checkVec(t, a.Lerp(b, 0.5), vec.Vec2{0.5, 0})
}
