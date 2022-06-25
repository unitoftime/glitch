package glitch

var (
	White = RGBA{1, 1, 1, 1}
	Black = RGBA{0, 0, 0, 1}
)

// Premultipled RGBA value scaled from [0, 1.0]
type RGBA struct {
	R,G,B,A float32
}

// TODO - conversion from golang colors
func FromUint8(r, g, b, a uint8) RGBA {
	return RGBA{
		float32(r)/255.0,
		float32(g)/255.0,
		float32(b)/255.0,
		float32(a)/255.0,
	}
}
