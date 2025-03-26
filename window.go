package glitch

import (
	"fmt"
	"image"
	"time"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/glfw"
	"github.com/unitoftime/glitch/internal/mainthread"
)

type WindowConfig struct {
	Fullscreen  bool
	Undecorated bool
	Maximized   bool
	Vsync       bool
	// Resizable bool
	Samples int
	Icons []image.Image
}

type Window struct {
	window *glfw.Window

	vsync bool
	closed bool

	width, height int

	tmpInput, input struct {
		justPressed  [KeyLast + 1]bool
		justReleased [KeyLast + 1]bool
		repeated     [KeyLast + 1]bool

		scroll struct {
			X, Y float64
		}
	}

	// TODO: Could use an array? I think gamepad indexes have restrictions in glfw and browser
	cachedGamepadStates map[Gamepad]*glfw.GamepadState

	mousePosition         Vec2
	currentPrimaryGamepad Gamepad              // Tracks the current primarily used gamepad
	justPressedGamepad    [ButtonLast + 1]bool // Tracks buttons that were just pressed this frame

	mainthreadUpdateConnectedGamepads func()
	connectedGamepads                 []Gamepad
	// TODO: Technically you could just store the last primaryGamepadState
	pressedGamepad [ButtonLast + 1]bool  // Tracks buttons that are currently pressed this frame
	gamepadAxis    [AxisLast + 1]float64 // Tracks the current axis state

	// The back and front buffers for tracking typed characters
	typedBack, typedFront []rune

	mainthreadUpdate  func()
	mainthreadPressed func()
	pressedKeyCheck   Key
	pressedKeyReturn  bool

	scrollCallbacks      []glfw.ScrollCallback
	keyCallbacks         []glfw.KeyCallback
	mouseButtonCallbacks []glfw.MouseButtonCallback
	charCallbacks        []glfw.CharCallback

	repeatTracker     []repeatData
	mouseRepeatDelay  time.Duration // amount of time delay to wait after a mouse hold to consider it repeated
	mouseRepeatPeriod time.Duration // amount of time in between consecutive repeats after a repeat has started

	lastUpdateTime time.Time
}

type repeatData struct {
	key Key
	dur time.Duration
}

