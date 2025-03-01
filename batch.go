package glitch

import "fmt"

type meshDraw struct {
	filler      GeometryFiller
	matrix      glMat4
	mask        RGBA
	material    Material
	translucent bool
}

// For batching multiple sprites into one
type DrawBatch struct {
	draws []meshDraw

	boundsSet bool
	bounds    Box
}

func NewDrawBatch() *DrawBatch {
	return &DrawBatch{
		draws: make([]meshDraw, 0),
	}
}

// func (b *DrawBatch) Buffer(pass *RenderPass) *DrawBatch {
// 	return &DrawBatch{
// 		mesh: b.mesh.Buffer(pass, b.material, b.Translucent),
// 		material: b.material,
// 		Translucent: b.Translucent,
// 	}
// }

func (b *DrawBatch) Add(filler GeometryFiller, matrix glMat4, mask RGBA, material Material, translucent bool) {
	b.draws = append(b.draws, meshDraw{
		filler:      filler,
		matrix:      matrix,
		mask:        mask,
		material:    material,
		translucent: translucent,
	})

	newBounds := filler.Bounds().Apply(matrix.Mat4())
	// TODO: Does this improve performance?
	// if matrix != glMat4Ident {
	// 	newBounds = newBounds.Apply(matrix)
	// }
	if b.boundsSet {
		b.bounds = b.bounds.Union(newBounds)
	} else {
		b.boundsSet = true
		b.bounds = newBounds
	}
}

func (b *DrawBatch) Clear() {
	b.draws = b.draws[:0]
	b.boundsSet = false
	b.bounds = Box{}
}

func (b *DrawBatch) Draw(target BatchTarget, matrix Mat4) {
	for i := range b.draws {
		mat := glm4(matrix)
		mat.Mul(&b.draws[i].matrix)
		target.Add(b.draws[i].filler, mat, b.draws[i].mask, b.draws[i].material, b.draws[i].translucent)
	}
	// target.Add(b.mesh, matrix.gl(), RGBA{1.0, 1.0, 1.0, 1.0}, b.material, b.Translucent)
	// b.DrawColorMask(target, matrix, White)
}

func (b *DrawBatch) DrawColorMask(target BatchTarget, matrix Mat4, color RGBA) {
	for i := range b.draws {
		mat := glm4(matrix)
		mat.Mul(&b.draws[i].matrix)

		mask := b.draws[i].mask.Mult(color)
		target.Add(b.draws[i].filler, mat, mask, b.draws[i].material, b.draws[i].translucent)
	}

	// target.Add(b.mesh, matrix.gl(), color, b.material, b.Translucent)
	// for i := range b.draws {
	// 	target.Add(b.draws[i].mesh, b.draws[i].matrix, b.draws[i].color, b.draws[i].material, b.draws[i].translucent)
	// }
}

func (b *DrawBatch) RectDraw(target BatchTarget, bounds Rect) {
	batchBounds := b.Bounds().Rect()
	matrix := Mat4Ident
	matrix.Scale(bounds.W()/batchBounds.W(), bounds.H()/batchBounds.H(), 1).Translate(bounds.W()/2+bounds.Min.X, bounds.H()/2+bounds.Min.Y, 0)

	b.Draw(target, matrix)

	// b.RectDrawColorMask(target, bounds, RGBA{1, 1, 1, 1})
}

// TODO: Generalize this rectdraw logic. Copy paseted from Sprite
func (b *DrawBatch) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	batchBounds := b.Bounds().Rect()
	matrix := Mat4Ident
	matrix.Scale(bounds.W()/batchBounds.W(), bounds.H()/batchBounds.H(), 1).Translate(bounds.W()/2+bounds.Min.X, bounds.H()/2+bounds.Min.Y, 0)

	b.DrawColorMask(target, matrix, mask)

	// // pass.SetTexture(0, s.texture)
	// // pass.Add(s.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, s.material)

	// batchBounds := b.Bounds().Rect()
	// matrix := Mat4Ident
	// matrix.Scale(bounds.W() / batchBounds.W(), bounds.H() / batchBounds.H(), 1).Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)
	// target.Add(b.mesh, matrix.gl(), mask, b.material, false)
}

func (b *DrawBatch) Bounds() Box {
	return b.bounds
}

// type Batcher struct {
// 	shader *Shader
// 	lastBuffer *VertexBuffer
// 	target Target
// }

// func NewBatcher() *Batcher {
// 	return &Batcher{} // TODO: Default case for shader?
// }

// func (b *Batcher) SetShader(shader *Shader) {
// 	b.Flush() // TODO: You technically only need to do this if it will change the uniform

// 	b.shader = shader
// }

// func (b *Batcher) SetUniform(name string, val any) {
// 	b.Flush() // TODO: You technically only need to do this if it will change the uniform

// 	b.shader.SetUniform(name, val)
// }

// func (b *Batcher) Clear() {

// }

