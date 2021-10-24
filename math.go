package glitch

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ungerik/go3d/mat4"
)

type Vec4 mgl32.Vec4

type Vec2 struct {
	X, Y float32
}

type Rect struct {
	Min, Max Vec2
}
func R(minX, minY, maxX, maxY float32) Rect {
	// TODO - guarantee min is less than max
	return Rect{
		Min: Vec2{minX, minY},
		Max: Vec2{maxX, maxY},
	}
}

func (r *Rect) W() float32 {
	return r.Max.X - r.Min.X
}

func (r *Rect) H() float32 {
	return r.Max.Y - r.Min.Y
}

// TODO - Not sure I like the type alias here. Eventually rewrite
type Mat4 = mat4.T
func Mat4Ident() *Mat4 {
	mat := mat4.From(&mat4.Ident)
	return &mat
}
