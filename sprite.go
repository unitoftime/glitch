package glitch

type Sprite struct {
	mesh *Mesh
	bounds Rect
	texture *Texture
	material Material
	// origin Vec3 // This is used to skew the center of the sprite (which helps with sorting sprites who shouldn't be sorted based on their center points.
}

func NewSprite(texture *Texture, bounds Rect) *Sprite {
	uvBounds := R(
		bounds.Min[0] / float32(texture.width),
		bounds.Min[1] / float32(texture.height),
		bounds.Max[0] / float32(texture.width),
		bounds.Max[1] / float32(texture.height),
	)
	// Note: I tried biasing this for tilemaps, but it doesn't seem to work very well.
	// uvBounds := R(
	// 	(1./2. + bounds.Min[0]) / float32(texture.width),
	// 	(1./2. + bounds.Min[1]) / float32(texture.height),
	// 	(-1./2. + bounds.Max[0]) / float32(texture.width),
	// 	(-1./2. + bounds.Max[1]) / float32(texture.height),
	// )

	return &Sprite{
		mesh: NewSpriteMesh(bounds.W(), bounds.H(), uvBounds),
		bounds: bounds,
		texture: texture,
		material: NewSpriteMaterial(texture),
	}
}

// Translates the underlying geometry by the requested position
// func (s *Sprite) SetTranslation(pos Vec3) {
// 	// TODO - push logic to mesh?
// 	s.mesh.SetTranslation(pos)
// }

func (s *Sprite) Draw(target BatchTarget, matrix Mat4) {
	// pass.SetTexture(0, s.texture)
	target.Add(s.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, s.material)
}
func (s *Sprite) DrawColorMask(target BatchTarget, matrix Mat4, mask RGBA) {
	// pass.SetTexture(0, s.texture)
	target.Add(s.mesh, matrix, mask, s.material)
}

func (s *Sprite) RectDraw(target BatchTarget, bounds Rect) {
	s.RectDrawColorMask(target, bounds, RGBA{1, 1, 1, 1})
}
func (s *Sprite) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	// pass.SetTexture(0, s.texture)
	// pass.Add(s.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, s.material)

	matrix := Mat4Ident
	matrix.Scale(bounds.W() / s.bounds.W(), bounds.H() / s.bounds.H(), 1).Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)
	target.Add(s.mesh, matrix, mask, s.material)
}

func (s *Sprite) Bounds() Rect {
	return s.bounds
}

// // // Add another sprite on top of this sprite
// // // TODO - Include matrix transformation
// func (s *Sprite) DrawToSprite(destSprite *Sprite, mat Mat4) {
// 	if destSprite.texture != s.texture { panic("Error DrawToSprite, textures must match!") }
// 	if destSprite.material != s.material { panic("Error DrawToSprite, materials must match!") }

// 	baseSprite.bounds = baseSprite.bounds.Union(s)

// 	baseSprite.mesh.Append(s.mesh)
// }

type NinePanelSprite struct {
	sprites []*Sprite
	border Rect
	bounds Rect
	Mask RGBA // This represents the default color mask to draw with (unless one is passed in via a draw function, Example: *Mask)
	Scale float32
}

func SpriteToNinePanel(sprite *Sprite, border Rect) *NinePanelSprite {
	return NewNinePanelSprite(sprite.texture, sprite.bounds, border)
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
		Mask: White,
		Scale: 1,
	}
}

// Should 'matrix' just be scale and rotation? to scale up and down border pieces
// func (s *NinePanelSprite) Draw(pass *RenderPass, bounds Rect, matrix Mat4) {
// 	s.DrawColorMask(pass, bounds, matrix, RGBA{1.0, 1.0, 1.0, 1.0})
// }

func (s *NinePanelSprite) RectDraw(pass *RenderPass, bounds Rect) {
	s.RectDrawColorMask(pass, bounds, RGBA{1,1,1,1})
}

func (s *NinePanelSprite) RectDrawColorMask(pass *RenderPass, bounds Rect, mask RGBA) {
	// fmt.Println("here")
	// fmt.Println(bounds.W(), bounds.H())

	border := R(
		s.Scale * s.border.Min[0],
		s.Scale * s.border.Min[1],
		s.Scale * s.border.Max[0],
		s.Scale * s.border.Max[1])

	top := bounds.CutTop(border.Max[1])
	bot := bounds.CutBottom(border.Min[1])

	topLeft := top.CutLeft(border.Min[0])
	topRight := top.CutRight(border.Max[0])
	botLeft := bot.CutLeft(border.Min[0])
	botRight := bot.CutRight(border.Max[0])
	left := bounds.CutLeft(border.Min[0])
	right := bounds.CutRight(border.Max[0])

	destRects := [9]Rect{
		bounds, //center

		top, // Top
		bot, // Bot
		left, // Left
		right, // Right

		topLeft, // TL
		topRight, // TR
		botLeft, // BL
		botRight, // BR
	}


	for i := range s.sprites {
		// fmt.Println(destRects[i].W(), destRects[i].H())
		matrix := Mat4Ident
		matrix.Scale(destRects[i].W() / s.sprites[i].bounds.W(), destRects[i].H() / s.sprites[i].bounds.H(), 1).Translate(destRects[i].W()/2 + destRects[i].Min[0], destRects[i].H()/2 + destRects[i].Min[1], 0)
		pass.Add(s.sprites[i].mesh, matrix, s.Mask, s.sprites[i].material)
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