func NewWindow(width, height int, title string, config WindowConfig) (*Window, error) {
	win := &Window{
		scrollCallbacks:      make([]glfw.ScrollCallback, 0),
		keyCallbacks:         make([]glfw.KeyCallback, 0),
		mouseButtonCallbacks: make([]glfw.MouseButtonCallback, 0),
		cachedGamepadStates:  make(map[Gamepad]*glfw.GamepadState),

		repeatTracker: []repeatData{
			{key: MouseButtonLeft},
			{key: MouseButtonRight},
			{key: MouseButtonMiddle},
			// TODO: Make this configurable
			// TODO: Add all mousebuttons
			// TODO: Add all gamepad buttons
		},
		mouseRepeatDelay:  350 * time.Millisecond, // Note: The total delay for first repeat is delay + period
		mouseRepeatPeriod: 150 * time.Millisecond,

		lastUpdateTime: time.Now(),
	}

	err := mainthread.CallErr(func() error {
		err := glfw.Init(gl.ContextWatcher)
		if err != nil {
			return err
		}

		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		// glfw.WindowHint(glfw.Resizable, config.Resizable)
		if config.Samples > 0 {
			glfw.WindowHint(glfw.Samples, config.Samples)
		}
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True) // Compatibility - For Mac only?

		// Disables the ability to minimze a fullscreen window
		// glfw.WindowHint(glfw.AutoIconify, glfw.False)

		var monitor *glfw.Monitor
		if config.Fullscreen {
			monitor = glfw.GetPrimaryMonitor()
		}
		if config.Undecorated {
			glfw.WindowHint(glfw.Decorated, glfw.False)
		}
		if config.Maximized {
			glfw.WindowHint(glfw.Maximized, glfw.True)
		}

		win.window, err = glfw.CreateWindow(width, height, title, monitor, nil)
		if err != nil {
			return err
		}

		win.window.MakeContextCurrent()

		// log.Printf("OpenGL: %s %s %s; %v samples.\n", gl.GetString(gl.VENDOR), gl.GetString(gl.RENDERER), gl.GetString(gl.VERSION), gl.GetInteger(gl.SAMPLES))
		// log.Printf("GLSL: %s.\n", gl.GetString(gl.SHADING_LANGUAGE_VERSION))

		if len(config.Icons) > 0 {
			win.window.SetIcon(config.Icons)
		}

		if config.Samples > 0 {
			// TODO - But how to work with wasm (which enables multisample in the context?)
			gl.Enable(gl.MULTISAMPLE)
		}

		// gl.Enable(gl.DEPTH_TEST)
		// gl.Enable(gl.CULL_FACE)
		// gl.CullFace(gl.BACK)
		// gl.FrontFace(gl.CCW) // Default

		// gl.Enable(gl.BLEND) // TODO: Will this ever need to be disabled?
		// // gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA); // Non premult
		// gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA) // Premult

		win.vsync = config.Vsync
		if config.Vsync {
			glfw.SwapInterval(1)
		} else {
			glfw.SwapInterval(0)
		}

		win.width = width
		win.height = height
		gl.Viewport(0, 0, int(width), int(height))

		win.window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
			// fmt.Println("Framebuffer size callback", width, height)
			win.width = width
			win.height = height
			gl.Viewport(0, 0, int(win.width), int(win.height))
		})

		win.AddScrollCallback(func(w *glfw.Window, xoff float64, yoff float64) {
			// win.window.SetScrollCallback(func(w *glfw.Window, xoff float64, yoff float64) {
			win.tmpInput.scroll.X += xoff
			win.tmpInput.scroll.Y += yoff
		})

		win.AddMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
			// win.window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
			switch action {
			case glfw.Press:
				win.tmpInput.justPressed[Key(button)] = true
			case glfw.Release:
				win.tmpInput.justReleased[Key(button)] = true

				// Warning: Repeat events aren't returned for mouse callbacks. so we track them with repeatTracker
				// case glfw.Repeat:
				// 	win.tmpInput.justReleased[Key(button)] = true
			}
		})

		win.AddKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			// win.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if key == glfw.KeyUnknown {
				return
			}

			switch action {
			case glfw.Press:
				win.tmpInput.justPressed[Key(key)] = true
			case glfw.Release:
				win.tmpInput.justReleased[Key(key)] = true
			case glfw.Repeat:
				win.tmpInput.repeated[Key(key)] = true
			}
		})

		win.AddCharCallback(func(w *glfw.Window, char rune) {
			// win.window.SetCharCallback(func(w *glfw.Window, char rune) {
			win.typedBack = append(win.typedBack, char)
		})

		win.window.SetScrollCallback(func(w *glfw.Window, xoff float64, yoff float64) {
			for _, c := range win.scrollCallbacks {
				c(w, xoff, yoff)
			}
		})
		win.window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
			for _, c := range win.mouseButtonCallbacks {
				c(w, button, action, mods)
			}
		})
		win.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			for _, c := range win.keyCallbacks {
				c(w, key, scancode, action, mods)
			}
		})
		win.window.SetCharCallback(func(w *glfw.Window, char rune) {
			for _, c := range win.charCallbacks {
				c(w, char)
			}
		})

		// win.window.SetIconifyCallback(func(w *glfw.Window, iconified bool) {
		// 	if iconified {
		// 		// Window was iconified (ie minimized)
		// 	} else {
		// 		// Window was restored
		// 		gl.Viewport(0, 0, win.width, win.height)
		// 	}
		// })

		// TODO - other callbacks?

		// TODO - A hack for wasm - where the framebuffer doesn't trigger until the view gets resized once, we just set the size based on what the browser window says
		{
			w, h := win.window.GetFramebufferSize()
			win.width = w
			win.height = h
			gl.Viewport(0, 0, w, h)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	win.mainthreadUpdate = func() {
		// TODO - I think this is only useful for webgl because of how my RAF works I think in Firefox it sometimes swaps buffer before finishing the opengl stuff. Weird!
		// TODO - using gl.Finish is bad? https://www.khronos.org/opengl/wiki/Swap_Interval
		// gl.Flush()
		// gl.Finish()
		win.window.SwapBuffers()
		glfw.PollEvents()

		// errorEnum := gl.GetError()
		// if errorEnum != 0 {
		// 	fmt.Println("GL ERROR: ", errorEnum)
		// 	panic("GL ERROR")
		// }

		win.mainthreadCacheMousePosition()
	}

	win.mainthreadPressed = func() {
		win.pressed()
	}

	// Gets and caches the connected gamepads
	win.mainthreadUpdateConnectedGamepads = func() {
		joysticks := win.window.GetConnectedGamepads()

		win.connectedGamepads = win.connectedGamepads[:0]
		for i := range joysticks {
			win.connectedGamepads = append(win.connectedGamepads, Gamepad(joysticks[i]))
		}
	}

	win.Update()

	return win, nil
}

