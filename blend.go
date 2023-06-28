package glitch

import (
	"github.com/unitoftime/glitch/internal/gl"
)

// https://registry.khronos.org/OpenGL-Refpages/gl4/html/glBlendFunc.xhtml
type BlendMode struct {
	src, dst gl.Enum
}
var BlendModeNormal = BlendMode{
	// Note: This is what I used before premult: gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA
	gl.ONE,
	gl.ONE_MINUS_SRC_ALPHA,
}

// // TODO: Untested
// var BlendModeAdd = BlendMode{
// gl.SRC_ALPHA, gl.ONE
// }

var BlendModeMultiply = BlendMode{
	gl.DST_COLOR, gl.ZERO,
}
