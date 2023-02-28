package glitch

import (
	// "os"
	// "image/png"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"

	"golang.org/x/image/math/fixed"

	"unicode"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

func DefaultAtlas() (*Atlas, error) {
	runes := make([]rune, unicode.MaxASCII - 32)
	for i := range runes {
		runes[i] = rune(32 + i)
	}

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}
	fontFace := truetype.NewFace(font, &truetype.Options{
		Size: 64,
		// GlyphCacheEntries: 1,
	})
	atlas := NewAtlas(fontFace, runes, true, 0)
	return atlas, nil
}

type Glyph struct {
	Advance float64
	Bearing Vec2
	BoundsUV Rect
}

// TODO - instead of creating a single atlas ahead of time. I should just load the font and then dynamically create the atlas as needed. This should probably change once you add automatic texture batching.
type Atlas struct {
	face font.Face
	mapping map[rune]Glyph
	ascent fixed.Int26_6 // Distance from top of line to baseline
	descent fixed.Int26_6 // Distance from bottom of line to baseline
	lineGap fixed.Int26_6 // The recommended gap between two lines
	texture *Texture
	border int // Specifies a border on the font.
	pixelPerfect bool // if true anti-aliasing will be disabled
}

func fixedToFloat(val fixed.Int26_6) float64 {
	// Shift to the left by 6 then convert to an int, then to a float, then shift right by 6
	// TODO - How to handle overruns?
	// intVal := val.Mul(fixed.I(1_000_000)).Floor()
	// return float32(intVal) / 1_000_000.0
	return float64(val) / (1 << 6)
}

