package graph

import (
	// "fmt"
	"math"
	"github.com/unitoftime/glitch"
)

// 2D Graph
type Graph struct {
	geom *glitch.GeomDraw
	mesh *glitch.Mesh
	bounds glitch.Rect
	axes glitch.Rect
}

func NewGraph(bounds glitch.Rect) *Graph {
	g := &Graph{
		geom: glitch.NewGeomDraw(),
		mesh: glitch.NewMesh(),
		bounds: bounds,
	}
	return g
}

func (g *Graph) Clear() {
	g.mesh.Clear()
}

func (g *Graph) DrawColorMask(pass *glitch.RenderPass, matrix glitch.Mat4, mask glitch.RGBA) {
	pass.Add(g.mesh, matrix, mask, glitch.DefaultMaterial())
}

// TODO - Assumes sorted?
func (g *Graph) Line(series []glitch.Vec2) {
	minDomain := math.MaxFloat64
	maxDomain := -math.MaxFloat64
	minRange := math.MaxFloat64
	maxRange := -math.MaxFloat64
	for _, p := range series {
		minDomain = math.Min(minDomain, float64(p[0]))
		maxDomain = math.Max(maxDomain, float64(p[0]))

		minRange = math.Min(minRange, float64(p[1]))
		maxRange = math.Max(maxRange, float64(p[1]))
	}

	g.axes = glitch.R(minDomain, minRange, maxDomain, maxRange)

	dx := g.bounds.W() / (maxDomain - minDomain)
	dy := g.bounds.H() / (maxRange - minRange)
	// fmt.Println(dx, dy, rect.H(), maxDomain, minDomain, minRange, maxRange)
	points := make([]glitch.Vec3, 0)
	for _, p := range series {
		points = append(points, glitch.Vec3{
			g.bounds.Min[0] + (p[0] - (minDomain)) * dx,
			g.bounds.Min[1] + (p[1] - (minRange)) * dy,
			0,
		})
	}

	g.mesh.Append(g.geom.LineStrip(points, 1))
}

func (g *Graph) Axes() {
	g.mesh.Append(g.geom.LineStrip(
		[]glitch.Vec3{
			glitch.Vec3{g.bounds.Min[0], g.bounds.Max[1], 0},
			glitch.Vec3{g.bounds.Min[0], g.bounds.Min[1], 0},
			glitch.Vec3{g.bounds.Max[0], g.bounds.Min[1], 0},
		},
		2,
	))
}

func (g *Graph) GetAxes() glitch.Rect {
	return g.axes
}
