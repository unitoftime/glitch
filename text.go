package glitch

import (
	"os"

	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"golang.org/x/image/font"
	// "github.com/golang/freetype/truetype"

	"golang.org/x/image/math/fixed"
	// "github.com/go-gl/mathgl/mgl32"
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
	outputFile, err := os.Create("test.png")
	if err != nil { panic(err) }
	png.Encode(outputFile, img)
	outputFile.Close()

	atlas.texture = NewTexture(img)
	fmt.Println("TextAtlas: ", atlas.texture.width, atlas.texture.height)
	return atlas
}

func (a *Atlas) StringVerts(text string, size float32) *Mesh {
	dot := Vec2{0,0}

	mesh := NewMesh()
	for _,r := range text {
		runeMesh, newDot := a.RuneVerts(r, dot, size)
		mesh.Append(runeMesh)
		dot = newDot
	}
	return mesh
}

func (a *Atlas) RuneVerts(r rune, dot Vec2, size float32) (*Mesh, Vec2) {
	//	size := float32(0.5) // TODO - hardcoded
	//	size = size / 32 // TODO - hardcoding. Not actually computing the right font size
	//	size = float32(1)
	size = size * 12 // // TODO - hardcoding. Not actually computing the right font size

	glyph, ok := a.mapping[r]
	if !ok { panic("Missing Rune!") }

	//	log.Println(glyph.Bearing)

	// UV coordinates of the quad
	u1 := glyph.BoundsUV.Min[0]
	u2 := glyph.BoundsUV.Max[0]
	v1 := glyph.BoundsUV.Min[1]
	v2 := glyph.BoundsUV.Max[1]

	// x1 := dot.X
	// x2 := dot.X + size * (u2 - u1)
	// y1 := dot.Y
	// y2 := dot.Y + size * (v2 - v1)

	// gl position coordinates of the quad
	// Note: we scale by size here because the atlas isn't necessarily 1:1 with the fontsize
	x1 := dot[0] + (size * glyph.Bearing[0])
	x2 := x1 + (size * (u2 - u1))
	y1 := dot[1] + (size * glyph.Bearing[1])
	y2 := y1 + (size * (v2 - v1))

	mesh := NewQuadMesh(R(x1, y1, x2, y2), R(u1, v1, u2, v2))

	dot[0] += (size * glyph.Advance)

	return mesh, dot
}

func (a *Atlas) Text(str string) *Text {
	return &Text{
		currentString: str,
		mesh: a.StringVerts(str, 36),
		texture: a.texture,
		material: NewSpriteMaterial(a.texture),
	}
}

type Text struct {
	currentString string
	mesh *Mesh
	bounds Rect
	texture *Texture
	material Material
}

func (t *Text) Draw(pass *RenderPass, matrix Mat4) {
	pass.Add(t.mesh, matrix, RGBA{1.0, 1.0, 1.0, 1.0}, t.material)
}