func (w *Window) Update() {
	// Track frame times
	nextLastUpdate := time.Now()
	dt := nextLastUpdate.Sub(w.lastUpdateTime)
	w.lastUpdateTime = nextLastUpdate

	global.finish()

	mainthread.Call(w.mainthreadUpdate)

	w.input = w.tmpInput
	w.tmpInput.scroll.X = 0
	w.tmpInput.scroll.Y = 0

	w.tmpInput.justPressed = [KeyLast + 1]bool{}
	w.tmpInput.justReleased = [KeyLast + 1]bool{}
	w.tmpInput.repeated = [KeyLast + 1]bool{}

	for i := range w.repeatTracker {
		key := w.repeatTracker[i].key

		if w.Pressed(key) {
			w.repeatTracker[i].dur += dt

			if w.repeatTracker[i].dur > w.mouseRepeatDelay {
				if w.repeatTracker[i].dur > w.mouseRepeatPeriod {
					w.repeatTracker[i].dur -= w.mouseRepeatPeriod
					w.tmpInput.repeated[key] = true
				}
			}
		} else {
			w.repeatTracker[i].dur = 0
		}
	}

	// --- Swap the typed buffers ---
	{
		backBuf := w.typedBack
		w.typedBack = w.typedFront[:0]
		w.typedFront = backBuf
	}

	// Note: Disabled this logic because the wasm performance is really bad
	// // --- Gamepad ---
	// // Recalculate all active gamepads
	// mainthread.Call(w.mainthreadUpdateConnectedGamepads)

	// // 1. I need to decide the current active gamepad
	// primaryGamepadState := w.getGamepadState(w.currentPrimaryGamepad)
	// primaryStillActive := checkGamepadActive(primaryGamepadState)
	// if !primaryStillActive {
	// 	newActiveGamepad := w.findNewActiveGamepad()
	// 	if newActiveGamepad != GamepadNone {
	// 		w.currentPrimaryGamepad = newActiveGamepad
	// 		primaryGamepadState = w.getGamepadState(w.currentPrimaryGamepad)
	// 	}
	// }

	// // 2. I need to track their input for JustPressed
	// // Note: Here we calculate the *newly* pressed buttons.
	// // - buttons that arent pressed in the last frame, but are pressed in this new gamepad state
	// if primaryGamepadState == nil {
	// 	// If the primary gamepad is nil, then set everything to false, it maybe got disconnected
	// 	w.pressedGamepad = [ButtonLast + 1]bool{}
	// 	w.justPressedGamepad = [ButtonLast + 1]bool{}
	// 	w.gamepadAxis = [AxisLast + 1]float64{}
	// } else {
	// 	for i := range primaryGamepadState.Buttons {
	// 		if w.pressedGamepad[i] {
	// 			// If currently pressed, then you couldn't possibly have just pressed it
	// 			w.justPressedGamepad[i] = false
	// 		} else {
	// 			// Else, If the new gamepad state has this button pressed, it means this is the first frame it was pressed
	// 			w.justPressedGamepad[i] = (primaryGamepadState.Buttons[i] == glfw.Press)
	// 		}

	// 		// Finally, set the current button state
	// 		w.pressedGamepad[i] = (primaryGamepadState.Buttons[i] == glfw.Press)
	// 	}

	// 	// And set the current axis state
	// 	for i := range primaryGamepadState.Axes {
	// 		w.gamepadAxis[i] = float64(primaryGamepadState.Axes[i])
	// 	}
	// }
}

func (w *Window) Closed() bool {
	shouldClose := false
	mainthread.Call(func() {
		shouldClose = w.window.ShouldClose()
	})
	if shouldClose {
		w.closed = true
	}
	return w.closed
}

func (w *Window) SetClose(close bool) {
	w.closed = close
	mainthread.Call(func() {
		w.window.SetShouldClose(close)
	})
}

func (w *Window) Close() {
	w.SetClose(true)
}

func (w *Window) Bounds() Rect {
	return glm.R(0, 0, float64(w.width), float64(w.height))
}

