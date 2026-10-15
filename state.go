package glitch

// TODO - maybe push this up into internal/gl?
// TODO - this might lock us up into a single window? That doesn't seem like too bad of a requirement though

import (
	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/mainthread"
)

// The stateTracker is the single owner of GL binding state (render target + viewport, shader program, texture, depth, blend, cull).
// Every change to that state must go through it, otherwise its cache won't match the real GL state and needed binds will be skipped.
// The one exception is resource creation: code running on the mainthread may bind objects directly, as long as it calls the
// matching mainthreadApply* function before returning, which restores the tracked binding.
// Note: Vertex arrays and buffers are not tracked. They are always bound right before they are used.
type stateTracker struct {
	// Render Target (FBO + Viewport)
	target    Target // The currently bound target. Nil if the next setTarget must rebind
	fbo       gl.Framebuffer
	fboBounds Rect
	fboBinder func()

	// Shader Program
	shader       *Shader
	shaderBinder func()

	// Depth Test
	// depthTest       bool
	// enableDepthFunc func()
	// depthFunc       gl.Enum
	// depthFuncBinder func()
	depthMode       DepthMode
	depthModeBinder func()

	// Texture
	texture       *Texture
	textureBinder func()

	// BlendFunc
	// blendSrc, blendDst gl.Enum
	blendMode       BlendMode
	blendModeBinder func()

	// Cull Mode
	// cullModeEnable bool
	cullMode       CullMode
	cullModeBinder func()

	// Vert Buffer
	vertBuf       *VertexBuffer
	vertBufDrawer func()

	clearColor RGBA
	clearMode  gl.Enum
	clearFunc  func()
}

var state *stateTracker

func init() {
	state = &stateTracker{
		fbo: gl.NoFramebuffer,
	}

	state.depthModeBinder = func() {
		if state.depthMode == DepthModeNone {
			gl.Disable(gl.DEPTH_TEST)
		} else {
			gl.Enable(gl.DEPTH_TEST)
			data := depthModeLut[state.depthMode]
			gl.DepthFunc(data.mode)
		}
	}

	state.blendModeBinder = func() {
		if state.blendMode == BlendModeNone {
			gl.Disable(gl.BLEND)
			return
		}

		gl.Enable(gl.BLEND) // TODO: This only needs to run if it was disabled previously

		data := blendModeLut[state.blendMode]
		gl.BlendFunc(data.src, data.dst)
	}

	state.cullModeBinder = func() {
		if state.cullMode == CullModeNone {
			gl.Disable(gl.CULL_FACE)
		} else {
			gl.Enable(gl.CULL_FACE)
			data := cullModeLut[state.cullMode]
			gl.CullFace(data.face)
			gl.FrontFace(data.dir)
		}
	}

	// state.enableDepthFunc = func() {
	// 	if state.depthTest {
	// 		gl.Enable(gl.DEPTH_TEST)
	// 	} else {
	// 		gl.Disable(gl.DEPTH_TEST)
	// 	}
	// }

	// state.depthFuncBinder = func() {
	// 	gl.DepthFunc(state.depthFunc)
	// }

	state.fboBinder = func() {
		// TODO - Note: I set the viewport when I bind the framebuffer. Is this okay?
		gl.Viewport(0, 0, int(state.fboBounds.W()), int(state.fboBounds.H()))
		state.mainthreadApplyFramebuffer()
	}

	state.shaderBinder = state.mainthreadApplyShader

	state.textureBinder = func() {
		gl.ActiveTexture(gl.TEXTURE0)
		// gl.ActiveTexture(gl.TEXTURE0 + position); // TODO - include position
		state.mainthreadApplyTexture()
	}

	// state.blendFuncBinder = func() {
	// 	gl.BlendFunc(state.blendSrc, state.blendDst)
	// }

	// state.cullModeBinder = func() {
	// 	if state.cullModeEnable {
	// 		gl.Enable(gl.CULL_FACE)
	// 		gl.CullFace(state.cullMode.face)
	// 		gl.FrontFace(state.cullMode.dir)
	// 	} else {
	// 		gl.Disable(gl.CULL_FACE)
	// 	}
	// }

	state.vertBufDrawer = func() {
		state.vertBuf.mainthreadDraw()
	}

	state.clearFunc = func() {
		gl.ClearColor(float32(state.clearColor.R), float32(state.clearColor.G), float32(state.clearColor.B), float32(state.clearColor.A))
		// gl.Clear(gl.COLOR_BUFFER_BIT) // Make configurable?
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	}
}

