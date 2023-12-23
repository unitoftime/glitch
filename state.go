package glitch

// TODO - maybe push this up into internal/gl?
// TODO - this might lock us up into a single window? That doesn't seem like too bad of a requirement though

import (
	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/mainthread"
)

type stateTracker struct {
	// FBO
	fbo       gl.Framebuffer
	fboBounds Rect
	fboBinder func()

	// Depth Test
	depthTest       bool
	enableDepthFunc func()
	depthFunc       gl.Enum
	depthFuncBinder func()

	// Texture
	texture       *Texture
	textureBinder func()

	// BlendFunc
	blendSrc, blendDst gl.Enum
	blendFuncBinder    func()

	// Vert Buffer
	vertBuf       *VertexBuffer
	vertBufDrawer func()

	clearColor RGBA
	clearMode  gl.Enum
	clearFunc  func()
}

var state *stateTracker

func init() {
	state = &stateTracker{}

	state.enableDepthFunc = func() {
		if state.depthTest {
			gl.Enable(gl.DEPTH_TEST)
		} else {
			gl.Disable(gl.DEPTH_TEST)
		}
	}

	state.depthFuncBinder = func() {
		gl.DepthFunc(state.depthFunc)
	}

	state.fboBinder = func() {
		// TODO - Note: I set the viewport when I bind the framebuffer. Is this okay?
		gl.Viewport(0, 0, int(state.fboBounds.W()), int(state.fboBounds.H()))
		gl.BindFramebuffer(gl.FRAMEBUFFER, state.fbo)
	}

	state.textureBinder = func() {
		gl.ActiveTexture(gl.TEXTURE0)
		// gl.ActiveTexture(gl.TEXTURE0 + position); // TODO - include position
		gl.BindTexture(gl.TEXTURE_2D, state.texture.texture)
	}

	state.blendFuncBinder = func() {
		gl.BlendFunc(state.blendSrc, state.blendDst)
	}

	state.vertBufDrawer = func() {
		state.vertBuf.mainthreadDraw()
	}

	state.clearFunc = func() {
		gl.ClearColor(float32(state.clearColor.R), float32(state.clearColor.G), float32(state.clearColor.B), float32(state.clearColor.A))
		// gl.Clear(gl.COLOR_BUFFER_BIT) // Make configurable?
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
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

func (s *stateTracker) enableDepthTest(enable bool) {
	if s.depthTest == enable {
		return // Skip if state already matches
	}
	s.depthTest = enable
	mainthread.Call(s.enableDepthFunc)
}

func (s *stateTracker) setDepthFunc(depthFunc gl.Enum) {
	if s.depthFunc == depthFunc {
		return // Skip if already enabled and depth functions match
	}

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

func (s *stateTracker) clearTarget(color RGBA) {
	s.clearColor = color
	mainthread.Call(s.clearFunc)
}
