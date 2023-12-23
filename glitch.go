package glitch

import (
	"github.com/unitoftime/glitch/internal/mainthread"
)

func Run(function func()) {
	mainthread.Run(function)
}

// const batchSizeTris int = 1000
// const batchSizeVerts int = 1000

// type renderContext struct {
// 	target *Window
// 	shader *Shader
// 	textureUnit []*Texture
// 	bufferPool map[*Shader]*VertexBuffer
// 	vertexBuffer *VertexBuffer
// }

// var context renderContext

// func init() {
// 	// context = renderContext{
// 	// 	target: nil,
// 	// 	shader: nil,
// 	// 	textureUnit: make([]*Texture, 16), // TODO - can I get this from opengl?
// 	// 	bufferPool: make(map[*Shader]*VertexBuffer),
// 	// 	vertexBuffer: nil,
// 	// }

// 	targetClearer.Func = func() {
// 		targetClearer.Run()
// 	}
// }

// Sets the current target
// func SetTarget(win *Window) {
// 	context.target = win
// 	// TODO - set framebuffer or w/e
// }

// Clears the current target
// TODO - depthbuffer and stuff?
// func Clear(rgba RGBA) {
// context.target.Clear(rgba)
// }

// type targetClear struct {
// 	color RGBA
// 	Func  func()
// }

// var targetClearer targetClear

// func (t *targetClear) Run() {
// 	color := t.color
// 	gl.ClearColor(float32(color.R), float32(color.G), float32(color.B), float32(color.A))
// 	// gl.Clear(gl.COLOR_BUFFER_BIT) // Make configurable?
// 	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
// }

// func Clear(target Target, color RGBA) {
// 	target.Bind() // TODO: Push into state tracker?

// 	targetClearer.color = color
// 	mainthread.Call(targetClearer.Func)
// }

func Clear(target Target, color RGBA) {
	target.Bind() // TODO: Push into state tracker?
	state.clearTarget(color)
}

type Material interface {
	Bind()
}

type SpriteMaterial struct {
	texture *Texture
}

func NewSpriteMaterial(texture *Texture) SpriteMaterial {
	return SpriteMaterial{
		texture: texture,
	}
}

func (m SpriteMaterial) Bind() {
	m.texture.Bind(0) // Direct opengl? Or should I call through shader?
	// pass.SetTexture(0, m.texture) // TODO - hardcoded slot?
}

func DefaultMaterial() SpriteMaterial {
	return SpriteMaterial{
		texture: WhiteTexture(),
	}
}
