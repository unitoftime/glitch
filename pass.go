package glitch

// Sort by:
// - Front-to-back vs Back-to-front (single bit)
// - Depth bits
// - Material / Uniforms / Textures
type drawCommand struct {
	command uint64
	mesh *Mesh
	matrix Mat4
	mask RGBA
	texMat Mat3
}

type RenderPass struct {
	shader *Shader
	texture *Texture
	uniforms map[string]interface{}
	buffer *BufferPool
	commands []drawCommand
}

func NewRenderPass(shader *Shader) *RenderPass {
	defaultBatchSize := 100000
	return &RenderPass{
		shader: shader,
		texture: nil,
		uniforms: make(map[string]interface{}),
		// buffer: NewVertexBuffer(shader, 10000, 10000),
		buffer: NewBufferPool(shader, defaultBatchSize),
		commands: make([]drawCommand, 0),
	}
}

func (r *RenderPass) Clear() {
	// Clear stuff
	r.buffer.Clear()
	r.commands = r.commands[:0]
}

// TODO - Mat?
func (r *RenderPass) Draw(win *Window) {
	r.shader.Bind()
	r.texture.Bind(0) // TODO - hardcoded texture slot
	for k,v := range r.uniforms {
		ok := r.shader.SetUniform(k, v)
		if !ok {
			panic("Error setting uniform - todo decrease this to log")
		}
	}

	destBuffs := make([]interface{}, 3) // TODO - hardcoded
	destBuffs[0] = &[]Vec3{}
	destBuffs[1] = &[]Vec4{}
	destBuffs[2] = &[]Vec2{}
	for _, c := range r.commands {
		numVerts := len(c.mesh.positions)
		r.buffer.Reserve(c.mesh.indices, numVerts, destBuffs)

		// work and append
		posBuf := *(destBuffs[0]).(*[]Vec3)
		for i := range c.mesh.positions {
			vec := c.matrix.Apply(c.mesh.positions[i])
			posBuf[i] = vec
		}

		colBuf := *(destBuffs[1]).(*[]Vec4)
		for i := range c.mesh.colors {
			colBuf[i] = Vec4{
				c.mesh.colors[i][0] * c.mask.R,
				c.mesh.colors[i][1] * c.mask.G,
				c.mesh.colors[i][2] * c.mask.B,
				c.mesh.colors[i][2] * c.mask.A,
			}
		}

		texBuf := *(destBuffs[2]).(*[]Vec2)
		for i := range c.mesh.texCoords {
			texBuf[i] = c.texMat.Apply(c.mesh.texCoords[i])
		}
	}

	r.buffer.Draw()
}

func (r *RenderPass) SetTexture(slot int, texture *Texture) {
	// TODO - use correct texture slot
	r.texture = texture
}

func (r *RenderPass) SetUniform(name string, value interface{}) {
	r.uniforms[name] = value
}

func (r *RenderPass) Add(mesh *Mesh, mat Mat4, mask RGBA, texMat Mat3) {
	r.commands = append(r.commands, drawCommand{
		0, mesh, mat, mask, texMat,
	})
}
