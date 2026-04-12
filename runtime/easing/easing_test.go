package easing

import "testing"

func TestClampAndCurves(t *testing.T) {
	if Linear(-1) != 0 || Linear(2) != 1 {
		t.Fatal("expected linear to clamp into [0,1]")
	}
	if InOutQuad(0) != 0 || InOutQuad(1) != 1 {
		t.Fatal("expected InOutQuad endpoints to be stable")
	}
	if InOutCubic(0) != 0 || InOutCubic(1) != 1 {
		t.Fatal("expected InOutCubic endpoints to be stable")
	}
	if OutBack(0) != 0 {
		t.Fatal("expected OutBack(0) == 0")
	}
	if out := OutBack(1); out < 0.999 || out > 1.001 {
		t.Fatalf("expected OutBack(1) ~= 1, got %f", out)
	}
}

func TestSteps(t *testing.T) {
	steps := Steps(4)
	if got := steps(0.49); got != 0.25 {
		t.Fatalf("expected stepped value 0.25, got %f", got)
	}
	if got := steps(1); got != 1 {
		t.Fatalf("expected stepped value 1, got %f", got)
	}
}