func (w *Window) MousePosition() (float64, float64) {
	// w.mainthreadCacheMousePosition() // This would decrease cursor lag a *little* bit
	return w.mousePosition.X, w.mousePosition.Y
}

func (w *Window) ContentScale() (float64, float64) {
	x, y := w.window.GetContentScale()
	return float64(x), float64(y)
}

func (w *Window) mainthreadCacheMousePosition() {
	var x, y float64
	var sx, sy float32
	x, y = w.window.GetCursorPos()

	// TODO - Use callback to get contentScale. There is a function available in glfw library. In javascript though, I'm not sure if there's a way to detect content scale (other than maybe in the framebuffer size callback) But if a window is dragged to another monitor which has a different content scale, then the framebuffer size callback may not trigger, but the content scale will be updated.
	sx, sy = w.window.GetContentScale()
	// We scale the mouse position (which is in window pixel coords) into framebuffer pixel coords by multiplying it by the content scale.

	xPos := x * float64(sx)
	yPos := float64(w.height) - (y * float64(sy)) // This flips the coordinate to quadrant 1
	// return xPos, yPos
	w.mousePosition.X = xPos
	w.mousePosition.Y = yPos

	// fmt.Println("ContentScale:", sx, sy)
	// fmt.Println("CursorPos:", x, y)
	// fmt.Println("FinalMousePos:", w.mousePosition)
}

// // Returns true if the key was pressed in the last frame
func (w *Window) JustPressed(key Key) bool {
	if key == KeyUnknown {
		return false
	}
	return w.input.justPressed[key]
}

func (w *Window) JustReleased(key Key) bool {
	if key == KeyUnknown {
		return false
	}
	return w.input.justReleased[key]
}

func (w *Window) Repeated(key Key) bool {
	if key == KeyUnknown {
		return false
	}
	return w.input.repeated[key]
}

// Binds the window as the OpenGL render targe
func (w *Window) Bind() {
	state.bindFramebuffer(gl.NoFramebuffer, w.Bounds())
}

// Reads a rectangle of the window's frame as a collection of bytes
// func (w *Window) ReadFrame(rect Rect, dst []byte) {
// 	mainthreadCall(func() {
// 		gl.BindFramebuffer(gl.FRAMEBUFFER, gl.NoFramebuffer)
// 		// TODO Note: https://docs.gl/es3/glReadPixels#:~:text=glReadPixels%20returns%20pixel%20data%20from,parameters%20are%20set%20with%20glPixelStorei.
// 		// Format and Type Enums define the expected pixel format and type to return to the byte buffer. Right now I have that hardcoded to gl.RGBA and gl.UNSIGNED_BYTE, respectively
// 		gl.ReadPixels(dst, int(rect.Min[0]), int(rect.Min[1]), int(rect.W()), int(rect.H()), gl.RGBA, gl.UNSIGNED_BYTE)
// 	})
// }

func (w *Window) Pressed(key Key) bool {
	if key == KeyUnknown {
		return false
	}
	// TODO: You could cache these every frame to avoid calling mainthread call (ie {unset, pressed, unpressed})

	w.pressedKeyCheck = key
	mainthread.Call(w.mainthreadPressed)
	return w.pressedKeyReturn
}

func (w *Window) pressed() {
	key := w.pressedKeyCheck

	var action glfw.Action
	if isMouseKey(key) {
		action = w.window.GetMouseButton(glfw.MouseButton(key))
	} else {
		action = w.window.GetKey(glfw.Key(key))
	}

	w.pressedKeyReturn = false
	if action == glfw.Press || action == glfw.Repeat {
		w.pressedKeyReturn = true
	}
}

// func (w *Window) Pressed(key Key) bool {
// 	var action glfw.Action
// 	mainthreadCall(func() {
// 		if isMouseKey(key) {
// 			action = w.window.GetMouseButton(glfw.MouseButton(key))
// 		} else {
// 			action = w.window.GetKey(glfw.Key(key))
// 		}
// 	})

// 	if action == glfw.Press || action == glfw.Repeat {
// 		return true
// 	}
// 	return false
// }

// Don't cache the returned buffer because it gets overwritten
func (w *Window) Typed() []rune {
	// TODO - should I copy just to be safe?
	return w.typedFront
}

func (w *Window) MouseScroll() (float64, float64) {
	return w.input.scroll.X, w.input.scroll.Y
}

type CursorMode uint8

const (
	CursorNormal   CursorMode = iota // A normal cursor
	CursorHidden                     // A normal cursor, but not rendered
	CursorDisabled                   // Hides and locks the cursor
)

