package glitch

type Sprite struct {
	mesh *Mesh
	bounds Rect // Represents the bounds centered on (0, 0)
	frame Rect // Represents the bounds inside the spritesheet
	texture *Texture
	material Material
	Translucent bool
	uvBounds Rect
}

func NewSprite(texture *Texture, bounds Rect) *Sprite {
	uvBounds := R(
		bounds.Min[0] / float64(texture.width),
		bounds.Min[1] / float64(texture.height),
		bounds.Max[0] / float64(texture.width),
		bounds.Max[1] / float64(texture.height),
	)

	mesh := NewSpriteMesh(bounds.W(), bounds.H(), uvBounds)
	return &Sprite{
		mesh: mesh,
		// bounds: bounds,
		frame: bounds,
		bounds: bounds.Moved(bounds.Center().Scaled(-1)),
		// bounds: mesh.Bounds().Rect(),
		texture: texture,
		material: NewSpriteMaterial(texture),
		uvBounds: uvBounds,
	}
}

// Changes the origin point of the sprite by translating all the geometry to the new origin. This shouldn't be called frequently. The default origin is around the center of the sprite
// Returns a newly allocated mesh and does not modify the original
func (s Sprite) WithSetOrigin(origin Vec3) Sprite {
	s.mesh = s.mesh.WithSetOrigin(origin)
	return s
}

// func (s *Sprite) DrawToBatch(target *Batch, matrix Mat4) {
// 	target.Add(s.mesh, matrix, White, s.material, s.Translucent)
// 	// s.DrawColorMask(target, matrix, White)
// }

func (s *Sprite) Draw(target BatchTarget, matrix Mat4) {
	s.DrawColorMask(target, matrix, White)
}
func (s *Sprite) DrawColorMask(target BatchTarget, matrix Mat4, mask RGBA) {
	target.Add(s.mesh, matrix.gl(), mask, s.material, s.Translucent)
}

func (s *Sprite) RectDraw(target BatchTarget, bounds Rect) {
	s.RectDrawColorMask(target, bounds, White)
}
func (s *Sprite) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	matrix := Mat4Ident
	matrix.Scale(bounds.W() / s.bounds.W(), bounds.H() / s.bounds.H(), 1).Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)
	s.DrawColorMask(target, matrix, mask)
}

func (s *Sprite) Bounds() Rect {
	return s.bounds
}
func (s *Sprite) Frame() Rect {
	return s.frame
}

func (s *Sprite) SetTextureBounds(bounds Rect) {
	s.frame = bounds
	s.bounds = bounds.Moved(bounds.Center().Scaled(-1))
	s.uvBounds = R(
		bounds.Min[0] / float64(s.texture.width),
		bounds.Min[1] / float64(s.texture.height),
		bounds.Max[0] / float64(s.texture.width),
		bounds.Max[1] / float64(s.texture.height),
	)

	w := bounds.W()
	h := bounds.H()
	meshBounds := R(-w/2, -h/2, w/2, h/2)

	s.mesh.positions[0] = glVec3{float32(meshBounds.Max[0]), float32(meshBounds.Max[1]), float32(0.0)}
	s.mesh.positions[1] = glVec3{float32(meshBounds.Max[0]), float32(meshBounds.Min[1]), float32(0.0)}
	s.mesh.positions[2] = glVec3{float32(meshBounds.Min[0]), float32(meshBounds.Min[1]), float32(0.0)}
	s.mesh.positions[3] = glVec3{float32(meshBounds.Min[0]), float32(meshBounds.Max[1]), float32(0.0)}

	s.mesh.texCoords[0] = glVec2{float32(s.uvBounds.Max[0]), float32(s.uvBounds.Min[1])}
	s.mesh.texCoords[1] = glVec2{float32(s.uvBounds.Max[0]), float32(s.uvBounds.Max[1])}
	s.mesh.texCoords[2] = glVec2{float32(s.uvBounds.Min[0]), float32(s.uvBounds.Max[1])}
	s.mesh.texCoords[3] = glVec2{float32(s.uvBounds.Min[0]), float32(s.uvBounds.Min[1])}
}

// // TODO: This stuff was somehow just about the same speed as the mesh fill function. Not sure if its worth it unless I can make it way faster
// func (s *Sprite) GetBuffer() *VertexBuffer {
// 	return nil
// }
// // Note: For caching purposes
// var spriteQuadIndices = []uint32{
// 	0, 1, 3,
// 	1, 2, 3,
// }
// func (s *Sprite) Fill(pass *RenderPass, mat glMat4, mask RGBA, state BufferState) *VertexBuffer {
// 	numVerts := 4
// 	vertexBuffer := pass.buffer.Reserve(state, spriteQuadIndices, numVerts, pass.shader.tmpBuffers)

