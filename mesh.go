package glitch

import "fmt"

// For batching multiple sprites into one
type Batch struct {
	mesh        *Mesh
	material    Material
	Translucent bool
}

func NewBatch() *Batch {
	return &Batch{
		mesh:     NewMesh(),
		material: nil,
	}
}

func (b *Batch) Buffer(pass *RenderPass) *Batch {
	return &Batch{
		mesh:        b.mesh.Buffer(pass, b.material, b.Translucent),
		material:    b.material,
		Translucent: b.Translucent,
	}
}

// TODO - It may be faster to copy all the bufs to the destination and then operate on them there. that might save you a copy
// TODO: should I maintain a translucent and non-translucent batch mesh?
func (b *Batch) Add(mesh *Mesh, matrix glMat4, mask RGBA, material Material, translucent bool) {
	if b.material == nil {
		b.material = material
	} else {
		if b.material != material {
			panic("Materials must match inside a batch!")
		}
	}

	// If anything translucent is added to the batch, then we will consider the entire thing translucent
	b.Translucent = b.Translucent || translucent

	// mat := matrix.gl()

	// Append each index
	currentElement := uint32(len(b.mesh.positions))
	for i := range mesh.indices {
		b.mesh.indices = append(b.mesh.indices, currentElement+mesh.indices[i])
	}

	// Append each position
	for i := range mesh.positions {
		b.mesh.positions = append(b.mesh.positions, matrix.Apply(mesh.positions[i]))
	}

	// Calculate the bounding box of the mesh we just merged in
	// Because we already figured out the first index of the new mesh (ie `currentElement`) we can just slice off the end of the new mesh
	posBuf := b.mesh.positions[int(currentElement):]
	min := posBuf[0]
	max := posBuf[0]
	for i := range posBuf {
		// X
		if posBuf[i][0] < min[0] {
			min[0] = posBuf[i][0]
		}
		if posBuf[i][0] > max[0] {
			max[0] = posBuf[i][0]
		}

		// Y
		if posBuf[i][1] < min[1] {
			min[1] = posBuf[i][1]
		}
		if posBuf[i][1] > max[1] {
			max[1] = posBuf[i][1]
		}

		// Z
		if posBuf[i][2] < min[2] {
			min[2] = posBuf[i][2]
		}
		if posBuf[i][2] > max[2] {
			max[2] = posBuf[i][2]
		}
	}

	newBounds := Box{
		Min: Vec3{float64(min[0]), float64(min[1]), float64(min[2])},
		Max: Vec3{float64(max[0]), float64(max[1]), float64(max[2])},
	}
	b.mesh.bounds = b.mesh.bounds.Union(newBounds)

	if len(mesh.normals) > 0 {
		renormalizeMat := matrix.Inv().Transpose()
		for i := range mesh.normals {
			b.mesh.normals = append(b.mesh.normals, renormalizeMat.Apply(mesh.normals[i]))
		}
	}

	for i := range mesh.colors {
		// TODO - vec4 mult function
		b.mesh.colors = append(b.mesh.colors, glVec4{
			mesh.colors[i][0] * float32(mask.R),
			mesh.colors[i][1] * float32(mask.G),
			mesh.colors[i][2] * float32(mask.B),
			mesh.colors[i][3] * float32(mask.A),
		})
	}

	// TODO - is a copy faster?
	for i := range mesh.texCoords {
		b.mesh.texCoords = append(b.mesh.texCoords, mesh.texCoords[i])
	}

	// if b.material == nil {
	// 	b.material = material
	// } else {
	// 	if b.material != material {
	// 		panic("Materials must match inside a batch!")
	// 	}
	// }

	// mat := matrix.gl()

	// posBuf := make([]glVec3, len(mesh.positions))
	// for i := range mesh.positions {
	// 	posBuf[i] = mat.Apply(mesh.positions[i])
	// }

	// min := posBuf[0]
	// max := posBuf[0]
	// for i := range posBuf {
	// 	// X
	// 	if posBuf[i][0] < min[0] {
	// 		min[0] = posBuf[i][0]
	// 	}
	// 	if posBuf[i][0] > max[0] {
	// 		max[0] = posBuf[i][0]
	// 	}

	// 	// Y
	// 	if posBuf[i][1] < min[1] {
	// 		min[1] = posBuf[i][1]
	// 	}
	// 	if posBuf[i][1] > max[1] {
	// 		max[1] = posBuf[i][1]
	// 	}

	// 	// Z
	// 	if posBuf[i][2] < min[2] {
	// 		min[2] = posBuf[i][2]
	// 	}
	// 	if posBuf[i][2] > max[2] {
	// 		max[2] = posBuf[i][2]
	// 	}
	// }

	// newBounds := Box{
	// 	Min: Vec3{float64(min[0]), float64(min[1]), float64(min[2])},
	// 	Max: Vec3{float64(max[0]), float64(max[1]), float64(max[2])},
	// }

	// renormalizeMat := matrix.Inv().Transpose().gl()
	// normBuf := make([]glVec3, len(mesh.normals))
	// for i := range mesh.normals {
	// 	normBuf[i] = renormalizeMat.Apply(mesh.normals[i])
	// }

	// colBuf := make([]glVec4, len(mesh.colors))
	// for i := range mesh.colors {
	// 	// TODO - vec4 mult function
	// 	colBuf[i] = glVec4{
	// 		mesh.colors[i][0] * float32(mask.R),
	// 		mesh.colors[i][1] * float32(mask.G),
	// 		mesh.colors[i][2] * float32(mask.B),
	// 		mesh.colors[i][3] * float32(mask.A),
	// 	}
	// }

	// // TODO - is a copy faster?
	// texBuf := make([]glVec2, len(mesh.texCoords))
	// for i := range mesh.texCoords {
	// 	texBuf[i] = mesh.texCoords[i]
	// }

	// indices := make([]uint32, len(mesh.indices))
	// for i := range mesh.indices {
	// 	indices[i] = mesh.indices[i]
	// }

	// m2 := &Mesh{
	// 	positions: posBuf,
	// 	normals: normBuf,
	// 	colors: colBuf,
	// 	texCoords: texBuf,
	// 	indices: indices,
	// 	bounds: newBounds,
	// }

	// b.mesh.Append(m2)
}

