package glitch

import (
	"image/color"
)

var (
	White = RGBA{1, 1, 1, 1}
	Black = RGBA{0, 0, 0, 1}
)

// Premultipled RGBA value scaled from [0, 1.0]
type RGBA struct {
	R,G,B,A float64
}

// TODO - conversion from golang colors
func FromUint8(r, g, b, a uint8) RGBA {
	return RGBA{
		float64(r)/255.0,
		float64(g)/255.0,
		float64(b)/255.0,
		float64(a)/255.0,
	}
}

func FromNRGBA(c color.NRGBA) RGBA {
	// TODO!!!!!! - premultiply alpha bug?
	return FromUint8(c.R, c.G, c.B, c.A)
}