// 	destBuffs := pass.shader.tmpBuffers
// 	for bufIdx, attr := range pass.shader.attrFmt {
// 		// TODO - I'm not sure of a good way to break up this switch statement
// 		switch attr.Swizzle {
// 		case PositionXYZ:
// 			bounds := s.bounds.Box()
// 			min := bounds.Min.gl()
// 			max := bounds.Max.gl()
// 			if mat != glMat4Ident {
// 				min = mat.Apply(min)
// 				max = mat.Apply(max)
// 			}

// 			// TODO: Depth? Right now I just do min[2] b/c max and min should be on same Z axis
// 			posBuf := *(destBuffs[bufIdx]).(*[]glVec3)
// 			posBuf[0] = glVec3{float32(max[0]), float32(max[1]), float32(min[2])}
// 			posBuf[1] = glVec3{float32(max[0]), float32(min[1]), float32(min[2])}
// 			posBuf[2] = glVec3{float32(min[0]), float32(min[1]), float32(min[2])}
// 			posBuf[3] = glVec3{float32(min[0]), float32(max[1]), float32(min[2])}

// 		case ColorRGBA:
// 			colBuf := *(destBuffs[bufIdx]).(*[]glVec4)
// 			color := mask.gl()
// 			colBuf[0] = color
// 			colBuf[1] = color
// 			colBuf[2] = color
// 			colBuf[3] = color
// 		case TexCoordXY:
// 			texBuf := *(destBuffs[bufIdx]).(*[]glVec2)
// 			texBuf[0] = glVec2{float32(s.uvBounds.Max[0]), float32(s.uvBounds.Min[1])}
// 			texBuf[1] = glVec2{float32(s.uvBounds.Max[0]), float32(s.uvBounds.Max[1])}
// 			texBuf[2] = glVec2{float32(s.uvBounds.Min[0]), float32(s.uvBounds.Max[1])}
// 			texBuf[3] = glVec2{float32(s.uvBounds.Min[0]), float32(s.uvBounds.Min[1])}
// 		default:
// 			panic("Unsupported")
// 		}
// 	}

// 	return vertexBuffer
// }

//--------------------------------------------------------------------------------

type NinePanelSprite struct {
	sprites []*Sprite
	border Rect
	bounds Rect
	// Mask RGBA // This represents the default color mask to draw with (unless one is passed in via a draw function, Example: *Mask)
	Scale float64
}

func SpriteToNinePanel(sprite *Sprite, border Rect) *NinePanelSprite {
	return NewNinePanelSprite(sprite.texture, sprite.frame, border)
}

func NewNinePanelSprite(texture *Texture, bounds Rect, border Rect) *NinePanelSprite {
	fullBounds := bounds

	top := bounds.CutBottom(border.Max[1])
	bot := bounds.CutTop(border.Min[1])

	topLeft := top.CutLeft(border.Min[0])
	topRight := top.CutRight(border.Max[0])

	botLeft := bot.CutLeft(border.Min[0])
	botRight := bot.CutRight(border.Max[0])
	left := bounds.CutLeft(border.Min[0])
	right := bounds.CutRight(border.Max[0])

	rects := []Rect{
		bounds, // Center

		top, // Top
		bot, // Bot
		left, // Left
		right, // Right

		topLeft, // TL
		topRight, // TR
		botLeft, // BL
		botRight, // BR
	}

	sprites := make([]*Sprite, 9)
	for i := range rects {
		// TODO - instead of sprites, use quadmesh so I can manage a little more tightly
		sprites[i] = NewSprite(texture, rects[i])
	}

	return &NinePanelSprite{
		sprites: sprites,
		bounds: fullBounds,
		border: border,
		// Mask: White,
		Scale: 1,
	}
}

func (s *NinePanelSprite) SetTranslucent(translucent bool) {
	for i := range s.sprites {
		s.sprites[i].Translucent = translucent
	}
}

// Should 'matrix' just be scale and rotation? to scale up and down border pieces
// func (s *NinePanelSprite) Draw(pass *RenderPass, bounds Rect, matrix Mat4) {
// 	s.DrawColorMask(pass, bounds, matrix, RGBA{1.0, 1.0, 1.0, 1.0})
// }

func (s *NinePanelSprite) RectDraw(pass BatchTarget, bounds Rect) {
	s.RectDrawColorMask(pass, bounds, White)
}

