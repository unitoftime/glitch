package glitch

// type Buffer struct {
// 	buf *VertexBuffer
// }

// func NewBuffer() *Buffer {
// 	return &Buffer{}
// }

// func (b *Buffer) Add(mesh *Mesh, matrix Mat4, mask RGBA, material Material, translucent bool) {
// 	if b.material == nil {
// 		b.material = material
// 	} else {
// 		if b.material != material {
// 			panic("Materials must match inside a batch!")
// 		}
// 	}

// 	b.mesh.generation++

// 	// If anything translucent is added to the batch, then we will consider the entire thing translucent
// 	b.Translucent = b.Translucent || translucent

// 	mat := matrix.gl()

// 	// Append each index
// 	currentElement := uint32(len(b.mesh.positions))
// 	for i := range mesh.indices {
// 		b.mesh.indices = append(b.mesh.indices, currentElement + mesh.indices[i])
// 	}

// 	// Append each position
// 	for i := range mesh.positions {
// 		b.mesh.positions = append(b.mesh.positions, mat.Apply(mesh.positions[i]))
// 	}

// 	// Calculate the bounding box of the mesh we just merged in
// 	// Because we already figured out the first index of the new mesh (ie `currentElement`) we can just slice off the end of the new mesh
// 	posBuf := b.mesh.positions[int(currentElement):]
// 	min := posBuf[0]
// 	max := posBuf[0]
// 	for i := range posBuf {
// 		// X
// 		if posBuf[i][0] < min[0] {
// 			min[0] = posBuf[i][0]
// 		}
// 		if posBuf[i][0] > max[0] {
// 			max[0] = posBuf[i][0]
// 		}

// 		// Y
// 		if posBuf[i][1] < min[1] {
// 			min[1] = posBuf[i][1]
// 		}
// 		if posBuf[i][1] > max[1] {
// 			max[1] = posBuf[i][1]
// 		}

// 		// Z
// 		if posBuf[i][2] < min[2] {
// 			min[2] = posBuf[i][2]
// 		}
// 		if posBuf[i][2] > max[2] {
// 			max[2] = posBuf[i][2]
// 		}
// 	}

// 	newBounds := Box{
// 		Min: Vec3{float64(min[0]), float64(min[1]), float64(min[2])},
// 		Max: Vec3{float64(max[0]), float64(max[1]), float64(max[2])},
// 	}
// 	b.mesh.bounds = b.mesh.bounds.Union(newBounds)

// 	renormalizeMat := matrix.Inv().Transpose().gl()
// 	for i := range mesh.normals {
// 		b.mesh.normals = append(b.mesh.normals, renormalizeMat.Apply(mesh.normals[i]))
// 	}

// 	for i := range mesh.colors {
// 		// TODO - vec4 mult function
// 		b.mesh.colors = append(b.mesh.colors, glVec4{
// 			mesh.colors[i][0] * float32(mask.R),
// 			mesh.colors[i][1] * float32(mask.G),
// 			mesh.colors[i][2] * float32(mask.B),
// 			mesh.colors[i][3] * float32(mask.A),
// 		})
// 	}

// 	// TODO - is a copy faster?
// 	for i := range mesh.texCoords {
// 		b.mesh.texCoords = append(b.mesh.texCoords, mesh.texCoords[i])
// 	}

// 	// if b.material == nil {
// 	// 	b.material = material
// 	// } else {
// 	// 	if b.material != material {
// 	// 		panic("Materials must match inside a batch!")
// 	// 	}
// 	// }

// 	// mat := matrix.gl()

// 	// posBuf := make([]glVec3, len(mesh.positions))
// 	// for i := range mesh.positions {
// 	// 	posBuf[i] = mat.Apply(mesh.positions[i])
// 	// }

