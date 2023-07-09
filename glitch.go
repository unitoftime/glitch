package glitch

import (
	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/mainthread"
)

func Run(function func()) {
	mainthread.Run(function)
}

// ***** Should these all be a part of the shader object?
// Render same geom with diff shader. Ie batching across shaders?
// - Shaders need to share same vertex format
// {
// 	setshader(1)
// 	drawGeom(a)
// 	drawGeom(b)
// 	drawGeom(c)
// 	setshader(2)
// 	drawGeom(a)
// 	drawGeom(b)
// 	drawGeom(c)
// 	setshader(3)
// 	drawGeom(a)
// 	drawGeom(b)
// 	drawGeom(c)

// 	// 1. Redraw function to draw exactly what the last shader had
// 	//   - Pro: Simple
// 	//   - Con: Prevents user from doing partial draws
// 	setshader(1)
// 	drawGeom(a)
// 	drawGeom(b)
// 	drawGeom(c)
// 	setshader(2)
// 	redraw()
// 	setshader(3)
// 	redraw()

// 	// 2. Specify batches manually? - Choosing this because I can still batch stuff as needed in a single shader. but this lets me batch select things across single shaders
// 	// - Pro: also kinda simple
// 	// - Con: Pushes responsibility to user for how they batch.
// 	z := NewGroup(shader, a, b, c)
// 	setshader(1)
// 	drawGroup(z)
// 	setshader(2)
// 	drawGroup(z)
// 	setshader(3)
// 	drawGroup(z)

// 	// 3. Pull in all of the data and track everything then at some final signal send everything to the gpu. Then batch things that have some sort of similarity based on an algo
// 	// - Pro: Really hard
// 	// - Con: Simple for user. Not much control over batching though
// }

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

func init() {
	// context = renderContext{
	// 	target: nil,
	// 	shader: nil,
	// 	textureUnit: make([]*Texture, 16), // TODO - can I get this from opengl?
	// 	bufferPool: make(map[*Shader]*VertexBuffer),
	// 	vertexBuffer: nil,
	// }

	targetClearer.Func = func() {
		targetClearer.Run()
	}
}

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

type targetClear struct {
	color RGBA
	Func func()
}
var targetClearer targetClear

func (t *targetClear) Run() {
	color := t.color
	gl.ClearColor(float32(color.R), float32(color.G), float32(color.B), float32(color.A))
	// gl.Clear(gl.COLOR_BUFFER_BIT) // Make configurable?
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func Clear(target Target, color RGBA) {
	target.Bind()

	targetClearer.color = color
	mainthread.Call(targetClearer.Func)
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