func (s *NinePanelSprite) RectDrawColorMask(pass BatchTarget, rect Rect, mask RGBA) {
	// fmt.Println("here")
	// fmt.Println(bounds.W(), bounds.H())

	border := R(
		s.Scale * s.border.Min[0],
		s.Scale * s.border.Min[1],
		s.Scale * s.border.Max[0],
		s.Scale * s.border.Max[1])

	top := rect.CutTop(border.Max[1])
	bot := rect.CutBottom(border.Min[1])

	topLeft := top.CutLeft(border.Min[0])
	topRight := top.CutRight(border.Max[0])
	botLeft := bot.CutLeft(border.Min[0])
	botRight := bot.CutRight(border.Max[0])
	left := rect.CutLeft(border.Min[0])
	right := rect.CutRight(border.Max[0])

	destRects := [9]Rect{
		rect, //center

		top, // Top
		bot, // Bot
		left, // Left
		right, // Right

		topLeft, // TL
		topRight, // TR
		botLeft, // BL
		botRight, // BR
	}

	matrix := Mat4Ident
	for i := range s.sprites {
		// fmt.Println(destRects[i].W(), destRects[i].H())
		matrix = Mat4Ident
		matrix.Scale(destRects[i].W() / s.sprites[i].bounds.W(), destRects[i].H() / s.sprites[i].bounds.H(), 1).Translate(destRects[i].W()/2 + destRects[i].Min[0], destRects[i].H()/2 + destRects[i].Min[1], 0)
		// pass.Add(s.sprites[i], matrix, mask, s.sprites[i].material, false)
		s.sprites[i].DrawColorMask(pass, matrix, mask)
	}
}

func (s *NinePanelSprite) Bounds() Rect {
	return s.bounds
}

// The bounds of the borders rect
func (s *NinePanelSprite) Border() Rect {
	return s.border.Scaled(s.Scale)
	// return s.bounds.Pad(s.border.Scaled(-1))
}

// // type Geometry struct {
// // 	format []GeomFormat
// // }

// // func NewGeometry() *Geometry {
	
// // }

// // type Sprite struct {
// // 	texture *Texture
// // 	bounds Rect
// // 	// Indices    []uint32
// // 	positions  []float32
// // 	// // Normals []float32
// // 	// TexCoords  []float32
// // 	geomFormat []GeomFormat
// // }

// type GeomFormat int
// const (
// 	GeomPosX GeomFormat = iota
// 	GeomPosY
// 	GeomPosZ
// 	GeomPosW
// 	GeomColR
// 	GeomColG
// 	GeomColB
// 	GeomColA
// 	GeomTexU
// 	GeomTexV
// 	GeomLast
// )

// // func NewSprite(texture *Texture, bounds Rect) *Sprite {
// // 	return &Sprite{
// // 		texture: texture,
// // 		bounds: bounds,

// // 		// positions: []float32{
// // 		// 	s.bounds.W()	, s.bounds.H(), 0.0,
// // 		// 	s.bounds.W()	, 0           , 0.0,
// // 		// 	0							, 0           , 0.0,
// // 		// 	0							, s.bounds.H(), 0.0,
// // 		// },
// // 	}
// // 	// return Geometry{
// // 	// 	Indices: []uint32{
// // 	// 		0, 1, 2,
// // 	// 		1, 2, 3,
// // 	// 	},
// // 	// 	Positions: []float32{
// // 	// 		1.0, 1.0, 0.0, // Top Right
// // 	// 		1.0, 0.0, 0.0, // Bot Right
// // 	// 		0.0, 0.0, 0.0, // Bot Left
// // 	// 		0.0, 1.0, 0.0, // Top Left
// // 	// 	},
// // 	// 	// Notably, texture coords are flipped
// // 	// 	TexCoords: []float32{
// // 	// 		1.0, 0.0, // Top Right
// // 	// 		1.0, 1.0, // Bot Right
// // 	// 		0.0, 1.0, // Bot Left
// // 	// 		0.0, 0.0, // Top Left
// // 	// 	},
// // 	// }
// // }

// // // func (s *Sprite) Draw(buffer *VertexBuffer, x, y float32) {
// // // 	color := RGBA{1.0, 1.0, 1.0, 1.0}

// // // 	// verts := []float32{
// // // 	// 	x + s.bounds.W(), y + s.bounds.H(), 0.0,  color.R, color.G, color.B,  1.0, 0.0,
// // // 	// 	x + s.bounds.W(), y               , 0.0,  color.R, color.G, color.B,  1.0, 1.0,
// // // 	// 	x               , y               , 0.0,  color.R, color.G, color.B,  0.0, 1.0,
// // // 	// 	x               , y + s.bounds.H(), 0.0,  color.R, color.G, color.B,  0.0, 0.0,
// // // 	// }
// // // 	// inds := []uint32{
// // // 	// 	0, 1, 3,
// // // 	// 	1, 2, 3,
// // // 	// }
// // // 	// buffer.Add(verts, inds)

