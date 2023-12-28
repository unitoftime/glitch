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
	points []glitch.Vec3
	bounds glitch.Rect
	axes glitch.Rect
}

func NewGraph(bounds glitch.Rect) *Graph {
	g := &Graph{
		geom: glitch.NewGeomDraw(),
		mesh: glitch.NewMesh(),
		points: make([]glitch.Vec3, 0),
		bounds: bounds,
	}
	return g
}

func (g *Graph) Clear() {
	g.mesh.Clear()
}

func (g *Graph) SetBounds(bounds glitch.Rect) {
	g.bounds = bounds
}

func (g *Graph) DrawColorMask(pass *glitch.RenderPass, matrix glitch.Mat4, mask glitch.RGBA) {
	g.mesh.DrawColorMask(pass, matrix, mask)
	// pass.Add(g.mesh, matrix, mask, glitch.DefaultMaterial(), false)
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
	g.points = g.points[:0]
	for _, p := range series {
		g.points = append(g.points, glitch.Vec3{
			g.bounds.Min[0] + (p[0] - (minDomain)) * dx,
			g.bounds.Min[1] + (p[1] - (minRange)) * dy,
			0,
		})
	}

	g.geom.LineStrip(g.mesh, g.points, 1)
}

func (g *Graph) Axes() {
	g.geom.LineStrip(g.mesh,
		[]glitch.Vec3{
			glitch.Vec3{g.bounds.Min[0], g.bounds.Max[1], 0},
			glitch.Vec3{g.bounds.Min[0], g.bounds.Min[1], 0},
			glitch.Vec3{g.bounds.Max[0], g.bounds.Min[1], 0},
		},
		2,
	)

	// g.mesh.Append(g.geom.LineStrip(
	// 	[]glitch.Vec3{
	// 		glitch.Vec3{g.bounds.Min[0], g.bounds.Max[1], 0},
	// 		glitch.Vec3{g.bounds.Min[0], g.bounds.Min[1], 0},
	// 		glitch.Vec3{g.bounds.Max[0], g.bounds.Min[1], 0},
	// 	},
	// 	2,
	// ))
}

func (g *Graph) GetAxes() glitch.Rect {
	return g.axes
}
