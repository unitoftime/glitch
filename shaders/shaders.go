package shaders

import (
	"github.com/jstewart7/glitch"
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
