package glitch

import (
	// "fmt"
	"math"
	"github.com/go-gl/mathgl/mgl32"
)

type Vec2 [2]float32
type Vec3 [3]float32
type Vec4 [4]float32

func (v Vec2) Add(u Vec2) Vec2 {
	return Vec2{v[0] + u[0], v[1] + u[1]}
}

func (v Vec2) Sub(u Vec2) Vec2 {
	return Vec2{v[0] - u[0], v[1] - u[1]}
}

func (v Vec2) Len() float32 {
	return float32(math.Hypot(float64(v[0]), float64(v[1])))
}

func (v Vec2) Scaled(s float32) Vec2 {
	return Vec2{s * v[0], s * v[1]}
}

func (v Vec3) Scale(x, y, z float32) Vec3 {
	v[0] *= x
	v[1] *= y
	v[2] *= z

	return v
}

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

// Note: Does not modify in place
func (m *Mat4) Mul(n *Mat4) Mat4 {
	// This is in column major order
	return Mat4{
		// Column 0
		m[i4_0_0] * n[i4_0_0] + m[i4_1_0] * n[i4_0_1] + m[i4_2_0] * n[i4_0_2] + m[i4_3_0] * n[i4_0_3],
		m[i4_0_1] * n[i4_0_0] + m[i4_1_1] * n[i4_0_1] + m[i4_2_1] * n[i4_0_2] + m[i4_3_1] * n[i4_0_3],
		m[i4_0_2] * n[i4_0_0] + m[i4_1_2] * n[i4_0_1] + m[i4_2_2] * n[i4_0_2] + m[i4_3_2] * n[i4_0_3],
		m[i4_0_3] * n[i4_0_0] + m[i4_1_3] * n[i4_0_1] + m[i4_2_3] * n[i4_0_2] + m[i4_3_3] * n[i4_0_3],

		// Column 1
		m[i4_0_0] * n[i4_1_0] + m[i4_1_0] * n[i4_1_1] + m[i4_2_0] * n[i4_1_2] + m[i4_3_0] * n[i4_1_3],
		m[i4_0_1] * n[i4_1_0] + m[i4_1_1] * n[i4_1_1] + m[i4_2_1] * n[i4_1_2] + m[i4_3_1] * n[i4_1_3],
		m[i4_0_2] * n[i4_1_0] + m[i4_1_2] * n[i4_1_1] + m[i4_2_2] * n[i4_1_2] + m[i4_3_2] * n[i4_1_3],
		m[i4_0_3] * n[i4_1_0] + m[i4_1_3] * n[i4_1_1] + m[i4_2_3] * n[i4_1_2] + m[i4_3_3] * n[i4_1_3],

		// Column 2
		m[i4_0_0] * n[i4_2_0] + m[i4_1_0] * n[i4_2_1] + m[i4_2_0] * n[i4_2_2] + m[i4_3_0] * n[i4_2_3],
		m[i4_0_1] * n[i4_2_0] + m[i4_1_1] * n[i4_2_1] + m[i4_2_1] * n[i4_2_2] + m[i4_3_1] * n[i4_2_3],
		m[i4_0_2] * n[i4_2_0] + m[i4_1_2] * n[i4_2_1] + m[i4_2_2] * n[i4_2_2] + m[i4_3_2] * n[i4_2_3],
		m[i4_0_3] * n[i4_2_0] + m[i4_1_3] * n[i4_2_1] + m[i4_2_3] * n[i4_2_2] + m[i4_3_3] * n[i4_2_3],

		// Column 3
		m[i4_0_0] * n[i4_3_0] + m[i4_1_0] * n[i4_3_1] + m[i4_2_0] * n[i4_3_2] + m[i4_3_0] * n[i4_3_3],
		m[i4_0_1] * n[i4_3_0] + m[i4_1_1] * n[i4_3_1] + m[i4_2_1] * n[i4_3_2] + m[i4_3_1] * n[i4_3_3],
		m[i4_0_2] * n[i4_3_0] + m[i4_1_2] * n[i4_3_1] + m[i4_2_2] * n[i4_3_2] + m[i4_3_2] * n[i4_3_3],
		m[i4_0_3] * n[i4_3_0] + m[i4_1_3] * n[i4_3_1] + m[i4_2_3] * n[i4_3_2] + m[i4_3_3] * n[i4_3_3],
	}
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

func (r Rect) Center() Vec2 {
	return Vec2{r.Min[0] + (r.W()/2), r.Min[1] + (r.H()/2)}
}

func (r Rect) Norm() Rect {
	x1, x2 := minMax(r.Min[0], r.Max[0])
	y1, y2 := minMax(r.Min[1], r.Max[1])
	return R(x1, y1, x2, y2)
}

func (r Rect) Contains(x, y float32) bool {
	return x > r.Min[0] && x < r.Max[0] && y > r.Min[1] && y < r.Max[1]
}

func (r *Rect) CutLeft(amount float32) Rect {
	cutRect := *r
	cutRect.Max[0] = cutRect.Min[0] + amount
	r.Min[0] += amount
	return cutRect
}

func (r *Rect) CutRight(amount float32) Rect {
	cutRect := *r
	cutRect.Min[0] = cutRect.Max[0] - amount
	r.Max[0] -= amount
	return cutRect
}

func (r *Rect) CutBottom(amount float32) Rect {
	cutRect := *r
	cutRect.Max[1] = cutRect.Min[1] + amount
	r.Min[1] += amount
	return cutRect
}

func (r *Rect) CutTop(amount float32) Rect {
	cutRect := *r
	cutRect.Min[1] = cutRect.Max[1] - amount
	r.Max[1] -= amount
	return cutRect
}

// Returns a centered horizontal sliver with height set by amount
func (r Rect) SliceHorizontal(amount float32) Rect {
	r.CutTop((r.H() - amount) / 2)
	return r.CutTop(amount)
}

// Returns a centered vertical sliver with width set by amount
func (r Rect) SliceVertical(amount float32) Rect {
	r.CutRight((r.W() - amount) / 2)
	return r.CutRight(amount)
}

// Takes r2 and places it in r based on the alignment
func (r Rect) Anchor(r2 Rect, anchor Vec2) Rect {
	// Anchor point is the position in r that we are anchoring to
	anchorPoint := Vec2{r.Min[0] + (anchor[0] * r.W()) , r.Min[1] + (anchor[1] * r.H())}
	pivotPoint := Vec2{r2.Min[0] + (anchor[0] * r2.W()) , r2.Min[1] + (anchor[1] * r2.H())}

	// fmt.Println("Anchor:", anchorPoint)
	// fmt.Println("Pivot:", pivotPoint)

	a := Vec2{anchorPoint[0] - pivotPoint[0], anchorPoint[1] - pivotPoint[1]}
	return R(a[0], a[1], a[0] + r2.W(), a[1] + r2.H()).Norm()
}

func lerp(a, b float32, t float32) float32 {
	m := b - a // Slope = Rise over run | Note: Run = (1 - 0)
	y := (m * t) + a
	return y
}

// returns the min, max of the two numbers
func minMax(a, b float32) (float32, float32) {
	if a > b {
		return b, a
	}
	return a, b
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


type CameraOrtho struct {
	Projection Mat4
	View Mat4

	position Vec3
}

func NewCameraOrtho() *CameraOrtho {
	return &CameraOrtho{
		Projection: Mat4Ident,
		View: Mat4Ident,
	}
}

func (c *CameraOrtho) SetOrtho2D(win *Window) {
	bounds := win.Bounds()
	// c.Projection = Mat4(mgl32.Ortho2D(0, bounds.W(), 0, bounds.H()))
	c.Projection = Mat4(mgl32.Ortho(0, bounds.W(), 0, bounds.H(), -1, 1))
}

func (c *CameraOrtho) SetView2D(x, y, scaleX, scaleY float32) {
	c.View = Mat4Ident
	c.View.Scale(scaleX, scaleY, 1.0).Translate(-x, -y, 0)
}

type Camera struct {
	Projection Mat4
	View Mat4

	position Vec3
	target Vec3
}

func NewCamera() *Camera {
	return &Camera{
		Projection: Mat4Ident,
		View: Mat4Ident,
		position: Vec3{0, 0, 0},
		target: Vec3{0, 0, 0},
	}
}

func (c *Camera) SetPerspective(win *Window) {
	bounds := win.Bounds()
	aspect := bounds.W() / bounds.H()
	// c.Projection = Mat4(mgl32.Ortho2D(0, bounds.W(), 0, bounds.H()))
	// c.Projection = Mat4(mgl32.Ortho(0, bounds.W(), 0, bounds.H(), -1080, 1080))
	c.Projection = Mat4(mgl32.Perspective(math.Pi/4, aspect, 0.1, 1000))
}

func (c *Camera) SetViewLookAt(win *Window) {
	c.View = Mat4(mgl32.LookAt(
		c.position[0], c.position[1], c.position[2],
		200, 200, 0, // target
		0, 0, 1,
	))
}
