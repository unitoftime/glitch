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
	if u == nil { return }
	for k, v := range u.set {
		shader.setUniform(k, v)
	}
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

// TODO: You could pack this down even more
// Shader:  shader slot lut ID 256 maximum
// Texture: texture slot lut ID 256 maximum
// Uniform: uniform slot lut ID 256 maximum
type Material struct {
	shader *Shader
	texture *Texture
	uniforms *Uniforms // TODO: Generic binder (eg old Material interface)?

	blend BlendMode
	depth DepthMode
	cull CullMode
}

func NewMaterial(shader *Shader) Material {
	return Material{
		shader: shader,
		blend: BlendModeNormal,
		uniforms: nil,
	}
}

// TODO: Implement
// func (m Material) Copy() Material {
// 	m2 := NewMaterial(m.shader)
// 	// m2.SetTexture(m.texture)
// 	// TODO: SetBlendMode()
// 	m2.uniforms = m.uniforms.Copy()
// 	return m2
// }

func (m *Material) SetShader(shader *Shader) *Material {
	m.shader = shader
	return m
}

func (m *Material) SetUniform(name string, val any) *Material {
	if m.uniforms == nil {
		m.uniforms = &Uniforms{}
	}
	m.uniforms.SetUniform(name, val)
	return m
}

func (m *Material) SetTexture(/* slot int, */ texture *Texture) {
	m.texture = texture
}

func (m *Material) SetCullMode(cullMode CullMode) *Material {
	m.cull = cullMode
	return m
}

func (m *Material) SetDepthMode(depthMode DepthMode) *Material {
	m.depth = depthMode
	return m
}

func (m *Material) SetBlendMode(blendMode BlendMode) *Material {
	m.blend = blendMode
	return m
}

func (m Material) Bind() {
	setShader(m.shader)
	// m.shader.Use()

	if m.texture != nil {
		// texSlot := 0
		// m.texture.bind(texSlot)

		state.bindTexture(m.texture)
	}

	state.setBlendMode(m.blend)
	state.setDepthMode(m.depth)
	state.setCullMode(m.cull)


	// // Bind Depthmode
	// if m.depth == DepthModeNone {
	// 	state.enableDepthTest(false)
	// } else {
	// 	state.enableDepthTest(true)
	// 	state.setDepthFunc(m.depth.mode)
	// }

	// // Bind CullMode
	// if m.cull == CullModeNone {
	// 	state.disableCullMode()
	// } else {
	// 	state.enableCullMode(m.cull)
	// }

	// // Bind Blendmode
	// state.setBlendFunc(m.blend.src, m.blend.dst)

	// // // Bind Camera (ie global material)
	// // // TODO: m.camera.Bind(m.shader)
	// // if m.camera != nil {
	// // 	m.shader.SetUniform("projection", m.camera.Projection.gl())
	// // 	m.shader.SetUniform("view", m.camera.View.gl())
	// // }

	// Bind uniforms (ie local material)
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

type Metrics struct {
	setShader int
	setCamera int
	clearTarget int
	setTarget int
	setMaterial int
	add int
	flushAttempt int
	flush int
	finish int
	draw int

	vertsTotal int // The total number of vertices drawn
	vertsAvg int // The average number of vertices drawn per drawCall

	// Note: Disabled because this didn't really give me any insight
	// vertsMin int
	// vertsMax int
}

func GetMetrics() Metrics {
	metric := global.metric
	global.metric = Metrics{}

	return metric
}

//--------------------------------------------------------------------------------
type CameraMaterial struct {
	Projection, View glMat4
}
//--------------------------------------------------------------------------------

var global = &globalBatcher{
	shaderCache: make(map[*Shader]struct{}), // TODO: Does this cause shaders to not cleanup?
	// camera: NewCameraOrtho(), // Identity camera
	camera: CameraMaterial{
		glMat4Ident, glMat4Ident,
	},
} // TODO: Default case for shader?

type globalBatcher struct {
	shader *Shader
	camera CameraMaterial
	lastBuffer *VertexBuffer
	target Target
	texture *Texture
	blend BlendMode

	material Material

	shaderCache map[*Shader]struct{}

	metric Metrics
}

func Clear(target Target, color RGBA) {
	setTarget(target)
	state.clearTarget(color)

	global.metric.clearTarget++
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

func SetCameraMaterial(camMaterial CameraMaterial) {
	if global.camera == camMaterial {
		return
	}

	global.flush() // TODO: You technically only need to do this if it will change the uniform
	global.camera = camMaterial
	if global.shader != nil {
		global.shader.setUniform("projection", global.camera.Projection)
		global.shader.setUniform("view", global.camera.View)
	}

	global.metric.setCamera++
}

func SetCamera(camera *CameraOrtho) {
	camMaterial := CameraMaterial{
		Projection: glm4(camera.Projection),
		View: glm4(camera.View),
	}
	SetCameraMaterial(camMaterial)
}

func setTarget(target Target) {
	if global.target == target {
		return
	}

	global.flush() // TODO: You technically only need to do this if it will change the uniform
	global.target = target
	target.Bind()

	global.metric.setTarget++
}

func setShader(shader *Shader) {
	if global.shader == shader {
		return
	}

	global.flush()
	global.shader = shader
	mainthread.Call(shader.mainthreadBind)

	global.shader.setUniform("projection", global.camera.Projection)
	global.shader.setUniform("view", global.camera.View)

	global.shaderCache[shader] = struct{}{}
	global.metric.setShader++
}

func (g *globalBatcher) Add(filler GeometryFiller, mat glMat4, mask RGBA, material Material, translucent bool) {
	if filler == nil { return } // Skip nil meshes

	global.metric.add++

	// 1. If you switch materials, then draw the last one
	if material != g.material {
		// fmt.Printf("setmaterial (old -> new):\n%+v\n%+v\n", g.material, material)

		global.metric.setMaterial++
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
	// clear(g.shaderCache) // TODO: the shaderCache leaks right now, but only grows to as many shaders as the user loads which isn't that much. You cant clear here because in single shader scenarios itll never get set back again
	g.metric.finish++
}

// Draws the current buffer and progress the shader pool to the next available
func (g *globalBatcher) flush() {
	g.metric.flushAttempt++
	if g.lastBuffer == nil { return }

	g.drawCall(g.lastBuffer, glMat4Ident)
	g.lastBuffer = nil
	g.shader.pool.gotoNextClean()

	g.metric.flush++
}

// Executes a drawcall with ...
func (g *globalBatcher) drawCall(buffer *VertexBuffer, mat glMat4) {
	// // TODO: rewrite how buffer state works for immediate mode case
	// buffer.state.Bind(g.shader)

	// TOOD: Maybe pass this into VertexBuffer.Draw() func
	ok := g.shader.setUniform("model", mat)
	if !ok {
		panic("Error setting model uniform - all shaders must have 'model' uniform")
	}

	buffer.Draw()
	g.metric.draw++

	vertCount := int(buffer.numVerts)
	g.metric.vertsTotal += vertCount
	g.metric.vertsAvg = g.metric.vertsTotal / g.metric.draw
	// g.metric.vertsMax = max(vertCount, g.metric.vertsMax)
	// g.metric.vertsMin = min(vertCount, g.metric.vertsMin)
	// if g.metric.vertsMin == 0 {
	// 	g.metric.vertsMin = vertCount
	// }
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
