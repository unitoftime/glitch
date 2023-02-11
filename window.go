package glitch

import (
	"fmt"

	"github.com/unitoftime/glfw"
	"github.com/unitoftime/gl"
)


type WindowConfig struct {
	Fullscreen bool
	Vsync bool
	// Resizable bool
	Samples int
}

type Window struct {
	window *glfw.Window

	width, height int

	tmpInput, input struct {
		justPressed [KeyLast + 1]bool
		justReleased [KeyLast + 1]bool
		repeated [KeyLast + 1]bool

		scroll struct {
			X, Y float64
		}
	}

	mousePosition Vec2

	// The back and front buffers for tracking typed characters
	typedBack, typedFront []rune

	mainthreadUpdate func()
	mainthreadBind func()
	mainthreadPressed func()
	pressedKeyCheck Key
	pressedKeyReturn bool
}

func NewWindow(width, height int, title string, config WindowConfig) (*Window, error) {
	win := &Window{}

	err := mainthreadCallErr(func() error {
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

		var monitor *glfw.Monitor
		if config.Fullscreen {
			monitor = glfw.GetPrimaryMonitor()
		}
		win.window, err = glfw.CreateWindow(width, height, title, monitor, nil)
		if err != nil {
			return err
		}

		win.window.MakeContextCurrent()

		// log.Printf("OpenGL: %s %s %s; %v samples.\n", gl.GetString(gl.VENDOR), gl.GetString(gl.RENDERER), gl.GetString(gl.VERSION), gl.GetInteger(gl.SAMPLES))
		// log.Printf("GLSL: %s.\n", gl.GetString(gl.SHADING_LANGUAGE_VERSION))

		if config.Samples > 0 {
			// TODO - But how to work with wasm (which enables multisample in the context?)
			gl.Enable(gl.MULTISAMPLE)
		}

		// gl.Enable(gl.DEPTH_TEST)
		// gl.Enable(gl.CULL_FACE)
		// gl.CullFace(gl.BACK)
		// gl.FrontFace(gl.CCW) // Default

		gl.Enable(gl.BLEND)
		// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA); // Non premult
		gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA); // Premult

		if config.Vsync {
			glfw.SwapInterval(1)
		} else {
			glfw.SwapInterval(0)
		}

		win.width = width
		win.height = height
		gl.Viewport(0, 0, int(width), int(height))

		win.window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
			// log.Println("Framebuffer size callback")
			win.width = width
			win.height = height
			gl.Viewport(0, 0, int(win.width), int(win.height))
		})

		win.window.SetScrollCallback(func(w *glfw.Window, xoff float64, yoff float64) {
			win.tmpInput.scroll.X += xoff
			win.tmpInput.scroll.Y += yoff
		})

		win.window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
			switch action {
			case glfw.Press:
				win.tmpInput.justPressed[Key(button)] = true
			case glfw.Release:
				win.tmpInput.justReleased[Key(button)] = true
			}
		})

		win.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
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

		win.window.SetCharCallback(func(w *glfw.Window, char rune) {
			win.typedBack = append(win.typedBack, char)
		})

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
		return nil, fmt.Errorf("Failed CreateWindow: %w", err)
	}

	win.mainthreadUpdate = func() {
		win.window.SwapBuffers()
		glfw.PollEvents()

		win.mainthreadCacheMousePosition()
	}

	win.mainthreadBind = func() {
		win.bind()
	}

	win.mainthreadPressed = func() {
		win.pressed()
	}

	win.Update()

	return win, nil
}

func (w *Window) Update() {
	mainthreadCall(w.mainthreadUpdate)

	w.input = w.tmpInput
	w.tmpInput.scroll.X = 0
	w.tmpInput.scroll.Y = 0

	w.tmpInput.justPressed = [KeyLast + 1]bool{}
	w.tmpInput.justReleased = [KeyLast + 1]bool{}
	w.tmpInput.repeated = [KeyLast + 1]bool{}

	// Swap the typed buffers
	{
		backBuf := w.typedBack
		w.typedBack = w.typedFront[:0]
		w.typedFront = backBuf
	}
}

func (w *Window) Close() {
	mainthreadCall(func() {
		w.window.SetShouldClose(true)
	})
}

func (w *Window) ShouldClose() bool {
	var value bool
	mainthreadCall(func() {
		value = w.window.ShouldClose()
	})
	return value
}

func (w *Window) Bounds() Rect {
	return R(0, 0, float64(w.width), float64(w.height))
}

func (w *Window) MousePosition() (float64, float64) {
	return w.mousePosition[0], w.mousePosition[1]
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
	w.mousePosition[0] = xPos
	w.mousePosition[1] = yPos
}

// // Returns true if the key was pressed in the last frame
func (w *Window) JustPressed(key Key) bool {
	return w.input.justPressed[key]
}

func (w *Window) Repeated(key Key) bool {
	return w.input.repeated[key]
}

// Binds the window as the OpenGL render targe
func (w *Window) Bind() {
	mainthreadCall(w.mainthreadBind)
}
func (w *Window) bind() {
	// TODO - Note: I set the viewport when I bind the framebuffer. Is this okay?
	gl.Viewport(0, 0, int(w.width), int(w.height))
	// Note: 0 (gl.NoFramebuffer) is the window's framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, gl.NoFramebuffer)
}

// func (w *Window) Bind() {
// 	mainthreadCall(func() {
// 		// TODO - Note: I set the viewport when I bind the framebuffer. Is this okay?
// 		gl.Viewport(0, 0, int(w.width), int(w.height))
// 		// Note: 0 (gl.NoFramebuffer) is the window's framebuffer
// 		gl.BindFramebuffer(gl.FRAMEBUFFER, gl.NoFramebuffer)
// 	})
// }

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
	w.pressedKeyCheck = key
	mainthreadCall(w.mainthreadPressed)
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
	CursorNormal CursorMode = iota // A normal cursor
	CursorHidden // A normal cursor, but not rendered
	CursorDisabled // Hides and locks the cursor
)
func (w *Window) SetCursor(mode CursorMode) {
	mainthreadCall(func() {
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
