package shaders

import (
	"github.com/unitoftime/glitch"
	_ "embed"
)

//go:embed sprite.vs
var SpriteVertexShader string;

//go:embed sprite.fs
var SpriteFragmentShader string;

var SpriteShader = glitch.ShaderConfig{
	VertexShader: SpriteVertexShader,
	FragmentShader: SpriteFragmentShader,
	VertexFormat: glitch.VertexFormat{
		glitch.Attrib{"aPos", glitch.AttrVec3},
		glitch.Attrib{"aColor", glitch.AttrVec4},
		glitch.Attrib{"aTexCoord", glitch.AttrVec2},
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attrib{"projection", glitch.AttrMat4},
		glitch.Attrib{"view", glitch.AttrMat4},
	},
}

//go:embed subPixel.fs
var SubPixelAntiAliased string;

var PixelArtShader = glitch.ShaderConfig{
	VertexShader: SpriteVertexShader,
	FragmentShader: SubPixelAntiAliased,
	VertexFormat: glitch.VertexFormat{
		glitch.Attrib{"aPos", glitch.AttrVec3},
		glitch.Attrib{"aColor", glitch.AttrVec4},
		glitch.Attrib{"aTexCoord", glitch.AttrVec2},
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attrib{"projection", glitch.AttrMat4},
		glitch.Attrib{"view", glitch.AttrMat4},
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
		glitch.Attrib{"position", glitch.AttrVec3},
		// glitch.Attrib{"normal", glitch.AttrVec3},
		glitch.Attrib{"color", glitch.AttrVec4},
		glitch.Attrib{"texture", glitch.AttrVec2},
	},
	UniformFormat: glitch.UniformFormat{
		glitch.Attrib{"model", glitch.AttrMat4},
		glitch.Attrib{"view", glitch.AttrMat4},
		glitch.Attrib{"projection", glitch.AttrMat4},
		// glitch.Attrib{"dirlight.direction", glitch.AttrVec3},
		// glitch.Attrib{"dirlight.ambient", glitch.AttrVec3},
		// glitch.Attrib{"dirlight.diffuse", glitch.AttrVec3},
		// glitch.Attrib{"dirlight.specular", glitch.AttrVec3},
	},
}
