package glitch

import (
	// "github.com/go-gl/mathgl/mgl32"
)

type Vec struct {
	X, Y float32
}

type Rect struct {
	Min, Max Vec
}
func R(minX, minY, maxX, maxY float32) Rect {
	// TODO - guarantee min is less than max
	return Rect{
		Min: Vec{minX, minY},
		Max: Vec{maxX, maxY},
	}
}

func (r *Rect) W() float32 {
	return r.Max.X - r.Min.X
}

func (r *Rect) H() float32 {
	return r.Max.Y - r.Min.Y
}

type Sprite struct {
	texture *Texture
	bounds Rect
	// Indices    []uint32
	positions  []float32
	// // Normals []float32
	// TexCoords  []float32
}

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

	verts := []float32{
		x + s.bounds.W(), y + s.bounds.H(), 0.0,  color.R, color.G, color.B,  1.0, 0.0,
		x + s.bounds.W(), y               , 0.0,  color.R, color.G, color.B,  1.0, 1.0,
		x               , y               , 0.0,  color.R, color.G, color.B,  0.0, 1.0,
		x               , y + s.bounds.H(), 0.0,  color.R, color.G, color.B,  0.0, 0.0,
	}
	inds := []uint32{
		0, 1, 3,
		1, 2, 3,
	}
	buffer.Add(verts, inds)

	// positions := []float32{
	// 	x + s.bounds.W()	, y + s.bounds.H(), 0.0,
	// 	x + s.bounds.W()	, y               , 0.0,
	// 	x							    , y               , 0.0,
	// 	x							    , y + s.bounds.H(), 0.0,
	// },
	// colors := []float32{
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
