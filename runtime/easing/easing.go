package easing

import "math"

type Func func(float64) float64

func Clamp01(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

func Linear(t float64) float64 {
	return Clamp01(t)
}

func InOutQuad(t float64) float64 {
	t = Clamp01(t)
	if t < 0.5 {
		return 2 * t * t
	}
	u := -2*t + 2
	return 1 - (u*u)/2
}

func InOutCubic(t float64) float64 {
	t = Clamp01(t)
	if t < 0.5 {
		return 4 * t * t * t
	}
	u := -2*t + 2
	return 1 - (u*u*u)/2
}

func OutBack(t float64) float64 {
	t = Clamp01(t)
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	const c1 = 1.70158
	const c3 = c1 + 1
	u := t - 1
	return 1 + c3*u*u*u + c1*u*u
}

func Steps(n int) Func {
	if n <= 1 {
		return Linear
	}
	return func(t float64) float64 {
		t = Clamp01(t)
		return math.Floor(t*float64(n)) / float64(n)
	}
}
