package glitch

// Sort by:
// - Front-to-back vs Back-to-front (single bit)
// - Depth bits
// - Material / Uniforms / Textures
type drawCommand struct {
	command uint64
	mesh *Mesh
	matrix *Mat4
}

type RenderPass struct {
	target *Window
	shader *Shader
	texture *Texture
	buffer *BufferPool
	commands []drawCommand
}

func NewRenderPass(target *Window, shader *Shader) *RenderPass {
	defaultBatchSize := 10000
	return &RenderPass{
		target: target,
		shader: shader,
		texture: nil,
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

func (r *RenderPass) Execute() {
	r.shader.Bind()
	r.texture.Bind(0) // TODO - hardcoded texture slot

	for _, c := range r.commands {
		positions := make([]float32, len(c.mesh.positions) * 3) // 3 b/c vec3
		for i := range c.mesh.positions {
			vec := c.matrix.MulVec3(&c.mesh.positions[i])
			positions[(i * 3) + 0] = vec[0]
			positions[(i * 3) + 1] = vec[1]
			positions[(i * 3) + 2] = vec[2]
		}
		r.buffer.Add(positions, c.mesh.colors, c.mesh.texCoords, c.mesh.indices)
	}

	r.buffer.Draw()
}

func (r *RenderPass) SetTexture(slot int, texture *Texture) {
	// TODO - use correct texture slot
	r.texture = texture
}

func (r *RenderPass) Draw(mesh *Mesh, mat *Mat4) {
	r.commands = append(r.commands, drawCommand{
		0, mesh, mat,
	})
}