// --- Mainthread: Apply the tracked bindings to GL ---
// Note: Resource creation uses these to restore the tracked bindings after binding objects directly

func (s *stateTracker) mainthreadApplyFramebuffer() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, s.fbo)
}

func (s *stateTracker) mainthreadApplyShader() {
	if s.shader == nil {
		gl.UseProgram(gl.NoProgram)
	} else {
		gl.UseProgram(s.shader.program)
	}
}

func (s *stateTracker) mainthreadApplyTexture() {
	if s.texture == nil {
		gl.BindTexture(gl.TEXTURE_2D, gl.NoTexture)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, s.texture.texture)
	}
}

// --- Tracked state changes ---

// TODO - maybe allow for more than 15 if the platform supports it? TODO - max texture units
func (s *stateTracker) bindTexture(texture *Texture) {
	if s.texture == texture {
		return // Skip: State already matches
	}
	s.texture = texture
	mainthread.Call(s.textureBinder)
}

// Note: Callers are responsible for checking if the shader has changed, because they need to flush before changing it
func (s *stateTracker) bindShader(shader *Shader) {
	s.shader = shader
	mainthread.Call(s.shaderBinder)
}

func (s *stateTracker) bindFramebuffer(target Target, fbo gl.Framebuffer, bounds Rect) {
	s.target = target
	if s.fbo.Equal(fbo) && s.fboBounds == bounds {
		return // Skip: State already matches
	}
	s.fbo = fbo
	s.fboBounds = bounds

	mainthread.Call(s.fboBinder)
}

// Forces the next setTarget to rebind its target, so that a changed target size (ie its viewport) gets applied
func (s *stateTracker) invalidateTarget() {
	s.target = nil
}

func (s *stateTracker) setDepthMode(depth DepthMode) {
	if s.depthMode == depth {
		return // Skip: State already matches
	}
	s.depthMode = depth

	mainthread.Call(s.depthModeBinder)
}

func (s *stateTracker) setBlendMode(blend BlendMode) {
	if s.blendMode == blend {
		return // Skip: State already matches
	}
	s.blendMode = blend

	mainthread.Call(s.blendModeBinder)
}

func (s *stateTracker) setCullMode(cull CullMode) {
	if s.cullMode == cull {
		return // Skip: State already matches
	}
	s.cullMode = cull

	mainthread.Call(s.cullModeBinder)
}

// func (s *stateTracker) enableDepthTest(enable bool) {
// 	if s.depthTest == enable {
// 		return // Skip if state already matches
// 	}
// 	s.depthTest = enable
// 	mainthread.Call(s.enableDepthFunc)
// }

// func (s *stateTracker) setDepthFunc(depthFunc gl.Enum) {
// 	if s.depthFunc == depthFunc {
// 		return // Skip if already enabled and depth functions match
// 	}

// 	s.depthFunc = depthFunc
// 	mainthread.Call(s.depthFuncBinder)
// }

// func (s *stateTracker) setBlendFunc(src, dst gl.Enum) {
// 	s.blendSrc = src
// 	s.blendDst = dst
// 	mainthread.Call(s.blendFuncBinder)
// }

func (s *stateTracker) drawVertBuffer(vb *VertexBuffer) {
	s.vertBuf = vb
	mainthread.Call(s.vertBufDrawer)
}

func (s *stateTracker) clearTarget(color RGBA) {
	s.clearColor = color
	mainthread.Call(s.clearFunc)
}

// func (s *stateTracker) disableCullMode() {
// 	if s.cullModeEnable == false {
// 		return // Skip if state already matches
// 	}

// 	s.cullModeEnable = false
// 	mainthread.Call(s.cullModeBinder)
// }

// func (s *stateTracker) enableCullMode(cullMode CullMode) {
// 	if s.cullModeEnable == true && s.cullMode == cullMode {
// 		return // Skip if state already matches
// 	}

// 	s.cullModeEnable = true
// 	s.cullMode = cullMode
// 	mainthread.Call(s.cullModeBinder)
// }
