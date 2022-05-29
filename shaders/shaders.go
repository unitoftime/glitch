package shaders

import (
	"github.com/unitoftime/glitch"
	_ "embed"
)

func VertexAttribute(name string, Type glitch.AttrType, swizzle glitch.SwizzleType) glitch.VertexAttr {
	return glitch.VertexAttr{
		Attr: glitch.Attr{
			Name: name,
			Type: Type,
		},
		Swizzle: swizzle,
	}
}

//go:embed sprite.vs
var SpriteVertexShader string;

//go:embed sprite.fs
var SpriteFragmentShader string;

var SpriteShader = glitch.ShaderConfig{
	VertexShader: SpriteVertexShader,
	FragmentShader: SpriteFragmentShader,
	VertexFormat: glitch.VertexFormat{
		VertexAttribute("positionIn", glitch.AttrVec2, glitch.PositionXY),
		VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attr{"projection", glitch.AttrMat4},
		glitch.Attr{"view", glitch.AttrMat4},
	},
}

//go:embed subPixel.fs
var SubPixelAntiAliased string;

var PixelArtShader = glitch.ShaderConfig{
	VertexShader: SpriteVertexShader,
	FragmentShader: SubPixelAntiAliased,
	VertexFormat: glitch.VertexFormat{
		VertexAttribute("positionIn", glitch.AttrVec2, glitch.PositionXY),
		VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attr{"projection", glitch.AttrMat4},
		glitch.Attr{"view", glitch.AttrMat4},
	},
}

//go:embed mesh.vs
var DiffuseVertexShader string;

//go:embed flat.fs
var DiffuseFragmentShader string;

var DiffuseShader = glitch.ShaderConfig{
	VertexShader: DiffuseVertexShader,
	FragmentShader: DiffuseFragmentShader,
	VertexFormat: glitch.VertexFormat{
		VertexAttribute("positionIn", glitch.AttrVec3, glitch.PositionXYZ),
		// glitch.Attr{"normal", glitch.AttrVec3},
		VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attr{"model", glitch.AttrMat4},
		glitch.Attr{"view", glitch.AttrMat4},
		glitch.Attr{"projection", glitch.AttrMat4},
		// glitch.Attr{"dirlight.direction", glitch.AttrVec3},
		// glitch.Attr{"dirlight.ambient", glitch.AttrVec3},
		// glitch.Attr{"dirlight.diffuse", glitch.AttrVec3},
		// glitch.Attr{"dirlight.specular", glitch.AttrVec3},
	},
}
