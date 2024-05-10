package glitch

type meshDraw struct {
	mesh *Mesh
	matrix glMat4
	mask RGBA
	material Material
	translucent bool
}

// For batching multiple sprites into one
type DrawBatch struct {
	draws []meshDraw

	boundsSet bool
	bounds Box
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

func (b *DrawBatch) Add(mesh *Mesh, matrix glMat4, mask RGBA, material Material, translucent bool) {
	b.draws = append(b.draws, meshDraw{
		mesh: mesh,
		matrix: matrix,
		mask: mask,
		material: material,
		translucent: translucent,
	})

	newBounds := mesh.Bounds().Apply(matrix)
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
		mat := matrix.gl()
		mat.Mul(&b.draws[i].matrix)
		target.Add(b.draws[i].mesh, mat, b.draws[i].mask, b.draws[i].material, b.draws[i].translucent)
	}
	// target.Add(b.mesh, matrix.gl(), RGBA{1.0, 1.0, 1.0, 1.0}, b.material, b.Translucent)
	// b.DrawColorMask(target, matrix, White)
}

func (b *DrawBatch) DrawColorMask(target BatchTarget, matrix Mat4, color RGBA) {
	for i := range b.draws {
		mat := matrix.gl()
		mat.Mul(&b.draws[i].matrix)

		mask := b.draws[i].mask.Mult(color)
		target.Add(b.draws[i].mesh, mat, mask, b.draws[i].material, b.draws[i].translucent)
	}

	// target.Add(b.mesh, matrix.gl(), color, b.material, b.Translucent)
	// for i := range b.draws {
	// 	target.Add(b.draws[i].mesh, b.draws[i].matrix, b.draws[i].color, b.draws[i].material, b.draws[i].translucent)
	// }
}

func (b *DrawBatch) RectDraw(target BatchTarget, bounds Rect) {
	batchBounds := b.Bounds().Rect()
	matrix := Mat4Ident
	matrix.Scale(bounds.W() / batchBounds.W(), bounds.H() / batchBounds.H(), 1).Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)

	b.Draw(target, matrix)

	// b.RectDrawColorMask(target, bounds, RGBA{1, 1, 1, 1})
}

// TODO: Generalize this rectdraw logic. Copy paseted from Sprite
func (b *DrawBatch) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	batchBounds := b.Bounds().Rect()
	matrix := Mat4Ident
	matrix.Scale(bounds.W() / batchBounds.W(), bounds.H() / batchBounds.H(), 1).Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)

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