// // // 	// Works
// // // 	// positions := []float32{
// // // 	// 	x + s.bounds.W()	, y + s.bounds.H(), 0.0,
// // // 	// 	x + s.bounds.W()	, y               , 0.0,
// // // 	// 	x							    , y               , 0.0,
// // // 	// 	x							    , y + s.bounds.H(), 0.0,
// // // 	// }
// // // 	// colors := []float32{
// // // 	// 	color.R, color.G, color.B,
// // // 	// 	color.R, color.G, color.B,
// // // 	// 	color.R, color.G, color.B,
// // // 	// 	color.R, color.G, color.B,
// // // 	// }
// // // 	// texCoords := []float32{
// // // 	// 	1.0, 0.0,
// // // 	// 	1.0, 1.0,
// // // 	// 	0.0, 1.0,
// // // 	// 	0.0, 0.0,
// // // 	// }

// // // 	// inds := []uint32{
// // // 	// 	0, 1, 3,
// // // 	// 	1, 2, 3,
// // // 	// }

// // // 	// buffer.Add(positions, colors, texCoords, inds)

// // // 	// Maybe continue down this path, but instead of multiple buffers, just interleave?
// // // 	// I think since most geom will be pulled from files we can read the geom in and build out the dataset fairly easily? I don't like that this uses GeomFormat tho
// // // 	// how do you handle dynamic things, like position shifts and colors though?
// // // 	// Note - this is in memory batching, rather than just doing a different draw call.
// // // 	// geometry := []map[GeomFormat]float32{
// // // 	// 	GeomPosX: x + s.bounds.W(),
// // // 	// 	GeomPosY: y + s.bounds.H(),
// // // 	// 	GeomPosZ: 0.0,
// // // 	// }

// // // 	// numVerts := 4
// // // 	// geometry := make([][]float32, GeomLast)
// // // 	// for i := range geometry {
// // // 	// 	geometry[i] = make([]float32, numVerts)
// // // 	// }

// // // 	// geometry[GeomPosX] = []float32{ s.bounds.W(), s.bounds.W(), 0, 0 }
// // // 	// geometry[GeomPosY] = []float32{ s.bounds.H(), 0, 0, s.bounds.H() }
// // // 	// geometry[GeomPosZ] = []float32{ 0, 0, 0, 0 }
// // // 	// geometry[GeomPosW] = []float32{ 1.0, 1.0, 1.0, 1.0 }

// // // 	// geomFormat := []GeomFormat{
// // // 	// 	GeomPosX, GeomPosY, GeomPosZ,
// // // 	// }

// // // 	// positions := make([]float32, 0, numVerts * len(geomFormat))
// // // 	// for i := 0; i < numVerts; i++ {
// // // 	// 	pos := mgl32.Vec4{geometry[GeomPosX][i], geometry[GeomPosY][i], geometry[GeomPosZ][i], geometry[GeomPosW][i]}
// // // 	// 	pos = matrix.Mul4x1(pos)
// // // 	// 	for _, g := range geomFormat {
// // // 	// 		positions = append(positions, )
// // // 	// 	}
// // // 	// }

// // // 	positions := []float32{
// // // 		x + s.bounds.W()	, y + s.bounds.H(), 0.0,
// // // 		x + s.bounds.W()	, y               , 0.0,
// // // 		x							    , y               , 0.0,
// // // 		x							    , y + s.bounds.H(), 0.0,
// // // 	}
// // // 	colors := []float32{
// // // 		color.R, color.G, color.B,
// // // 		color.R, color.G, color.B,
// // // 		color.R, color.G, color.B,
// // // 		color.R, color.G, color.B,
// // // 	}
// // // 	texCoords := []float32{
// // // 		1.0, 0.0,
// // // 		1.0, 1.0,
// // // 		0.0, 1.0,
// // // 		0.0, 0.0,
// // // 	}

// // // 	inds := []uint32{
// // // 		0, 1, 3,
// // // 		1, 2, 3,
// // // 	}

// // // 	buffer.Add(positions, colors, texCoords, inds)
// // // }

// // func (s *Sprite) NumVertices() int {
// // 	return 4
// // }
// // func (s *Sprite) Indices() []uint32 {
// // 	return []uint32{
// // 		0, 1, 3,
// // 		1, 2, 3,
// // 	}
// // }

// // func (s *Sprite) PositionXYZ(n int) (float32, float32, float32) {
// // 	start := n * 3 // 3 because there are three floats per vert
// // 	return s.positions[start], s.positions[start + 1], s.positions[start + 2]
// // }
