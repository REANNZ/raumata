package vec_test

import (
	"math"
	"testing"

	"github.com/REANNZ/raumata/internal/f32"
	"github.com/REANNZ/raumata/vec"
)

func TestPolylineLength(t *testing.T) {
	checkLen := func(pl vec.Polyline, expected float32) {
		t.Helper()
		actual := pl.Length()
		if !f32.ApproxEq(actual, expected, 1e-12) {
			t.Errorf("Line incorrect length, expected %f, got %f",
				expected, actual)
		}
	}

	checkLen(nil, 0)
	checkLen([]vec.Vec2{{0, 0}}, 0)

	checkLen([]vec.Vec2{
		{0, 0},
		{0, 0},
	}, 0)

	checkLen([]vec.Vec2{
		{0, 0},
		{1, 0},
	}, 1)
	checkLen([]vec.Vec2{
		{0, 0},
		{0, 0},
		{1, 0},
		{1, 0},
	}, 1)

	checkLen([]vec.Vec2{
		{0, 0},
		{1, 0},
		{2, 0},
	}, 2)

	checkLen([]vec.Vec2{
		{0, 0},
		{1, 1},
	}, math.Sqrt2)

	checkLen([]vec.Vec2{
		{0, 0},
		{1, 1},
		{0, 2},
	}, 2*math.Sqrt2)
}

func TestPolylineFix(t *testing.T) {
	check := func(actual, expected vec.Polyline) {
		t.Helper()

		if len(actual) != len(expected) {
			t.Errorf("Expected %d, got %d", len(expected), len(actual))
		}
		for i := range actual {
			if actual[i] != expected[i] {
				t.Errorf("Expected %s, got %s at index %d", expected[i], actual[i], i)
			}
		}
	}

	var line vec.Polyline = []vec.Vec2{
		{0, 0},
		{1, 1},
	}

	var line2 vec.Polyline = []vec.Vec2{
		{0, 0},
		{0, 0},
		{1, 1},
	}

	check(line.Fix(), line)
	check(line2.Fix(), line)
}

func TestPolylineSimplify(t *testing.T) {
	check := func(actual, expected vec.Polyline) {
		t.Helper()

		if len(actual) != len(expected) {
			t.Errorf("Expected %d, got %d", len(expected), len(actual))
		}
		for i := range actual {
			if actual[i] != expected[i] {
				t.Errorf("Expected %s, got %s at index %d", expected[i], actual[i], i)
			}
		}
	}

	var line vec.Polyline = []vec.Vec2{
		{0, 0},
		{1, 1},
	}
	check(line.Simplify(), line)

	line = []vec.Vec2{
		{0, 0},
		{1, 0},
		{2, 0},
		{3, 0},
	}

	check(line.Simplify(), []vec.Vec2{
		{0, 0},
		{3, 0},
	})

	line = []vec.Vec2{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 2},
	}

	check(line.Simplify(), []vec.Vec2{
		{0, 0},
		{1, 0},
		{3, 2},
	})

	line = []vec.Vec2{
		{0, 0},
		{1, 0},
		{2, 1},
		{1, 2},
	}

	check(line.Simplify(), []vec.Vec2{
		{0, 0},
		{1, 0},
		{2, 1},
		{1, 2},
	})

	line = []vec.Vec2{
		{0, 0},
		{1, 0},
		{2, 0},
		{2, 1},
		{2, 2},
		{2, 3},
	}

	simpLine := line.Simplify()
	check(simpLine, []vec.Vec2{
		{0, 0},
		{2, 0},
		{2, 3},
	})

	p1 := line.Interpolate(0.5)
	p2 := simpLine.Interpolate(0.5)

	if !p1.ApproxEq(p2, 1e-7) {
		t.Errorf("Expected %s, got %s", p1, p2)
	}
}

