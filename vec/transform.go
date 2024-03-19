package vec

import "github.com/REANNZ/raumata/internal/f32"

// Transform is an affine transformation matrix
// It is an augmented matrix:
//
//	A C | E
//	B D | F
//	0 0 | 1
//
// Where (E, F) is the translation component of the
// transformation
type Transform struct {
	//   A  C  E
	//   B  D  F
	//   0  0  1
	A, B, C, D, E, F float32
}

func NewTransform(a, b, c, d, e, f float32) *Transform {
	return &Transform{
		a, b, c, d, e, f,
	}
}

// Returns a new transformation that doesn't modify any
// vector.
//
//	Ix = x
func NewIdentityTransform() *Transform {
	return NewTransform(1, 0, 0, 1, 0, 0)
}

// Returns a new transformation that represents the
// translation of a point by v
//
//	Tx = x + v
func NewTranslate(v Vec2) *Transform {
	return NewTransform(1, 0, 0, 1, v.X, v.Y)
}

// Returns a new transformation that represents scaling
// a point on the x direction by v.X and the y direction
// by v.Y
//
//	Tx = (x.X * v.X, x.Y * v.Y)
func NewScale(v Vec2) *Transform {
	return NewTransform(v.X, 0, 0, v.Y, 0, 0)
}

// Returns a new transformation that represents rotating
// a point around the origin counterclockwise by the given 
// angle
//
//	Tx = x.Rotate(θ)
func NewRotate(angle float32) *Transform {
	cosAngle := f32.Cos(angle)
	sinAngle := f32.Sin(angle)
	return NewTransform(cosAngle, sinAngle, -sinAngle, cosAngle, 0, 0)
}

// Apply applies the transform to v
func (t *Transform) Apply(v Vec2) Vec2 {
	x := t.E + v.X * t.A + v.Y * t.C
	y := t.F + v.X * t.B + v.Y * t.D

	return Vec2{ X: x, Y: y }
}

// Combine the two transforms.
//
// Returns a transform that is equivalent to applying
// t1 then t2. Mathematically this is:
//
//	T = T2*T1
func (t1 *Transform) Combine(t2 *Transform) *Transform {
	// Post-multiply t1 and t2 (t2*t1)
	a := t2.A * t1.A + t2.C * t1.B
	b := t2.B * t1.A + t2.D * t1.B
	c := t2.A * t1.C + t2.C * t1.D
	d := t2.B * t1.C + t2.D * t1.D
	e := t2.A * t1.E + t2.C * t1.F + t2.E
	f := t2.B * t1.E + t2.D * t1.F + t2.F

	return NewTransform(a, b, c, d, e, f)
}

// Returns whether this transform is exactly the
// identity
func (t *Transform) IsIdentity() bool {
	return *t == Transform{
		A: 1, B: 0, C: 0, D: 1, E: 0, F: 0,
	}
}

// If the transformation is a pure translation, this
// method returns the translation.
//
// If ok is false, this transform is not a pure translation
func (t *Transform) GetTranslation() (v Vec2, ok bool) {
	if t.A == 1 && t.B == 0 && t.C == 0 && t.D == 1 {
		return Vec2{ X: t.E, Y: t.F }, true
	}
	return Vec2{}, false
}

// If the transformation is a pure rotation, this
// method returns the rotation.
//
// If ok is false, this transform is not a pure rotation
func (t *Transform) GetRotation() (float32, bool) {
	// Check that the transform has A translation
	// component
	if t.E != 0 || t.F != 0 {
		return 0, false
	}

	// A rotation matrix has A == D && B == -C
	if t.A != t.D || t.B != -t.C {
		return 0, false
	}

	// A matrix with A determinant of 1 has no scale
	det := t.determinant()
	if f32.ApproxEq(det, 1, 1e-8) {
		return f32.Acos(t.A), true
	}

	return 0, false
}

func (t *Transform) determinant() float32 {
	// The determinant of the matrix
	//   A  C  E
	//   B  D  F
	//   0  0  1
	//
	// Reduces to the determinant of
	//   A  C
	//   B  D
	return t.A*t.D - t.B*t.C
}
