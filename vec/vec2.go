package vec

import (
	"fmt"
	"strconv"

	"github.com/REANNZ/raumata/internal/f32"
)

// A 2D vector, can represent either a point or
// a direction
type Vec2 struct {
	X float32
	Y float32
}

// Returns the length of the vector v
func (v Vec2) Length() float32 {
	return f32.Hypot(v.X, v.Y)
}

// Returns the vector with the same direction as v
// but has a length of 1
// If v is the zero vector, this returns the zero
// vector
func (v Vec2) Normalized() Vec2 {
	len := v.Length()
	if len == 0 {
		return Vec2{}
	}
	return v.Div(len)
}

// Vector addition a + b
func (a Vec2) Add(b Vec2) Vec2 {
	return Vec2{
		X: a.X + b.X,
		Y: a.Y + b.Y,
	}
}

// Vector subtraction a - b
func (a Vec2) Sub(b Vec2) Vec2 {
	return Vec2{
		X: a.X - b.X,
		Y: a.Y - b.Y,
	}
}

// Multiplies both components of v by m
func (v Vec2) Mul(m float32) Vec2 {
	return Vec2{
		X: v.X * m,
		Y: v.Y * m,
	}
}

// Divides both components of v by d
func (v Vec2) Div(d float32) Vec2 {
	return Vec2{
		X: v.X / d,
		Y: v.Y / d,
	}
}

// Returns the dot product of a and b
//
//	a.X * b.X + a.Y*b.Y
//
// The dot product is also equivalent to
//
//	cos(t)*a.Length()*b.Length()
//
// Where t is the angle between a and b
func (a Vec2) Dot(b Vec2) float32 {
	return a.X*b.X + a.Y*b.Y
}

// Returns v * -1
func (v Vec2) Neg() Vec2 {
	return Vec2{
		X: -v.X,
		Y: -v.Y,
	}
}

// Returns the vector 90 degrees counterclockwise to v
func (v Vec2) Norm() Vec2 {
	return Vec2{
		X: -v.Y,
		Y: v.X,
	}
}

// Returns the component-wise minimum of a and b
func (a Vec2) Min(b Vec2) Vec2 {
	return Vec2{
		X: f32.Min(a.X, b.X),
		Y: f32.Min(a.Y, b.Y),
	}
}

// Returns the component-wise maximum of a and b
func (a Vec2) Max(b Vec2) Vec2 {
	return Vec2{
		X: f32.Max(a.X, b.X),
		Y: f32.Max(a.Y, b.Y),
	}
}

// Rounds each component to the nearest integer, rounding half away from zero
func (v Vec2) Round() Vec2 {
	return Vec2{
		X: f32.Round(v.X),
		Y: f32.Round(v.Y),
	}
}

// Rounds each component to the next largest integer
func (v Vec2) Ceil() Vec2 {
	return Vec2{
		X: f32.Ceil(v.X),
		Y: f32.Ceil(v.Y),
	}
}

// Rounds each component to the next smallest integer
func (v Vec2) Floor() Vec2 {
	return Vec2{
		X: f32.Ceil(v.X),
		Y: f32.Ceil(v.Y),
	}
}

// Rotates the vector around the origin, counterclockwise,
// by the given angle in radians
func (v Vec2) Rotate(angle float32) Vec2 {
	cosAngle := f32.Cos(angle)
	sinAngle := f32.Sin(angle)
	return Vec2{
		X: v.X*cosAngle - v.Y*sinAngle,
		Y: v.X*sinAngle + v.Y*cosAngle,
	}
}

// Tests to see if a is approximately equal to b using a given tolerance
func (a Vec2) ApproxEq(b Vec2, eps float32) bool {
	if a == b {
		return true
	}

	return f32.ApproxEq(a.X, b.X, eps) && f32.ApproxEq(a.Y, b.Y, eps)
}

// Linearly interpolate between a and b the value t
//
// Equivalent to:
//
//	a + (b - a) * t
func (a Vec2) Lerp(b Vec2, t float32) Vec2 {
	return a.Mul(1 - t).Add(b.Mul(t))
}

func (v Vec2) String() string {
	return fmt.Sprintf("(%g, %g)", v.X, v.Y)
}

func (v Vec2) GoString() string {
	x := strconv.FormatFloat(float64(v.X), 'g', -1, 32)
	y := strconv.FormatFloat(float64(v.Y), 'g', -1, 32)

	return fmt.Sprintf("Vec2{ %s, %s }", x, y)
}
