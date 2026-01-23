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
		fillType: fillTypeProgrammatic,
		prog:     s,
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
	numVerts := 4
	vertexBuffer := pool.Reserve(quadIndices, numVerts, pool.shader.tmpBuffers)

	destBuffs := pool.shader.tmpBuffers
	for bufIdx, attr := range pool.shader.attrFmt {
		// TODO - I'm not sure of a good way to break up this switch statement
		switch attr.Swizzle {
		case shaders.PositionXYZ:
			bounds := s.Bounds().Moved(s.Origin.Scaled(-1, -1, -1))
			min := glv3(bounds.Min)
			max := glv3(bounds.Max)

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
			texture := s.material.texture
			uvBounds := glm.R(
				s.Frame.Min.X/float64(texture.width),
				s.Frame.Min.Y/float64(texture.height),
				s.Frame.Max.X/float64(texture.width),
				s.Frame.Max.Y/float64(texture.height),
			)

			texBuf := *(destBuffs[bufIdx]).(*[]glVec2)
			texBuf[0] = glVec2{float32(uvBounds.Max.X), float32(uvBounds.Min.Y)}
			texBuf[1] = glVec2{float32(uvBounds.Max.X), float32(uvBounds.Max.Y)}
			texBuf[2] = glVec2{float32(uvBounds.Min.X), float32(uvBounds.Max.Y)}
			texBuf[3] = glVec2{float32(uvBounds.Min.X), float32(uvBounds.Min.Y)}
		default:
			panic("Unsupported")
		}
	}

	return vertexBuffer
}
