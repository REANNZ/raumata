package vec_test

import (
	"math"
	"testing"

	"github.com/REANNZ/raumata/internal/f32"
	"github.com/REANNZ/raumata/vec"
)

func TestTransform(t *testing.T) {

	v := vec.Vec2{ 1, 1 }

	transform := vec.NewIdentityTransform()
	checkVec(t, transform.Apply(v), v)

	transform = vec.NewTranslate(vec.Vec2{ 1, 1 })
	checkVec(t, transform.Apply(v), vec.Vec2{ 2, 2 })

	transform = vec.NewScale(vec.Vec2{ 5, 5 })
	checkVec(t, transform.Apply(v), vec.Vec2{ 5, 5 })
	checkVec(t, transform.Apply(vec.Vec2{ 2, 2 }), vec.Vec2{ 10, 10 })

	transform = vec.NewRotate(math.Pi / 2)

	// Rotations are a little less accurate, so use a lower epsilon
	expected := vec.Vec2{ -1, 1 }
	actual := transform.Apply(v)
	if !expected.ApproxEq(actual, 1e-7) {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
	expected = vec.Vec2{ -1, -1 }
	actual = transform.Apply(vec.Vec2{ -1, 1 })
	if !expected.ApproxEq(actual, 1e-7) {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestTransformDecompose(t *testing.T) {

	transform := vec.NewTranslate(vec.Vec2{ 1, 1 })

	trans, ok := transform.GetTranslation()
	if !ok {
		t.Errorf("Transform is not a translation!")
	}

	checkVec(t, trans, vec.Vec2{ 1, 1 })

	transform = vec.NewRotate(math.Pi)

	rot, ok := transform.GetRotation()
	if !ok {
		t.Errorf("Transform is not a rotation!")
	}

	if !f32.ApproxEq(rot, math.Pi, 1e-7) {
		t.Errorf("Rotation amount incorrect, expected π, got %g", rot) 
	}
}

func TestTransformCombine(t *testing.T) {
	trans := vec.NewTranslate(vec.Vec2{ 1, 2 })
	scale := vec.NewScale(vec.Vec2{ 5, 5 })

	combined := trans.Combine(scale)

	v := vec.Vec2{ 1, 1 }

	vTrans := trans.Apply(v)
	vScale := scale.Apply(vTrans)

	checkVec(t, vScale, vec.Vec2{ 10, 15})
	checkVec(t, combined.Apply(v), vec.Vec2{ 10, 15})

	combined = scale.Combine(trans)

	vScale = scale.Apply(v)
	vTrans = trans.Apply(vScale)

	checkVec(t, vTrans, vec.Vec2{ 6, 7})
	checkVec(t, combined.Apply(v), vec.Vec2{ 6, 7})
}