func (b *Batch) Clear() {
	b.mesh.Clear()
	b.material = nil
	b.Translucent = false
}

func (b *Batch) Draw(target BatchTarget, matrix Mat4) {
	target.Add(b.mesh, matrix.gl(), RGBA{1.0, 1.0, 1.0, 1.0}, b.material, b.Translucent)
	// b.DrawColorMask(target, matrix, White)
}

func (b *Batch) DrawColorMask(target BatchTarget, matrix Mat4, color RGBA) {
	target.Add(b.mesh, matrix.gl(), color, b.material, b.Translucent)
}

func (b *Batch) RectDraw(target BatchTarget, bounds Rect) {
	b.RectDrawColorMask(target, bounds, RGBA{1, 1, 1, 1})
}

// TODO: Generalize this rectdraw logic. Copy paseted from Sprite
func (b *Batch) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	// pass.SetTexture(0, s.texture)
	// pass.Add(s.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, s.material)

	batchBounds := b.Bounds().Rect()
	matrix := Mat4Ident
	matrix.Scale(bounds.W()/batchBounds.W(), bounds.H()/batchBounds.H(), 1).Translate(bounds.W()/2+bounds.Min[0], bounds.H()/2+bounds.Min[1], 0)
	target.Add(b.mesh, matrix.gl(), mask, b.material, false)
}

func (b *Batch) Bounds() Box {
	return b.mesh.Bounds()
}

type Mesh struct {
	positions []glVec3
	normals   []glVec3
	colors    []glVec4
	texCoords []glVec2
	indices   []uint32
	bounds    Box

	origin Vec3

	// TODO: migrate towards
	buffer *VertexBuffer
}

