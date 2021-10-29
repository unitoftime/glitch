package glitch

// Sort by:
// - Front-to-back vs Back-to-front (single bit)
// - Depth bits
// - Material / Uniforms / Textures
type drawCommand struct {
	command uint64
	mesh *Mesh
	matrix *Mat4
	mask RGBA
}

type RenderPass struct {
	shader *Shader
	texture *Texture
	uniforms map[string]interface{}
	buffer *BufferPool
	commands []drawCommand
}

func NewRenderPass(shader *Shader) *RenderPass {
	defaultBatchSize := 1000000
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

func (r *RenderPass) Draw(win *Window) {
	r.shader.Bind()
	r.texture.Bind(0) // TODO - hardcoded texture slot
	for k,v := range r.uniforms {
		ok := r.shader.SetUniform(k, v)
		if !ok {
			panic("Error setting uniform - todo decrease this to log")
		}
	}

	destBuffs := make([]interface{}, 3) // TODO -hardcode
	for _, c := range r.commands {
		// positions := make([]float32, len(c.mesh.positions) * 3) // 3 b/c vec3
		// for i := range c.mesh.positions {
		// 	vec := c.matrix.MulVec3(&c.mesh.positions[i])
		// 	positions[(i * 3) + 0] = vec[0]
		// 	positions[(i * 3) + 1] = vec[1]
		// 	positions[(i * 3) + 2] = vec[2]
		// }
		// r.buffer.Add(positions, c.mesh.colors, c.mesh.texCoords, c.mesh.indices)

		numVerts := len(c.mesh.positions)
		r.buffer.Reserve(c.mesh.indices, numVerts, destBuffs)
		posBuff := (destBuffs[0]).([]Vec3)
		for i := range c.mesh.positions {
			vec := MatMul(c.matrix, c.mesh.positions[i])
			posBuff[i] = vec
		}

		colBuf := (destBuffs[1]).([]Vec3)
		colBuf = append(colBuf[:0], c.mesh.colors...)
		texBuf := (destBuffs[2]).([]Vec2)
		texBuf = append(texBuf[:0], c.mesh.texCoords...)

		// r.buffer.Add2(c.mesh.indices,
		// 	Vec3Add{
		// 		c.mesh.positions,
		// 		func(in *Vec3) {
		// 			*in = MatMul(c.matrix, *in)
		// 		},
		// 	}, c.mesh.colors, c.mesh.texCoords)

		// r.buffer.Add(c.mesh.positions, c.mesh.colors, c.mesh.texCoords, c.mesh.indices, c.matrix, c.mask)
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

func (r *RenderPass) Add(mesh *Mesh, mat *Mat4, mask RGBA) {
	r.commands = append(r.commands, drawCommand{
		0, mesh, mat, mask,
	})
}
