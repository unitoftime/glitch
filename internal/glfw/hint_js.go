// +build js

package glfw

var hints = make(map[Hint]int)

const (
	True = iota
	False
	OpenGLCoreProfile
)

type Hint int

const (
	AlphaBits Hint = iota
	DepthBits
	StencilBits
	Samples
	Resizable

	ContextVersionMajor // TODO
	ContextVersionMinor // TODO
	OpenGLProfile // TODO
	OpenGLForwardCompatible // TODO

	// goxjs/glfw-specific hints for WebGL.
	PremultipliedAlpha
	PreserveDrawingBuffer
	PreferLowPowerToHighPerformance
	FailIfMajorPerformanceCaveat

	Decorated // TODO
	Floating // TODO
	AutoIconify // TODO
	TransparentFramebuffer // TODO
	Maximized // TODO
	Visible // TODO
)

func WindowHint(target Hint, hint int) {
	hints[target] = hint
}
