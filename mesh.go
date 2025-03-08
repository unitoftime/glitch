package glitch

import (
	"fmt"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch/shaders"
)

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

func (m *Mesh) Buffer(shader *Shader) *Mesh {
	return &Mesh{
		buffer: shader.BufferMesh(m),
	}
}

func (m *Mesh) g() GeometryFiller {
	return GeometryFiller{
		mesh: m,
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

func (m *Mesh) Draw(target BatchTarget, matrix Mat4) {
	// pass.Add(m, matrix.gl(), RGBA{1.0, 1.0, 1.0, 1.0}, DefaultMaterial(), false)
	m.DrawColorMask(target, matrix, White)
}

// TODO - This should accept image/color and call RGBA(). Would that be slower?
func (m *Mesh) DrawColorMask(target BatchTarget, matrix Mat4, mask RGBA) {
	target.Add(m.g(), glm4(matrix), mask, DefaultMaterial(WhiteTexture()))
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
	delta := glv3(newMesh.origin.Sub(newOrigin))
	for i := range newMesh.positions {
		newMesh.positions[i] = delta.Add(newMesh.positions[i])
	}

	newMesh.origin = newOrigin

	newMesh.bounds = Box{
		Min: delta.Add(glv3(newMesh.bounds.Min)).Float64(),
		Max: delta.Add(glv3(newMesh.bounds.Max)).Float64(),
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
		glVec3{float32(bounds.Max.X), float32(bounds.Max.Y), float32(0.0)},
		glVec3{float32(bounds.Max.X), float32(bounds.Min.Y), float32(0.0)},
		glVec3{float32(bounds.Min.X), float32(bounds.Min.Y), float32(0.0)},
		glVec3{float32(bounds.Min.X), float32(bounds.Max.Y), float32(0.0)},
	}
	// TODO normals
	colors := []glVec4{
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
	}
	texCoords := []glVec2{
		glVec2{float32(uvBounds.Max.X), float32(uvBounds.Min.Y)},
		glVec2{float32(uvBounds.Max.X), float32(uvBounds.Max.Y)},
		glVec2{float32(uvBounds.Min.X), float32(uvBounds.Max.Y)},
		glVec2{float32(uvBounds.Min.X), float32(uvBounds.Min.Y)},
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
	return NewQuadMesh(glm.R(-w/2, -h/2, w/2, h/2), uvBounds)
}

func NewQuadMesh(bounds Rect, uvBounds Rect) *Mesh {
	color := RGBA{1.0, 1.0, 1.0, 1.0}
	positions := []glVec3{
		glVec3{float32(bounds.Max.X), float32(bounds.Max.Y), float32(0.0)},
		glVec3{float32(bounds.Max.X), float32(bounds.Min.Y), float32(0.0)},
		glVec3{float32(bounds.Min.X), float32(bounds.Min.Y), float32(0.0)},
		glVec3{float32(bounds.Min.X), float32(bounds.Max.Y), float32(0.0)},
	}
	// TODO normals
	colors := []glVec4{
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
		glVec4{float32(color.R), float32(color.G), float32(color.B), float32(color.A)},
	}
	texCoords := []glVec2{
		glVec2{float32(uvBounds.Max.X), float32(uvBounds.Min.Y)},
		glVec2{float32(uvBounds.Max.X), float32(uvBounds.Max.Y)},
		glVec2{float32(uvBounds.Min.X), float32(uvBounds.Max.Y)},
		glVec2{float32(uvBounds.Min.X), float32(uvBounds.Min.Y)},
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

func (m *Mesh) Fill(bufferPool *BufferPool, mat glMat4, mask RGBA) *VertexBuffer {
	numVerts := m.NumVerts()
	indices := m.Indices()
	vertexBuffer := bufferPool.Reserve(indices, numVerts, bufferPool.shader.tmpBuffers)
	batchToBuffers(bufferPool.shader, m, mat, mask)

	return vertexBuffer
}

func batchToBuffers(shader *Shader, mesh *Mesh, mat32 glMat4, mask RGBA) {
	destBuffs := shader.tmpBuffers

	// Append all mesh buffers to shader buffers
	for bufIdx, attr := range shader.attrFmt {
		// TODO - I'm not sure of a good way to break up this switch statement
		switch attr.Swizzle {
		// Positions

		// TODO: This is a pretty untested swizzle
		case shaders.PositionXY:
			posBuf := *(destBuffs[bufIdx]).(*[]glVec2)
			if mat32 == glMat4Ident {
				// If matrix is identity, don't transform anything
				for i := range mesh.positions {
					posBuf[i] = *(*glVec2)(mesh.positions[i][:2])
				}
			} else {
				for i := range mesh.positions {
					// vec := mat32.Apply(mesh.positions[i])
					// posBuf[i] = *(*glVec2)(vec[:2])
					vec := mat32.ApplyVec2(glVec2{mesh.positions[i][0], mesh.positions[i][1]})
					posBuf[i] = vec
				}
			}

		case shaders.PositionXYZ:
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

		case shaders.NormalXYZ:
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
		case shaders.ColorR:
			colBuf := *(destBuffs[bufIdx]).(*[]float32)
			for i := range mesh.colors {
				colBuf[i] = mesh.colors[i][0] * float32(mask.R)
			}
		case shaders.ColorRG:
			colBuf := *(destBuffs[bufIdx]).(*[]glVec2)
			for i := range mesh.colors {
				colBuf[i] = glVec2{
					mesh.colors[i][0] * float32(mask.R),
					mesh.colors[i][1] * float32(mask.G),
				}
			}
		case shaders.ColorRGB:
			colBuf := *(destBuffs[bufIdx]).(*[]glVec3)
			for i := range mesh.colors {
				colBuf[i] = glVec3{
					mesh.colors[i][0] * float32(mask.R),
					mesh.colors[i][1] * float32(mask.G),
					mesh.colors[i][2] * float32(mask.B),
				}
			}
		case shaders.ColorRGBA:
			colBuf := *(destBuffs[bufIdx]).(*[]glVec4)
			for i := range mesh.colors {
				colBuf[i] = glVec4{
					mesh.colors[i][0] * float32(mask.R),
					mesh.colors[i][1] * float32(mask.G),
					mesh.colors[i][2] * float32(mask.B),
					mesh.colors[i][3] * float32(mask.A),
				}
			}

		case shaders.TexCoordXY:
			texBuf := *(destBuffs[bufIdx]).(*[]glVec2)
			copy(texBuf, mesh.texCoords)
		default:
			panic(fmt.Sprintf("Unsupported %T: %+v", attr, attr))
		}
	}

	// TODO: Could use copy funcs if you want to restrict buffer types
	// for bufIdx, attr := range pass.shader.attrFmt {
	// 	switch attr.Swizzle {
	// 	case PositionXYZ:
	// 		buffer.buffers[bufIdx].SetData(mesh.positions)
	// 	case ColorRGBA:
	// 		buffer.buffers[bufIdx].SetData(mesh.colors)
	// 	case TexCoordXY:
	// 		buffer.buffers[bufIdx].SetData(mesh.texCoords)
	// 	default:
	// 		panic("unsupported")
	// 	}
	// }

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
