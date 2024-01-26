package glitch

import (
	"image"
	"image/color"
	"image/draw"
	"runtime"

	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/mainthread"
)

// TODO - Should I use this as default? Or is there a way to do null textures for textureless things?
var whiteTexture *Texture
func WhiteTexture() *Texture {
	if whiteTexture != nil { return whiteTexture }
	max := 128 // TODO - webgl forces textures to be power of 2 - maybe I can go smaller though

	whiteColor := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	whiteTexture = NewRGBATexture(max, max, whiteColor, true)
	return whiteTexture
}

func NewRGBATexture(width, height int, col color.RGBA, smooth bool) *Texture {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for x:=0; x<width; x++ {
		for y:=0; y<height; y++ {
			img.SetRGBA(x,y, col)
		}
	}
	return NewTexture(img, smooth)
}

func NewTransparentTexture64() *Texture {
	max := 64 // TODO - webgl forces textures to be power of 2 - maybe I can go smaller though
	img := image.NewRGBA(image.Rect(0,0,max,max))

	for x:=0; x<max; x++ {
		for y:=0; y<max; y++ {
			img.SetRGBA(x,y, color.RGBA{255,255,255, 255})
			// img.SetRGBA(x,y, color.RGBA{0, 0, 0, 0})
		}
	}

	return NewTexture(img, true)
}

type Texture struct {
	texture gl.Texture
	width, height int
	smooth bool
}

func toRgba(img image.Image) *image.RGBA {
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
	return rgba
}

func NewEmptyTexture(width, height int, smooth bool) *Texture {
	// We can only use RGBA images right now.
	t := &Texture{
		width: width,
		height: height,
		smooth: smooth,
	}

	t.initialize(nil)
	return t
}

func NewTexture(img image.Image, smooth bool) *Texture {
	// We can only use RGBA images right now.
	rgba := toRgba(img)

	width := rgba.Bounds().Dx()
	height := rgba.Bounds().Dy()
	t := &Texture{
		width: width,
		height: height,
		smooth: smooth,
	}

	t.initialize(rgba.Pix)

	return t
}

func (t *Texture) initialize(pixels []uint8) {
	mainthread.Call(func() {
		t.texture = gl.CreateTexture()
		gl.BindTexture(gl.TEXTURE_2D, t.texture)

		gl.TexImage2DFull(gl.TEXTURE_2D, 0, gl.RGBA, t.width, t.height, gl.RGBA, gl.UNSIGNED_BYTE, pixels)

		// TODO - webgl doesn't support CLAMP_TO_BORDER
		// GL_CLAMP_TO_EDGE: The coordinate will simply be clamped between 0 and 1.
		// GL_CLAMP_TO_BORDER: The coordinates that fall outside the range will be given a specified border color.

		// TODO - Do I need this all the time? Probably not, especially when I'm doing texture atlases
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

		// TODO - pass smooth in as a parameter
		if t.smooth {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		} else {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		}
	})

	runtime.SetFinalizer(t, (*Texture).delete)
}

// TODO: This needs to be combined into the NewTexture function, or this needs to bind the texture
func (t *Texture) GenerateMipmap() {
	mainthread.Call(func() {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
		gl.GenerateMipmap(gl.TEXTURE_2D);
	})
}

// Sets the texture to be this image.
// Texture size must match img size or this will panic!
// TODO - Should I just try and set it? or do nothing?
func (t *Texture) SetImage(img image.Image) {
	if img == nil { return }

	if t.width != img.Bounds().Dx() || t.height != img.Bounds().Dy() {
		panic("SetImage: img bounds are not equal to texture bounds!")
	}

	// // TODO - skip this if already an NRGBA
	// nrgba := image.NewNRGBA(img.Bounds())
	// draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)

	// pixels := nrgba.Pix

	rgba := toRgba(img)
	pixels := rgba.Pix
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
	return R(0, 0, float64(t.width), float64(t.height))
}

func (t *Texture) Bind(position int) {
	// TODO - maybe allow for more than 15 if the platform supports it? TODO - max texture units
	state.bindTexture(t)
}

func (t *Texture) delete() {
	mainthread.CallNonBlock(func() {
		gl.DeleteTexture(t.texture)
	})
}
