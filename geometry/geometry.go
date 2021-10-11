package geometry

// import (
// 	// "github.com/go-gl/mathgl/mgl32"
// 	"github.com/jstewart7/glitch"
// )

// type Geometry struct {
// 	Indices    []uint32
// 	Positions  []float32
// 	// Normals []float32
// 	TexCoords  []float32
// }



// func Quad() Geometry {
// 	return Geometry{
// 		Indices: []uint32{
// 			0, 1, 2,
// 			1, 2, 3,
// 		},
// 		Positions: []float32{
// 			1.0, 1.0, 0.0, // Top Right
// 			1.0, 0.0, 0.0, // Bot Right
// 			0.0, 0.0, 0.0, // Bot Left
// 			0.0, 1.0, 0.0, // Top Left
// 		},
// 		// Notably, texture coords are flipped
// 		TexCoords: []float32{
// 			1.0, 0.0, // Top Right
// 			1.0, 1.0, // Bot Right
// 			0.0, 1.0, // Bot Left
// 			0.0, 0.0, // Top Left
// 		},
// 	}
// }
