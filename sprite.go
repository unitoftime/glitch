package glitch

// type Geometry struct {
// 	format []GeomFormat
// }

// func NewGeometry() *Geometry {
	
// }

type Sprite struct {
	texture *Texture
	bounds Rect
	// Indices    []uint32
	positions  []float32
	// // Normals []float32
	// TexCoords  []float32
	geomFormat []GeomFormat
}

type GeomFormat int
const (
	GeomPosX GeomFormat = iota
	GeomPosY
	GeomPosZ
	GeomPosW
	GeomColR
	GeomColG
	GeomColB
	GeomColA
	GeomTexU
	GeomTexV
	GeomLast
)

func NewSprite(texture *Texture, bounds Rect) *Sprite {
	return &Sprite{
		texture: texture,
		bounds: bounds,

		// positions: []float32{
		// 	s.bounds.W()	, s.bounds.H(), 0.0,
		// 	s.bounds.W()	, 0           , 0.0,
		// 	0							, 0           , 0.0,
		// 	0							, s.bounds.H(), 0.0,
		// },
	}
	// return Geometry{
	// 	Indices: []uint32{
	// 		0, 1, 2,
	// 		1, 2, 3,
	// 	},
	// 	Positions: []float32{
	// 		1.0, 1.0, 0.0, // Top Right
	// 		1.0, 0.0, 0.0, // Bot Right
	// 		0.0, 0.0, 0.0, // Bot Left
	// 		0.0, 1.0, 0.0, // Top Left
	// 	},
	// 	// Notably, texture coords are flipped
	// 	TexCoords: []float32{
	// 		1.0, 0.0, // Top Right
	// 		1.0, 1.0, // Bot Right
	// 		0.0, 1.0, // Bot Left
	// 		0.0, 0.0, // Top Left
	// 	},
	// }
}

func (s *Sprite) Draw(buffer *VertexBuffer, x, y float32) {
	color := RGBA{1.0, 1.0, 1.0, 1.0}

	// verts := []float32{
	// 	x + s.bounds.W(), y + s.bounds.H(), 0.0,  color.R, color.G, color.B,  1.0, 0.0,
	// 	x + s.bounds.W(), y               , 0.0,  color.R, color.G, color.B,  1.0, 1.0,
	// 	x               , y               , 0.0,  color.R, color.G, color.B,  0.0, 1.0,
	// 	x               , y + s.bounds.H(), 0.0,  color.R, color.G, color.B,  0.0, 0.0,
	// }
	// inds := []uint32{
	// 	0, 1, 3,
	// 	1, 2, 3,
	// }
	// buffer.Add(verts, inds)

	// Works
	// positions := []float32{
	// 	x + s.bounds.W()	, y + s.bounds.H(), 0.0,
	// 	x + s.bounds.W()	, y               , 0.0,
	// 	x							    , y               , 0.0,
	// 	x							    , y + s.bounds.H(), 0.0,
	// }
	// colors := []float32{
	// 	color.R, color.G, color.B,
	// 	color.R, color.G, color.B,
	// 	color.R, color.G, color.B,
	// 	color.R, color.G, color.B,
	// }
	// texCoords := []float32{
	// 	1.0, 0.0,
	// 	1.0, 1.0,
	// 	0.0, 1.0,
	// 	0.0, 0.0,
	// }

	// inds := []uint32{
	// 	0, 1, 3,
	// 	1, 2, 3,
	// }

	// buffer.Add(positions, colors, texCoords, inds)

	// Maybe continue down this path, but instead of multiple buffers, just interleave?
	// I think since most geom will be pulled from files we can read the geom in and build out the dataset fairly easily? I don't like that this uses GeomFormat tho
	// how do you handle dynamic things, like position shifts and colors though?
	// Note - this is in memory batching, rather than just doing a different draw call.
	// geometry := []map[GeomFormat]float32{
	// 	GeomPosX: x + s.bounds.W(),
	// 	GeomPosY: y + s.bounds.H(),
	// 	GeomPosZ: 0.0,
	// }

	// numVerts := 4
	// geometry := make([][]float32, GeomLast)
	// for i := range geometry {
	// 	geometry[i] = make([]float32, numVerts)
	// }

	// geometry[GeomPosX] = []float32{ s.bounds.W(), s.bounds.W(), 0, 0 }
	// geometry[GeomPosY] = []float32{ s.bounds.H(), 0, 0, s.bounds.H() }
	// geometry[GeomPosZ] = []float32{ 0, 0, 0, 0 }
	// geometry[GeomPosW] = []float32{ 1.0, 1.0, 1.0, 1.0 }

	// geomFormat := []GeomFormat{
	// 	GeomPosX, GeomPosY, GeomPosZ,
	// }

	// positions := make([]float32, 0, numVerts * len(geomFormat))
	// for i := 0; i < numVerts; i++ {
	// 	pos := mgl32.Vec4{geometry[GeomPosX][i], geometry[GeomPosY][i], geometry[GeomPosZ][i], geometry[GeomPosW][i]}
	// 	pos = matrix.Mul4x1(pos)
	// 	for _, g := range geomFormat {
	// 		positions = append(positions, )
	// 	}
	// }

	positions := []float32{
		x + s.bounds.W()	, y + s.bounds.H(), 0.0,
		x + s.bounds.W()	, y               , 0.0,
		x							    , y               , 0.0,
		x							    , y + s.bounds.H(), 0.0,
	}
	colors := []float32{
		color.R, color.G, color.B,
		color.R, color.G, color.B,
		color.R, color.G, color.B,
		color.R, color.G, color.B,
	}
	texCoords := []float32{
		1.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
		0.0, 0.0,
	}

	inds := []uint32{
		0, 1, 3,
		1, 2, 3,
	}

	buffer.Add(positions, colors, texCoords, inds)
}

func (s *Sprite) NumVertices() int {
	return 4
}
func (s *Sprite) Indices() []uint32 {
	return []uint32{
		0, 1, 3,
		1, 2, 3,
	}
}

func (s *Sprite) PositionXYZ(n int) (float32, float32, float32) {
	start := n * 3 // 3 because there are three floats per vert
	return s.positions[start], s.positions[start + 1], s.positions[start + 2]
}