func NewAtlas(face font.Face, runes []rune, smooth bool, border int) *Atlas {
	metrics := face.Metrics()
	atlas := &Atlas{
		face: face,
		mapping: make(map[rune]Glyph),
		ascent: metrics.Ascent,
		descent: metrics.Descent,
		lineGap: metrics.Height,
		border: int(border),
		pixelPerfect: !smooth, // TODO - not sure this is exactly right. You could presumably want a bilinear filtered texture but anti-aliasing turned off on the text.
	}

	size := 512
	fixedSize := fixed.I(size)
	fSize := float64(size)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// draw.Draw(img, img.Bounds(), image.NewUniform(color.Alpha{0}), image.ZP, draw.Src)
	// Note: In case you want to see the boundary of each rune, uncomment this
	// draw.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)

	padding := fixed.I(2 + (2 * atlas.border)) // Padding for runes drawn to atlas
	startDot := fixed.P(padding.Floor(), (atlas.ascent + padding).Floor()) // Starting point of the dot
	dot := startDot
	for i, r := range runes {
		bounds, mask, maskp, adv, ok := face.Glyph(dot, r)
		if !ok { panic("Missing rune!") }
		bearingRect, _, _ := face.GlyphBounds(r)

		// if r == 'R' {
		// 	fmt.Printf("%T\n", mask)
		// 	// fmt.Println(mask)
		// 	outputFile, err := os.Create("testR.png")
		// 	if err != nil { panic(err) }
		// 	png.Encode(outputFile, mask)
		// 	outputFile.Close()
		// }

		// Instead of flooring we convert from fixed int to float manually (mult by 10^6 then floor, cast and divide by 10^6). I think this is slightly more accurate but it's hard to tell so I'll leave old code below
		//		log.Println("Rune: ", string(r), " - BearingRect: ", bearingRect)
		bearingX := float64((bearingRect.Min.X * 1000000).Floor()) / (1000000 * fSize)
		bearingY := float64((-bearingRect.Max.Y * 1000000).Floor()) / (1000000 * fSize)
		//		advance := float32((adv * 1000000).Floor())/(1000000 * fSize) // TODO - why doesn't this work?
		// log.Println("Rune: ", string(r), " - BearingX: ", float32(bearingRect.Min.X.Floor())/fSize)
		// log.Println("Rune: ", string(r), " - BearingX: ", bearingX)
		// log.Println("Rune: ", string(r), " - BearingY: ", float32(-bearingRect.Max.Y.Floor())/fSize)
		// log.Println("Rune: ", string(r), " - BearingY: ", bearingY)

		draw.Draw(img, bounds, mask, maskp, draw.Src)
		atlas.mapping[r] = Glyph{
			Advance: float64(adv.Floor())/fSize,
			//			Bearing: Vec2{float32(bearingRect.Min.X.Floor())/fSize, float32((-bearingRect.Max.Y).Floor())/fSize},
			//Advance: advance,
			Bearing: Vec2{bearingX, bearingY},
			BoundsUV: R(
				float64(bounds.Min.X - atlas.border)/fSize, float64(bounds.Min.Y - atlas.border)/fSize,
				float64(bounds.Max.X + atlas.border)/fSize, float64(bounds.Max.Y + atlas.border)/fSize),
		}

		// Usual next dot location
		nextDotX := dot.X + adv + padding
		nextDotY := dot.Y

		// Exit if we are at the end
		if (i+1) >= len(runes) { break }

		// If the rune after this one pushes us too far then loop around
		nextAdv, ok := face.GlyphAdvance(runes[i+1])
		if !ok { panic("Missing rune!") }
		if nextDotX + nextAdv >= fixedSize {
			// log.Println("Ascending!")
			nextDotX = startDot.X
			nextDotY = dot.Y + atlas.ascent + padding
		}
		// log.Println(nextDotX, nextDotY)
		dot = fixed.Point26_6{nextDotX, nextDotY}
	}

	// This just disables anti-aliasing by snapping pixels to either white or transparent
	if atlas.pixelPerfect {
		imgBounds := img.Bounds()
		for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
			for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
				rgba := img.RGBAAt(x, y)
				if rgba.A > 0 {
					rgba.A = 255
					img.Set(x, y, color.White)
				}
			}
		}
	}

	// TODO - border messes up glyph advance I think. Makes kerning and lineheight different
	// This runs a box filter based on the border side
	if atlas.border != 0 {
		imgBounds := img.Bounds()
		for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
			for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
				rgba := img.RGBAAt(x, y)
				if (rgba != color.RGBA{255, 255, 255, 255}) {
					continue // If the pixel is transparent then it doesn't trigger a border
				}

				box := image.Rect(x-atlas.border, y-atlas.border, x+atlas.border, y+atlas.border)
				for xx := box.Min.X; xx <= box.Max.X; xx++ {
					for yy := box.Min.Y; yy <= box.Max.Y; yy++ {
						rgba := img.RGBAAt(xx, yy)
						if rgba.A == 0 {
							// Only add a border to transparent pixels
							img.Set(xx, yy, color.Black)
						}
					}
				}
			}
		}
	}

	// // outputFile is a File type which satisfies Writer interface
	// outputFile, err := os.Create("test.png")
	// if err != nil { panic(err) }
	// png.Encode(outputFile, img)
	// outputFile.Close()

	atlas.texture = NewTexture(img, smooth)
	// fmt.Println("TextAtlas: ", atlas.texture.width, atlas.texture.height)
	return atlas
}

func (a *Atlas) LineHeight() float64 {
	// TODO - scale?
	// TODO - I'm not sure why, but I have to multiply by 2 here
	return 2 * (-fixedToFloat(a.ascent) + fixedToFloat(a.descent) - fixedToFloat(a.lineGap))
}

