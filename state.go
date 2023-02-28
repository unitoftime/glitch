package glitch
// TODO - maybe push this up into internal/gl?
// TODO - this might lock us up into a single window? That doesn't seem like too bad of a requirement though

import (
	"github.com/unitoftime/glitch/internal/gl"
)

type stateTracker struct {
	fbo gl.Framebuffer
	fboBounds Rect
	fboBinder func()

	depthTest bool
	depthFunc gl.Enum
}

var state *stateTracker
func init() {
	state = &stateTracker{
	}

	state.fboBinder = func() {
		// TODO - Note: I set the viewport when I bind the framebuffer. Is this okay?
		gl.Viewport(0, 0, int(state.fboBounds.W()), int(state.fboBounds.H()))
		gl.BindFramebuffer(gl.FRAMEBUFFER, state.fbo)
	}
}

func (s *stateTracker) bindFramebuffer(fbo gl.Framebuffer, bounds Rect) {
	if s.fbo.Equal(fbo) && s.fboBounds == bounds {
		return
	}
	state.fbo = fbo
	state.fboBounds = bounds

	mainthreadCall(s.fboBinder)
}

func (s *stateTracker) enableDepthTest(depthFunc gl.Enum) {
	if s.depthTest && s.depthFunc == depthFunc {
		return // Skip if already enabled and depth functions match
	}
	s.depthTest = true
	s.depthFunc = depthFunc

	mainthreadCall(func() {
		gl.Enable(gl.DEPTH_TEST)
		gl.DepthFunc(depthFunc)
	})
}