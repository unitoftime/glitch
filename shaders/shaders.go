package shaders

import (
	_ "embed"
	"fmt"
)

type ShaderConfig struct {
	VertexShader, FragmentShader string
	VertexFormat                 VertexFormat
	UniformFormat                UniformFormat
}

// TODO - right now we only support floats (for simplicity)
type VertexFormat []VertexAttr
type UniformFormat []Attr

type VertexAttr struct {
	Attr                // The underlying Attribute
	Swizzle SwizzleType // This defines how the shader wants to map a generic object (like a mesh, to the shader buffers)
}

type Attr struct {
	Name string
	Type AttrType
}

// Returns the size of the attribute type
func (a Attr) Size() int {
	switch a.Type {
	case AttrInt:
		return 1
	case AttrFloat:
		return 1
	case AttrVec2:
		return 2
	case AttrVec3:
		return 3
	case AttrVec4:
		return 4
	case AttrMat2:
		return 2 * 2
	case AttrMat23:
		return 2 * 3
	case AttrMat24:
		return 2 * 4
	case AttrMat3:
		return 3 * 3
	case AttrMat32:
		return 3 * 2
	case AttrMat34:
		return 3 * 4
	case AttrMat4:
		return 4 * 4
	case AttrMat42:
		return 4 * 2
	case AttrMat43:
		return 4 * 3
	default:
		panic(fmt.Sprintf("Invalid Attribute: %v", a))
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

func VertexAttribute(name string, Type AttrType, swizzle SwizzleType) VertexAttr {
	return VertexAttr{
		Attr: Attr{
			Name: name,
			Type: Type,
		},
		Swizzle: swizzle,
	}
}

// Commented out: This was for webgl1 mode
// //go:embed sprite_100.vs
// var SpriteVertexShaderWebGL1 string;

// //go:embed sprite_100.fs
// var SpriteFragmentShaderWebGL1 string;

// var SpriteShader = ShaderConfig{
// 	VertexShader: SpriteVertexShaderWebGL1,
// 	FragmentShader: SpriteFragmentShaderWebGL1,
// 	VertexFormat: VertexFormat{
// 		VertexAttribute("positionIn", AttrVec2, PositionXY),
// 		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
// 		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
// 	},
// 	UniformFormat: UniformFormat{
// 		Attr{"projection", AttrMat4},
// 		Attr{"view", AttrMat4},
// 	},
// }

//go:embed sprite.vs
var SpriteVertexShader string

//go:embed sprite.fs
var SpriteFragmentShader string

var SpriteShader = ShaderConfig{
	VertexShader:   SpriteVertexShader,
	FragmentShader: SpriteFragmentShader,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"projection", AttrMat4},
		Attr{"view", AttrMat4},
		// Attr{"silhouetteMix", AttrFloat},
	},
}

//go:embed msdf.fs
var MSDFFragmentShader string

var MSDFShader = ShaderConfig{
	VertexShader:   SpriteVertexShader,
	FragmentShader: MSDFFragmentShader,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"projection", AttrMat4},
		Attr{"view", AttrMat4},
		Attr{"u_threshold", AttrFloat},
		Attr{"u_outline_width_relative", AttrFloat},
		Attr{"u_outline_blur", AttrFloat},
		Attr{"u_outline_color", AttrVec4},
	},
}

//go:embed sdf.fs
var SDFFragmentShader string

var SDFShader = ShaderConfig{
	VertexShader:   SpriteVertexShader,
	FragmentShader: SDFFragmentShader,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"projection", AttrMat4},
		Attr{"view", AttrMat4},
	},
}

//go:embed minimap.fs
var MinimapFragmentShader string

var MinimapShader = ShaderConfig{
	VertexShader:   SpriteVertexShader,
	FragmentShader: MinimapFragmentShader,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"projection", AttrMat4},
		Attr{"view", AttrMat4},
	},
}

//go:embed subPixel.fs
var SubPixelAntiAliased string

var PixelArtShader = ShaderConfig{
	VertexShader:   PixelArtVert,
	FragmentShader: SubPixelAntiAliased,
	// FragmentShader: SubPixelAntiAliased,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"projection", AttrMat4},
		Attr{"view", AttrMat4},
		Attr{"texelsPerPixel", AttrFloat},
	},
}

//go:embed pixel.vs
var PixelArtVert string

//go:embed pixel.fs
var PixelArtFrag string

var PixelArtShader2 = ShaderConfig{
	VertexShader:   PixelArtVert,
	FragmentShader: PixelArtFrag,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"projection", AttrMat4},
		Attr{"view", AttrMat4},
	},
}

//go:embed mesh.vs
var DiffuseVertexShader string

//go:embed flat.fs
var DiffuseFragmentShader string

var DiffuseShader = ShaderConfig{
	VertexShader:   DiffuseVertexShader,
	FragmentShader: DiffuseFragmentShader,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("normalIn", AttrVec3, NormalXYZ),
		// VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"view", AttrMat4},
		Attr{"projection", AttrMat4},

		Attr{"viewPos", AttrVec3},

		Attr{"material.ambient", AttrVec3},
		Attr{"material.diffuse", AttrVec3},
		Attr{"material.specular", AttrVec3},
		Attr{"material.shininess", AttrFloat},

		Attr{"dirLight.direction", AttrVec3},
		Attr{"dirLight.ambient", AttrVec3},
		Attr{"dirLight.diffuse", AttrVec3},
		Attr{"dirLight.specular", AttrVec3},
	},
}

//go:embed sprite-repeat.fs
var SpriteRepeatFragmentShader string

var SpriteRepeatShader = ShaderConfig{
	VertexShader:   SpriteVertexShader,
	FragmentShader: SpriteRepeatFragmentShader,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"projection", AttrMat4},
		Attr{"view", AttrMat4},
		Attr{"iTime", AttrFloat},
		Attr{"zoom", AttrVec2},
		Attr{"repeatRect", AttrVec4},
	},
}

//go:embed sprite-water.fs
var SpriteWaterFragmentShader string

var SpriteWaterShader = ShaderConfig{
	VertexShader:   SpriteVertexShader,
	FragmentShader: SpriteWaterFragmentShader,
	VertexFormat: VertexFormat{
		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
	},
	UniformFormat: UniformFormat{
		Attr{"model", AttrMat4},
		Attr{"projection", AttrMat4},
		Attr{"view", AttrMat4},
		Attr{"iTime", AttrFloat},
		Attr{"zoom", AttrVec2},
		Attr{"repeatRect", AttrVec4},
		Attr{"scaleVal", AttrFloat},
		Attr{"scaleFreq", AttrFloat},
		Attr{"noiseMoveBias", AttrVec2},
		Attr{"moveBias", AttrVec2},
	},
}
