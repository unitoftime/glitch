package glitch

import (
	"math"
)

// TODO - Rename
// type Geometry struct {
// 	mesh *Mesh
// 	bound Rect
// 	// texture *Texture // Possible?
// 	material Material
// }

// func NewGeometry() {
// 	// uvBounds := R(
// 	// 	bounds.Min[0] / float32(texture.width),
// 	// 	bounds.Min[1] / float32(texture.height),
// 	// 	bounds.Max[0] / float32(texture.width),
// 	// 	bounds.Max[1] / float32(texture.height),
// 	// )

// 	return &Geometry{
// 		mesh: NewMesh(),
// 		bounds: R(0,0,0,0),
// 		// texture: texture,
// 		material: NewSpriteMaterial(nil),
// 	}
// }

// func (g *Geometry) Rectangle(r Rect) {
// 	g.mesh.Append(NewQuadMesh(r, R(0,0,1,1))) // Fill
// 	// g.mesh.AddTriangle(
// }

// type Geometry struct {
// 	mesh *Mesh
// }

// func NewGeometry() *Geometry {
// 	return &Geometry{
// 		mesh: NewMesh(),
// 	}
// }


// TODO - Can I just pass meshes into these functions to have them get drawn to?
type GeomDraw struct {
	color RGBA
}

func NewGeomDraw() *GeomDraw {
	return &GeomDraw{
		color: RGBA{1, 1, 1, 1},
	}
}

func (g *GeomDraw) SetColor(color RGBA) {
	g.color = color
}

// if width == 0, then fill the rect
func (g *GeomDraw) Rectangle(rect Rect, width float32) *Mesh {
	if width <= 0 {
		panic("TODO - Fill Rect")
	}

	points := []Vec3{
		Vec3{rect.Min[0], rect.Min[1], 0},
		Vec3{rect.Min[0], rect.Max[1], 0},
		Vec3{rect.Max[0], rect.Max[1], 0},
		Vec3{rect.Max[0], rect.Min[1], 0},
		Vec3{rect.Min[0], rect.Min[1], 0},
	}

	return g.LineStrip(points, width)
}

// TODO - Should I pass in number of divisions?
func (g *GeomDraw) Circle(center Vec3, radius float32, width float32) *Mesh {
	if width <= 0 {
		panic("TODO - Fill Circle")
	}

	maxDivisions := 100
	points := make([]Vec3, maxDivisions, maxDivisions)
	radians := 0.0
	for i := range points {
		points[i] = center.Add(Vec3{
			radius * float32(math.Cos(radians)),
			radius * float32(math.Sin(radians)),
			0,
		})
		radians += 2 * math.Pi / float64(maxDivisions)
	}
	// Append last point
	points = append(points, center.Add(Vec3{radius, 0, 0}))

	return g.LineStrip(points, width)
}

func (g *GeomDraw) LineStrip(points []Vec3, width float32) *Mesh {
	m := NewMesh()
	for i := 0; i < len(points)-1; i++ {
		m.Append(g.Line(points[i], points[i+1], width))
	}
	return m
}

func (g *GeomDraw) Line(a, b Vec3, width float32) *Mesh {
	line := a.Sub(b).Unit()

	// Shift the point over by width
	a = a.Add(line.Scaled(width/2, width/2, width/2))
	b = b.Sub(line.Scaled(width/2, width/2, width/2))

	// A line along the width of the line
	// (x, y) rotated 90 degrees around (0, 0) is (-y, x)
	// TODO - 3D version of this 90 degree rotation?
	wLineUp := Vec3{-line[1], line[0], line[2]}.Scaled(width/2, width/2, width/2)
	wLineDown := wLineUp.Scaled(-1, -1, -1)
	a1 := a.Add(wLineUp)
	a2 := a.Add(wLineDown)
	b1 := b.Add(wLineUp)
	b2 := b.Add(wLineDown)

	positions := []Vec3{
		b1,
		b2,
		a2,
		a1,
	}
	colors := []Vec4{
		Vec4{g.color.R, g.color.G, g.color.B, g.color.A},
		Vec4{g.color.R, g.color.G, g.color.B, g.color.A},
		Vec4{g.color.R, g.color.G, g.color.B, g.color.A},
		Vec4{g.color.R, g.color.G, g.color.B, g.color.A},
	}

	// TODO - Finalize what these should be
	texCoords := []Vec2{
		Vec2{1, 0},
		Vec2{1, 1},
		Vec2{0, 1},
		Vec2{0, 0},
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
	}
}
