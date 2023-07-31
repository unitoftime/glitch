// +build js,wasm

package glfw

import (
	"fmt"
	"syscall/js"
)


type webglSupport struct {
	WebGLRenderingContext bool // WebGL1
	WebGLRenderingContext2 bool // WebGl2
}
func (s webglSupport) String() string {
	return fmt.Sprintf(
		"WebGLRenderingContext: %v | WebGLRenderingContext2: %v",
		s.WebGLRenderingContext, s.WebGLRenderingContext2,
	)
}

func getWebglSupport() webglSupport {
	support := webglSupport{}

	{
		webgl := js.Global().Get("WebGLRenderingContext")
		support.WebGLRenderingContext = false
		if !webgl.Equal(js.Null()) {
			if !webgl.Equal(js.Undefined()) {
				support.WebGLRenderingContext = true
			}
		}
	}

	{
		webgl2 := js.Global().Get("WebGLRenderingContext2")
		support.WebGLRenderingContext2 = false
		if !webgl2.Equal(js.Null()) {
			if !webgl2.Equal(js.Undefined()) {
				support.WebGLRenderingContext2 = true
			}
		}
	}

	return support
}

func newContext(canvas js.Value, ca *contextAttributes) (js.Value, error) {
	support := getWebglSupport()

	attrs := map[string]interface{}{
		"alpha":                           ca.Alpha,
		"depth":                           ca.Depth,
		"stencil":                         ca.Stencil,
		"antialias":                       ca.Antialias,
		"premultipliedAlpha":              ca.PremultipliedAlpha,
		"preserveDrawingBuffer":           ca.PreserveDrawingBuffer,
		"preferLowPowerToHighPerformance": ca.PreferLowPowerToHighPerformance,
		"failIfMajorPerformanceCaveat":    ca.FailIfMajorPerformanceCaveat,
	}

	gl := canvas.Call("getContext", "webgl2", attrs)
	if !gl.Equal(js.Null()) {
		return gl, nil
	}

	// if !gl.Equal(js.Null()) {
	// 	//		ext := gl.Call("getExtension", "WEBGL_lose_context")
	// 	// TODO - this isn't working?
	// 	ext := gl.Call("getExtension", "WEBKIT_WEBGL_depth_texture")
	// 	//		ext := gl.Call("getExtension", "WEBGL_depth_texture") // windows?
	// 	log.Println("DepthTexture Extension: ", ext)
	// }

	// --- Fallbacks ---
	// TODO: Is this worth falling back to?
	// fmt.Println("Failed to create webgl2 context, trying experimental-webgl2")
	// gl = canvas.Call("getContext", "experimental-webgl2", attrs)
	// if !gl.Equal(js.Null()) {
	// 	return gl, nil
	// }

	fmt.Println("Failed to create webgl2 context, trying webgl1")
	gl = canvas.Call("getContext", "webgl", attrs)
	if !gl.Equal(js.Null()) {
		return gl, nil
	}

	fmt.Println("Failed to create webgl2 and webgl1 context, trying experimental-webgl")
	gl = canvas.Call("getContext", "experimental-webgl", attrs)
	if !gl.Equal(js.Null()) {
		return gl, nil
	}

	// TODO: Not sure what WebGLDebugUtils is, or if its worth falling back to
	// if !gl.Equal(js.Null()) {
	// 	debug := js.Global().Get("WebGLDebugUtils")
	// 	if debug.Equal(js.Undefined()) {
	// 		return gl, errors.New("No debugging for WebGL.")
	// 	}
	// 	gl = debug.Call("makeDebugContext", gl)
	// 	return gl, nil
	// }

	return js.Value{}, fmt.Errorf("webgl context creation error. Browser Support: %s", support.String())
}

type contextAttributes struct {
	Alpha                           bool
	Depth                           bool
	Stencil                         bool
	Antialias                       bool
	PremultipliedAlpha              bool
	PreserveDrawingBuffer           bool
	PreferLowPowerToHighPerformance bool
	FailIfMajorPerformanceCaveat    bool
}

// https://www.glfw.org/docs/3.3/window_guide.html
func defaultAttributes() *contextAttributes {
	return &contextAttributes{
		Alpha:                 false,
		Depth:                 true,
		Stencil:               false,
		Antialias:             false,
		PremultipliedAlpha:    false,
		PreserveDrawingBuffer: false,
	}
}
