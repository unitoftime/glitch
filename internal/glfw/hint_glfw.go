// +build !js

package glfw

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	True = glfw.True
	False = glfw.False
	OpenGLCoreProfile = glfw.OpenGLCoreProfile
)

type Hint int

const (
	AlphaBits   = Hint(glfw.AlphaBits)
	DepthBits   = Hint(glfw.DepthBits)
	StencilBits = Hint(glfw.StencilBits)
	Samples     = Hint(glfw.Samples)
	Resizable   = Hint(glfw.Resizable)

	ContextVersionMajor = Hint(glfw.ContextVersionMajor)
	ContextVersionMinor = Hint(glfw.ContextVersionMinor)
	OpenGLProfile = Hint(glfw.OpenGLProfile)
	OpenGLForwardCompatible = Hint(glfw.OpenGLForwardCompatible)

	// These hints used for WebGL contexts, ignored on desktop.
	PremultipliedAlpha = noopHint
	PreserveDrawingBuffer
	PreferLowPowerToHighPerformance
	FailIfMajorPerformanceCaveat

	Focused = Hint(glfw.Focused)
	Decorated = Hint(glfw.Decorated)
	Floating = Hint(glfw.Floating)
	AutoIconify = Hint(glfw.AutoIconify)
	TransparentFramebuffer = Hint(glfw.TransparentFramebuffer)
	Maximized = Hint(glfw.Maximized)
	Visible = Hint(glfw.Visible)
)

// noopHint is ignored.
const noopHint Hint = -1

func WindowHint(target Hint, hint int) {
	if target == noopHint {
		return
	}

	glfw.WindowHint(glfw.Hint(target), hint)
}
