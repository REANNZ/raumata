// Functions for operating on 32bit floats, mostly just
// wrappers around the functions from std/math
package f32

import "math"

// Returns the absoulate value of x.
func Abs(x float32) float32 {
	return float32(math.Abs(float64(x)))
}

// Returns the arccosine, in radians, of x.
func Acos(x float32) float32 {
	return float32(math.Acos(float64(x)))
}

// Returns the arcsine, in radians, of x.
func Asin(x float32) float32 {
	return float32(math.Asin(float64(x)))
}

// Returns the arctangent, in radians, of x.
func Atan(x float32) float32 {
	return float32(math.Atan(float64(x)))
}

// Returns the least integer value greather than or equal to x.
func Ceil(x float32) float32 {
	return float32(math.Ceil(float64(x)))
}

// Returns the cosine of the radian argument x.
func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

// Returns the greatest integer value less than or equal to x
func Floor(x float32) float32 {
	return float32(math.Floor(float64(x)))
}

// Returns [Sqrt](p*p + q*q)
func Hypot(p, q float32) float32 {
	return float32(math.Hypot(float64(p), float64(q)))
}

// Returns positive infinity if sign >= 0, negative infinity if sign < 0.
func Inf(sign int) float32 {
	return float32(math.Inf(sign))
}

// Reports whether f is an infinity according to sign.
// sign > 0  reports x == +Inf
// sign < 0  reports x == -Inf
// sign == 0 reports x == +Inf || x == -Inf
func IsInf(f float32, sign int) bool {
	return math.IsInf(float64(f), sign)
}

// Reports whether f is NaN
func IsNaN(f float32) bool {
	return math.IsNaN(float64(f))
}

// Returns the maximum value of the arguments
// If any argument is NaN, Max returns NaN
func Max(args ...float32) float32 {
	if len(args) == 0 {
		return NaN()
	}

	mx := args[0]
	for _, v := range args {
		if IsInf(v, 1) || IsNaN(v) {
			return v
		}

		if v > mx {
			mx = v
		}
	}

	return mx
}

// Returns the minimum value of the arguments
// If any argument is NaN, Min returns NaN
func Min(args ...float32) float32 {
	if len(args) == 0 {
		return NaN()
	}

	mn := args[0]
	for _, v := range args {
		if IsInf(v, -1) || IsNaN(v) {
			return v
		}

		if v < mn {
			mn = v
		}
	}

	return mn
}

// Returns a NaN value
func NaN() float32 {
	return float32(math.NaN())
}

// Returns x**y, the base-x exponential of y
func Pow(x, y float32) float32 {
	return float32(math.Pow(float64(x), float64(y)))
}

// Returns the nearest integer, rounding half away from zero
func Round(x float32) float32 {
	return float32(math.Round(float64(x)))
}

// Returns the sine of the radian argument x
func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

// Returns the square root of x
func Sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

// Returns the tangent of the radian argument x
func Tan(x float32) float32 {
	return float32(math.Tan(float64(x)))
}

// Returns whether the difference between p and q is less
// than eps
func ApproxEq(p, q, eps float32) bool {
	if p == q {
		return true
	}
	diff := Abs(p - q)
	return diff < eps
}

// Sums `vals` using pairwise summation to reduce round-off error
func Sum(vals []float32) float32 {
	if len(vals) <= 8 {
		var s float32 = 0.0
		for _, v := range vals {
			s += v
		}
		return s
	}

	splitAt := len(vals) / 2

	return Sum(vals[:splitAt]) + Sum(vals[splitAt:])
}
