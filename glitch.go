package glitch

import (
	"github.com/jstewart7/gl"
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
func Clear(color RGBA) {
	mainthread.Call(func() {
		gl.ClearColor(color.R, color.G, color.B, color.A)
		gl.Clear(gl.COLOR_BUFFER_BIT)
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

type Sprite struct {
	mesh *Mesh
	bounds Rect
	texture *Texture
	textureMatrix Mat3
}

func NewSprite(texture *Texture, bounds Rect) *Sprite {
	texMat := Mat3Ident
	ScaleMat3(&texMat,
		bounds.W() / float32(texture.width),
		bounds.H() / float32(texture.height),
		1.0)
	TranslateMat3(&texMat,
		bounds.Min[0] / float32(texture.width),
		bounds.Min[1] / float32(texture.height),
	)

	return &Sprite{
		mesh: spriteMesh,
		bounds: bounds,
		texture: texture,
		textureMatrix: texMat,
	}
}

func (s *Sprite) Draw(pass *RenderPass, matrix Mat4) {
	pass.Add(s.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, s.textureMatrix)
}
func (s *Sprite) DrawColorMask(pass *RenderPass, matrix Mat4, mask RGBA) {
	pass.Add(s.mesh, matrix, mask, s.textureMatrix)
}

type Mesh struct {
	positions []Vec3
	colors []Vec3
	texCoords []Vec2
	indices []uint32
}

func (m *Mesh) Draw(pass *RenderPass, matrix Mat4) {
	pass.Add(m, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, Mat3Ident)
}

func (m *Mesh) DrawColorMask(pass *RenderPass, matrix Mat4, mask RGBA) {
	pass.Add(m, matrix, mask, Mat3Ident)
}

var spriteMesh = NewQuadMesh()

func NewQuadMesh() *Mesh {
	color := RGBA{1.0, 1.0, 1.0, 1.0}
	positions := []Vec3{
		Vec3{0.5  , 0.5,  0.0},
		Vec3{0.5  , -0.5, 0.0},
		Vec3{-0.5 , -0.5, 0.0},
		Vec3{-0.5 , 0.5,  0.0},
	}
	colors := []Vec3{
		Vec3{color.R, color.G, color.B},
		Vec3{color.R, color.G, color.B},
		Vec3{color.R, color.G, color.B},
		Vec3{color.R, color.G, color.B},
	}
	texCoords := []Vec2{
		Vec2{1.0, 0.0},
		Vec2{1.0, 1.0},
		Vec2{0.0, 1.0},
		Vec2{0.0, 0.0},
	}

	inds := []uint32{
		0, 1, 3,
		1, 2, 3,
	}

	return &Mesh{
		positions: positions,
		colors: colors,
		texCoords: texCoords,
		indices: inds,
	}
}

// type Material struct {
// 	uniforms map[string]interface{}
// }

// type Model struct {
// 	meshes []Mesh
// 	materials []Material
// }
