package glitch

import (
	_ "embed"

	"github.com/unitoftime/glitch/shaders"
)

// //go:embed shaders/sprite.vs
// var spriteVertexShader string;

// //go:embed shaders/sprite.fs
// var spriteFragmentShader string;

// var spriteShader = ShaderConfig{
// 	VertexShader: spriteVertexShader,
// 	FragmentShader: spriteFragmentShader,
// 	VertexFormat: VertexFormat{
// 		VertexAttribute("positionIn", AttrVec3, PositionXYZ),
// 		VertexAttribute("colorIn", AttrVec4, ColorRGBA),
// 		VertexAttribute("texCoordIn", AttrVec2, TexCoordXY),
// 	},
// 	UniformFormat: UniformFormat{
// 		Attr{"model", AttrMat4},
// 		Attr{"projection", AttrMat4},
// 		Attr{"view", AttrMat4},
// 	},
// }
// func VertexAttribute(name string, Type AttrType, swizzle SwizzleType) VertexAttr {
// 	return VertexAttr{
// 		Attr: Attr{
// 			Name: name,
// 			Type: Type,
// 		},
// 		Swizzle: swizzle,
// 	}
// }

var defaultSpriteShader *Shader // Can set this to whatever you want

func SetDefaultSpriteShader(shader *Shader) {
	defaultSpriteShader = shader
}

func GetDefaultSpriteShader() *Shader {
	if defaultSpriteShader != nil {
		return defaultSpriteShader
	}

	// Note: We snuff the error here. If the user wants they can pre-supply a defaultspriteshader so this one never loads
	defaultSpriteShader, _ = NewShader(shaders.SpriteShader)
	return defaultSpriteShader
}

func DefaultMaterial(texture *Texture) Material {
	material := NewMaterial(GetDefaultSpriteShader())
	material.texture = texture
	return material
}

var defaultMsdfShader *Shader // Can set this to whatever you want

func SetDefaultMsdfShader(shader *Shader) {
	defaultSpriteShader = shader
}

func GetDefaultMsdfShader() *Shader {
	if defaultMsdfShader != nil {
		return defaultMsdfShader
	}

	// Note: We snuff the error here. If the user wants they can pre-supply a defaultmsdfshader so this one never loads
	var err error
	defaultMsdfShader, err = NewShader(shaders.MSDFShader)
	if err != nil {
		panic(err)
	}
	return defaultMsdfShader
}

func DefaultMsdfMaterial(texture *Texture) Material {
	material := NewMaterial(GetDefaultMsdfShader())
	material.texture = texture
	material.
		SetUniform("u_threshold", 0.5).
		SetUniform("u_outline_width_relative", 0.1).
		SetUniform("u_outline_blur", 0.0).
		SetUniform("u_outline_color", RGBA{0, 0, 0, 1})
	return material
}