func (a *Atlas) RuneVerts(mesh *Mesh, r rune, dot Vec2, scale float64) (Vec2, float64) {
	// multiplying by texture sizes converts from UV to pixel coords
	scaleX := scale * float64(a.texture.width)
	scaleY := scale * float64(a.texture.height)

	glyph, ok := a.mapping[r]
	// if !ok { panic(fmt.Sprintf("Missing Rune: %v", r)) }
	if !ok {
		// fmt.Printf("Missing Rune: %v", r)
		// Replace rune with '?'
		oldR := r
		r = '?' // TODO - Pick some other rune. TODO - require this rune to be in the atlas!
		glyph, ok = a.mapping[r]
		if !ok {
			panic(fmt.Sprintf("Missing Rune: %v and replacement%v", oldR, r))
		}
	}

	//	log.Println(glyph.Bearing)

	// UV coordinates of the quad
	u1 := glyph.BoundsUV.Min[0]
	u2 := glyph.BoundsUV.Max[0]
	v1 := glyph.BoundsUV.Min[1]
	v2 := glyph.BoundsUV.Max[1]

	// Pixel coordinates of the quad (scaled by scale)
	x1 := dot[0] + (scaleX * glyph.Bearing[0])
	x2 := x1 + (scaleX * (u2 - u1))

	// Note: Commented out the downard shift here, and I'm doing it in the above func
	y1 := dot[1] + (scaleY * glyph.Bearing[1])// - (2*fixedToFloat(a.descent))
	y2 := y1 + (scaleY * (v2 - v1))

	mesh.AppendQuadMesh(R(x1, y1, x2, y2), R(u1, v1, u2, v2))
	// mesh := NewQuadMesh(R(x1, y1, x2, y2), R(u1, v1, u2, v2))

	dot[0] += (scaleX * glyph.Advance)

	return dot, y2
}

func (a *Atlas) Text(str string) *Text {
	t := &Text{
		currentString: "",
		atlas: a,
		texture: a.texture,
		material: NewSpriteMaterial(a.texture),
		scale: 1.0,
		LineHeight: a.LineHeight(),
		mesh: NewMesh(),
		tmpMesh: NewMesh(),

		Color: RGBA{1, 1, 1, 1},
	}

	t.Set(str)

	return t
}

type Text struct {
	currentString string
	mesh *Mesh
	tmpMesh *Mesh // For temporarily buffering data. TODO - would be more efficient just to append the quads directly to the mesh rather than buffering them here
	atlas *Atlas
	bounds Rect
	texture *Texture
	material Material
	scale float64 // TODO - is this useful? Play around with different scaling methods.
	LineHeight float64

	Orig Vec2 // The baseline starting point from which to draw the text
	Dot Vec2 // The location of the next rune to draw
	Color RGBA // The color with which to draw the next text
}

func (t *Text) Bounds() Rect {
	return t.bounds
}

func (t *Text) MeshBounds() Rect {
	return t.mesh.Bounds().Rect()
}

func (t *Text) Clear() {
	t.Orig = Vec2{}
	t.Dot = t.Orig
}

// This resets the text and sets it to the passed in string (if the passed in string is different!)
// TODO - I need to deprecate this in favor of a better interface
func (t *Text) Set(str string) {
	if t.currentString != str {
		t.currentString = str
		t.mesh.Clear()
		t.bounds = t.AppendStringVerts(str)
	}
}

// This appends the list of bytes onto the end of the string
// Note: implements io.Writer interface
func (t *Text) Write(p []byte) (n int, err error) {
	appendedStr := string(p)

	if t.mesh == nil {
		t.Set(appendedStr)
		return len(p), nil
	}

	t.currentString = t.currentString + appendedStr
	newBounds := t.AppendStringVerts(appendedStr)
	t.bounds = t.bounds.Union(newBounds)
	return len(p), nil
}

func (t *Text) Draw(target BatchTarget, matrix Mat4) {
	target.Add(t.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, t.material)
}

func (t *Text) DrawColorMask(target BatchTarget, matrix Mat4, color RGBA) {
	target.Add(t.mesh, matrix, color, t.material)
}

func (t *Text) DrawRect(target BatchTarget, rect Rect, color RGBA) {
	mat := Mat4Ident
	mat.Scale(1.0, 1.0, 1.0).Translate(rect.Min[0], rect.Min[1], 0)
	target.Add(t.mesh, mat, color, t.material)
}

