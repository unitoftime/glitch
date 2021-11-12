package glitch

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ungerik/go3d/mat4"
)

type Vec2 = [2]float32
type Vec3 = [3]float32

type Mat2 = [2][2]float32
type Mat3 = [3][3]float32
// type Mat4 = [4]float32
// var Mat4Ident Mat4 = mgl32.Ident4()

// Notably this definition looks transposed, I'm building columns here
var Mat3Ident Mat3 = Mat3{
	Vec3{1.0, 0.0, 0.0},
	Vec3{0.0, 1.0, 0.0},
	Vec3{0.0, 0.0, 1.0},
}

func ScaleMat3(m *Mat3, x, y, z float32) {
	m[0][0] = m[0][0] * x
	m[1][1] = m[1][1] * y
	m[2][2] = m[2][2] * z
}

func TranslateMat3(m *Mat3, x, y float32) {
	m[2][0] = m[2][0] + x
	m[2][1] = m[2][1] + y
}

type Vec4 mgl32.Vec4

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
	return r.Max[0] - r.Min[0]
}

func (r *Rect) H() float32 {
	return r.Max[1] - r.Min[1]
}

// TODO - Not sure I like the type alias here. Eventually rewrite
type Mat4 = mat4.T
var Mat4Ident Mat4 = mat4.Ident
// func Mat4Ident() Mat4 {
// 	return mat4.Ident
// }

func MatApply4x3(m *Mat4, v Vec3) Vec3 {
	return Vec3{
		m[0][0]*v[0] + m[1][0]*v[1] + m[2][0]*v[2] + m[3][0], // w = 1.0
		m[0][1]*v[0] + m[1][1]*v[1] + m[2][1]*v[2] + m[3][1], // w = 1.0
		m[0][2]*v[0] + m[1][2]*v[1] + m[2][2]*v[2] + m[3][2], // w = 1.0
	}
}

func MatApply3x2(m *Mat3, v Vec2) Vec2 {
	return Vec2{
		m[0][0]*v[0] + m[1][0]*v[1] + m[2][0],
		m[0][1]*v[0] + m[1][1]*v[1] + m[2][1],
	}
}


// Camera??
type Camera struct {
	Projection mgl32.Mat4
	View mgl32.Mat4

	position mgl32.Vec3
}

func NewCamera() *Camera {
	return &Camera{
		Projection: mgl32.Ident4(),
		View: mgl32.Ident4(),
	}
}

func (c *Camera) SetOrtho2D(win *Window) {
	bounds := win.Bounds()
	c.Projection = mgl32.Ortho2D(0, bounds.W(), 0, bounds.H())
}

func (c *Camera) SetView2D(x, y, scaleX, scaleY float32) {
	// c.View = mgl32.Translate3D(-x, -y, 0).Mul4(mgl32.Scale3D(scale, scale, 1.0))
	c.View = mgl32.Scale3D(scaleX, scaleY, 1.0).Mul4(mgl32.Translate3D(-x, -y, 0))
}