func (w *Window) SetCursor(mode CursorMode) {
	mainthread.Call(func() {
		if mode == CursorNormal {
			w.window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		} else if mode == CursorHidden {
			w.window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
		} else if mode == CursorDisabled {
			w.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		}
	})
}

// Used to check if we are in browser and we are hidden. If not in a web browser this will always return false. Should be used to selectively disable code that shouldn't be run when the browser is hidden
func (w *Window) BrowserHidden() bool {
	return w.window.BrowserHidden()
}

// TODO - rename. Also potentially allow for multiple swaps?
func (w *Window) SetVSync(enable bool) {
	mainthread.Call(func() {
		if enable {
			w.vsync = true
			glfw.SwapInterval(1)
		} else {
			w.vsync = false
			glfw.SwapInterval(0)
		}
	})
}

// Returns true if vsync is enabled
func (w *Window) VSync() bool {
	return w.vsync
}

type ScreenModeType glfw.ScreenModeType

const (
	ScreenModeFull     = ScreenModeType(glfw.ScreenModeFull)
	ScreenModeWindowed = ScreenModeType(glfw.ScreenModeWindowed)
)

// TODO - rename. also maybe do modes - window, borderless, full, etc.
func (w *Window) SetScreenMode(smt ScreenModeType) {
	mainthread.Call(func() {
		w.window.SetScreenMode(glfw.ScreenModeType(smt))

		// Note: In windows, when going to fullscreen, it looks like vsync data gets lost during SetMonitor, So I will manually reset swapinterval here
		// GLFW Issue: https://github.com/glfw/glfw/issues/1072

		if w.vsync {
			glfw.SwapInterval(1)
		} else {
			glfw.SwapInterval(0)
		}
	})
}

func (w *Window) ScreenMode() ScreenModeType {
	return ScreenModeType(w.window.ScreenMode())
}

// Returns true if the window is embedded in an iframe, else returns false.
// Always returns false on desktop
func (w *Window) EmbeddedIframe() bool {
	return w.window.EmbeddedIframe()
}

func (w *Window) Add(filler GeometryFiller, mat glMat4, mask RGBA, material Material) {
	setTarget(w)
	global.Add(filler, mat, mask, material)
}

func (w *Window) GetConnectedGamepads() []Gamepad {
	return w.connectedGamepads
}

func (w *Window) findNewActiveGamepad() Gamepad {
	gamepads := w.GetConnectedGamepads()

	for _, gp := range gamepads {
		state := w.getGamepadState(gp)
		if checkGamepadActive(state) {
			return gp
		}
	}
	return GamepadNone
}

// Sets the icon of the window
// Note: Currently in browser, this does nothing
func (w *Window) SetIcon(images []image.Image) {
	mainthread.Call(func() {
		w.window.SetIcon(images)
	})
}

// --- Dear Imgui required ---
func (w *Window) GetMouse() (x, y float64) {
	return w.window.GetCursorPos()
}

func (w *Window) GetMouseButton(b glfw.MouseButton) glfw.Action {
	return w.window.GetMouseButton(b)
}

func (w *Window) DisplaySize() [2]float32 {
	// width, height := w.Window.GetSize()
	return [2]float32{float32(w.width), float32(w.height)}
}
func (w *Window) FramebufferSize() [2]float32 {
	return w.DisplaySize()
}

func (w *Window) AddScrollCallback(cb glfw.ScrollCallback) {
	w.scrollCallbacks = append(w.scrollCallbacks, cb)
	// fmt.Println("Adding new scroll callback. Currently: ", len(w.scrollCallbacks))
}

func (w *Window) AddKeyCallback(cb glfw.KeyCallback) {
	w.keyCallbacks = append(w.keyCallbacks, cb)
	// fmt.Println("Adding new key callback. Currently: ", len(w.keyCallbacks))
}

func (w *Window) AddCharCallback(cb glfw.CharCallback) {
	w.charCallbacks = append(w.charCallbacks, cb)
	// fmt.Println("Adding new char callback. Currently: ", len(w.charCallbacks))
}

func (w *Window) AddMouseButtonCallback(cb glfw.MouseButtonCallback) {
	w.mouseButtonCallbacks = append(w.mouseButtonCallbacks, cb)
	// fmt.Println("Adding new mouse button callback. Currently: ", len(w.mouseButtonCallbacks))
}
