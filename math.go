package glitch

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

type glVec2 [2]float32
type glVec3 [3]float32
type glVec4 [4]float32

func (v glVec3) Add(u glVec3) glVec3 {
	return glVec3{v[0] + u[0], v[1] + u[1], v[2] + u[2]}
}
func (v glVec3) Float64() Vec3 {
	return Vec3{float64(v[0]), float64(v[1]), float64(v[2])}
}

type glMat4 [16]float32

var glMat4Ident = Mat4Ident.gl()

func (m *glMat4) Apply(v glVec3) glVec3 {
	return glVec3{
		m[i4_0_0]*v[0] + m[i4_1_0]*v[1] + m[i4_2_0]*v[2] + m[i4_3_0], // w = 1.0
		m[i4_0_1]*v[0] + m[i4_1_1]*v[1] + m[i4_2_1]*v[2] + m[i4_3_1], // w = 1.0
		m[i4_0_2]*v[0] + m[i4_1_2]*v[1] + m[i4_2_2]*v[2] + m[i4_3_2], // w = 1.0
	}
}

// TODO: untested. Is this right? I guess v[0] = 0?
func (m *glMat4) ApplyVec2(v glVec2) glVec2 {
	return glVec2{
		m[i4_0_0]*v[0] + m[i4_1_0]*v[1] + m[i4_3_0], // w = 1.0
		m[i4_0_1]*v[0] + m[i4_1_1]*v[1] + m[i4_3_1], // w = 1.0
	}
}

