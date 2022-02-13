package glitch

import (
	"fmt"
	"github.com/faiface/mainthread"

	"github.com/unitoftime/glfw"
	"github.com/unitoftime/gl"
)


type WindowConfig struct {
	Vsync bool
	// Resizable bool
	Samples int
}

type Window struct {
	window *glfw.Window

	width, height int

	tmpInput, input struct {
		scroll struct {
			X, Y float64
		}
	}
}

func NewWindow(width, height int, title string, config WindowConfig) (*Window, error) {
	win := &Window{}

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

		// TODO - For fullscreen: glfw.GetPrimaryMonitor()
		// monitor := glfw.GetPrimaryMonitor()
		// win.window, err = glfw.CreateWindow(width, height, title, monitor, nil)
		win.window, err = glfw.CreateWindow(width, height, title, nil, nil)
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
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

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

		// TODO - other callbacks?

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed CreateWindow: %w", err)
	}

	win.Update()
	return win, nil
}

func (w *Window) Update() {
	mainthread.Call(func() {
		w.window.SwapBuffers()
		glfw.PollEvents()
	})

	w.input = w.tmpInput
	w.tmpInput.scroll.X = 0
	w.tmpInput.scroll.Y = 0
}

func (w *Window) Close() {
	mainthread.Call(func() {
		w.window.SetShouldClose(true)
	})
}

func (w *Window) ShouldClose() bool {
	var value bool
	mainthread.Call(func() {
		value = w.window.ShouldClose()
	})
	return value
}

func (w *Window) Bounds() Rect {
	return R(0, 0, float32(w.width), float32(w.height))
}

func (w *Window) MousePosition() (float32, float32) {
	var x, y float64
	mainthread.Call(func() {
		x, y = w.window.GetCursorPos()
	})
	return float32(x), float32(float64(w.height) - y) // This flips the coordinate to quadrant 1
}

// // Returns true if the key was pressed in the last frame
// func (w *Window) JustPressed(key Key) bool {
	
// }

// Binds the window as the OpenGL render targe
func (w *Window) Bind() {
	mainthread.Call(func() {
		// Note: 0 (gl.NoFramebuffer) is the window's framebuffer
		gl.BindFramebuffer(gl.FRAMEBUFFER, gl.NoFramebuffer)
	})
}

func (w *Window) Pressed(key Key) bool {
	var action glfw.Action
	mainthread.Call(func() {
		if isMouseKey(key) {
			action = w.window.GetMouseButton(glfw.MouseButton(key))
		} else {
			action = w.window.GetKey(glfw.Key(key))
		}
	})

	if action == glfw.Press || action == glfw.Repeat {
		return true
	}
	return false
}

func (w *Window) MouseScroll() (float64, float64) {
	return w.input.scroll.X, w.input.scroll.Y
}
