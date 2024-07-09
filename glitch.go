package glitch

import (
	"github.com/unitoftime/glitch/internal/mainthread"
)

func Run(function func()) {
	mainthread.Run(function)
}

// type Material interface {
// 	Bind(*Shader)
// }

// type SpriteMaterial struct {
// 	texture *Texture
// }

// func NewSpriteMaterial(texture *Texture) SpriteMaterial {
// 	return SpriteMaterial{
// 		texture: texture,
// 	}
// }

// func (m SpriteMaterial) Bind(shader *Shader) {
// 	m.texture.Bind(0) // Direct opengl? Or should I call through shader?
// 	// pass.SetTexture(0, m.texture) // TODO - hardcoded slot?
// }

// func DefaultMaterial() SpriteMaterial {
// 	return SpriteMaterial{
// 		texture: WhiteTexture(),
// 	}
// }

type Uniforms struct {
	set map[string]any
}
func (u *Uniforms) Bind(shader *Shader) {

}

func (u *Uniforms) SetUniform(name string, val any) {
	if u.set == nil {
		u.set = make(map[string]any)
	}
	u.set[name] = val
}

func (u *Uniforms) Copy() *Uniforms {
	u2 := &Uniforms{}
	for k,v := range u.set {
		u2.SetUniform(k, v)
	}
	return u2
}

type Material struct {
	shader *Shader
	texture *Texture
	blend BlendMode
	uniforms *Uniforms
}

func NewMaterial(shader *Shader) Material {
	return Material{
		shader: shader,
		blend: BlendModeNormal,
		uniforms: &Uniforms{},
	}
}

func (m Material) Copy() Material {
	m2 := NewMaterial(m.shader)
	// m2.SetTexture(m.texture)
	// TODO: SetBlendMode()
	m2.uniforms = m.uniforms.Copy()
	return m2
}

func (m *Material) SetShader(shader *Shader) *Material {
	m.shader = shader
	return m
}

func (m *Material) SetUniform(name string, val any) *Material {
	m.uniforms.SetUniform(name, val)
	return m
}

func (m *Material) SetTexture(/* slot int, */ texture *Texture) {
	m.texture = texture
}

func (m Material) Bind() {
	m.shader.Use()

	if m.texture != nil {
		texSlot := 0
		m.texture.Bind(texSlot)
	}

	// TODO: Blendmode

	m.uniforms.Bind(m.shader)
}

// type materialGroup struct {
// 	globalMaterial Material
// 	localMaterial Material
// }
// func (m materialGroup) Bind(shader *Shader) {
// 	if m.globalMaterial != nil {
// 		m.globalMaterial.Bind(shader)
// 	}
// 	if m.localMaterial != nil {
// 		m.localMaterial.Bind(shader)
// 	}
// }


//--------------------------------------------------------------------------------

var global = &globalBatcher{
	shaderCache: make(map[*Shader]struct{}), // TODO: Does this cause shaders to not cleanup?
} // TODO: Default case for shader?

type globalBatcher struct {
	shader *Shader
	lastBuffer *VertexBuffer
	target Target
	texture *Texture
	blend BlendMode

	material Material

	shaderCache map[*Shader]struct{}
}

func Clear(target Target, color RGBA) {
	setTarget(target)
	state.clearTarget(color)
}

// func setBlendMode(blend BlendMode) {
// 	global.flush() // TODO: You technically only need to do this if it will change the uniform
// 	global.blend = blend

// 	state.setBlendFunc(blend.src, blend.dst)
// }

// func setTexture(texture *Texture) {
// 	global.flush() // TODO: You technically only need to do this if it will change the uniform
// 	global.texture = texture
// 	texSlot := 0 // TODO: Implement Texture slots
// 	texture.Bind(texSlot)
// }

func setTarget(target Target) {
	global.flush() // TODO: You technically only need to do this if it will change the uniform
	global.target = target
	target.Bind()
}

func setShader(shader *Shader) {
	global.flush() // TODO: You technically only need to do this if it will change the uniform
	global.shader = shader
	shader.Bind()

	global.shaderCache[shader] = struct{}{}
}

func (g *globalBatcher) Add(filler GeometryFiller, mat glMat4, mask RGBA, material Material, translucent bool) {
	if filler == nil { return } // Skip nil meshes

	// 1. If you switch materials, then draw the last one
	if material != g.material {
		// Note: This is kindof different from a global material. it's more like a local material
		g.flush()
		g.material = material
		g.material.Bind()
	}

	buffer := filler.GetBuffer()
	if buffer != nil {
		global.drawCall(buffer, mat)
		return
	}

	// Note: Captured in shader.pool
	// 1. If you fill up then draw the last one
	vertexBuffer := filler.Fill(global.shader.pool, mat, mask)

	// If vertexBuffer has changed then we want to draw the last one
	if global.lastBuffer != nil && vertexBuffer != global.lastBuffer {
		global.drawCall(global.lastBuffer, glMat4Ident)
	}

	global.lastBuffer = vertexBuffer
}

func (g *globalBatcher) finish() {
	g.flush()
	for shader := range g.shaderCache {
		shader.pool.Clear()
	}
	clear(g.shaderCache)
}

// Draws the current buffer and progress the shader pool to the next available
func (g *globalBatcher) flush() {
	if g.lastBuffer == nil { return }

	g.drawCall(g.lastBuffer, glMat4Ident)
	g.lastBuffer = nil
	g.shader.pool.gotoNextClean()
}

// Executes a drawcall with ...
func (g *globalBatcher) drawCall(buffer *VertexBuffer, mat glMat4) {
	// // TODO: rewrite how buffer state works for immediate mode case
	// buffer.state.Bind(g.shader)

	// TOOD: Maybe pass this into VertexBuffer.Draw() func
	ok := g.shader.SetUniform("model", mat)
	if !ok {
		panic("Error setting model uniform - all shaders must have 'model' uniform")
	}

	buffer.Draw()
}

// //--------------------------------------------------------------------------------
// // Holds the invariant state of the buffer (ie the configurations required for batching other draws into this buffer)
// // Note: Everything you put in here must be comparable, and if there is any mismatch of data, it will force a new buffer
// type BufferState struct{
// 	material Material
// 	blend BlendMode
// }
// func (b BufferState) Bind(shader *Shader) {
// 	// TODO: combine these into the same mainthread call?
// 	if b.material != nil {
// 		b.material.Bind(shader)
// 	}

// 	state.setBlendFunc(b.blend.src, b.blend.dst)
// }
