package glitch

import (
	"image/color"
	"math"
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

func Greyscale(g float64) RGBA {
	return RGBA{g, g, g, 1.0}
}

func FromStraightRGBA(r, g, b float64, a float64) RGBA {
	return RGBA{r * a, g * a, b * a, a}
}

func FromNRGBA(c color.NRGBA) RGBA {
	r, g, b, a := c.RGBA()

	return RGBA{
		float64(r) / float64(math.MaxUint16),
		float64(g) / float64(math.MaxUint16),
		float64(b) / float64(math.MaxUint16),
		float64(a) / float64(math.MaxUint16),
	}
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
func (c1 RGBA) Mult(c2 RGBA) RGBA {
	return RGBA{
		c1.R * c2.R,
		c1.G * c2.G,
		c1.B * c2.B,
		c1.A * c2.A,
	}
}

func (c RGBA) gl() glVec4 {
	return glVec4{float32(c.R), float32(c.G), float32(c.B), float32(c.A)}
}
