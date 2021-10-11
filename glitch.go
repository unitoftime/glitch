package glitch

import (
	// "github.com/jstewart7/gl"
	"github.com/faiface/mainthread"
)

//type Attrib gl.Attrib
//type Program gl.Program
//type Buffer gl.Buffer
//type Framebuffer gl.Framebuffer
//type Renderbuffer gl.Renderbuffer
//type Texture gl.Texture
//type Uniform gl.Uniform

func Run(function func()) {
	mainthread.Run(function)
}