func (m *glMat4) Inv() *glMat4 {
	retMat := glMat4(mgl32.Mat4(*m).Inv())
	return &retMat
}
func (m *glMat4) Transpose() *glMat4 {
	retMat := glMat4(mgl32.Mat4(*m).Transpose())
	return &retMat
}
func (m *glMat4) Mul(n *glMat4) *glMat4 {
	// TODO: Does this improve performance?
	// if *m == glMat4Ident {
	// 	*m = *n
	// 	return m
	// } else if *n == glMat4Ident {
	// 	return m
	// }

	// This is in column major order
	*m = glMat4{
	// return &Mat4{
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
	return m
}


func (v Vec2) gl() glVec2 {
	return glVec2{float32(v.X), float32(v.Y)}
}
func (v Vec3) gl() glVec3 {
	return glVec3{float32(v.X), float32(v.Y), float32(v.Z)}
}
func (v Vec4) gl() glVec4 {
	return glVec4{float32(v[0]), float32(v[1]), float32(v[2]), float32(v[3])}
}
func (m Mat4) gl() glMat4 {
	ret := glMat4{}
	for i := range m {
		ret[i] = float32(m[i])
	}
	return ret
}

func (m glMat4) writeToFloat32(s []float32) []float32 {
	// TODO: Replace with copy()?
	for i := range m {
		s = append(s, m[i])
	}
	return s
}

func (m Mat4) writeToFloat32(s []float32) []float32 {
	for i := range m {
		s = append(s, float32(m[i]))
	}
	return s
}

// TODO - conver these to structs {x, y}
type Vec2 struct {
	X, Y float64
}
type Vec3 struct {
	X, Y, Z float64
}
type Vec4 [4]float64

func (v Vec2) Add(u Vec2) Vec2 {
	return Vec2{v.X + u.X, v.Y + u.Y}
}

func (v Vec2) Sub(u Vec2) Vec2 {
	return Vec2{v.X - u.X, v.Y - u.Y}
}

func (v Vec2) Snap() Vec2 {
	return Vec2{
		math.Round(v.X),
		math.Round(v.Y),
	}
}

func (v Vec2) Unit() Vec2 {
	len := v.Len()
	return Vec2{v.X/len, v.Y/len}
}

func (v Vec2) Len() float64 {
	return math.Hypot(float64(v.X), float64(v.Y))
}

func (v Vec2) Scaled(s float64) Vec2 {
	return Vec2{s * v.X, s * v.Y}
}
func (v Vec2) ScaledXY(s Vec2) Vec2 {
	return Vec2{v.X * s.X, v.Y * s.Y}
}

func (v Vec2) Vec3() Vec3 {
	return Vec3{v.X, v.Y, 0}
}

func (v Vec3) Add(u Vec3) Vec3 {
	return Vec3{v.X + u.X, v.Y + u.Y, v.Z + u.Z}
}

func (v Vec3) Sub(u Vec3) Vec3 {
	return Vec3{v.X - u.X, v.Y - u.Y, v.Z - u.Z}
}

// Finds the dot product of two vectors
func (v Vec3) Dot(u Vec3) float64 {
	return (v.X * u.X) + (v.Y * u.Y) + (v.Z * u.Z)
}

// Finds the angle between two vectors
func (v Vec3) Angle(u Vec3) float64 {
	return  math.Acos(v.Dot(u) / (v.Len() * u.Len()))
}

func (v Vec3) Theta() float64 {
	return math.Atan2(v.Y, v.X)
}

// Rotates the vector by theta on the XY 2d plane
func (v Vec3) Rotate2D(theta float64) Vec3 {
	t := theta
	x := v.X
	y := v.Y
	x1 := x * math.Cos(t) - y * math.Sin(t)
	y1 := x * math.Sin(t) + y * math.Cos(t)
	return Vec3{x1, y1, v.Z}
}

func (v Vec3) Len() float64 {
	// return float32(math.Hypot(float64(v.X), float64(v.Y)))
	a := v.X
	b := v.Y
	c := v.Z
	return math.Sqrt((a * a) + (b * b) + (c * c))
}

func (v Vec3) Vec2() Vec2 {
	return Vec2{v.X, v.Y}
}

func (v Vec3) Unit() Vec3 {
	len := v.Len()
	return Vec3{v.X/len, v.Y/len, v.Z/len}
}

func (v Vec3) Scaled(x, y, z float64) Vec3 {
	v.X *= x
	v.Y *= y
	v.Z *= z

	return v
}

// All Matrices are in column-major order
type Mat2 [4]float64
type Mat3 [9]float64
type Mat4 [16]float64

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

func (m *Mat3) Translate(x, y float64) *Mat3 {
	m[i3_2_0] = m[i3_2_0] + x
	m[i3_2_1] = m[i3_2_1] + y
	return m
}

// Note: Scales around 0,0
func (m *Mat4) Scale(x, y, z float64) *Mat4 {
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

func (m *Mat4) Translate(x, y, z float64) *Mat4 {
	m[i4_3_0] = m[i4_3_0] + x
	m[i4_3_1] = m[i4_3_1] + y
	m[i4_3_2] = m[i4_3_2] + z
	return m
}

func (m *Mat4) GetTranslation() Vec3 {
	return Vec3{m[i4_3_0], m[i4_3_1], m[i4_3_2]}
}

// https://github.com/go-gl/mathgl/blob/v1.0.0/mgl32/transform.go#L159
func (m *Mat4) Rotate(angle float64, axis Vec3) *Mat4 {
	// quat := mgl32.Mat4ToQuat(mgl32.Mat4(*m))
	// return &retMat
	rotation := Mat4(mgl64.HomogRotate3D(angle, mgl64.Vec3{axis.X, axis.Y, axis.Z}))
	// retMat := Mat4(mgl32.Mat4(*m).)
	// return &retMat
	mNew := m.Mul(&rotation)
	*m = *mNew
	return m
}

// Note: This modifies in place
func (m *Mat4) Mul(n *Mat4) *Mat4 {
	// This is in column major order
	*m = Mat4{
	// return &Mat4{
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

type Box struct {
	Min, Max Vec3
}
func (b Box) Rect() Rect {
	return Rect{
		Min: Vec2{b.Min.X, b.Min.Y},
		Max: Vec2{b.Max.X, b.Max.Y},
	}
}


func (a Box) Union(b Box) Box {
	x1, _ := minMax(a.Min.X, b.Min.X)
	_, x2 := minMax(a.Max.X, b.Max.X)
	y1, _ := minMax(a.Min.Y, b.Min.Y)
	_, y2 := minMax(a.Max.Y, b.Max.Y)
	z1, _ := minMax(a.Min.Z, b.Min.Z)
	_, z2 := minMax(a.Max.Z, b.Max.Z)
	return Box{
		Min: Vec3{x1, y1, z1},
		Max: Vec3{x2, y2, z2},
	}
}

// TODO: This is the wrong input matrix type
func (b Box) Apply(mat glMat4) Box {
	return Box{
		Min: mat.Apply(b.Min.gl()).Float64(),
		Max: mat.Apply(b.Max.gl()).Float64(),
	}
}


type Rect struct {
	Min, Max Vec2
}
func R(minX, minY, maxX, maxY float64) Rect {
	// TODO - guarantee min is less than max
	return Rect{
		Min: Vec2{minX, minY},
		Max: Vec2{maxX, maxY},
	}
}

// Creates a quad mesh from this rect
func (r Rect) ToMesh() *Mesh {
	return NewQuadMesh(r, R(0, 0, 1, 1))
}

// Returns a box that holds this rect. The Z axis is 0
func (r Rect) Box() Box {
	return r.ToBox()
}
func (r Rect) ToBox() Box {
	return Box{
		Min: Vec3{r.Min.X, r.Min.Y, 0},
		Max: Vec3{r.Max.X, r.Max.Y, 0},
	}
}

func (r Rect) W() float64 {
	return r.Max.X - r.Min.X
}

func (r Rect) H() float64 {
	return r.Max.Y - r.Min.Y
}

func (r Rect) Center() Vec2 {
	return Vec2{r.Min.X + (r.W()/2), r.Min.Y + (r.H()/2)}
}

// func (r Rect) CenterAt(v Vec2) Rect {
// 	return r.Moved(r.Center().Scaled(-1)).Moved(v)
// }
func (r Rect) WithCenter(v Vec2) Rect {
	w := r.W()/2
	h := r.H()/2
	return R(v.X - w, v.Y - h, v.X + w, v.Y + h)
}

// TODO: Should I make a pointer version of this that handles the nil case too?
// Returns the smallest rect which contains both input rects
func (r Rect) Union(s Rect) Rect {
	r = r.Norm()
	s = s.Norm()
	x1, _ := minMax(r.Min.X, s.Min.X)
	_, x2 := minMax(r.Max.X, s.Max.X)
	y1, _ := minMax(r.Min.Y, s.Min.Y)
	_, y2 := minMax(r.Max.Y, s.Max.Y)
	return R(x1, y1, x2, y2)
}

func (r Rect) Moved(v Vec2) Rect {
	return Rect{
		Min: r.Min.Add(v),
		Max: r.Max.Add(v),
	}
}

// Calculates the scale required to fit rect r inside r2
func (r Rect) FitScale(r2 Rect) float64 {
	scaleX := r2.W() / r.W()
	scaleY := r2.H() / r.H()

	min := min(scaleX, scaleY)
	return min
}

// Scales rect r uniformly to fit inside rect r2
// TODO This only scales around {0, 0}
func (r Rect) ScaledToFit(r2 Rect) Rect {
	return r.Scaled(r.FitScale(r2))
}

// Returns the largest square that fits inside the rectangle
func (r Rect) SubSquare() Rect {
	w := r.W()
	h := r.H()
	min, _ := minMax(w, h)
	m2 := min/2
	return R(-m2, -m2, m2, m2).Moved(r.Center())
}

func (r Rect) CenterScaled(scale float64) Rect {
	c := r.Center()
	w := r.W() * scale / 2.0
	h := r.H() * scale / 2.0
	return R(c.X - w, c.Y - h, c.X + w, c.Y + h)
}

// Note: This scales around the center
// func (r Rect) ScaledXY(scale Vec2) Rect {
// 	c := r.Center()
// 	w := r.W() * scale.X / 2.0
// 	h := r.H() * scale.Y / 2.0
// 	return R(c.X - w, c.Y - h, c.X + w, c.Y + h)
// }

// TODO: I need to deprecate this. This currently just indepentently scales the min and max point which is only useful if the center, min, or max is on (0, 0)
func (r Rect) Scaled(scale float64) Rect {
	// center := r.Center()
	// r = r.Moved(center.Scaled(-1))
	r = Rect{
		Min: r.Min.Scaled(scale),
		Max: r.Max.Scaled(scale),
	}
	// r = r.Moved(center)
	return r
}

func (r Rect) ScaledXY(scale Vec2) Rect {
	r = Rect{
		Min: r.Min.ScaledXY(scale),
		Max: r.Max.ScaledXY(scale),
	}
	return r
}

func (r Rect) Norm() Rect {
	x1, x2 := minMax(r.Min.X, r.Max.X)
	y1, y2 := minMax(r.Min.Y, r.Max.Y)
	return R(x1, y1, x2, y2)
}

func (r Rect) Contains(x, y float64) bool {
	return x > r.Min.X && x < r.Max.X && y > r.Min.Y && y < r.Max.Y
}

func (r Rect) Intersects(r2 Rect) bool {
	return (
		r.Min.X <= r2.Max.X &&
			r.Max.X >= r2.Min.X &&
			r.Min.Y <= r2.Max.Y &&
			r.Max.Y >= r2.Min.Y)
}

// Layous out 'n' rectangles horizontally with specified padding between them and returns that rect
// The returned rectangle has a min point of 0,0
func (r Rect) LayoutHorizontal(n int, padding float64) Rect {
	return R(
		0,
		0,
		float64(n) * r.W() + float64(n-1) * padding,
		r.H(),
	)
}

func (r *Rect) CutLeft(amount float64) Rect {
	cutRect := *r
	cutRect.Max.X = cutRect.Min.X + amount
	r.Min.X += amount
	return cutRect
}

func (r *Rect) CutRight(amount float64) Rect {
	cutRect := *r
	cutRect.Min.X = cutRect.Max.X - amount
	r.Max.X -= amount
	return cutRect
}

func (r *Rect) CutBottom(amount float64) Rect {
	cutRect := *r
	cutRect.Max.Y = cutRect.Min.Y + amount
	r.Min.Y += amount
	return cutRect
}

func (r *Rect) CutTop(amount float64) Rect {
	cutRect := *r
	cutRect.Min.Y = cutRect.Max.Y - amount
	r.Max.Y -= amount
	return cutRect
}

// Returns a centered horizontal sliver with height set by amount
func (r Rect) SliceHorizontal(amount float64) Rect {
	r.CutTop((r.H() - amount) / 2)
	return r.CutTop(amount)
}

// Returns a centered vertical sliver with width set by amount
func (r Rect) SliceVertical(amount float64) Rect {
	r.CutRight((r.W() - amount) / 2)
	return r.CutRight(amount)
}

func (r Rect) Snap() Rect {
	r.Min = r.Min.Snap()
	r.Max = r.Max.Snap()
	return r
}

// Adds padding to a rectangle consistently
func (r Rect) PadAll(padding float64) Rect {
	return r.Pad(R(padding, padding, padding, padding))
}

// Adds padding to a rectangle (pads inward if padding is negative)
func (r Rect) Pad(pad Rect) Rect {
	return R(r.Min.X - pad.Min.X, r.Min.Y - pad.Min.Y, r.Max.X + pad.Max.X, r.Max.Y + pad.Max.Y)
}

// Removes padding from a rectangle (pads outward if padding is negative). Essentially calls pad but with negative values
func (r Rect) Unpad(pad Rect) Rect {
	return r.Pad(pad.Scaled(-1))
}

// Takes r2 and places it in r based on the alignment
// TODO - rename to InnerAnchor?
func (r Rect) Anchor(r2 Rect, anchor Vec2) Rect {
	// Anchor point is the position in r that we are anchoring to
	anchorPoint := Vec2{r.Min.X + (anchor.X * r.W()) , r.Min.Y + (anchor.Y * r.H())}
	pivotPoint := Vec2{r2.Min.X + (anchor.X * r2.W()) , r2.Min.Y + (anchor.Y * r2.H())}

	// fmt.Println("Anchor:", anchorPoint)
	// fmt.Println("Pivot:", pivotPoint)

	a := Vec2{anchorPoint.X - pivotPoint.X, anchorPoint.Y - pivotPoint.Y}
	return R(a.X, a.Y, a.X + r2.W(), a.Y + r2.H()).Norm()
}

// Anchors r2 to r1 based on two anchors, one for r and one for r2
// TODO - rename to Anchor?
func (r Rect) FullAnchor(r2 Rect, anchor, pivot Vec2) Rect {
	anchorPoint := Vec2{r.Min.X + (anchor.X * r.W()), r.Min.Y + (anchor.Y * r.H())}
	pivotPoint := Vec2{r2.Min.X + (pivot.X * r2.W()) , r2.Min.Y + (pivot.Y * r2.H())}

	a := Vec2{anchorPoint.X - pivotPoint.X, anchorPoint.Y - pivotPoint.Y}
	return R(a.X, a.Y, a.X + r2.W(), a.Y + r2.H()).Norm()
}

// Move the min point of the rect to a certain position
func (r Rect) MoveMin(pos Vec2) Rect {
	dv := r.Min.Sub(pos)
	return r.Moved(dv)
}


func lerp(a, b float64, t float64) float64 {
	m := b - a // Slope = Rise over run | Note: Run = (1 - 0)
	y := (m * t) + a
	return y
}

// returns the min, max of the two numbers
func minMax(a, b float64) (float64, float64) {
	if a > b {
		return b, a
	}
	return a, b
}

func (m *Mat4) Apply(v Vec3) Vec3 {
	return Vec3{
		m[i4_0_0]*v.X + m[i4_1_0]*v.Y + m[i4_2_0]*v.Z + m[i4_3_0], // w = 1.0
		m[i4_0_1]*v.X + m[i4_1_1]*v.Y + m[i4_2_1]*v.Z + m[i4_3_1], // w = 1.0
		m[i4_0_2]*v.X + m[i4_1_2]*v.Y + m[i4_2_2]*v.Z + m[i4_3_2], // w = 1.0
	}
}

func (m *Mat3) Apply( v Vec2) Vec2 {
	return Vec2{
		m[i3_0_0]*v.X + m[i3_1_0]*v.Y + m[i3_2_0],
		m[i3_0_1]*v.X + m[i3_1_1]*v.Y + m[i3_2_1],
	}
}

// Note: Returns a new Mat4
func (m *Mat4) Inv() *Mat4 {
	retMat := Mat4(mgl64.Mat4(*m).Inv())
	return &retMat
}

func (m *Mat4) Transpose() *Mat4 {
	retMat := Mat4(mgl64.Mat4(*m).Transpose())
	return &retMat
}

func (r Rect) RectDraw(r2 Rect) Mat4 {
	mat := Mat4Ident
	mat.
		Scale(r2.W() / r.W(), r2.H() / r.H(), 1).
		Translate(r2.Min.X, r2.Min.Y, 0)
	return mat
}

// TODO - I feel like camera should be a higher-up abstraction and not held here
type CameraOrtho struct {
	Projection Mat4
	View Mat4
	// ViewSnapped Mat4
	bounds Rect
	DepthRange Vec2 // Specifies the near and far plane of the camera, defaults to (-1, 1)

	// Tracks the view inverse and whether or not its been recalculated or not
	ViewInv Mat4
	dirtyViewInv bool
}

func NewCameraOrtho() *CameraOrtho {
	return &CameraOrtho{
		Projection: Mat4Ident,
		View: Mat4Ident,
		// ViewSnapped: Mat4Ident,
		bounds: R(0,0,1,1),
		DepthRange: Vec2{-1, 1},
	}
}

func (c *CameraOrtho) Bounds() Rect {
	return c.bounds
}

func (c *CameraOrtho) SetOrtho2D(bounds Rect) {
	c.dirtyViewInv = true

	c.bounds = bounds

	c.Projection = Mat4(mgl64.Ortho(0, c.bounds.W(), 0, c.bounds.H(), c.DepthRange.X, c.DepthRange.Y))
}

// Helpful: https://stackoverflow.com/questions/2346238/opengl-how-do-i-avoid-rounding-errors-when-specifying-uv-co-ordinates
func (c *CameraOrtho) SetView2D(x, y, scaleX, scaleY float64) {
	c.dirtyViewInv = true

	c.View = Mat4Ident
	cameraCenter := c.bounds.Center()
	// c.View.
	// 	// Translate by x, y of the camera
	// 	Translate(-x, -y, 0).
	// 	// Scale around the center of the camera
	// 	Translate(-cameraCenter.X, -cameraCenter.Y, 0).
	// 	Scale(scaleX, scaleY, 1.0).
	// 	Translate(cameraCenter.X, cameraCenter.Y, 0)

	// Rounding the cameraCenter position helps fix scaling issues where we might have scaled around a non integer position
	cX := math.Round(cameraCenter.X)
	cY := math.Round(cameraCenter.Y)
	c.View.
		// Translate by x, y of the camera
		Translate(-x, -y, 0).
		// Scale around the center of the camera
		Translate(-cX, -cY, 0).
		Scale(scaleX, scaleY, 1.0).
		Translate(cX, cY, 0)


	// // TODO - this is literally only for pixel art
	// c.ViewSnapped = Mat4Ident
	// centerX := math.Round(cameraCenter[0])
	// centerY := math.Round(cameraCenter[1])
	// pX := math.Round(x)
	// pY := math.Round(y)
	// c.ViewSnapped.
	// 	Translate(-pX - centerX, -pY - centerY, 0).
	// 	Scale(scaleX, scaleY, 1.0).
	// 	Translate(centerX, centerY, 0)




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
	// TODO - This logic breaks down if someone modifies camera internals. Ie I need better protection. for private members
	if c.dirtyViewInv == true {
		c.ViewInv = *c.View.Inv()
		c.dirtyViewInv = false
	}
	return c.ViewInv.Apply(point)

	// p := c.View.Inv().Apply(point)
	// return p
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
	c.Projection = Mat4(mgl64.Perspective(math.Pi/4, aspect, 0.1, 1000))
}

func (c *Camera) SetViewLookAt(win *Window) {
	c.View = Mat4(mgl64.LookAt(
		c.Position.X, c.Position.Y, c.Position.Z,
		c.Target.X, c.Target.Y, c.Target.Z,
		0, 0, 1,
	))
}

func (c *Camera) Material() CameraMaterial {
	return CameraMaterial{
		Projection: c.Projection.gl(),
		View: c.View.gl(),
	}
}