// func (b *Batcher) Add(filler GeometryFiller, mat glMat4, mask RGBA, material Material, translucent bool) {
// 	if filler == nil { return } // Skip nil meshes

// 	buffer := filler.GetBuffer()
// 	if buffer != nil {
// 		b.drawCall(buffer, mat)
// 		return
// 	}

// 	// Note: Captured in shader.pool
// 	// 1. If you switch materials, then draw the last one
// 	// 2. If you fill up then draw the last one
// 	state := BufferState{material, BlendModeNormal} // TODO: blendmode and track full state some better way
// 	vertexBuffer := filler.Fill(b.shader.pool, mat, mask, state)

// 	// If vertexBuffer has changed then we want to draw the last one
// 	if b.lastBuffer != nil && vertexBuffer != b.lastBuffer {
// 		b.drawCall(b.lastBuffer, glMat4Ident)
// 	}

// 	b.lastBuffer = vertexBuffer
// }

// // Draws the current buffer and progress the shader pool to the next available
// func (b *Batcher) Flush() {
// 	if b.lastBuffer == nil { return }

// 	b.drawCall(b.lastBuffer, glMat4Ident)
// 	b.lastBuffer = nil
// 	b.shader.pool.gotoNextClean()
// }

// // Executes a drawcall with ...
// func (b *Batcher) drawCall(buffer *VertexBuffer, mat glMat4) {
// 	if b.target != nil {
// 		b.target.Bind()
// 	}

// 	// TODO: Set all uniforms
// 	// 1. camera
// 	// 2. materials

// 	b.shader.Bind() // TODO: global State cache

// 	// TODO: rewrite how buffer state works for immediate mode case
// 	buffer.state.Bind(b.shader)

// 	// TOOD: Maybe pass this into VertexBuffer.Draw() func
// 	ok := b.shader.SetUniform("model", mat)
// 	if !ok {
// 		panic("Error setting model uniform - all shaders must have 'model' uniform")
// 	}

// 	buffer.Draw()
// }

// --------------------------------------------------------------------------------
// For batching multiple meshes into one mesh
type Batch struct {
	mesh        *Mesh
	material    Material
	materialSet bool
	Translucent bool
}

func NewBatch() *Batch {
	return &Batch{
		mesh:     NewMesh(),
		materialSet: false,
	}
}

func (b *Batch) Buffer() *Batch {
	if !b.materialSet {
		return NewBatch() // Nothing was ever set, just return a new blank batch
	}

	shader := b.material.shader
	return &Batch{
		mesh:        b.mesh.Buffer(shader, b.Translucent),
		materialSet: true,
		material:    b.material,
		Translucent: b.Translucent,
	}
}

// TODO - It may be faster to copy all the bufs to the destination and then operate on them there. that might save you a copy
// TODO: should I maintain a translucent and non-translucent batch mesh?
// TODO: Fix the interface here to only allow meshes to be drawn to batches
func (b *Batch) Add(filler GeometryFiller, matrix glMat4, mask RGBA, material Material, translucent bool) {
	// mesh := filler.(*Mesh) // TODO: Hack
	mesh := filler.mesh // TODO: Is this safe? Assumes everything added has a mesh

	if !b.materialSet {
		b.materialSet = true
		b.material = material
	} else {
		if b.material != material {
			fmt.Printf("setmaterial (old -> new):\n%+v\n%+v\n", b.material, material)
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
	b.materialSet = false
	b.material = Material{}
	b.Translucent = false
}

func (b *Batch) Draw(target BatchTarget, matrix Mat4) {
	if !b.materialSet {
		return
	}
	target.Add(b.mesh.g(), glm4(matrix), RGBA{1.0, 1.0, 1.0, 1.0}, b.material, b.Translucent)
	// b.DrawColorMask(target, matrix, White)
}

func (b *Batch) DrawColorMask(target BatchTarget, matrix Mat4, color RGBA) {
	if !b.materialSet {
		return
	}
	target.Add(b.mesh.g(), glm4(matrix), color, b.material, b.Translucent)
}

func (b *Batch) RectDraw(target BatchTarget, bounds Rect) {
	if !b.materialSet {
		return
	}
	b.RectDrawColorMask(target, bounds, RGBA{1, 1, 1, 1})
}

// TODO: Generalize this rectdraw logic. Copy paseted from Sprite
func (b *Batch) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	if !b.materialSet {
		return
	}
	// pass.SetTexture(0, s.texture)
	// pass.Add(s.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, s.material)

	batchBounds := b.Bounds().Rect()
	matrix := Mat4Ident
	matrix.Scale(bounds.W()/batchBounds.W(), bounds.H()/batchBounds.H(), 1).Translate(bounds.W()/2+bounds.Min.X, bounds.H()/2+bounds.Min.Y, 0)
	target.Add(b.mesh.g(), glm4(matrix), mask, b.material, false)
}

func (b *Batch) Bounds() Box {
	return b.mesh.Bounds()
}
