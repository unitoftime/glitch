package glitch

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/vec3"
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
var Mat4Ident Mat4 = mat4.Ident
// func Mat4Ident() Mat4 {
// 	return mat4.Ident
// }

func MatMul(m *Mat4, v vec3.T) vec3.T {
	return vec3.T{
		m[0][0]*v[0] + m[1][0]*v[1] + m[2][0]*v[2] + m[3][0], // w = 1.0
		m[0][1]*v[0] + m[1][1]*v[1] + m[2][1]*v[2] + m[3][1], // w = 1.0
		m[0][2]*v[0] + m[1][2]*v[1] + m[2][2]*v[2] + m[3][2], // w = 1.0
	}
}
