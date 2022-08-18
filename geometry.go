package glitch

import (
	"math"
)

type GeomDraw struct {
	color RGBA
	Divisions int
}

func NewGeomDraw() *GeomDraw {
	return &GeomDraw{
		color: RGBA{1, 1, 1, 1},
		Divisions: 100,
	}
}

func (g *GeomDraw) SetColor(color RGBA) {
	g.color = color
}

func (g *GeomDraw) FillRect(rect Rect) *Mesh {
	positions := []Vec3{
		Vec3{rect.Min[0], rect.Max[1], 0},
		Vec3{rect.Min[0], rect.Min[1], 0},
		Vec3{rect.Max[0], rect.Min[1], 0},
		Vec3{rect.Max[0], rect.Max[1], 0},
	}
	colors := []Vec4{
		Vec4{g.color.R, g.color.G, g.color.B, g.color.A},
		Vec4{g.color.R, g.color.G, g.color.B, g.color.A},
		Vec4{g.color.R, g.color.G, g.color.B, g.color.A},
		Vec4{g.color.R, g.color.G, g.color.B, g.color.A},
	}

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

// if width == 0, then fill the rect
func (g *GeomDraw) Rectangle(rect Rect, width float32) *Mesh {
	if width <= 0 {
		return g.FillRect(rect)
	}

	t := rect.CutTop(width)
	b := rect.CutBottom(width)
	l := rect.CutLeft(width)
	r := rect.CutRight(width)

	m := NewMesh()
	m.Append(g.FillRect(t))
	m.Append(g.FillRect(b))
	m.Append(g.FillRect(l))
	m.Append(g.FillRect(r))

	return m
}

func (g *GeomDraw) Circle(center Vec3, radius float32, width float32) *Mesh {
	return g.Ellipse(center, Vec2{radius, radius}, 0, width)
}

func (g *GeomDraw) Ellipse(center Vec3, size Vec2, rotation float32, width float32) *Mesh {
	if width <= 0 {
		panic("TODO - Fill Ellipse")
	}

	alpha := float64(rotation)

	a := math.Max(float64(size[0]), float64(size[1])) // SemiMajorAxis
	b := math.Min(float64(size[0]), float64(size[1])) // SemiMinorAxis
	// TODO - Rotate pi/2 if width < height?
	e := math.Sqrt(1 - (b*b)/(a*a)) // Eccintricity

	points := make([]Vec3, g.Divisions, g.Divisions)
	radians := 0.0
	for i := range points {
		eCos := (e * math.Cos(radians))
		r := b / (math.Sqrt(1 - (eCos * eCos)))

		points[i] = center.Add(Vec3{
		float32(r * math.Cos(radians - alpha)),
		float32(r * math.Sin(radians - alpha)),
		0,
		})
		radians += 2 * math.Pi / float64(g.Divisions)
	}
	// Append last point
	{
		eCos := (e * math.Cos(radians))
		r := b / (math.Sqrt(1 - (eCos * eCos)))
		// r := a * (1 - e * e) / (1 + (e * math.Cos(radians - alpha)))
		// r := l / (1 + (e * math.Cos(radians - alpha)))
		lastPoint := center.Add(Vec3{
		float32(r * math.Cos(radians - alpha)),
		float32(r * math.Sin(radians - alpha)),
		0,
		})
		points = append(points, lastPoint)
	}

	// // Circle only
	// points := make([]Vec3, g.Divisions, g.Divisions)
	// radians := 0.0
	// for i := range points {
	// 	points[i] = center.Add(Vec3{
	// 		radius * float32(math.Cos(radians)),
	// 		(22.0/32.0) * radius * float32(math.Sin(radians)),
	// 		0,
	// 	})
	// 	radians += 2 * math.Pi / float64(g.Divisions)
	// }
	// // Append last point
	// points = append(points, center.Add(Vec3{radius, 0, 0}))

	return g.LineStrip(points, width)
}

func (g *GeomDraw) LineStrip(points []Vec3, width float32) *Mesh {
	m := NewMesh()
	for i := 0; i < len(points)-1; i++ {
		m.Append(g.Line(points[i], points[i+1], width))
	}
	return m
}

// TODO different line endings
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

