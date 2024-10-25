package glitch

import (
	"github.com/unitoftime/glitch/internal/glfw"
	"github.com/unitoftime/glitch/internal/mainthread"
)

func GetConnectedGamepads() []Gamepad {
	var joysticks []glfw.Joystick

	mainthread.Call(func() {
		joysticks = glfw.GetConnectedGamepads()
	})

	ret := make([]Gamepad, len(joysticks))
	for i := range joysticks {
		ret[i] = Gamepad(joysticks[i])
	}
	return ret
}

type Gamepad int
const GamepadNone = Gamepad(-1)
// TODO: Comment out and add function that says: GetAllActiveGamepads() or something
// Note: I don't think that users should be statically referring to gamepads. I can't think of a use for that
// const (
// 	Gamepad1  = Gamepad(glfw.Joystick1)
// 	Gamepad2  = Gamepad(glfw.Joystick2)
// 	Gamepad3  = Gamepad(glfw.Joystick3)
// 	Gamepad4  = Gamepad(glfw.Joystick4)
// 	Gamepad5  = Gamepad(glfw.Joystick5)
// 	Gamepad6  = Gamepad(glfw.Joystick6)
// 	Gamepad7  = Gamepad(glfw.Joystick7)
// 	Gamepad8  = Gamepad(glfw.Joystick8)
// 	Gamepad9  = Gamepad(glfw.Joystick9)
// 	Gamepad10 = Gamepad(glfw.Joystick10)
// 	Gamepad11 = Gamepad(glfw.Joystick11)
// 	Gamepad12 = Gamepad(glfw.Joystick12)
// 	Gamepad13 = Gamepad(glfw.Joystick13)
// 	Gamepad14 = Gamepad(glfw.Joystick14)
// 	Gamepad15 = Gamepad(glfw.Joystick15)
// 	Gamepad16 = Gamepad(glfw.Joystick16)

// 	GamepadLast = Gamepad(glfw.JoystickLast)
// )

type GamepadAxis int
const (
	AxisLeftX        = GamepadAxis(glfw.AxisLeftX)
	AxisLeftY        = GamepadAxis(glfw.AxisLeftY)
	AxisRightX       = GamepadAxis(glfw.AxisRightX)
	AxisRightY       = GamepadAxis(glfw.AxisRightY)
	AxisLeftTrigger  = GamepadAxis(glfw.AxisLeftTrigger)
	AxisRightTrigger = GamepadAxis(glfw.AxisRightTrigger)
	AxisLast         = GamepadAxis(glfw.AxisLast)
)

type GamepadButton int
const (
	ButtonA           = GamepadButton(glfw.ButtonA)
	ButtonB           = GamepadButton(glfw.ButtonB)
	ButtonX           = GamepadButton(glfw.ButtonX)
	ButtonY           = GamepadButton(glfw.ButtonY)
	ButtonLeftBumper  = GamepadButton(glfw.ButtonLeftBumper)
	ButtonRightBumper = GamepadButton(glfw.ButtonRightBumper)
	ButtonBack        = GamepadButton(glfw.ButtonBack)
	ButtonStart       = GamepadButton(glfw.ButtonStart)
	ButtonGuide       = GamepadButton(glfw.ButtonGuide)
	ButtonLeftThumb   = GamepadButton(glfw.ButtonLeftThumb)
	ButtonRightThumb  = GamepadButton(glfw.ButtonRightThumb)
	ButtonDpadUp      = GamepadButton(glfw.ButtonDpadUp)
	ButtonDpadRight   = GamepadButton(glfw.ButtonDpadRight)
	ButtonDpadDown    = GamepadButton(glfw.ButtonDpadDown)
	ButtonDpadLeft    = GamepadButton(glfw.ButtonDpadLeft)
	ButtonCross       = GamepadButton(glfw.ButtonCross)
	ButtonCircle      = GamepadButton(glfw.ButtonCircle)
	ButtonSquare      = GamepadButton(glfw.ButtonSquare)
	ButtonTriangle    = GamepadButton(glfw.ButtonTriangle)

	ButtonFirst       = ButtonA
	ButtonLast        = GamepadButton(glfw.ButtonLast)
)

// Returns the primary gamepad
func (w *Window) GetPrimaryGamepad() Gamepad {
	return w.currentPrimaryGamepad
}

func findNewActiveGamepad() Gamepad {
	gamepads := GetConnectedGamepads()

	for _, gp := range gamepads {
		state := gp.getGamepadState()
		if checkGamepadActive(state) {
			return gp
		}
	}
	return GamepadNone
}

// Returns true if the gamepad state is considered active this frame
func checkGamepadActive(state *glfw.GamepadState) bool {
	if state == nil { return false }

	for i := range state.Buttons {
		if state.Buttons[i] == glfw.Press {
			return true
		}
	}
	return false
}

func (g Gamepad) getGamepadState() *glfw.GamepadState {
	if g == GamepadNone { return nil }

	var ret *glfw.GamepadState
	mainthread.Call(func() {
		ret = glfw.Joystick(g).GetGamepadState()
	})
	return ret
}

// Returns true if the gamepad button is pressed, else returns false
func (w *Window) GetGamepadPressed(g Gamepad, button GamepadButton) bool {
	if g == GamepadNone { return false }

	return w.pressedGamepad[button]
}

// Returns true if the gamepad button was just pressed this frame, else returns false
func (w *Window) GetGamepadJustPressed(g Gamepad, button GamepadButton) bool {
	if g == GamepadNone { return false }

	return w.justPressedGamepad[button]
}

// Returns the gamepad axis value, ranging on -1 to +1
func (w *Window) GetGamepadAxis(g Gamepad, axis GamepadAxis) float64 {
	if g == GamepadNone { return 0 }

	return w.gamepadAxis[axis]
}
