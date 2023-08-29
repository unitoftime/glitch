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

// Commented out: This was for webgl1 mode
// //go:embed sprite_100.vs
// var SpriteVertexShaderWebGL1 string;

// //go:embed sprite_100.fs
// var SpriteFragmentShaderWebGL1 string;

// var SpriteShader = glitch.ShaderConfig{
// 	VertexShader: SpriteVertexShaderWebGL1,
// 	FragmentShader: SpriteFragmentShaderWebGL1,
// 	VertexFormat: glitch.VertexFormat{
// 		VertexAttribute("positionIn", glitch.AttrVec2, glitch.PositionXY),
// 		VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
// 		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
// 	},
// 	UniformFormat: glitch.UniformFormat{
// 		glitch.Attr{"projection", glitch.AttrMat4},
// 		glitch.Attr{"view", glitch.AttrMat4},
// 	},
// }

//go:embed sprite.vs
var SpriteVertexShader string;

//go:embed sprite.fs
var SpriteFragmentShader string;

var SpriteShader = glitch.ShaderConfig{
	VertexShader: SpriteVertexShader,
	FragmentShader: SpriteFragmentShader,
	VertexFormat: glitch.VertexFormat{
		VertexAttribute("positionIn", glitch.AttrVec3, glitch.PositionXYZ),
		VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attr{"model", glitch.AttrMat4},
		glitch.Attr{"projection", glitch.AttrMat4},
		glitch.Attr{"view", glitch.AttrMat4},
	},
}

//go:embed minimap.fs
var MinimapFragmentShader string;

var MinimapShader = glitch.ShaderConfig{
	VertexShader: SpriteVertexShader,
	FragmentShader: MinimapFragmentShader,
	VertexFormat: glitch.VertexFormat{
		VertexAttribute("positionIn", glitch.AttrVec3, glitch.PositionXYZ),
		VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attr{"model", glitch.AttrMat4},
		glitch.Attr{"projection", glitch.AttrMat4},
		glitch.Attr{"view", glitch.AttrMat4},
	},
}

//go:embed subPixel.fs
var SubPixelAntiAliased string;

var PixelArtShader = glitch.ShaderConfig{
	VertexShader: PixelArtVert,
	FragmentShader: SubPixelAntiAliased,
	// FragmentShader: SubPixelAntiAliased,
	VertexFormat: glitch.VertexFormat{
		VertexAttribute("positionIn", glitch.AttrVec3, glitch.PositionXYZ),
		VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attr{"model", glitch.AttrMat4},
		glitch.Attr{"projection", glitch.AttrMat4},
		glitch.Attr{"view", glitch.AttrMat4},
		glitch.Attr{"texelsPerPixel", glitch.AttrFloat},
	},
}

//go:embed pixel.vs
var PixelArtVert string;
//go:embed pixel.fs
var PixelArtFrag string;

var PixelArtShader2 = glitch.ShaderConfig{
	VertexShader: PixelArtVert,
	FragmentShader: PixelArtFrag,
	VertexFormat: glitch.VertexFormat{
		VertexAttribute("positionIn", glitch.AttrVec3, glitch.PositionXYZ),
		VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attr{"model", glitch.AttrMat4},
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
		VertexAttribute("normalIn", glitch.AttrVec3, glitch.NormalXYZ),
		// VertexAttribute("colorIn", glitch.AttrVec4, glitch.ColorRGBA),
		VertexAttribute("texCoordIn", glitch.AttrVec2, glitch.TexCoordXY),
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attr{"model", glitch.AttrMat4},
		glitch.Attr{"view", glitch.AttrMat4},
		glitch.Attr{"projection", glitch.AttrMat4},

		glitch.Attr{"viewPos", glitch.AttrVec3},

		glitch.Attr{"material.ambient", glitch.AttrVec3},
		glitch.Attr{"material.diffuse", glitch.AttrVec3},
		glitch.Attr{"material.specular", glitch.AttrVec3},
		glitch.Attr{"material.shininess", glitch.AttrFloat},

		glitch.Attr{"dirLight.direction", glitch.AttrVec3},
		glitch.Attr{"dirLight.ambient", glitch.AttrVec3},
		glitch.Attr{"dirLight.diffuse", glitch.AttrVec3},
		glitch.Attr{"dirLight.specular", glitch.AttrVec3},
	},
}
