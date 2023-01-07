package glitch

import (
	"fmt"
)

// TODO - right now we only support floats (for simplicity)
type VertexFormat []VertexAttr
type UniformFormat []Attr

type VertexAttr struct {
	Attr // The underlying Attribute
	Swizzle SwizzleType // This defines how the shader wants to map a generic object (like a mesh, to the shader buffers)
}

type Attr struct {
	Name string
	Type AttrType
}
func (a Attr) GetBuffer() any {
	switch a.Type {
	case AttrFloat:
		return &[]float32{}
	case AttrVec2:
		return &[]glVec2{}
	case AttrVec3:
		return &[]glVec3{}
	case AttrVec4:
		return &[]glVec4{}
	default:
		panic(fmt.Sprintf("Attr not valid for GetBuffer: %v", a))
	}
}

// Returns the size of the attribute type
func (a Attr) Size() int {
	switch a.Type {
	case AttrInt: return 1
	case AttrFloat: return 1
	case AttrVec2: return 2
	case AttrVec3: return 3
	case AttrVec4: return 4
	case AttrMat2: return 2 * 2
	case AttrMat23: return 2 * 3
	case AttrMat24: return 2 * 4
	case AttrMat3: return 3 * 3
	case AttrMat32: return 3 * 2
	case AttrMat34: return 3 * 4
	case AttrMat4: return 4 * 4
	case AttrMat42: return 4 * 2
	case AttrMat43: return 4 * 3
	default: panic(fmt.Sprintf("Invalid Attribute: %v", a))
	}
}

// This type is used to define the underlying data type of a vertex attribute or uniform attribute
type AttrType uint8
const (
	// TODO - others
	AttrInt AttrType = iota
	AttrFloat
	AttrVec2
	AttrVec3
	AttrVec4
	AttrMat2
	AttrMat23
	AttrMat24
	AttrMat3
	AttrMat32
	AttrMat34
	AttrMat4
	AttrMat42
	AttrMat43
)

// This type is used to define how generic meshes map into specific shader buffers
type SwizzleType uint8
const (
	PositionXY SwizzleType = iota
	PositionXYZ
	NormalXY
	NormalXYZ
	ColorR
	ColorRG
	ColorRGB
	ColorRGBA
	TexCoordXY
	// TexCoordXYZ // Is this a thing?
)
