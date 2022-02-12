package glitch

import (
	"image"
	"runtime"

	"github.com/faiface/mainthread"
	"github.com/unitoftime/gl"
)

type Frame struct {
	fbo gl.Framebuffer
	tex *Texture
	mesh *Mesh
	material Material
}

// Type? Color, depth, stencil?
func NewFrame(bounds Rect) *Frame {
	var frame Frame

	// Create texture
	// TODO - Note: I'm passing actual data to the texture object, rather than null. That might be suboptimal. This fills the GPU memory, whereas if I pass null I can just allocate it.
	img := image.NewNRGBA(image.Rect(int(bounds.Min[0]), int(bounds.Min[1]), int(bounds.Max[0]), int(bounds.Max[1])))
	frame.tex = NewTexture(img)

	// Create mesh (in case we want to draw the fbo to another target)
	frame.mesh = NewQuadMesh(R(-1, -1, 1, 1), R(0, 1, 1, 0))
	frame.material = NewSpriteMaterial(frame.tex)

	// frame.tex.Bind(0)///??????
	mainthread.Call(func() {
		frame.fbo = gl.CreateFramebuffer()
		gl.BindFramebuffer(gl.FRAMEBUFFER, frame.fbo)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, frame.tex.texture, 0)
	})

	runtime.SetFinalizer(&frame, (*Frame).delete)
	return &frame
}

func (f *Frame) Texture() *Texture {
	return f.tex
}

func (f *Frame) Draw(pass *RenderPass, matrix Mat4) {
	f.DrawColorMask(pass, matrix, RGBA{1.0, 1.0, 1.0, 1.0})
}
func (f *Frame) DrawColorMask(pass *RenderPass, matrix Mat4, mask RGBA) {
	// pass.SetTexture(0, s.texture)
	pass.Add(f.mesh, matrix, mask, f.material)
}

func (f *Frame) delete() {
	mainthread.CallNonBlock(func() {
		gl.DeleteFramebuffer(f.fbo)
	})
}

func (f *Frame) Bind() {
	mainthread.Call(func() {
		gl.BindFramebuffer(gl.FRAMEBUFFER, f.fbo)
	})
}

func (f *Frame) Clear() {
	mainthread.Call(func() {
		gl.BindFramebuffer(gl.FRAMEBUFFER, f.fbo)
	})
}