// 	// min := posBuf[0]
// 	// max := posBuf[0]
// 	// for i := range posBuf {
// 	// 	// X
// 	// 	if posBuf[i][0] < min[0] {
// 	// 		min[0] = posBuf[i][0]
// 	// 	}
// 	// 	if posBuf[i][0] > max[0] {
// 	// 		max[0] = posBuf[i][0]
// 	// 	}

// 	// 	// Y
// 	// 	if posBuf[i][1] < min[1] {
// 	// 		min[1] = posBuf[i][1]
// 	// 	}
// 	// 	if posBuf[i][1] > max[1] {
// 	// 		max[1] = posBuf[i][1]
// 	// 	}

// 	// 	// Z
// 	// 	if posBuf[i][2] < min[2] {
// 	// 		min[2] = posBuf[i][2]
// 	// 	}
// 	// 	if posBuf[i][2] > max[2] {
// 	// 		max[2] = posBuf[i][2]
// 	// 	}
// 	// }

// 	// newBounds := Box{
// 	// 	Min: Vec3{float64(min[0]), float64(min[1]), float64(min[2])},
// 	// 	Max: Vec3{float64(max[0]), float64(max[1]), float64(max[2])},
// 	// }

// 	// renormalizeMat := matrix.Inv().Transpose().gl()
// 	// normBuf := make([]glVec3, len(mesh.normals))
// 	// for i := range mesh.normals {
// 	// 	normBuf[i] = renormalizeMat.Apply(mesh.normals[i])
// 	// }

// 	// colBuf := make([]glVec4, len(mesh.colors))
// 	// for i := range mesh.colors {
// 	// 	// TODO - vec4 mult function
// 	// 	colBuf[i] = glVec4{
// 	// 		mesh.colors[i][0] * float32(mask.R),
// 	// 		mesh.colors[i][1] * float32(mask.G),
// 	// 		mesh.colors[i][2] * float32(mask.B),
// 	// 		mesh.colors[i][3] * float32(mask.A),
// 	// 	}
// 	// }

// 	// // TODO - is a copy faster?
// 	// texBuf := make([]glVec2, len(mesh.texCoords))
// 	// for i := range mesh.texCoords {
// 	// 	texBuf[i] = mesh.texCoords[i]
// 	// }

// 	// indices := make([]uint32, len(mesh.indices))
// 	// for i := range mesh.indices {
// 	// 	indices[i] = mesh.indices[i]
// 	// }

// 	// m2 := &Mesh{
// 	// 	positions: posBuf,
// 	// 	normals: normBuf,
// 	// 	colors: colBuf,
// 	// 	texCoords: texBuf,
// 	// 	indices: indices,
// 	// 	bounds: newBounds,
// 	// }

// 	// b.mesh.Append(m2)
// }

// func (b *Buffer) Clear() {
// 	b.mesh.Clear()
// 	b.material = nil
// 	b.Translucent = false
// }

// func (b *Buffer) Draw(target BatchTarget, matrix Mat4) {
// 	target.Add(b.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, b.material, b.Translucent)
// }

// func (b *Buffer) DrawColorMask(target BatchTarget, matrix Mat4, color RGBA) {
// 	target.Add(b.mesh, matrix, color, b.material, b.Translucent)
// }

// func (b *Buffer) RectDraw(target BatchTarget, bounds Rect) {
// 	b.RectDrawColorMask(target, bounds, RGBA{1, 1, 1, 1})
// }
// // TODO: Generalize this rectdraw logic. Copy paseted from Sprite
// func (b *Buffer) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
// 	// pass.SetTexture(0, s.texture)
// 	// pass.Add(s.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, s.material)

// 	batchBounds := b.Bounds().Rect()
// 	matrix := Mat4Ident
// 	matrix.Scale(bounds.W() / batchBounds.W(), bounds.H() / batchBounds.H(), 1).Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)
// 	target.Add(b.mesh, matrix, mask, b.material, false)
// }

// func (b *Buffer) Bounds() Box {
// 	return b.mesh.Bounds()
// }
