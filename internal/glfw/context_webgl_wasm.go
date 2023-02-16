// +build js,wasm

package glfw

import (
	"errors"
	"syscall/js"
)

func newContext(canvas js.Value, ca *contextAttributes) (context js.Value, err error) {
	if js.Global().Get("WebGLRenderingContext").Equal(js.Undefined()) {
		return js.Value{}, errors.New("Your browser doesn't appear to support WebGL.")
	}

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

	// if gl.Equal(js.Null()) {
	// 	// If gl context is null, then webgl2 creation failed. Let's try webgl1
	// 	log.Println("Failed to create Webgl2, trying webgl1")
	// 	gl = canvas.Call("getContext", "webgl", attrs)
	// }

	if !gl.Equal(js.Null()) {
		debug := js.Global().Get("WebGLDebugUtils")
		if debug.Equal(js.Undefined()) {
			return gl, errors.New("No debugging for WebGL.")
		}
		gl = debug.Call("makeDebugContext", gl)
		return gl, nil
	} else if gl := canvas.Call("getContext", "experimental-webgl", attrs); gl.Equal(js.Null()) {
		// log.Println("Failed to create, trying experimental-webgl")
		return gl, nil
	} else {
		return js.Value{}, errors.New("Creating a WebGL context has failed.")
	}
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
