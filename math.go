package glitch

import (
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

func (v Vec2) Vec3() Vec3 {
	return Vec3{v[0], v[1], 0}
}


func (v Vec3) Add(u Vec3) Vec3 {
	return Vec3{v[0] + u[0], v[1] + u[1], v[2] + u[2]}
}

func (v Vec3) Sub(u Vec3) Vec3 {
	return Vec3{v[0] - u[0], v[1] - u[1], v[2] - u[2]}
}

func (v Vec3) Len() float32 {
	// return float32(math.Hypot(float64(v[0]), float64(v[1])))
	a := float64(v[0])
	b := float64(v[1])
	c := float64(v[2])
	return float32(math.Sqrt((a * a) + (b * b) + (c * c)))
}

func (v Vec3) Vec2() Vec2 {
	return Vec2{v[0], v[1]}
}

func (v Vec3) Unit() Vec3 {
	len := v.Len()
	return Vec3{v[0]/len, v[1]/len, v[2]/len}
}

func (v Vec3) Scaled(x, y, z float32) Vec3 {
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

// TODO - This is wrong, need to rewrite from Mat4
// func (m *Mat3) Scale(x, y, z float32) *Mat3 {
// 	m[i3_0_0] = m[i3_0_0] * x
// 	m[i3_1_1] = m[i3_1_1] * y
// 	m[i3_2_2] = m[i3_2_2] * z
// 	return m
// }

func (m *Mat3) Translate(x, y float32) *Mat3 {
	m[i3_2_0] = m[i3_2_0] + x
	m[i3_2_1] = m[i3_2_1] + y
	return m
}

// Note: Scales around 0,0
func (m *Mat4) Scale(x, y, z float32) *Mat4 {
	m[i4_0_0] = m[i4_0_0] * x
	m[i4_1_0] = m[i4_1_0] * x
	m[i4_2_0] = m[i4_2_0] * x
	m[i4_3_0] = m[i4_3_0] * x

	m[i4_0_1] = m[i4_0_1] * y
	m[i4_1_1] = m[i4_1_1] * y
	m[i4_2_1] = m[i4_2_1] * y
	m[i4_3_1] = m[i4_3_1] * y

	m[i4_0_2] = m[i4_0_2] * z
	m[i4_1_2] = m[i4_1_2] * z
	m[i4_2_2] = m[i4_2_2] * z
	m[i4_3_2] = m[i4_3_2] * z

	return m
}

func (m *Mat4) Translate(x, y, z float32) *Mat4 {
	m[i4_3_0] = m[i4_3_0] + x
	m[i4_3_1] = m[i4_3_1] + y
	m[i4_3_2] = m[i4_3_2] + z
	return m
}

func (m *Mat4) GetTranslation() (float32, float32, float32) {
	return m[i4_3_0], m[i4_3_1], m[i4_3_2]
}

// TODO this doesn't modify the underlying matrix: https://github.com/go-gl/mathgl/blob/v1.0.0/mgl32/transform.go#L159
func (m *Mat4) Rotate(angle float32, axis Vec3) *Mat4 {
	// quat := mgl32.Mat4ToQuat(mgl32.Mat4(*m))
	// return &retMat
	rotation := Mat4(mgl32.HomogRotate3D(angle, mgl32.Vec3(axis)))
	// retMat := Mat4(mgl32.Mat4(*m).)
	// return &retMat
	return m.Mul(&rotation)
}

// TODO should this modify in place?
// Note: Does not modify in place
func (m *Mat4) Mul(n *Mat4) *Mat4 {
	// This is in column major order
	return &Mat4{
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

type Box struct {
	Min, Max Vec3
}
func (b Box) Rect() Rect {
	return Rect{
		Min: Vec2{b.Min[0], b.Min[1]},
		Max: Vec2{b.Max[0], b.Max[1]},
	}
}


func (a Box) Union(b Box) Box {
	x1, _ := minMax(a.Min[0], b.Min[0])
	_, x2 := minMax(a.Max[0], b.Max[0])
	y1, _ := minMax(a.Min[1], b.Min[1])
	_, y2 := minMax(a.Max[1], b.Max[1])
	z1, _ := minMax(a.Min[2], b.Min[2])
	_, z2 := minMax(a.Max[2], b.Max[2])
	return Box{
		Min: Vec3{x1, y1, z1},
		Max: Vec3{x2, y2, z2},
	}
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

// Returns a box that holds this rect. The Z axis is 0
func (r Rect) ToBox() Box {
	return Box{
		Min: Vec3{r.Min[0], r.Min[1], 0},
		Max: Vec3{r.Max[0], r.Max[1], 0},
	}
}

func (r Rect) W() float32 {
	return r.Max[0] - r.Min[0]
}

func (r Rect) H() float32 {
	return r.Max[1] - r.Min[1]
}

func (r Rect) Center() Vec2 {
	return Vec2{r.Min[0] + (r.W()/2), r.Min[1] + (r.H()/2)}
}

// Returns the smallest rect which contains both input rects
func (r Rect) Union(s Rect) Rect {
	r = r.Norm()
	s = s.Norm()
	x1, _ := minMax(r.Min[0], s.Min[0])
	_, x2 := minMax(r.Max[0], s.Max[0])
	y1, _ := minMax(r.Min[1], s.Min[1])
	_, y2 := minMax(r.Max[1], s.Max[1])
	return R(x1, y1, x2, y2)
}

func (r Rect) Moved(v Vec2) Rect {
	return Rect{
		Min: r.Min.Add(v),
		Max: r.Max.Add(v),
	}
}

// Scales rect r uniformly to fit inside rect r2
// TODO This only scales around {0, 0}
func (r Rect) ScaledToFit(r2 Rect) Rect {
	scaleX := r2.W() / r.W()
	scaleY := r2.H() / r.H()

	min, _ := minMax(scaleX, scaleY)
	return r.Scaled(min)
}

func (r Rect) Scaled(scale float32) Rect {
	return Rect{
		Min: r.Min.Scaled(scale),
		Max: r.Max.Scaled(scale),
	}
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

// Adds padding to a rectangle (pads inward if padding is negative)
func (r Rect) Pad(pad Rect) Rect {
	return R(r.Min[0] - pad.Min[0], r.Min[1] - pad.Min[1], r.Max[0] + pad.Max[0], r.Max[1] + pad.Max[1])
}

// Removes padding from a rectangle (pads outward if padding is negative). Essentially calls pad but with negative values
func (r Rect) Unpad(pad Rect) Rect {
	return r.Pad(pad.Scaled(-1))
}

// Takes r2 and places it in r based on the alignment
// TODO - rename to InnerAnchor?
func (r Rect) Anchor(r2 Rect, anchor Vec2) Rect {
	// Anchor point is the position in r that we are anchoring to
	anchorPoint := Vec2{r.Min[0] + (anchor[0] * r.W()) , r.Min[1] + (anchor[1] * r.H())}
	pivotPoint := Vec2{r2.Min[0] + (anchor[0] * r2.W()) , r2.Min[1] + (anchor[1] * r2.H())}

	// fmt.Println("Anchor:", anchorPoint)
	// fmt.Println("Pivot:", pivotPoint)

	a := Vec2{anchorPoint[0] - pivotPoint[0], anchorPoint[1] - pivotPoint[1]}
	return R(a[0], a[1], a[0] + r2.W(), a[1] + r2.H()).Norm()
}

// Anchors r2 to r1 based on two anchors, one for r and one for r2
// TODO - rename to Anchor?
func (r Rect) FullAnchor(r2 Rect, rAnchor, r2Anchor Vec2) Rect {
	anchorPoint := Vec2{r.Min[0] + (rAnchor[0] * r.W()), r.Min[1] + (rAnchor[1] * r.H())}
	pivotPoint := Vec2{r2.Min[0] + (r2Anchor[0] * r2.W()) , r2.Min[1] + (r2Anchor[1] * r2.H())}

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

// Note: Returns a new Mat4
func (m *Mat4) Inv() *Mat4 {
	retMat := Mat4(mgl32.Mat4(*m).Inv())
	return &retMat
}

func (m *Mat4) Transpose() *Mat4 {
	retMat := Mat4(mgl32.Mat4(*m).Transpose())
	return &retMat
}

// TODO - I feel like camera should be a higher-up abstraction and not held here
type CameraOrtho struct {
	Projection Mat4
	View, ViewSnapped Mat4
	bounds Rect
}

func NewCameraOrtho() *CameraOrtho {
	return &CameraOrtho{
		Projection: Mat4Ident,
		View: Mat4Ident,
		ViewSnapped: Mat4Ident,
		bounds: R(0,0,1,1),
	}
}

func (c *CameraOrtho) SetOrtho2D(bounds Rect) {
	c.bounds = bounds
	c.Projection = Mat4(mgl32.Ortho2D(0, c.bounds.W(), 0, c.bounds.H()))
	// c.Projection = Mat4(mgl32.Ortho(0, c.bounds.W(), 0, c.bounds.H(), -1, 1))
}

func (c *CameraOrtho) SetView2D(x, y, scaleX, scaleY float32) {
	c.View = Mat4Ident
	cameraCenter := c.bounds.Center()
	c.View.
		// Translate by x, y of the camera
		Translate(-x, -y, 0).
		// Scale around the center of the camera
		Translate(-cameraCenter[0], -cameraCenter[1], 0).
		Scale(scaleX, scaleY, 1.0).
		Translate(cameraCenter[0], cameraCenter[1], 0)

	// TODO - this is literally only for pixel art
	c.ViewSnapped = Mat4Ident
	centerX := float32(math.Round(float64(cameraCenter[0])))
	centerY := float32(math.Round(float64(cameraCenter[1])))
	pX := float32(math.Round(float64(x)))
	pY := float32(math.Round(float64(y)))
	c.ViewSnapped.
		Translate(-pX - centerX, -pY - centerY, 0).
		Scale(scaleX, scaleY, 1.0).
		Translate(centerX, centerY, 0)

	// centerX := float32(math.Round(float64(cameraCenter[0])))
	// centerY := float32(math.Round(float64(cameraCenter[1])))
	// pX := float32(math.Round(float64(x)))
	// pY := float32(math.Round(float64(y)))
	// c.View.
	// 	Translate(-pX - centerX, -pY - centerY, 0).
	// 	Scale(scaleX, scaleY, 1.0).
	// 	Translate(centerX, centerY, 0)

	// centerX := float64(cameraCenter[0])
	// centerY := float64(cameraCenter[1])
	// pX := float64(x)
	// pY := float64(y)
	// c.View.
	// 	Translate(float32(-pX - centerX), float32(-pY - centerY), 0).
	// 	Scale(scaleX, scaleY, 1.0).
	// 	Translate(float32(centerX), float32(centerY), 0)



	// c.View.
	// 	// Translate by x, y of the camera
	// 	Translate(-float32(math.Round(float64(x))), -float32(math.Round(float64(y))), 0).
	// 	// Scale around the center of the camera
	// 	Translate(-float32(math.Round(float64(cameraCenter[0]))), -float32(math.Round(float64(cameraCenter[1]))), 0).
	// 	Scale(scaleX, scaleY, 1.0).
	// 	Translate(float32(math.Round(float64(cameraCenter[0]))), float32(math.Round(float64(cameraCenter[1]))), 0)
}

func (c *CameraOrtho) Project(point Vec3) Vec3 {
	p := c.View.Apply(point)
	return p
}

func (c *CameraOrtho) Unproject(point Vec3) Vec3 {
	p := c.View.Inv().Apply(point)
	return p

	// return c.Projection.Mul(&c.View).Inv().Apply(point)
	// point[0] /= 1920
	// point[1] /= 1080

	// p := c.Projection.Mul(&c.View).Inv().Apply(point)
	// return p
	// return Vec3{
	// 	p[0] / 1920,
	// 	p[1] / 1080,
	// 	p[2] / 1,
	// }

	// m := c.Projection.Mul(&c.View)
	// det := m[i4_0_0]*m[i4_1_1] - m[i4_1_0]*m[i4_0_1]
	// return Vec3{
	// 	(m[i4_1_1]*(point[0] - m[i4_3_0]) - m[2]*(point[1] - m[i4_3_1])) / det,
	// 	(-m[i4_0_1]*(point[0] - m[i4_3_0]) + m[0]*(point[1] - m[i4_3_1])) / det,
	// 	0,
	// }
}

type Camera struct {
	Projection Mat4
	View Mat4

	Position Vec3
	Target Vec3
}

func NewCamera() *Camera {
	return &Camera{
		Projection: Mat4Ident,
		View: Mat4Ident,
		Position: Vec3{0, 0, 0},
		Target: Vec3{0, 0, 0},
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
		c.Position[0], c.Position[1], c.Position[2],
		c.Target[0], c.Target[1], c.Target[2],
		0, 0, 1,
	))
}