func TestPolylineInterpolate(t *testing.T) {
	var line vec.Polyline = []vec.Vec2{{0, 0}, {1, 1}}

	checkVec(t, line.Interpolate(-1), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(0), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(1), vec.Vec2{1, 1})
	checkVec(t, line.Interpolate(2), vec.Vec2{1, 1})
	checkVec(t, line.Interpolate(0.5), vec.Vec2{0.5, 0.5})
	checkVec(t, line.Interpolate(0.1), vec.Vec2{0.1, 0.1})

	line = []vec.Vec2{
		{0, 0},
		{1, 0},
		{1, 1},
	}
	checkVec(t, line.Interpolate(-1), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(0), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(1), vec.Vec2{1, 1})
	checkVec(t, line.Interpolate(2), vec.Vec2{1, 1})
	checkVec(t, line.Interpolate(0.5), vec.Vec2{1, 0})
	checkVec(t, line.Interpolate(0.25), vec.Vec2{0.5, 0})
	checkVec(t, line.Interpolate(0.75), vec.Vec2{1, 0.5})

	line = []vec.Vec2{
		{0, 0},
		{1, 0},
		{1, 0},
		{1, 1},
	}
	checkVec(t, line.Interpolate(-1), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(0), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(1), vec.Vec2{1, 1})
	checkVec(t, line.Interpolate(2), vec.Vec2{1, 1})
	checkVec(t, line.Interpolate(0.5), vec.Vec2{1, 0})
	checkVec(t, line.Interpolate(0.25), vec.Vec2{0.5, 0})
	checkVec(t, line.Interpolate(0.75), vec.Vec2{1, 0.5})

	line = []vec.Vec2{
		{0, 0},
		{1, 0},
		{2, 0},
		{3, 0},
	}
	checkVec(t, line.Interpolate(-1), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(0), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(1), vec.Vec2{3, 0})
	checkVec(t, line.Interpolate(2), vec.Vec2{3, 0})
	checkVec(t, line.Interpolate(0.5), vec.Vec2{1.5, 0})
	checkVec(t, line.Interpolate(0.25), vec.Vec2{0.75, 0})
	checkVec(t, line.Interpolate(0.75), vec.Vec2{2.25, 0})

	line = []vec.Vec2{
		{0, 0},
		{2, 0},
		{3, 0},
		{4, 0},
	}
	checkVec(t, line.Interpolate(-1), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(0), vec.Vec2{0, 0})
	checkVec(t, line.Interpolate(1), vec.Vec2{4, 0})
	checkVec(t, line.Interpolate(2), vec.Vec2{4, 0})
	checkVec(t, line.Interpolate(0.5), vec.Vec2{2.0, 0})
	checkVec(t, line.Interpolate(0.25), vec.Vec2{1.0, 0})
	checkVec(t, line.Interpolate(0.75), vec.Vec2{3.0, 0})
}

func TestPolylineSplitAt(t *testing.T) {
	var line vec.Polyline = []vec.Vec2{{0, 0}, {1, 1}}

	l1, l2 := line.SplitAt(0.5)

	checkVec(t, l1[0], vec.Vec2{0, 0})
	checkVec(t, l1[1], vec.Vec2{0.5, 0.5})
	checkVec(t, l2[0], vec.Vec2{0.5, 0.5})
	checkVec(t, l2[1], vec.Vec2{1, 1})

	line = []vec.Vec2{{0, 0}, {0, 1}, {1, 1}}
	l1, l2 = line.SplitAt(0.5)
	checkVec(t, l1[0], vec.Vec2{0, 0})
	checkVec(t, l1[1], vec.Vec2{0, 1})
	checkVec(t, l2[0], vec.Vec2{0, 1})
	checkVec(t, l2[1], vec.Vec2{1, 1})

	line = []vec.Vec2{{0, 0}, {0, 1}, {3, 1}}

	l1, l2 = line.SplitAt(0.5)

	checkVec(t, l1[0], vec.Vec2{0, 0})
	checkVec(t, l1[1], vec.Vec2{0, 1})
	checkVec(t, l1[2], vec.Vec2{1, 1})
	checkVec(t, l2[0], vec.Vec2{1, 1})
	checkVec(t, l2[1], vec.Vec2{3, 1})
}

func TestPolylineSubdivide(t *testing.T) {
	checkSubdivide := func(pl vec.Polyline, n int, rec bool) {
		t.Helper()
		expected := pl.Length()

		for i := 0; i < n; i++ {
			var subDiv vec.Polyline
			if rec {
				subDiv = pl.Subdivide(2)
			} else {
				subDiv = pl.Subdivide(i)
			}
			actual := subDiv.Length()
			if !f32.ApproxEq(actual, expected, 1e-12) {
				t.Errorf("Line incorrect length at subdivision level %d, expected %f, got %f (diff %g)",
					i, expected, actual, actual-expected)
			}

			if rec {
				pl = subDiv
			}
		}
	}

	var line vec.Polyline = []vec.Vec2{{0, 0}, {1, 0}}

	checkSubdivide(line, 8, false)

	line = []vec.Vec2{{0, 0}, {1, 0}, {2, 0}, {4, 2}}
	checkSubdivide(line, 8, false)
	checkSubdivide(line, 6, true)
}

func BenchmarkPolylineLength(b *testing.B) {
	var line vec.Polyline = []vec.Vec2{
		{0, 0},
		{0.5, 0},
		{1, 0},
		{1, 0.5},
		{1, 1},
		{0.5, 1},
		{0, 1},
	}

	for i := 0; i < b.N; i++ {
		line.Length()
	}
}

func BenchmarkPolylineInterpolate(b *testing.B) {
	var line vec.Polyline = []vec.Vec2{
		{0, 0},
		{1, 0},
		{1, 1},
		{0, 1},
	}

	for i := 0; i < b.N; i++ {
		line.Interpolate(0.5)
	}
}
