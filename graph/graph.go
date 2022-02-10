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
	rect glitch.Rect
}

func NewGraph(rect glitch.Rect) *Graph {
	g := &Graph{
		geom: glitch.NewGeomDraw(),
		mesh: glitch.NewMesh(),
		rect: rect,
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

	dx := g.rect.W() / float32(maxDomain - minDomain)
	dy := g.rect.H() / float32(maxRange - minRange)
	// fmt.Println(dx, dy, rect.H(), maxDomain, minDomain, minRange, maxRange)
	points := make([]glitch.Vec3, 0)
	for _, p := range series {
		points = append(points, glitch.Vec3{
			g.rect.Min[0] + float32(p[0] - float32(minDomain)) * dx,
			g.rect.Min[1] + float32(p[1] - float32(minRange)) * dy,
			0,
		})
	}

	g.mesh.Append(g.geom.LineStrip(points, 1))
}
