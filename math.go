package glitch

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Vec2 [2]float32
type Vec3 [3]float32
type Vec4 [4]float32

// All Matrices are in column-major order
type Mat2 [4]float32
type Mat3 [9]float32
type Mat4 [16]float32

// This is in column major order
var Mat3Ident Mat3 = Mat3{
	1.0, 0.0, 0.0,
	0.0, 1.0, 0.0,
	0.0, 0.0, 1.0,
}

// This is in column major order
var Mat4Ident Mat4 = Mat4{
	1.0, 0.0, 0.0, 0.0,
	0.0, 1.0, 0.0, 0.0,
	0.0, 0.0, 1.0, 0.0,
	0.0, 0.0, 0.0, 1.0,
}

func (m *Mat3) Scale(x, y, z float32) *Mat3 {
	m[i3_0_0] = m[i3_0_0] * x
	m[i3_1_1] = m[i3_1_1] * y
	m[i3_2_2] = m[i3_2_2] * z
	return m
}

func (m *Mat3) Translate(x, y float32) *Mat3 {
	m[i3_2_0] = m[i3_2_0] + x
	m[i3_2_1] = m[i3_2_1] + y
	return m
}

func (m *Mat4) Scale(x, y, z float32) *Mat4 {
	m[i4_0_0] = m[i4_0_0] * x
	m[i4_1_1] = m[i4_1_1] * y
	m[i4_2_2] = m[i4_2_2] * z
	return m
}

func (m *Mat4) Translate(x, y, z float32) *Mat4 {
	m[i4_3_0] = m[i4_3_0] + x
	m[i4_3_1] = m[i4_3_1] + y
	m[i4_3_2] = m[i4_3_2] + z
	return m
}

// Matrix Indices
const (
	// 4x4 - x_y
	i4_0_0 = 0
	i4_0_1 = 1
	i4_0_2 = 2
	i4_0_3 = 3
	i4_1_0 = 4
	i4_1_1 = 5
	i4_1_2 = 6
	i4_1_3 = 7
	i4_2_0 = 8
	i4_2_1 = 9
	i4_2_2 = 10
	i4_2_3 = 11
	i4_3_0 = 12
	i4_3_1 = 13
	i4_3_2 = 14
	i4_3_3 = 15

	// 3x3 - x_y
	i3_0_0 = 0
	i3_0_1 = 1
	i3_0_2 = 2
	i3_1_0 = 3
	i3_1_1 = 4
	i3_1_2 = 5
	i3_2_0 = 6
	i3_2_1 = 7
	i3_2_2 = 8
)

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

func (m *Mat4) Apply(v Vec3) Vec3 {
	return Vec3{
		m[i4_0_0]*v[0] + m[i4_1_0]*v[1] + m[i4_2_0]*v[2] + m[i4_3_0], // w = 1.0
		m[i4_0_1]*v[0] + m[i4_1_1]*v[1] + m[i4_2_1]*v[2] + m[i4_3_1], // w = 1.0
		m[i4_0_2]*v[0] + m[i4_1_2]*v[1] + m[i4_2_2]*v[2] + m[i4_3_2], // w = 1.0
	}
}

func (m *Mat3) Apply( v Vec2) Vec2 {
	return Vec2{
		m[i3_0_0]*v[0] + m[i3_1_0]*v[1] + m[i3_2_0],
		m[i3_0_1]*v[0] + m[i3_1_1]*v[1] + m[i3_2_1],
	}
}

// Camera??
type Camera struct {
	Projection Mat4
	View Mat4

	position Vec3
}

func NewCamera() *Camera {
	return &Camera{
		Projection: Mat4Ident,
		View: Mat4Ident,
	}
}

func (c *Camera) SetOrtho2D(win *Window) {
	bounds := win.Bounds()
	c.Projection = Mat4(mgl32.Ortho2D(0, bounds.W(), 0, bounds.H()))
}

func (c *Camera) SetView2D(x, y, scaleX, scaleY float32) {
	c.View = Mat4Ident
	c.View.Scale(scaleX, scaleY, 1.0).Translate(-x, -y, 0)
}
