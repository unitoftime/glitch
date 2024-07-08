package glitch

import (
	"image"
	"runtime"

	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/mainthread"
)

type Frame struct {
	fbo gl.Framebuffer
	*Batcher
	tex *Texture
	depth gl.Texture
	mesh *Mesh
	material Material
	bounds Rect
}

// Type? Color, depth, stencil?
func NewFrame(bounds Rect, smooth bool) *Frame {
	var frame = &Frame{
		bounds: bounds,
		Batcher: NewBatcher(),
	}
	frame.Batcher.target = frame

	// Create texture
	// TODO - Note: I'm passing actual data to the texture object, rather than null. That might be suboptimal. This fills the GPU memory, whereas if I pass null I can just allocate it.
	img := image.NewRGBA(image.Rect(int(bounds.Min.X), int(bounds.Min.Y), int(bounds.Max.X), int(bounds.Max.Y)))
	frame.tex = NewTexture(img, smooth)
	// frame.tex = NewEmptyTexture(int(bounds.W()), int(bounds.H()), smooth)

	// Create mesh (in case we want to draw the fbo to another target)
	// frame.mesh = NewQuadMesh(R(-1, -1, 1, 1), R(0, 1, 1, 0))
	frame.mesh = NewQuadMesh(bounds, R(0, 1, 1, 0))
	frame.material = NewSpriteMaterial(frame.tex)

	// frame.tex.Bind(0)///??????
	mainthread.Call(func() {
		frame.fbo = gl.CreateFramebuffer()
		gl.BindFramebuffer(gl.FRAMEBUFFER, frame.fbo)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, frame.tex.texture, 0)

		// https://webgl2fundamentals.org/webgl/lessons/webgl-render-to-texture.html
		// TODO - maybe centralize this into texture creation api
		// TODO - make fbo depth attachment optional
		frame.depth = gl.CreateTexture()
		gl.BindTexture(gl.TEXTURE_2D, frame.depth)
		// gl.TexImage2DFull(gl.TEXTURE_2D, 0, gl.DEPTH24_STENCIL8, frame.tex.width, frame.tex.height, gl.DEPTH_STENCIL, gl.UNSIGNED_INT_24_8, nil)
		gl.TexImage2DFull(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT24, frame.tex.width, frame.tex.height, gl.DEPTH_COMPONENT, gl.UNSIGNED_INT, nil)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, frame.depth, 0)
	})

	runtime.SetFinalizer(frame, (*Frame).delete)
	return frame
}

func (f *Frame) Bounds() Rect {
	return f.bounds
}

func (f *Frame) Texture() *Texture {
	return f.tex
}

func (f *Frame) Draw(target BatchTarget, matrix Mat4) {
	f.DrawColorMask(target, matrix, RGBA{1.0, 1.0, 1.0, 1.0})
}
func (f *Frame) DrawColorMask(target BatchTarget, matrix Mat4, mask RGBA) {
	// pass.SetTexture(0, s.texture)
	target.Add(f.mesh, matrix.gl(), mask, f.material, false)
}

func (f *Frame) RectDraw(target BatchTarget, bounds Rect) {
	f.RectDrawColorMask(target, bounds, White)
}

func (f *Frame) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	matrix := Mat4Ident
	matrix.Scale(bounds.W() / f.bounds.W(), bounds.H() / f.bounds.H(), 1).
		Translate(bounds.Min.X, bounds.Min.Y, 0)
	// Note: because frames are anchored to the bottom left, we don't have to shift by center
	// .Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)
	f.DrawColorMask(target, matrix, mask)
}


func (f *Frame) delete() {
	mainthread.CallNonBlock(func() {
		gl.DeleteFramebuffer(f.fbo)
	})
}

func (f *Frame) Bind() {
	state.bindFramebuffer(f.fbo, f.bounds)
}
