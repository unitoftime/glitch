package glitch

type Mesh struct {
	positions []Vec3
	normals []Vec3
	colors []Vec4
	texCoords []Vec2
	indices []uint32
}

func NewMesh() *Mesh {
	return &Mesh{
		positions: make([]Vec3, 0),
		normals: make([]Vec3, 0),
		colors: make([]Vec4, 0),
		texCoords: make([]Vec2, 0),
		indices: make([]uint32, 0),
	}
}

// TODO - clear function? Should append be more like draw?
func (m *Mesh) Clear() {
	m.positions = m.positions[:0]
	m.normals = m.normals[:0]
	m.colors = m.colors[:0]
	m.texCoords = m.texCoords[:0]
	m.indices = m.indices[:0]
}

func (m *Mesh) Draw(pass *RenderPass, matrix Mat4) {
	pass.Add(m, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, DefaultMaterial())
}

// TODO - This should accept image/color and call RGBA(). Would that be slower?
func (m *Mesh) DrawColorMask(pass *RenderPass, matrix Mat4, mask RGBA) {
	pass.Add(m, matrix, mask, DefaultMaterial())
}

// TODO - should this be more like draw?
func (m *Mesh) Append(m2 *Mesh) {
	currentElement := uint32(len(m.positions))
	for i := range m2.indices {
		m.indices = append(m.indices, currentElement + m2.indices[i])
	}

	m.positions = append(m.positions, m2.positions...)
	m.normals = append(m.normals, m2.normals...)
	m.colors = append(m.colors, m2.colors...)
	m.texCoords = append(m.texCoords, m2.texCoords...)
}

// TODO - Maybe this is faster in some scenarios?
// func (m *Mesh) AddTriangle(a, b, c Vec3, uv1, uv2, uv3 Vec2) {
// 	currentElement := uint32(len(m.positions))
// 	for i := range m2.indices {
// 		m.indices = append(m.indices, currentElement + m2.indices[i])
// 	}

// 	m.positions = append(m.positions, m2.positions...)
// 	m.colors = append(m.colors, m2.colors...)
// 	m.texCoords = append(m.texCoords, m2.texCoords...)
// }

// Basically a quad mesh, but with a centered position
func NewSpriteMesh(w, h float32, uvBounds Rect) *Mesh {
	return NewQuadMesh(R(-w/2, -h/2, w/2, h/2), uvBounds)
}

func NewQuadMesh(bounds Rect, uvBounds Rect) *Mesh {
	color := RGBA{1.0, 1.0, 1.0, 1.0}
	positions := []Vec3{
		Vec3{bounds.Max[0], bounds.Max[1], 0.0},
		Vec3{bounds.Max[0], bounds.Min[1], 0.0},
		Vec3{bounds.Min[0], bounds.Min[1], 0.0},
		Vec3{bounds.Min[0], bounds.Max[1], 0.0},
	}
	// TODO normals
	colors := []Vec4{
		Vec4{color.R, color.G, color.B, color.A},
		Vec4{color.R, color.G, color.B, color.A},
		Vec4{color.R, color.G, color.B, color.A},
		Vec4{color.R, color.G, color.B, color.A},
	}
	texCoords := []Vec2{
		Vec2{uvBounds.Max[0], uvBounds.Min[1]},
		Vec2{uvBounds.Max[0], uvBounds.Max[1]},
		Vec2{uvBounds.Min[0], uvBounds.Max[1]},
		Vec2{uvBounds.Min[0], uvBounds.Min[1]},
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

func NewCubeMesh(size float32) *Mesh {
	size = size/2

	positions := []Vec3{
		// Front face
		Vec3{-size, -size,  size},
		Vec3{size, -size,  size},
		Vec3{size,  size,  size},
		Vec3{-size,  size,  size},
		// Back face
		Vec3{-size, -size, -size},
		Vec3{-size,  size, -size},
		Vec3{size,  size, -size},
		Vec3{size, -size, -size},
		// Top face
		Vec3{-size,  size, -size},
		Vec3{-size,  size,  size},
		Vec3{size,  size,  size},
		Vec3{size,  size, -size},
		// Bottom face
		Vec3{-size, -size, -size},
		Vec3{size, -size, -size},
		Vec3{size, -size,  size},
		Vec3{-size, -size,  size},
		// Right face
		Vec3{size, -size, -size},
		Vec3{size,  size, -size},
		Vec3{size,  size,  size},
		Vec3{size, -size,  size},
		// Left face
		Vec3{-size, -size, -size},
		Vec3{-size, -size,  size},
		Vec3{-size,  size,  size},
		Vec3{-size,  size, -size},
	}

	col := Vec4{1.0, 1.0, 1.0, 1.0}
	colors := []Vec4{
		col, col, col, col,
		col, col, col, col,
		col, col, col, col,
		col, col, col, col,
		col, col, col, col,
	}

	// TODO normals
	normals := []Vec3{
		// Front face
		Vec3{0, 0, 1},
		Vec3{0, 0, 1},
		Vec3{0, 0, 1},
		Vec3{0, 0, 1},
		// Back face
		Vec3{0, 0, -1},
		Vec3{0, 0, -1},
		Vec3{0, 0, -1},
		Vec3{0, 0, -1},
		// Top face
		Vec3{0, 1, 0},
		Vec3{0, 1, 0},
		Vec3{0, 1, 0},
		Vec3{0, 1, 0},
		// Bottom face
		Vec3{0, -1, 0},
		Vec3{0, -1, 0},
		Vec3{0, -1, 0},
		Vec3{0, -1, 0},
		// Right face
		Vec3{1, 0, 0},
		Vec3{1, 0, 0},
		Vec3{1, 0, 0},
		Vec3{1, 0, 0},
		// Left face
		Vec3{-1, 0, 0},
		Vec3{-1, 0, 0},
		Vec3{-1, 0, 0},
		Vec3{-1, 0, 0},
	}

	// TODO texCoords
	texCoords := []Vec2{
		// Front face
		Vec2{-0, -0},
		Vec2{0, -0},
		Vec2{0,  0},
		Vec2{-0,  0},
		// Back face
		Vec2{-0, -0},
		Vec2{-0,  0},
		Vec2{0,  0},
		Vec2{0, -0},
		// Top face
		Vec2{-0,  0},
		Vec2{-0,  0},
		Vec2{0,  0},
		Vec2{0,  0},
		// Bottom face
		Vec2{-0, -0},
		Vec2{0, -0},
		Vec2{0, -0},
		Vec2{-0, -0},
		// Right face
		Vec2{0, -0},
		Vec2{0,  0},
		Vec2{0,  0},
		Vec2{0, -0},
		// Left face
		Vec2{-0, -0},
		Vec2{-0, -0},
		Vec2{-0,  0},
		Vec2{-0,  0},
	}

	indices := []uint32{
		0,  1,  2,      0,  2,  3,    // front
    4,  5,  6,      4,  6,  7,    // back
    8,  9,  10,     8,  10, 11,   // top
    12, 13, 14,     12, 14, 15,   // bottom
    16, 17, 18,     16, 18, 19,   // right
    20, 21, 22,     20, 22, 23,   // left
	}

	return &Mesh{
		positions: positions,
		normals: normals,
		colors: colors,
		texCoords: texCoords,
		indices: indices,
	}
}