func (t *Text) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	mat := Mat4Ident
	// TODO why shouldn't I be shifting to the middle?
	// mat.Scale(bounds.W() / t.bounds.W(), bounds.H() / t.bounds.H(), 1).Translate(bounds.W()/2 + bounds.Min[0], bounds.H()/2 + bounds.Min[1], 0)
	// mat.Scale(1.0, 1.0, 1.0).Translate(rect.Min[0], rect.Min[1], 0)

	// TODO!!! - There's something wrong with this
	mat.Scale(bounds.W() / t.bounds.W(), bounds.H() / t.bounds.H(), 1).Translate(bounds.Min[0], bounds.Min[1], 0)

	target.Add(t.mesh, mat, mask, t.material)
}

func (t *Text) AppendStringVerts(text string) Rect {
	// maxAscent := float32(0) // Tracks the maximum y point of the text block

	// initialDot := t.Dot

	for _,r := range text {
		// If the rune is a newline, then we need to reset the dot for the next line
		if r == '\n' {
			t.Dot[1] -= t.atlas.LineHeight()/2
			t.Dot[0] = t.Orig[0]
			continue
		}

		// runeMesh, newDot, ascent := a.RuneVerts(r, *dot, scale)
		// fmt.Println("dot", *dot)
		newDot, _ := t.atlas.RuneVerts(t.mesh, r, t.Dot, t.scale)
		// t.tmpMesh.SetColor(t.Color) // TODO - enable this back

		t.Dot = newDot

		// if maxAscent < ascent {
		// 	maxAscent = dot[1] + ascent
		// }
	}
	// return mesh, R(initialDot[0], initialDot[1], dot[0], dot[1] + maxAscent)

	// fmt.Println("-----")
	// fmt.Println(fixedToFloat(a.ascent))
	// fmt.Println(fixedToFloat(a.descent))
	// fmt.Println(fixedToFloat(a.lineGap))
	// fmt.Println(maxAscent)
	// fmt.Println(scale)
	// bounds := R(initialDot[0], initialDot[1], dot[0], dot[1] - a.LineHeight())

	// bounds := R(initialDot[0],
	// 	initialDot[1] - (2 * fixedToFloat(a.ascent)),
	// 	dot[0], // TODO - this is wrong if because this is the length of the last line, we need the length of the longest line
	// 	dot[1] - (2 * fixedToFloat(a.descent)))

	// TODO - this used the glyphs to determine bounds, below I use the mesh
	// // TODO - idk what I'm doing here, but it seems to work. Man text rendering is hard.
	// bounds := R(initialDot[0],
	// 	initialDot[1] - (fixedToFloat(t.atlas.ascent)),
	// 	t.Dot[0], // TODO - this is wrong if because this is the length of the last line, we need the length of the longest line
	// 	t.Dot[1] - (fixedToFloat(t.atlas.descent))).
	// 		Norm().
	// 		Moved(Vec2{0, fixedToFloat(t.atlas.ascent)})
	// return bounds

	// Attempt 2 - Use mesh bounds
	return t.mesh.Bounds().Rect()



	// fmt.Println(bounds)

	// bounds := R(initialDot[0],
	// 	initialDot[1] - (fixedToFloat(a.descent)),
	// 	dot[0], // TODO - this is wrong if because this is the length of the last line, we need the length of the longest line
	// 	dot[1] - (fixedToFloat(a.ascent))).Norm()
	// return mesh, bounds

	// return mesh, R(initialDot[0], initialDot[1], dot[0], dot[1] + fixedToFloat(a.lineHeight))
	// return mesh, R(initialDot[0], initialDot[1], dot[0], dot[1] - (fixedToFloat(a.ascent) - fixedToFloat(a.descent)))
	// return mesh, R(initialDot[0],
	// 	initialDot[1] - fixedToFloat(a.descent)/1024,
	// 	dot[0], // TODO - this is wrong if because this is the length of the last line, we need the length of the longest line
	// 	dot[1] + fixedToFloat(a.ascent)/1024)

}
