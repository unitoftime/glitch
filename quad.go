package glitch

import (
	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch/shaders"
)

type Quad struct {
	Frame    Rect     // The bounds inside the spritesheet in the material
	Origin   glm.Vec3 // TODO: Hack to allow for offsets in the frame other than (0, 0)
	material Material // Note: Texture is in here
}

func NewSpriteQuad(texture *Texture, frame Rect) Quad {
	// Note: Checked here because Fill (which needs the texture size) runs later, after sorting and batching
	if texture == nil {
		panic("NewSpriteQuad: texture must not be nil")
	}
	return Quad{
		Frame:    frame,
		material: DefaultMaterial(texture),
	}
}

// Convert a sprite to a quad
func (s *Sprite) ToQuad() Quad {
	return Quad{
		Frame:    s.frame,
		Origin:   s.mesh.origin,
		material: s.material,
	}
}

func (s Quad) Bounds() glm.Box {
	return s.Frame.WithCenter(glm.Vec2{}).Box()
}

func (s Quad) g() GeometryFiller {
	return GeometryFiller{
		fillType: fillTypeQuad,
		ref:      s.material.texture,
		quad:     s.geom(),
	}
}

// Returns just the geometry of the quad. The material is passed separately when drawing
func (s Quad) geom() quadGeom {
	return quadGeom{
		min:    glv2(s.Frame.Min),
		max:    glv2(s.Frame.Max),
		origin: glv3(s.Origin),
	}
}

// Cuts the bottom of the frame rectangle off by amount pixels
func (s Quad) CutBottom(amount float64) Quad {
	ret := s
	// ret.Frame.CutTop(amount)
	// cutRect.Min.Y = cutRect.Max.Y - amount
	ret.Frame.Max.Y -= amount
	ret.Frame.Max.Y = max(ret.Frame.Min.Y, ret.Frame.Max.Y)
	return ret
}

func (s Quad) Draw(target BatchTarget, matrix Mat4) {
	s.DrawColorMask(target, matrix, White)
}
func (s Quad) DrawColorMask(target BatchTarget, matrix Mat4, mask RGBA) {
	target.Add(s.g(), glm4(matrix), mask, s.material)
}

func (s Quad) RectDraw(target BatchTarget, bounds Rect) {
	s.RectDrawColorMask(target, bounds, White)
}
func (s Quad) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	matrix := Mat4Ident
	matrix.Scale(bounds.W()/s.Frame.W(), bounds.H()/s.Frame.H(), 1).
		Translate(bounds.W()/2+bounds.Min.X, bounds.H()/2+bounds.Min.Y, 0)
	s.DrawColorMask(target, matrix, mask)
}

// Note: For caching purposes
var quadIndices = []uint32{
	0, 1, 3,
	1, 2, 3,
}

func (s Quad) Fill(pool *BufferPool, mat glMat4, mask RGBA) *VertexBuffer {
	return s.geom().Fill(pool, mat, mask, s.material.texture)
}

// The geometry of a quad. This is what a GeometryFiller stores inline for quads (See: fillTypeQuad)
// Note: Stored as float32 to keep GeometryFiller small. The vertices are float32 anyways
type quadGeom struct {
	min, max glVec2 // The frame inside the texture
	origin   glVec3 // The offset of the quad from its center
}

// The half size of the frame
func (s quadGeom) halfSize() (float32, float32) {
	return (s.max[0] - s.min[0]) / 2, (s.max[1] - s.min[1]) / 2
}

// The frame, centered on (0, 0)
func (s quadGeom) Bounds() glm.Box {
	w, h := s.halfSize()
	return glm.R(float64(-w), float64(-h), float64(w), float64(h)).Box()
}

func (s quadGeom) Fill(pool *BufferPool, mat glMat4, mask RGBA, texture *Texture) *VertexBuffer {
	numVerts := 4
	vertexBuffer := pool.Reserve(quadIndices, numVerts, pool.shader.tmpBuffers)

	destBuffs := pool.shader.tmpBuffers
	for bufIdx, attr := range pool.shader.attrFmt {
		// TODO - I'm not sure of a good way to break up this switch statement
		switch attr.Swizzle {
		case shaders.PositionXYZ:
			// The frame centered on (0, 0), then moved by the origin
			w, h := s.halfSize()
			min := glVec3{-w - s.origin[0], -h - s.origin[1], -s.origin[2]}
			max := glVec3{w - s.origin[0], h - s.origin[1], -s.origin[2]}

			if mat != glMat4Ident {
				min = mat.Apply(min)
				max = mat.Apply(max)
			}

			// TODO: Depth? Right now I just do min[2] b/c max and min should be on same Z axis
			posBuf := *(destBuffs[bufIdx]).(*[]glVec3)
			posBuf[0] = glVec3{max[0], max[1], min[2]}
			posBuf[1] = glVec3{max[0], min[1], min[2]}
			posBuf[2] = glVec3{min[0], min[1], min[2]}
			posBuf[3] = glVec3{min[0], max[1], min[2]}

		case shaders.ColorRGBA:
			colBuf := *(destBuffs[bufIdx]).(*[]glVec4)
			color := glc4(mask)
			colBuf[0] = color
			colBuf[1] = color
			colBuf[2] = color
			colBuf[3] = color
		case shaders.TexCoordXY:
			texW := float32(texture.width)
			texH := float32(texture.height)
			uvMin := glVec2{s.min[0] / texW, s.min[1] / texH}
			uvMax := glVec2{s.max[0] / texW, s.max[1] / texH}

			texBuf := *(destBuffs[bufIdx]).(*[]glVec2)
			texBuf[0] = glVec2{uvMax[0], uvMin[1]}
			texBuf[1] = glVec2{uvMax[0], uvMax[1]}
			texBuf[2] = glVec2{uvMin[0], uvMax[1]}
			texBuf[3] = glVec2{uvMin[0], uvMin[1]}
		default:
			panic("Unsupported")
		}
	}

	return vertexBuffer
}
