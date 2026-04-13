/*
   Copyright 2021 Google LLC

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       https://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package device

import "fmt"

// Button represents a physical button on the Loupedeck Live.  This
// includes the 8 buttons at the bottom of the device as well as the
// 'click' function of the 6 dials.
type Button uint16

const (
	KnobPress1 Button = 1
	KnobPress2 Button = 2
	KnobPress3 Button = 3
	KnobPress4 Button = 4
	KnobPress5 Button = 5
	KnobPress6 Button = 6

	// Circle is sent when the left-most hardware button under the
	// display is clicked.  This has a circle icon on the
	// Loupedeck Live, but is unfortunately labeled "1" on the
	// Loupedeck CT.
	Circle  Button = 7
	Button1 Button = 8
	Button2 Button = 9
	Button3 Button = 10
	Button4 Button = 11
	Button5 Button = 12
	Button6 Button = 13
	Button7 Button = 14

	// CT-specific buttons.
	CTCircle Button = 15
	Undo     Button = 16
	Keyboard Button = 17
	Enter    Button = 18
	Save     Button = 19
	LeftFn   Button = 20
	Up       Button = 21
	A        Button = 21
	Left     Button = 22
	C        Button = 22
	RightFn  Button = 23
	Down     Button = 24
	B        Button = 24
	Right    Button = 25
	D        Button = 25
	E        Button = 26
)

// ButtonStatus represents the state of Buttons.
type ButtonStatus uint8

const (
	// ButtonDown indicates that a button has just been pressed.
	ButtonDown ButtonStatus = 0
	// ButtonUp indicates that a button was just released.
	ButtonUp ButtonStatus = 1
)

// ButtonFunc is a function signature used for callbacks on Button
// events.  When a specified event happens, the ButtonFunc is called
// with parameters specifying which button was pushed and what its
// current state is.
type ButtonFunc func(Button, ButtonStatus)

// Knob represents the 6 knobs on the Loupedeck Live.
type Knob uint16

const (
	// CTKnob is the large knob in the center of the Loupedeck CT
	CTKnob Knob = 0
	// Knob1 is the upper left knob.
	Knob1 = 1
	// Knob2 is the middle left knob.
	Knob2 = 2
	// Knob3 is the bottom left knob.
	Knob3 = 3
	// Knob4 is the upper right knob.
	Knob4 = 4
	// Knob5 is the middle right knob.
	Knob5 = 5
	// Knob6 is the bottom right knob.
	Knob6 = 6
)

// KnobFunc is a function signature used for callbacks on Knob events,
// similar to ButtonFunc's use with Button events.  The exact use of
// the second parameter depends on the use; in some cases it's simply
// +1/-1 (for right/left button turns) and in other cases it's the
// current value of the dial.
type KnobFunc func(Knob, int)

// TouchButton represents the regions of the touchpad on the Loupedeck Live.
type TouchButton uint16

const (
	// TouchLeft indicates that the left touchscreen area, near the leftmost knobs has been touched.
	TouchLeft TouchButton = 1
	// TouchRight indicates that hte right touchscreen area, near the rightmost knobs has been touched.
	TouchRight = 2
	Touch1     = 3
	Touch2     = 4
	Touch3     = 5
	Touch4     = 6
	Touch5     = 7
	Touch6     = 8
	Touch7     = 9
	Touch8     = 10
	Touch9     = 11
	Touch10    = 12
	Touch11    = 13
	Touch12    = 14
)

// TouchFunc is a function signature used for callbacks on TouchButton
// events, similar to ButtonFunc and KnobFunc.  The parameters are:
//
//   - The TouchButton touched
//   - The ButtonStatus (down/up)
//   - The X location touched (relative to the whole display)
//   - The Y location touched (relative to the whole display)
type TouchFunc func(TouchButton, ButtonStatus, uint16, uint16)

func (b Button) String() string {
	switch b {
	case KnobPress1:
		return "KnobPress1"
	case KnobPress2:
		return "KnobPress2"
	case KnobPress3:
		return "KnobPress3"
	case KnobPress4:
		return "KnobPress4"
	case KnobPress5:
		return "KnobPress5"
	case KnobPress6:
		return "KnobPress6"
	case Circle:
		return "Circle"
	case Button1:
		return "Button1"
	case Button2:
		return "Button2"
	case Button3:
		return "Button3"
	case Button4:
		return "Button4"
	case Button5:
		return "Button5"
	case Button6:
		return "Button6"
	case Button7:
		return "Button7"
	case CTCircle:
		return "CTCircle"
	case Undo:
		return "Undo"
	case Keyboard:
		return "Keyboard"
	case Enter:
		return "Enter"
	case Save:
		return "Save"
	case LeftFn:
		return "LeftFn"
	case Up:
		return "Up"
	case Left:
		return "Left"
	case RightFn:
		return "RightFn"
	case Down:
		return "Down"
	case Right:
		return "Right"
	case E:
		return "E"
	default:
		return fmt.Sprintf("Button(%d)", b)
	}
}

func ParseButton(name string) (Button, error) {
	switch name {
	case "KnobPress1":
		return KnobPress1, nil
	case "KnobPress2":
		return KnobPress2, nil
	case "KnobPress3":
		return KnobPress3, nil
	case "KnobPress4":
		return KnobPress4, nil
	case "KnobPress5":
		return KnobPress5, nil
	case "KnobPress6":
		return KnobPress6, nil
	case "Circle":
		return Circle, nil
	case "Button1":
		return Button1, nil
	case "Button2":
		return Button2, nil
	case "Button3":
		return Button3, nil
	case "Button4":
		return Button4, nil
	case "Button5":
		return Button5, nil
	case "Button6":
		return Button6, nil
	case "Button7":
		return Button7, nil
	case "CTCircle":
		return CTCircle, nil
	case "Undo":
		return Undo, nil
	case "Keyboard":
		return Keyboard, nil
	case "Enter":
		return Enter, nil
	case "Save":
		return Save, nil
	case "LeftFn":
		return LeftFn, nil
	case "Up", "A":
		return Up, nil
	case "Left", "C":
		return Left, nil
	case "RightFn":
		return RightFn, nil
	case "Down", "B":
		return Down, nil
	case "Right", "D":
		return Right, nil
	case "E":
		return E, nil
	default:
		return 0, fmt.Errorf("unknown button %q", name)
	}
}

func (s ButtonStatus) String() string {
	switch s {
	case ButtonDown:
		return "down"
	case ButtonUp:
		return "up"
	default:
		return fmt.Sprintf("ButtonStatus(%d)", s)
	}
}

func (k Knob) String() string {
	switch k {
	case CTKnob:
		return "CTKnob"
	case Knob1:
		return "Knob1"
	case Knob2:
		return "Knob2"
	case Knob3:
		return "Knob3"
	case Knob4:
		return "Knob4"
	case Knob5:
		return "Knob5"
	case Knob6:
		return "Knob6"
	default:
		return fmt.Sprintf("Knob(%d)", k)
	}
}

func ParseKnob(name string) (Knob, error) {
	switch name {
	case "CTKnob":
		return CTKnob, nil
	case "Knob1":
		return Knob1, nil
	case "Knob2":
		return Knob2, nil
	case "Knob3":
		return Knob3, nil
	case "Knob4":
		return Knob4, nil
	case "Knob5":
		return Knob5, nil
	case "Knob6":
		return Knob6, nil
	default:
		return 0, fmt.Errorf("unknown knob %q", name)
	}
}

func (t TouchButton) String() string {
	switch t {
	case TouchLeft:
		return "TouchLeft"
	case TouchRight:
		return "TouchRight"
	case Touch1:
		return "Touch1"
	case Touch2:
		return "Touch2"
	case Touch3:
		return "Touch3"
	case Touch4:
		return "Touch4"
	case Touch5:
		return "Touch5"
	case Touch6:
		return "Touch6"
	case Touch7:
		return "Touch7"
	case Touch8:
		return "Touch8"
	case Touch9:
		return "Touch9"
	case Touch10:
		return "Touch10"
	case Touch11:
		return "Touch11"
	case Touch12:
		return "Touch12"
	default:
		return fmt.Sprintf("TouchButton(%d)", t)
	}
}

func ParseTouchButton(name string) (TouchButton, error) {
	switch name {
	case "TouchLeft":
		return TouchLeft, nil
	case "TouchRight":
		return TouchRight, nil
	case "Touch1":
		return Touch1, nil
	case "Touch2":
		return Touch2, nil
	case "Touch3":
		return Touch3, nil
	case "Touch4":
		return Touch4, nil
	case "Touch5":
		return Touch5, nil
	case "Touch6":
		return Touch6, nil
	case "Touch7":
		return Touch7, nil
	case "Touch8":
		return Touch8, nil
	case "Touch9":
		return Touch9, nil
	case "Touch10":
		return Touch10, nil
	case "Touch11":
		return Touch11, nil
	case "Touch12":
		return Touch12, nil
	default:
		return 0, fmt.Errorf("unknown touch button %q", name)
	}
}

// touchCoordToButton translates an x,y coordinate on the
// touchscreen to a TouchButton.
func touchCoordToButton(x, y uint16) TouchButton {
	switch {
	case x < 60:
		return TouchLeft
	case x >= 420:
		return TouchRight
	}

	x -= 60
	x /= 90
	y /= 90

	return TouchButton(uint16(Touch1) + x + 4*y)
}
