package glitch

import (
	"math"
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
		float64(r) / float64(math.MaxUint8),
		float64(g) / float64(math.MaxUint8),
		float64(b) / float64(math.MaxUint8),
		float64(a) / float64(math.MaxUint8),
	}
}

func Alpha(a float64) RGBA {
	return RGBA{a, a, a, a}
}

func FromRGBA(c color.RGBA) RGBA {
	return FromUint8(c.R, c.G, c.B, c.A)
}

func FromColor(c color.Color) RGBA {
	r, g, b, a := c.RGBA()

	return RGBA{
		float64(r) / float64(math.MaxUint16),
		float64(g) / float64(math.MaxUint16),
		float64(b) / float64(math.MaxUint16),
		float64(a) / float64(math.MaxUint16),
	}
}
