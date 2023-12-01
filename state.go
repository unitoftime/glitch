package glitch

// TODO - maybe push this up into internal/gl?
// TODO - this might lock us up into a single window? That doesn't seem like too bad of a requirement though

import (
	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/mainthread"
)

type stateTracker struct {
	// FBO
	fbo gl.Framebuffer
	fboBounds Rect
	fboBinder func()

	// Depth Test
	depthTest bool
	depthFunc gl.Enum
	depthFuncBinder func()

	// Texture
	texture *Texture
	textureBinder func()

	// BlendFunc
	blendSrc, blendDst gl.Enum
	blendFuncBinder func()

	vertBuf *VertexBuffer
	vertBufDrawer func()
}

var state *stateTracker
func init() {
	state = &stateTracker{
	}

	state.depthFuncBinder = func() {
		gl.Enable(gl.DEPTH_TEST)
		gl.DepthFunc(state.depthFunc)
	}

	state.fboBinder = func() {
		// TODO - Note: I set the viewport when I bind the framebuffer. Is this okay?
		gl.Viewport(0, 0, int(state.fboBounds.W()), int(state.fboBounds.H()))
		gl.BindFramebuffer(gl.FRAMEBUFFER, state.fbo)
	}

	state.textureBinder = func() {
		gl.ActiveTexture(gl.TEXTURE0);
		// gl.ActiveTexture(gl.TEXTURE0 + position); // TODO - include position
		gl.BindTexture(gl.TEXTURE_2D, state.texture.texture)
	}

	state.blendFuncBinder = func() {
		gl.BlendFunc(state.blendSrc, state.blendDst)
	}

	state.vertBufDrawer = func() {
		state.vertBuf.mainthreadDraw()
	}
}

func (s *stateTracker) bindTexture(texture *Texture) {
	// TODO: Check if changed
	s.texture = texture
	mainthread.Call(state.textureBinder)
}

func (s *stateTracker) bindFramebuffer(fbo gl.Framebuffer, bounds Rect) {
	if s.fbo.Equal(fbo) && s.fboBounds == bounds {
		return
	}
	state.fbo = fbo
	state.fboBounds = bounds

	mainthread.Call(s.fboBinder)
}

// TODO: Enable disable
// 		gl.Enable(gl.DEPTH_TEST)
func (s *stateTracker) enableDepthTest(depthFunc gl.Enum) {
	if s.depthTest && s.depthFunc == depthFunc {
		return // Skip if already enabled and depth functions match
	}
	s.depthTest = true
	s.depthFunc = depthFunc

	mainthread.Call(s.depthFuncBinder)
}

func (s *stateTracker) setBlendFunc(src, dst gl.Enum) {
	s.blendSrc = src
	s.blendDst = dst
	mainthread.Call(s.blendFuncBinder)
}

func (s *stateTracker) drawVertBuffer(vb *VertexBuffer) {
	s.vertBuf = vb
	mainthread.Call(s.vertBufDrawer)
}
