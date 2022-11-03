package glitch

import (
	// "fmt"
	"image"
	"image/draw"
	"image/color"
	"runtime"
	"github.com/faiface/mainthread"
	"github.com/unitoftime/gl"
)

// TODO - Should I use this as default? Or is there a way to do null textures for textureless things?
var whiteTexture *Texture
func WhiteTexture() *Texture {
	if whiteTexture != nil { return whiteTexture }
	max := 128 // TODO - webgl forces textures to be power of 2 - maybe I can go smaller though
	img := image.NewNRGBA(image.Rect(0,0,max,max))

	col := uint8(255)
	for x:=0; x<max; x++ {
		for y:=0; y<max; y++ {
			img.SetNRGBA(x,y, color.NRGBA{col,col,col, 255})
		}
	}

	whiteTexture = NewTexture(img, true)
	return whiteTexture
}

type Texture struct {
	texture gl.Texture
	width, height int
}

// func NewTexture(img image.Image, smooth bool) *Texture {
// 	nrgba := image.NewNRGBA(img.Bounds())
// 	draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)

// 	width := nrgba.Bounds().Dx()
// 	height := nrgba.Bounds().Dy()
// 	pixels := nrgba.Pix
// 	t := &Texture{
// 		width: width,
// 		height: height,
// 	}

// 	mainthread.Call(func() {
// 		t.texture = gl.CreateTexture()
// 		gl.BindTexture(gl.TEXTURE_2D, t.texture)

// 		gl.TexImage2D(gl.TEXTURE_2D, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, pixels)

// 		// TODO - webgl doesn't support CLAMP_TO_BORDER
// 		// GL_CLAMP_TO_EDGE: The coordinate will simply be clamped between 0 and 1.
// 		// GL_CLAMP_TO_BORDER: The coordinates that fall outside the range will be given a specified border color.

// 		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
// 		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

// 		// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
// 		// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

// 		// TODO - pass smooth in as a parameter
// 		if smooth {
// 			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
// 			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
// 		} else {
// 			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
// 			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
// 		}
// 	})

// 	runtime.SetFinalizer(t, (*Texture).delete)

// 	return t
// }
func NewTexture(img image.Image, smooth bool) *Texture {
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)

	// Premultiply
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			// fmt.Println(img.At(x,y).RGBA())
			r,g,b,a := img.At(x,y).RGBA()
			// fmt.Println(r, g, b, a)
			rf := float64(r) / 256.0
			gf := float64(g) / 256.0
			bf := float64(b) / 256.0
			af := float64(a) / 256.0
			// fmt.Println(rf, gf, bf, af)
			premult := color.RGBA{uint8(rf * af), uint8(gf * af), uint8(bf * af), uint8(af)}
			// fmt.Println(premult)
			rgba.SetRGBA(x, y, premult)
		}
	}

	width := rgba.Bounds().Dx()
	height := rgba.Bounds().Dy()
	pixels := rgba.Pix
	t := &Texture{
		width: width,
		height: height,
	}

	mainthread.Call(func() {
		t.texture = gl.CreateTexture()
		gl.BindTexture(gl.TEXTURE_2D, t.texture)

		gl.TexImage2D(gl.TEXTURE_2D, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, pixels)

		// TODO - webgl doesn't support CLAMP_TO_BORDER
		// GL_CLAMP_TO_EDGE: The coordinate will simply be clamped between 0 and 1.
		// GL_CLAMP_TO_BORDER: The coordinates that fall outside the range will be given a specified border color.

		// TODO - Do I need this all the time? Probably not, especially when I'm doing texture atlases
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

		// TODO - pass smooth in as a parameter
		if smooth {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		} else {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		}
	})

	runtime.SetFinalizer(t, (*Texture).delete)

	return t
}

// Sets the texture to be this image.
// Texture size must match img size or this will panic!
// TODO - Should I just try and set it? or do nothing?
func (t *Texture) SetImage(img image.Image) {
	if img == nil { return }

	if t.width != img.Bounds().Dx() || t.height != img.Bounds().Dy() {
		panic("SetImage: img bounds are not equal to texture bounds!")
	}

	// TODO - skip this if already an NRGBA
	nrgba := image.NewNRGBA(img.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)

	pixels := nrgba.Pix
	t.SetPixels(0, 0, t.width, t.height, pixels)
}

// Sets the pixels of a section of a texture
func (t *Texture) SetPixels(x, y, w, h int, pixels []uint8) {
	if len(pixels) != w*h*4 {
		panic("set pixels: wrong number of pixels")
	}
	t.Bind(0)

	mainthread.Call(func() {
		gl.TexSubImage2D(
			gl.TEXTURE_2D,
			0,
			x,
			y,
			w,
			h,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			pixels,
		)
	})
}

func (t *Texture) Bounds() Rect {
	return R(0, 0, float32(t.width), float32(t.height))
}

func (t *Texture) Bind(position int) {
	mainthread.Call(func() {
		gl.ActiveTexture(gl.TEXTURE0);
		// gl.ActiveTexture(gl.TEXTURE0 + position); // TODO - include position
		gl.BindTexture(gl.TEXTURE_2D, t.texture)
	})
}

func (t *Texture) delete() {
	mainthread.CallNonBlock(func() {
		gl.DeleteTexture(t.texture)
	})
}
