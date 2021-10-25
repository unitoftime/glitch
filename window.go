package glitch

import (
	"fmt"
	"github.com/faiface/mainthread"

	"github.com/jstewart7/glfw"
	"github.com/jstewart7/gl"
)


type WindowConfig struct {
	Vsync bool
	// Resizable bool
	// Samples int
}

type Window struct {
	window *glfw.Window

	width, height int
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
		// glfw.WindowHint(glfw.Samples, config.Samples)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True) // Compatibility - For Mac only?

		// TODO - For fullscreen: glfw.GetPrimaryMonitor()
		win.window, err = glfw.CreateWindow(width, height, title, nil, nil)
		if err != nil {
			return err
		}

		win.window.MakeContextCurrent()

		// log.Printf("OpenGL: %s %s %s; %v samples.\n", gl.GetString(gl.VENDOR), gl.GetString(gl.RENDERER), gl.GetString(gl.VERSION), gl.GetInteger(gl.SAMPLES))
		// log.Printf("GLSL: %s.\n", gl.GetString(gl.SHADING_LANGUAGE_VERSION))

		// gl.Enable(gl.DEPTH_TEST)
		// gl.Enable(gl.MULTISAMPLE) // TODO - reenable? But how to work with wasm (which enables multisample in the context?)
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

		gl.Viewport(0, 0, int(width), int(height))

		win.window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
			// log.Println("Framebuffer size callback")
			win.width = width
			win.height = height
			// gl.Viewport(int32(0), int32(0), int32(win.width), int32(win.height))
			gl.Viewport(0, 0, int(win.width), int(win.height))
		})

		// TODO - other callbacks?

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed CreateWindow: %w", err)
	}
	return win, nil
}

func (w *Window) Update() {
	mainthread.Call(func() {
		w.window.SwapBuffers()
		glfw.PollEvents()
	})
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

func (w *Window) MousePosition() (float64, float64) {
	var x, y float64
	mainthread.Call(func() {
		x, y = w.window.GetCursorPos()
	})
	return x,y
}

// // Returns true if the key was pressed in the last frame
// func (w *Window) JustPressed(key Key) bool {
	// TODO 
// }

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
