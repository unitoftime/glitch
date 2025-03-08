package glitch

import "github.com/unitoftime/glitch/internal/gl"

// Note: These are all packed into uint8s to reduce size of the Material object
type BlendMode uint8

const (
	BlendModeNone BlendMode = iota
	BlendModeNormal
	BlendModeMultiply
)

type blendModeData struct {
	src, dst gl.Enum
}

var blendModeLut []blendModeData = []blendModeData{
	// Note: This is what I used before premult: gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA
	BlendModeNormal: {gl.ONE, gl.ONE_MINUS_SRC_ALPHA},
	// BlendModeNormal: {gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA},
	// BlendModeNormal: {gl.SRC_ALPHA, gl.ONE},
	BlendModeMultiply: {gl.DST_COLOR, gl.ZERO},
}

type DepthMode uint8

const (
	DepthModeNone DepthMode = iota
	DepthModeLess
	DepthModeLequal
)

type depthModeData struct {
	mode gl.Enum
}

var depthModeLut []depthModeData = []depthModeData{
	DepthModeNone:   {},
	DepthModeLess:   {gl.LESS},
	DepthModeLequal: {gl.LEQUAL},
}

type CullMode uint8

const (
	CullModeNone CullMode = iota
	CullModeNormal
)

type cullModeData struct {
	face, dir gl.Enum
}

var cullModeLut []cullModeData = []cullModeData{
	CullModeNone:   {},
	CullModeNormal: {gl.BACK, gl.CCW},
}

// // https://registry.khronos.org/OpenGL-Refpages/gl4/html/glBlendFunc.xhtml
// type BlendMode struct {
// 	src, dst gl.Enum
// }
// var BlendModeNormal = BlendMode{
// 	// Note: This is what I used before premult: gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA
// 	gl.ONE,
// 	gl.ONE_MINUS_SRC_ALPHA,
// }

// // // TODO: Untested
// // var BlendModeAdd = BlendMode{
// // gl.SRC_ALPHA, gl.ONE
// // }

// var BlendModeMultiply = BlendMode{
// 	gl.DST_COLOR, gl.ZERO,
// }

// type DepthMode struct {
// 	mode gl.Enum
// }
// var ( // TODO: Can these be constants? Will that break wasm?
// 	DepthModeNone = DepthMode{}
// 	DepthModeLess = DepthMode{gl.LESS}
// 	DepthModeLequal = DepthMode{gl.LEQUAL}
// )

// type CullMode struct {
// 	face gl.Enum
// 	dir gl.Enum
// }
// var ( // TODO: Can these be constants? Will that break wasm?
// 	CullModeNone = CullMode{}
// 	CullModeNormal = CullMode{gl.BACK, gl.CCW}
// )

// // https://registry.khronos.org/OpenGL-Refpages/gl4/html/glBlendFunc.xhtml
// type BlendMode struct {
// 	src, dst gl.Enum
// }
// var BlendModeNormal = BlendMode{
// 	// Note: This is what I used before premult: gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA
// 	gl.ONE,
// 	gl.ONE_MINUS_SRC_ALPHA,
// }

// // // TODO: Untested
// // var BlendModeAdd = BlendMode{
// // gl.SRC_ALPHA, gl.ONE
// // }

// var BlendModeMultiply = BlendMode{
// 	gl.DST_COLOR, gl.ZERO,
// }

// type DepthMode struct {
// 	mode gl.Enum
// }
// var ( // TODO: Can these be constants? Will that break wasm?
// 	DepthModeNone = DepthMode{}
// 	DepthModeLess = DepthMode{gl.LESS}
// 	DepthModeLequal = DepthMode{gl.LEQUAL}
// )

// type CullMode struct {
// 	face gl.Enum
// 	dir gl.Enum
// }
// var ( // TODO: Can these be constants? Will that break wasm?
// 	CullModeNone = CullMode{}
// 	CullModeNormal = CullMode{gl.BACK, gl.CCW}
// )
