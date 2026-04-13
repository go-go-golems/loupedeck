package device

import "testing"

func TestButtonStringAndParseRoundTrip(t *testing.T) {
	cases := []Button{Circle, Button1, Button7, CTCircle, Undo, Keyboard, Enter, Save, LeftFn, Up, Left, RightFn, Down, Right, E}
	for _, tc := range cases {
		name := tc.String()
		parsed, err := ParseButton(name)
		if err != nil {
			t.Fatalf("ParseButton(%q): %v", name, err)
		}
		if parsed != tc {
			t.Fatalf("roundtrip mismatch: got %v want %v", parsed, tc)
		}
	}
}

func TestButtonParseSupportsAliases(t *testing.T) {
	cases := map[string]Button{
		"A": Up,
		"B": Down,
		"C": Left,
		"D": Right,
	}
	for name, want := range cases {
		got, err := ParseButton(name)
		if err != nil {
			t.Fatalf("ParseButton(%q): %v", name, err)
		}
		if got != want {
			t.Fatalf("ParseButton(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestKnobStringAndParseRoundTrip(t *testing.T) {
	cases := []Knob{CTKnob, Knob1, Knob2, Knob3, Knob4, Knob5, Knob6}
	for _, tc := range cases {
		name := tc.String()
		parsed, err := ParseKnob(name)
		if err != nil {
			t.Fatalf("ParseKnob(%q): %v", name, err)
		}
		if parsed != tc {
			t.Fatalf("roundtrip mismatch: got %v want %v", parsed, tc)
		}
	}
}

func TestTouchButtonStringAndParseRoundTrip(t *testing.T) {
	cases := []TouchButton{TouchLeft, TouchRight, Touch1, Touch2, Touch3, Touch4, Touch5, Touch6, Touch7, Touch8, Touch9, Touch10, Touch11, Touch12}
	for _, tc := range cases {
		name := tc.String()
		parsed, err := ParseTouchButton(name)
		if err != nil {
			t.Fatalf("ParseTouchButton(%q): %v", name, err)
		}
		if parsed != tc {
			t.Fatalf("roundtrip mismatch: got %v want %v", parsed, tc)
		}
	}
}

func TestButtonStatusString(t *testing.T) {
	if ButtonDown.String() != "down" {
		t.Fatalf("ButtonDown.String() = %q", ButtonDown.String())
	}
	if ButtonUp.String() != "up" {
		t.Fatalf("ButtonUp.String() = %q", ButtonUp.String())
	}
}

func TestParseRejectsUnknownNames(t *testing.T) {
	if _, err := ParseButton("Nope"); err == nil {
		t.Fatalf("expected ParseButton to reject unknown name")
	}
	if _, err := ParseKnob("Nope"); err == nil {
		t.Fatalf("expected ParseKnob to reject unknown name")
	}
	if _, err := ParseTouchButton("Nope"); err == nil {
		t.Fatalf("expected ParseTouchButton to reject unknown name")
	}
}
