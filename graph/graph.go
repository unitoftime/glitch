package graph

import (
	// "fmt"
	"math"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
)

// 2D Graph
type Graph struct {
	geom   *glitch.GeomDraw
	mesh   *glitch.Mesh
	points []glitch.Vec3
	bounds glitch.Rect
	axes   glitch.Rect
}

func NewGraph(bounds glitch.Rect) *Graph {
	g := &Graph{
		geom:   glitch.NewGeomDraw(),
		mesh:   glitch.NewMesh(),
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

func (g *Graph) DrawColorMask(target glitch.BatchTarget, matrix glitch.Mat4, mask glitch.RGBA) {
	g.mesh.DrawColorMask(target, matrix, mask)
	// pass.Add(g.mesh, matrix, mask, glitch.DefaultMaterial(), false)
}

func (g *Graph) RectDrawColorMask(target glitch.BatchTarget, rect glitch.Rect, mask glitch.RGBA) {
	matrix := g.bounds.RectDraw(rect)
	g.mesh.DrawColorMask(target, matrix, mask)
}

func (g *Graph) RectDraw(target glitch.BatchTarget, rect glitch.Rect) {
	g.RectDrawColorMask(target, rect, glitch.White)
}

// TODO - Assumes sorted?
func (g *Graph) Line(series []glitch.Vec2) {
	minDomain := math.MaxFloat64
	maxDomain := -math.MaxFloat64
	minRange := math.MaxFloat64
	maxRange := -math.MaxFloat64
	for _, p := range series {
		minDomain = math.Min(minDomain, float64(p.X))
		maxDomain = math.Max(maxDomain, float64(p.X))

		minRange = math.Min(minRange, float64(p.Y))
		maxRange = math.Max(maxRange, float64(p.Y))
	}

	g.axes = glm.R(minDomain, minRange, maxDomain, maxRange)

	dx := g.bounds.W() / (maxDomain - minDomain)
	dy := g.bounds.H() / (maxRange - minRange)
	// fmt.Println(dx, dy, rect.H(), maxDomain, minDomain, minRange, maxRange)
	g.points = g.points[:0]
	for _, p := range series {
		g.points = append(g.points, glitch.Vec3{
			g.bounds.Min.X + (p.X-(minDomain))*dx,
			g.bounds.Min.Y + (p.Y-(minRange))*dy,
			0,
		})
	}

	g.geom.LineStrip(g.mesh, g.points, 1)
}

func (g *Graph) Axes() {
	g.geom.LineStrip(g.mesh,
		[]glitch.Vec3{
			glitch.Vec3{g.bounds.Min.X, g.bounds.Max.Y, 0},
			glitch.Vec3{g.bounds.Min.X, g.bounds.Min.Y, 0},
			glitch.Vec3{g.bounds.Max.X, g.bounds.Min.Y, 0},
		},
		2,
	)

	// g.mesh.Append(g.geom.LineStrip(
	// 	[]glitch.Vec3{
	// 		glitch.Vec3{g.bounds.Min.X, g.bounds.Max.Y, 0},
	// 		glitch.Vec3{g.bounds.Min.X, g.bounds.Min.Y, 0},
	// 		glitch.Vec3{g.bounds.Max.X, g.bounds.Min.Y, 0},
	// 	},
	// 	2,
	// ))
}

func (g *Graph) GetAxes() glitch.Rect {
	return g.axes
}
