package glitch

import (
	"github.com/unitoftime/gl"
	"github.com/faiface/mainthread"

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

const batchSizeTris int = 1000
const batchSizeVerts int = 1000

type renderContext struct {
	target *Window
	shader *Shader
	textureUnit []*Texture
	bufferPool map[*Shader]*VertexBuffer
	vertexBuffer *VertexBuffer
}

var context renderContext

func init() {
	context = renderContext{
		target: nil,
		shader: nil,
		textureUnit: make([]*Texture, 16), // TODO - can I get this from opengl?
		bufferPool: make(map[*Shader]*VertexBuffer),
		vertexBuffer: nil,
	}
}

// Sets the current target
func SetTarget(win *Window) {
	context.target = win
	// TODO - set framebuffer or w/e
}

// Clears the current target
// TODO - depthbuffer and stuff?
// func Clear(rgba RGBA) {
// context.target.Clear(rgba)
// }
func Clear(target Target, color RGBA) {
	target.Bind()
	mainthread.Call(func() {
		gl.ClearColor(color.R, color.G, color.B, color.A)
		// gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
// TODO - depth buffer bit?		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	})
}

/*
func FinalizeDraw() {
	context.vertexBuffer.Bind()
	context.vertexBuffer.Draw()
}

func SetShader(shader *Shader) {
	context.shader = shader
	context.shader.Bind()
	_, ok := context.bufferPool[shader]
	if !ok {
		context.bufferPool[shader] = NewVertexBuffer(shader, batchSizeVerts, batchSizeTris)
	}
	context.vertexBuffer, _ = context.bufferPool[shader]
}

// Instead of set camera?
// Why have this? Why not do shader.SetUniform()?
// func SetUniform(name string, value interface{}) {
// 	context.shader.SetUniform(name, value)
// }

// https://www.khronos.org/opengl/wiki/Shader#Resource_limitations
func SetTexture(position int, texture *Texture) {
	if position >= len(context.textureUnit) {
		panic("Can't support this many texture units!")
	}
	context.textureUnit[position] = texture
	context.textureUnit[position].Bind(position)
}

// Draws a mesh based on the currently set shader
func Draw(mesh *Mesh, mat Mat4) {
	positions := make([]float32, len(mesh.positions) * 3) // 3 b/c vec3
	for i := range mesh.positions {
		vec := mat.MulVec3(&mesh.positions[i])
		positions[(i * 3) + 0] = vec[0]
		positions[(i * 3) + 1] = vec[1]
		positions[(i * 3) + 2] = vec[2]
	}
	context.vertexBuffer.Add(positions, mesh.colors, mesh.texCoords, mesh.indices)
}
*/

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

// type Model struct {
// 	meshes []Mesh
// 	materials []Material
// }
