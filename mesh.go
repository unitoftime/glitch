package glitch

// For batching multiple sprites into one
type Batch struct {
	mesh *Mesh
	material Material
}

func NewBatch() *Batch {
	return &Batch{
		mesh: NewMesh(),
		material: nil,
	}
}

// TODO - It may be faster to copy all the bufs to the destination and then operate on them there. that might save you a copy
func (b *Batch) Add(mesh *Mesh, matrix Mat4, mask RGBA, material Material) {
	if b.material == nil {
		b.material = material
	} else {
		if b.material != material {
			panic("Materials must match inside a batch!")
		}
	}

	posBuf := make([]Vec3, len(mesh.positions))
	for i := range mesh.positions {
		posBuf[i] = matrix.Apply(mesh.positions[i])
	}

	renormalizeMat := matrix.Inv().Transpose()
	normBuf := make([]Vec3, len(mesh.normals))
	for i := range mesh.normals {
		normBuf[i] = renormalizeMat.Apply(mesh.normals[i])
	}

	colBuf := make([]Vec4, len(mesh.colors))
	for i := range mesh.colors {
		// TODO - vec4 mult function
		colBuf[i] = Vec4{
			mesh.colors[i][0] * mask.R,
			mesh.colors[i][1] * mask.G,
			mesh.colors[i][2] * mask.B,
			mesh.colors[i][3] * mask.A,
		}
	}

	// TODO - is a copy faster?
	texBuf := make([]Vec2, len(mesh.texCoords))
	for i := range mesh.texCoords {
		texBuf[i] = mesh.texCoords[i]
	}

	indices := make([]uint32, len(mesh.indices))
	for i := range mesh.indices {
		indices[i] = mesh.indices[i]
	}

	m2 := &Mesh{
		positions: posBuf,
		normals: normBuf,
		colors: colBuf,
		texCoords: texBuf,
		indices: indices,
		//bounds: todo,
	}

	b.mesh.Append(m2)
}

func (b *Batch) Clear() {
	b.mesh.Clear()
	b.material = nil
}

func (b *Batch) Draw(target BatchTarget, matrix Mat4) {
	target.Add(b.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, b.material)
}

type Mesh struct {
	positions []Vec3
	normals []Vec3
	colors []Vec4
	texCoords []Vec2
	indices []uint32
	bounds Box
	// translation Vec3
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
	m.bounds = Box{}
}

func (m *Mesh) Draw(pass *RenderPass, matrix Mat4) {
	pass.Add(m, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, DefaultMaterial())
}

// TODO - This should accept image/color and call RGBA(). Would that be slower?
func (m *Mesh) DrawColorMask(pass *RenderPass, matrix Mat4, mask RGBA) {
	pass.Add(m, matrix, mask, DefaultMaterial())
}

func (m *Mesh) Bounds() Box {
	return m.bounds
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

	m.bounds = m.bounds.Union(m2.bounds)
}

// func (m *Mesh) SetTranslation(pos Vec3) {
// 	if m.translation == pos { return } // Skip if we've already translated this amount
// 	delta := pos.Sub(m.translation)
// 	// delta := m.translation.Sub(pos)

// 	for i := range m.positions {
// 		m.positions[i] = delta.Add(m.positions[i])
// 	}

// 	m.translation = pos

// 	// TODO - recalculate mesh bounds/box
// }

// Sets the color of every vertex
func (m *Mesh) SetColor(col RGBA) {
	v4Color := Vec4{col.R, col.G, col.B, col.A}
	for i := range m.colors {
		m.colors[i] = v4Color
	}
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
		bounds: bounds.ToBox(),
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
		bounds: Box{
			Min: Vec3{-size, -size, -size},
			Max: Vec3{size, size, size},
		},
	}
}
