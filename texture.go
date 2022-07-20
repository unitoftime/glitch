package glitch

import (
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

func NewTexture(img image.Image, smooth bool) *Texture {
// func NewTexture(width, height int, pixels []uint8) *Texture {
	nrgba := image.NewNRGBA(img.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)

	width := nrgba.Bounds().Dx()
	height := nrgba.Bounds().Dy()
	pixels := nrgba.Pix
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
