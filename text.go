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
)

type Glyph struct {
	Advance float32
	Bearing Vec2
	BoundsUV Rect
}

type Atlas struct {
	face font.Face
	mapping map[rune]Glyph
	ascent fixed.Int26_6
	descent fixed.Int26_6
	lineHeight fixed.Int26_6
	texture *Texture
}

func NewAtlas(face font.Face, runes []rune) *Atlas {
	smooth := false // TODO - Should fonts always be smoothed?

	metrics := face.Metrics()
	atlas := &Atlas{
		face: face,
		mapping: make(map[rune]Glyph),
		ascent: metrics.Ascent,
		descent: metrics.Descent,
		lineHeight: metrics.Height,
	}

	size := 512
	fixedSize := fixed.I(size)
	fSize := float32(size)

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)

	padding := fixed.I(2) // Padding for runes drawn to atlas
	startDot := fixed.P(0, (atlas.ascent + padding).Floor()) // Starting point of the dot
	dot := startDot
	for i, r := range runes {
		bounds, mask, maskp, adv, ok := face.Glyph(dot, r)
		if !ok { panic("Missing rune!") }
		bearingRect, _, _ := face.GlyphBounds(r)

		// Instead of flooring we convert from fixed int to float manually (mult by 10^6 then floor, cast and divide by 10^6). I think this is slightly more accurate but it's hard to tell so I'll leave old code below
		//		log.Println("Rune: ", string(r), " - BearingRect: ", bearingRect)
		bearingX := float32((bearingRect.Min.X * 1000000).Floor())/(1000000 * fSize)
		bearingY := float32((-bearingRect.Max.Y * 1000000).Floor())/(1000000 * fSize)
		//		advance := float32((adv * 1000000).Floor())/(1000000 * fSize) // TODO - why doesn't this work?
		// log.Println("Rune: ", string(r), " - BearingX: ", float32(bearingRect.Min.X.Floor())/fSize)
		// log.Println("Rune: ", string(r), " - BearingX: ", bearingX)
		// log.Println("Rune: ", string(r), " - BearingY: ", float32(-bearingRect.Max.Y.Floor())/fSize)
		// log.Println("Rune: ", string(r), " - BearingY: ", bearingY)

		draw.Draw(img, bounds, mask, maskp, draw.Src)
		atlas.mapping[r] = Glyph{
			Advance: float32(adv.Floor())/fSize,
			//			Bearing: Vec2{float32(bearingRect.Min.X.Floor())/fSize, float32((-bearingRect.Max.Y).Floor())/fSize},
			//Advance: advance,
			Bearing: Vec2{bearingX, bearingY},
			BoundsUV: R(
				float32(bounds.Min.X)/fSize, float32(bounds.Min.Y)/fSize,
				float32(bounds.Max.X)/fSize, float32(bounds.Max.Y)/fSize),
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

	// outputFile is a File type which satisfies Writer interface
	// outputFile, err := os.Create("test.png")
	// if err != nil { panic(err) }
	// png.Encode(outputFile, img)
	// outputFile.Close()

	atlas.texture = NewTexture(img, smooth)
	fmt.Println("TextAtlas: ", atlas.texture.width, atlas.texture.height)
	return atlas
}

func (a *Atlas) StringVerts(text string, size float32) (*Mesh, Rect) {
	initialDot := Vec2{0, 0}
	dot := initialDot

	maxAscent := float32(0)

	mesh := NewMesh()
	for _,r := range text {
		runeMesh, newDot, ascent := a.RuneVerts(r, dot, size)
		mesh.Append(runeMesh)
		dot = newDot

		if maxAscent < ascent {
			maxAscent = ascent
		}
	}
	return mesh, R(initialDot[0], initialDot[1], dot[0], maxAscent)
}

func (a *Atlas) RuneVerts(r rune, dot Vec2, scale float32) (*Mesh, Vec2, float32) {
	// multiplying by texture sizes converts from UV to pixel coords
	scaleX := scale * float32(a.texture.width)
	scaleY := scale * float32(a.texture.height)

	glyph, ok := a.mapping[r]
	if !ok { panic("Missing Rune!") }

	//	log.Println(glyph.Bearing)

	// UV coordinates of the quad
	u1 := glyph.BoundsUV.Min[0]
	u2 := glyph.BoundsUV.Max[0]
	v1 := glyph.BoundsUV.Min[1]
	v2 := glyph.BoundsUV.Max[1]

	// Pixel coordinates of the quad (scaled by scale)
	x1 := dot[0] + (scaleX * glyph.Bearing[0])
	x2 := x1 + (scaleX * (u2 - u1))
	y1 := dot[1] + (scaleY * glyph.Bearing[1])
	y2 := y1 + (scaleY * (v2 - v1))

	mesh := NewQuadMesh(R(x1, y1, x2, y2), R(u1, v1, u2, v2))

	dot[0] += (scaleX * glyph.Advance)

	return mesh, dot, y2
}

func (a *Atlas) Text(str string) *Text {
	t := &Text{
		currentString: "",
		atlas: a,
		texture: a.texture,
		material: NewSpriteMaterial(a.texture),
		scale: 1.0,
	}

	t.Set(str)

	return t
}

type Text struct {
	currentString string
	mesh *Mesh
	atlas *Atlas
	bounds Rect
	texture *Texture
	material Material
	scale float32
}

func (t *Text) Bounds() Rect {
	return t.bounds
}

func (t *Text) Set(str string) {
	if t.currentString != str {
		t.currentString = str
		t.mesh, t.bounds = t.atlas.StringVerts(str, t.scale)
	}
}

func (t *Text) Draw(pass *RenderPass, matrix Mat4) {
	pass.Add(t.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, t.material)
}

func (t *Text) DrawColorMask(pass *RenderPass, matrix Mat4, color RGBA) {
	pass.Add(t.mesh, matrix, color, t.material)
}

func (t *Text) DrawRect(pass *RenderPass, rect Rect, color RGBA) {
	mat := Mat4Ident
	mat.Scale(1.0, 1.0, 1.0).Translate(rect.Min[0], rect.Min[1], 0)
	pass.Add(t.mesh, mat, color, t.material)
}