func NewMesh() *Mesh {
	return &Mesh{
		positions: make([]glVec3, 0),
		normals:   make([]glVec3, 0),
		colors:    make([]glVec4, 0),
		texCoords: make([]glVec2, 0),
		indices:   make([]uint32, 0),
	}
}

func (m *Mesh) Buffer(pass *RenderPass, material Material, translucent bool) *Mesh {
	return &Mesh{
		buffer: pass.BufferMesh(m, material, translucent),
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
	m.origin = Vec3{}

	m.buffer = nil // TODO: manually delete?
}

func (m *Mesh) Draw(pass *RenderPass, matrix Mat4) {
	// pass.Add(m, matrix.gl(), RGBA{1.0, 1.0, 1.0, 1.0}, DefaultMaterial(), false)
	m.DrawColorMask(pass, matrix, White)
}

// TODO - This should accept image/color and call RGBA(). Would that be slower?
func (m *Mesh) DrawColorMask(pass *RenderPass, matrix Mat4, mask RGBA) {
	pass.Add(m, matrix.gl(), mask, DefaultMaterial(), false)
}

func (m *Mesh) Bounds() Box {
	return m.bounds
}

// TODO - should this be more like draw?
func (m *Mesh) Append(m2 *Mesh) {
	currentElement := uint32(len(m.positions))
	for i := range m2.indices {
		m.indices = append(m.indices, currentElement+m2.indices[i])
	}

	m.positions = append(m.positions, m2.positions...)
	m.normals = append(m.normals, m2.normals...)
	m.colors = append(m.colors, m2.colors...)
	m.texCoords = append(m.texCoords, m2.texCoords...)

	m.bounds = m.bounds.Union(m2.bounds)
}

// Changes the origin point of the mesh by translating all the geometry to the new origin. This shouldn't be called frequently
// Returns a newly allocated mesh and does not modify the original
func (originalMesh *Mesh) WithSetOrigin(newOrigin Vec3) *Mesh {
	if originalMesh.origin == newOrigin {
		return originalMesh
	} // Skip if we've already translated this amount
	// delta := pos.Sub(m.translation)

	newMesh := NewMesh()
	newMesh.Append(originalMesh)
	// TODO - should I do this in a different order?
	delta := newMesh.origin.Sub(newOrigin).gl()
	for i := range newMesh.positions {
		newMesh.positions[i] = delta.Add(newMesh.positions[i])
	}

	newMesh.origin = newOrigin

	newMesh.bounds = Box{
		Min: delta.Add(newMesh.bounds.Min.gl()).Float64(),
		Max: delta.Add(newMesh.bounds.Max.gl()).Float64(),
	}

	return newMesh
}

// Sets the color of every vertex
func (m *Mesh) SetColor(col RGBA) {
	v4Color := glVec4{float32(col.R), float32(col.G), float32(col.B), float32(col.A)}
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

func (m *Mesh) AppendQuadMesh(bounds Rect, uvBounds Rect, color RGBA) {
	positions := []glVec3{
		glVec3{float32(bounds.Max[0]), float32(bounds.Max[1]), float32(0.0)},
		glVec3{float32(bounds.Max[0]), float32(bounds.Min[1]), float32(0.0)},
		glVec3{float32(bounds.Min[0]), float32(bounds.Min[1]), float32(0.0)},
		glVec3{float32(bounds.Min[0]), float32(bounds.Max[1]), float32(0.0)},
	}
	// TODO normals
	colors := []glVec4{
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
	}
	texCoords := []glVec2{
		glVec2{float32(uvBounds.Max[0]), float32(uvBounds.Min[1])},
		glVec2{float32(uvBounds.Max[0]), float32(uvBounds.Max[1])},
		glVec2{float32(uvBounds.Min[0]), float32(uvBounds.Max[1])},
		glVec2{float32(uvBounds.Min[0]), float32(uvBounds.Min[1])},
	}

	inds := []uint32{
		0, 1, 3,
		1, 2, 3,
	}

	currentElement := uint32(len(m.positions))
	for i := range inds {
		m.indices = append(m.indices, currentElement+inds[i])
	}

	m.positions = append(m.positions, positions...)
	m.colors = append(m.colors, colors...)
	m.texCoords = append(m.texCoords, texCoords...)

	m.bounds = m.bounds.Union(bounds.ToBox())
}

// --------------------------------------------------------------------------------
// - Standalone meshes
// --------------------------------------------------------------------------------

// Basically a quad mesh, but with a centered position
func NewSpriteMesh(w, h float64, uvBounds Rect) *Mesh {
	return NewQuadMesh(R(-w/2, -h/2, w/2, h/2), uvBounds)
}

func NewQuadMesh(bounds Rect, uvBounds Rect) *Mesh {
	color := RGBA{1.0, 1.0, 1.0, 1.0}
	positions := []glVec3{
		glVec3{float32(bounds.Max[0]), float32(bounds.Max[1]), float32(0.0)},
		glVec3{float32(bounds.Max[0]), float32(bounds.Min[1]), float32(0.0)},
		glVec3{float32(bounds.Min[0]), float32(bounds.Min[1]), float32(0.0)},
		glVec3{float32(bounds.Min[0]), float32(bounds.Max[1]), float32(0.0)},
	}
	// TODO normals
	colors := []glVec4{
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
	}
	texCoords := []glVec2{
		glVec2{float32(uvBounds.Max[0]), float32(uvBounds.Min[1])},
		glVec2{float32(uvBounds.Max[0]), float32(uvBounds.Max[1])},
		glVec2{float32(uvBounds.Min[0]), float32(uvBounds.Max[1])},
		glVec2{float32(uvBounds.Min[0]), float32(uvBounds.Min[1])},
	}

	inds := []uint32{
		0, 1, 3,
		1, 2, 3,
	}

	return &Mesh{
		positions: positions,
		colors:    colors,
		texCoords: texCoords,
		indices:   inds,
		bounds:    bounds.ToBox(),
	}
}

func NewCubeMesh(size float64) *Mesh {
	f32size := float32(size / 2)

	positions := []glVec3{
		// Front face
		glVec3{-f32size, -f32size, f32size},
		glVec3{f32size, -f32size, f32size},
		glVec3{f32size, f32size, f32size},
		glVec3{-f32size, f32size, f32size},
		// Back face
		glVec3{-f32size, -f32size, -f32size},
		glVec3{-f32size, f32size, -f32size},
		glVec3{f32size, f32size, -f32size},
		glVec3{f32size, -f32size, -f32size},
		// Top face
		glVec3{-f32size, f32size, -f32size},
		glVec3{-f32size, f32size, f32size},
		glVec3{f32size, f32size, f32size},
		glVec3{f32size, f32size, -f32size},
		// Bottom face
		glVec3{-f32size, -f32size, -f32size},
		glVec3{f32size, -f32size, -f32size},
		glVec3{f32size, -f32size, f32size},
		glVec3{-f32size, -f32size, f32size},
		// Right face
		glVec3{f32size, -f32size, -f32size},
		glVec3{f32size, f32size, -f32size},
		glVec3{f32size, f32size, f32size},
		glVec3{f32size, -f32size, f32size},
		// Left face
		glVec3{-f32size, -f32size, -f32size},
		glVec3{-f32size, -f32size, f32size},
		glVec3{-f32size, f32size, f32size},
		glVec3{-f32size, f32size, -f32size},
	}

	col := glVec4{1.0, 1.0, 1.0, 1.0}
	colors := []glVec4{
		col, col, col, col,
		col, col, col, col,
		col, col, col, col,
		col, col, col, col,
		col, col, col, col,
	}

	// TODO normals
	normals := []glVec3{
		// Front face
		glVec3{0, 0, 1},
		glVec3{0, 0, 1},
		glVec3{0, 0, 1},
		glVec3{0, 0, 1},
		// Back face
		glVec3{0, 0, -1},
		glVec3{0, 0, -1},
		glVec3{0, 0, -1},
		glVec3{0, 0, -1},
		// Top face
		glVec3{0, 1, 0},
		glVec3{0, 1, 0},
		glVec3{0, 1, 0},
		glVec3{0, 1, 0},
		// Bottom face
		glVec3{0, -1, 0},
		glVec3{0, -1, 0},
		glVec3{0, -1, 0},
		glVec3{0, -1, 0},
		// Right face
		glVec3{1, 0, 0},
		glVec3{1, 0, 0},
		glVec3{1, 0, 0},
		glVec3{1, 0, 0},
		// Left face
		glVec3{-1, 0, 0},
		glVec3{-1, 0, 0},
		glVec3{-1, 0, 0},
		glVec3{-1, 0, 0},
	}

	// TODO texCoords
	texCoords := []glVec2{
		// Front face
		glVec2{-0, -0},
		glVec2{0, -0},
		glVec2{0, 0},
		glVec2{-0, 0},
		// Back face
		glVec2{-0, -0},
		glVec2{-0, 0},
		glVec2{0, 0},
		glVec2{0, -0},
		// Top face
		glVec2{-0, 0},
		glVec2{-0, 0},
		glVec2{0, 0},
		glVec2{0, 0},
		// Bottom face
		glVec2{-0, -0},
		glVec2{0, -0},
		glVec2{0, -0},
		glVec2{-0, -0},
		// Right face
		glVec2{0, -0},
		glVec2{0, 0},
		glVec2{0, 0},
		glVec2{0, -0},
		// Left face
		glVec2{-0, -0},
		glVec2{-0, -0},
		glVec2{-0, 0},
		glVec2{-0, 0},
	}

	indices := []uint32{
		0, 1, 2, 0, 2, 3, // front
		4, 5, 6, 4, 6, 7, // back
		8, 9, 10, 8, 10, 11, // top
		12, 13, 14, 12, 14, 15, // bottom
		16, 17, 18, 16, 18, 19, // right
		20, 21, 22, 20, 22, 23, // left
	}

	return &Mesh{
		positions: positions,
		normals:   normals,
		colors:    colors,
		texCoords: texCoords,
		indices:   indices,
		bounds: Box{
			Min: Vec3{-size, -size, -size},
			Max: Vec3{size, size, size},
		},
	}
}

//--------------------------------------------------------------------------------

func (m *Mesh) GetBuffer() *VertexBuffer {
	return m.buffer
}

func (m *Mesh) NumVerts() int {
	return len(m.positions)
}
func (m *Mesh) Indices() []uint32 {
	return m.indices
}

func (m *Mesh) Fill(pass *RenderPass, mat glMat4, mask RGBA, state BufferState) *VertexBuffer {
	numVerts := m.NumVerts()
	indices := m.Indices()
	vertexBuffer := pass.buffer.Reserve(state, indices, numVerts, pass.shader.tmpBuffers)
	batchToBuffers(pass.shader, m, mat, mask)

	return vertexBuffer
}

func batchToBuffers(shader *Shader, mesh *Mesh, mat32 glMat4, mask RGBA) {
	destBuffs := shader.tmpBuffers

	// Append all mesh buffers to shader buffers
	for bufIdx, attr := range shader.attrFmt {
		// TODO - I'm not sure of a good way to break up this switch statement
		switch attr.Swizzle {
		// Positions
		case PositionXY:
			posBuf := *(destBuffs[bufIdx]).(*[]glVec2)
			if mat32 == glMat4Ident {
				// If matrix is identity, don't transform anything
				for i := range mesh.positions {
					posBuf[i] = *(*glVec2)(mesh.positions[i][:2])
				}
			} else {
				for i := range mesh.positions {
					vec := mat32.Apply(mesh.positions[i])
					posBuf[i] = *(*glVec2)(vec[:2])
				}
			}

		case PositionXYZ:
			posBuf := *(destBuffs[bufIdx]).(*[]glVec3)
			if mat32 == glMat4Ident {
				// If matrix is identity, don't transform anything
				copy(posBuf, mesh.positions)

			} else {
				for i := range mesh.positions {
					vec := mat32.Apply(mesh.positions[i])
					posBuf[i] = vec
				}
			}

			// Normals
			// TODO - Renormalize if batching
			// case NormalXY:
			// 	normBuf := *(destBuffs[bufIdx]).(*[]Vec2)
			// 	for i := range mesh.normals {
			// 		vec := mesh.normals[i]
			// 		normBuf[i] = *(*Vec2)(vec[:2])
			// 	}

		case NormalXYZ:
			posBuf := *(destBuffs[bufIdx]).(*[]glVec3)
			if mat32 == glMat4Ident {
				// If matrix is identity, don't transform anything
				copy(posBuf, mesh.positions)
			} else {
				normMat32 := mat32.Inv().Transpose()
				normBuf := *(destBuffs[bufIdx]).(*[]glVec3)
				for i := range mesh.normals {
					vec := normMat32.Apply(mesh.normals[i])
					normBuf[i] = vec
				}
			}

		// Colors
		case ColorR:
			colBuf := *(destBuffs[bufIdx]).(*[]float32)
			for i := range mesh.colors {
				colBuf[i] = mesh.colors[i][0] * float32(mask.R)
			}
		case ColorRG:
			colBuf := *(destBuffs[bufIdx]).(*[]glVec2)
			for i := range mesh.colors {
				colBuf[i] = glVec2{
					mesh.colors[i][0] * float32(mask.R),
					mesh.colors[i][1] * float32(mask.G),
				}
			}
		case ColorRGB:
			colBuf := *(destBuffs[bufIdx]).(*[]glVec3)
			for i := range mesh.colors {
				colBuf[i] = glVec3{
					mesh.colors[i][0] * float32(mask.R),
					mesh.colors[i][1] * float32(mask.G),
					mesh.colors[i][2] * float32(mask.B),
				}
			}
		case ColorRGBA:
			colBuf := *(destBuffs[bufIdx]).(*[]glVec4)
			for i := range mesh.colors {
				colBuf[i] = glVec4{
					mesh.colors[i][0] * float32(mask.R),
					mesh.colors[i][1] * float32(mask.G),
					mesh.colors[i][2] * float32(mask.B),
					mesh.colors[i][3] * float32(mask.A),
				}
			}

		case TexCoordXY:
			texBuf := *(destBuffs[bufIdx]).(*[]glVec2)
			copy(texBuf, mesh.texCoords)
		default:
			panic(fmt.Sprintf("Unsupported %T: %+v", attr, attr))
		}
	}

	//================================================================================
	// TODO The hardcoding is a bit slower. Keeping it around in case I want to do some performance analysis
	// Notes: Ran gophermark with 1000000 gophers.
	// - Hardcoded: ~ 120 to 125 ms range
	// - Switch Statement: ~ 125 to 130 ms range
	// - Switch Statement (with shader changed to use vec2s for position): ~ 122 to 127 ms range
	// work and append
	// 	posBuf := *(destBuffs[0]).(*[]Vec3)
	// 	for i := range mesh.positions {
	// 		vec := c.matrix.Apply(mesh.positions[i])
	// 		posBuf[i] = vec
	// 	}

	// 	colBuf := *(destBuffs[1]).(*[]Vec4)
	// 	for i := range mesh.colors {
	// 		colBuf[i] = Vec4{
	// 			mesh.colors[i][0] * mask.R,
	// 			mesh.colors[i][1] * mask.G,
	// 			mesh.colors[i][2] * mask.B,
	// 			mesh.colors[i][3] * mask.A,
	// 		}
	// 	}

	// 	texBuf := *(destBuffs[2]).(*[]Vec2)
	// 	for i := range mesh.texCoords {
	// 		texBuf[i] = mesh.texCoords[i]
	// 	}
	//================================================================================
}
