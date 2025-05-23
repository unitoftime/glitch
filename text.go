package glitch

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"golang.org/x/image/font"

	"golang.org/x/image/math/fixed"

	"unicode"

	"github.com/golang/freetype/truetype"
	"github.com/unitoftime/flow/glm"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
)

// TODO: Look into this: https://steamcdn-a.akamaihd.net/apps/valve/2007/SIGGRAPH2007_AlphaTestedMagnification.pdf
// TODO: And this: https://blog.mapbox.com/drawing-text-with-signed-distance-fields-in-mapbox-gl-b0933af6f817
// TODO: And this: https://www.youtube.com/watch?v=Y1kuhXtVAc4

// TODO: Ideally this wouldn't return an error
func DefaultAtlas() (*Atlas, error) {
	runes := make([]rune, unicode.MaxASCII-32)
	for i := range runes {
		runes[i] = rune(32 + i)
	}

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}
	fontFace := truetype.NewFace(font, &truetype.Options{
		Size: 32,
		// GlyphCacheEntries: 1,
	})
	cfg := AtlasConfig{
		Smooth:      true,
		TextureSize: 1024,
		Padding:     10,
	}

	atlas := NewAtlas(fontFace, runes, cfg)
	return atlas, nil
}

func BasicFontAtlas() (*Atlas, error) {
	runes := make([]rune, unicode.MaxASCII-32)
	for i := range runes {
		runes[i] = rune(32 + i)
	}

	// font, err := truetype.Parse(gofont.TTF)
	// if err != nil {
	// 	return nil, err
	// }
	// fontFace := truetype.NewFace(font, &truetype.Options{
	// 	Size: size,
	// 	// GlyphCacheEntries: 1,
	// })
	fontFace := basicfont.Face7x13
	cfg := AtlasConfig{
		Smooth:      true,
		TextureSize: 512,
		Padding:     10,
	}
	atlas := NewAtlas(fontFace, runes, cfg)
	return atlas, nil
}

type Glyph struct {
	Advance  float64
	Bearing  Vec2
	BoundsUV Rect
}

// TODO - instead of creating a single atlas ahead of time. I should just load the font and then dynamically create the atlas as needed. This should probably change once you add automatic texture batching.
type Atlas struct {
	// face font.Face
	mapping        map[rune]Glyph
	ascent         float64 // Distance from top of line to baseline
	descent        float64 // Distance from bottom of line to baseline
	height         float64 // The recommended gap between two lines
	texture        *Texture
	border         int  // Specifies a border on the font.
	pixelPerfect   bool // if true anti-aliasing will be disabled
	defaultKerning float64

	defaultMaterial Material

	// TODO: Kind of a hack, but lets me reuse code more easily. Used for measure and for drawText
	tmpText *Text
}

func fixedToFloat(val fixed.Int26_6) float64 {
	// Shift to the left by 6 then convert to an int, then to a float, then shift right by 6
	// TODO - How to handle overruns?
	// intVal := val.Mul(fixed.I(1_000_000)).Floor()
	// return float32(intVal) / 1_000_000.0
	return float64(val) / (1 << 6)
}
func floatToFixed(val float64) fixed.Int26_6 {
	// Shift to the left by 6 then convert to an int, then to a float, then shift right by 6
	// TODO - How to handle overruns?
	// intVal := val.Mul(fixed.I(1_000_000)).Floor()
	// return float32(intVal) / 1_000_000.0
	return fixed.Int26_6(val * (1 << 6))
}

type AtlasConfig struct {
	Border      float64
	Smooth      bool
	Padding     int
	Kerning     float64
	TextureSize int // TODO: automagically calculate
}

func NewAtlas(face font.Face, runes []rune, config AtlasConfig) *Atlas {
	metrics := face.Metrics()
	// fmt.Println("Metrics: ", fixedToFloat(metrics.Height), fixedToFloat(metrics.Ascent), fixedToFloat(metrics.Descent))
	atlas := &Atlas{
		// face: face,
		mapping:        make(map[rune]Glyph),
		ascent:         fixedToFloat(metrics.Ascent),
		descent:        fixedToFloat(metrics.Descent),
		height:         fixedToFloat(metrics.Height),
		border:         int(config.Border),
		pixelPerfect:   !config.Smooth, // TODO - not sure this is exactly right. You could presumably want a bilinear filtered texture but anti-aliasing turned off on the text.
		defaultKerning: config.Kerning,
	}

	border := int(config.Border)
	basePadding := config.Padding
	size := config.TextureSize
	fixedSize := fixed.I(size)
	fSize := float64(size)

	blackImg := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(blackImg, blackImg.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// draw.Draw(img, img.Bounds(), image.NewUniform(color.Alpha{0}), image.ZP, draw.Src)
	// Note: In case you want to see the boundary of each rune, uncomment this
	// draw.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)

	padding := fixed.I(basePadding + (2 * atlas.border))                     // Padding for runes drawn to atlas
	startDot := fixed.P(padding.Floor(), (metrics.Ascent + padding).Floor()) // Starting point of the dot
	dot := startDot
	for i, r := range runes {
		// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyphterms_2x.png
		bounds, mask, maskp, adv, ok := face.Glyph(dot, r)
		if !ok {
			panic("Missing rune!")
		}
		bearingRect, _, _ := face.GlyphBounds(r)

		// Instead of flooring we convert from fixed int to float manually (mult by 10^6 then floor, cast and divide by 10^6). I think this is slightly more accurate but it's hard to tell so I'll leave old code below
		//		log.Println("Rune: ", string(r), " - BearingRect: ", bearingRect)
		bearingX := float64((bearingRect.Min.X * 1000000).Floor()) / (1000000 * fSize)
		bearingY := float64((-bearingRect.Max.Y * 1000000).Floor()) / (1000000 * fSize)

		//		advance := float32((adv * 1000000).Floor())/(1000000 * fSize) // TODO - why doesn't this work?
		// log.Println("Rune: ", string(r), " - BearingX: ", float32(bearingRect.Min.X.Floor())/fSize)
		// log.Println("Rune: ", string(r), " - BearingX: ", bearingX)
		// log.Println("Rune: ", string(r), " - BearingY: ", float32(-bearingRect.Max.Y.Floor())/fSize)
		// log.Println("Rune: ", string(r), " - BearingY: ", bearingY)

		// Before: Single draw which wouldn't have a border
		// draw.Draw(img, bounds, mask, maskp, draw.Src)

		// After: 9 offset draws in every direction, then a normal draw
		// Draw nine slots around
		// // x = dist * cos(pi/2)
		// diagDist := int(float64(border) * 1.0 / math.Sqrt(2))
		// draw.DrawMask(img, bounds.Add(image.Point{border, 0}), blackImg, image.Point{}, mask, maskp, draw.Over)
		// draw.DrawMask(img, bounds.Add(image.Point{diagDist, diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)
		// draw.DrawMask(img, bounds.Add(image.Point{diagDist, -diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)

		// draw.DrawMask(img, bounds.Add(image.Point{-border, 0}), blackImg, image.Point{}, mask, maskp, draw.Over)
		// draw.DrawMask(img, bounds.Add(image.Point{-diagDist, diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)
		// draw.DrawMask(img, bounds.Add(image.Point{-diagDist, -diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)

		// draw.DrawMask(img, bounds.Add(image.Point{0, border}), blackImg, image.Point{}, mask, maskp, draw.Over)
		// draw.DrawMask(img, bounds.Add(image.Point{0, -border}), blackImg, image.Point{}, mask, maskp, draw.Over)

		// // Draw shadow
		// shadow := 1
		// draw.DrawMask(img, bounds.Add(image.Point{border + shadow, border + shadow}), blackImg, image.Point{}, mask, maskp, draw.Over)
		// // draw.DrawMask(img, bounds.Add(image.Point{0, -border-shadow}), blackImg, image.Point{}, mask, maskp, draw.Over)

		if border > 0 && config.Smooth {
			// Draw nine slots around
			// x = dist * cos(pi/2)
			diagDist := int(float64(border) * 1.0 / math.Sqrt(2))
			// diagDist := int(float64(border)/2)
			draw.DrawMask(img, bounds.Add(image.Point{border, 0}), blackImg, image.Point{}, mask, maskp, draw.Over)
			draw.DrawMask(img, bounds.Add(image.Point{diagDist, diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)
			draw.DrawMask(img, bounds.Add(image.Point{diagDist, -diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)

			draw.DrawMask(img, bounds.Add(image.Point{-border, 0}), blackImg, image.Point{}, mask, maskp, draw.Over)
			draw.DrawMask(img, bounds.Add(image.Point{-diagDist, diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)
			draw.DrawMask(img, bounds.Add(image.Point{-diagDist, -diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)

			draw.DrawMask(img, bounds.Add(image.Point{0, border}), blackImg, image.Point{}, mask, maskp, draw.Over)
			draw.DrawMask(img, bounds.Add(image.Point{0, -border}), blackImg, image.Point{}, mask, maskp, draw.Over)
		}

		draw.Draw(img, bounds, mask, maskp, draw.Over)
		// draw.DrawMask(img, bounds, blackImg, image.Point{}, mask, maskp, draw.Src)

		atlas.mapping[r] = Glyph{
			// Advance: float64(adv.Floor() + (2*border))/fSize,
			Advance: float64(adv.Floor()) / fSize,
			//			Bearing: Vec2{float32(bearingRect.Min.X.Floor())/fSize, float32((-bearingRect.Max.Y).Floor())/fSize},
			//Advance: advance,
			Bearing: Vec2{bearingX, bearingY},
			BoundsUV: glm.R(
				float64(bounds.Min.X-atlas.border)/fSize, float64(bounds.Min.Y-atlas.border)/fSize,
				float64(bounds.Max.X+atlas.border)/fSize, float64(bounds.Max.Y+atlas.border)/fSize,
			).Norm(),
		}

		// Usual next dot location
		nextDotX := dot.X + adv + padding
		nextDotY := dot.Y

		// Exit if we are at the end
		if (i + 1) >= len(runes) {
			break
		}

		// If the rune after this one pushes us too far then loop around
		nextAdv, ok := face.GlyphAdvance(runes[i+1])
		if !ok {
			panic("Missing rune!")
		}
		if nextDotX+nextAdv >= fixedSize {
			// log.Println("Ascending!")
			nextDotX = startDot.X
			nextDotY = dot.Y + metrics.Ascent + padding
		}
		// log.Println(nextDotX, nextDotY)
		dot = fixed.Point26_6{nextDotX, nextDotY}
	}

	// This just disables anti-aliasing by snapping pixels to either white or transparent
	// if atlas.pixelPerfect {
	// 	imgBounds := img.Bounds()
	// 	for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
	// 		for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
	// 			rgba := img.RGBAAt(x, y)
	// 			if rgba.A > 0 {
	// 				rgba.A = 255
	// 				img.Set(x, y, color.White)
	// 			}
	// 		}
	// 	}
	// }

	// This runs a box filter based on the border side
	if atlas.border != 0 && !config.Smooth {
		// Only border this way for pixel fonts
		// Finds white pixels and draws borders around the edges
		imgBounds := img.Bounds()
		for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
			for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
				rgba := img.RGBAAt(x, y)
				if (rgba != color.RGBA{255, 255, 255, 255}) {
					continue // If the pixel is not white, then it doesnt trigger a border
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

		// Finds transparent pixels and draws borders inward on non-transparent pixels
		// imgBounds := img.Bounds()
		// for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
		// 	for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
		// 		rgba := img.RGBAAt(x, y)
		// 		if rgba.A != 0 {
		// 			continue // Skip if pixel is not fully transparent
		// 		}

		// 		box := image.Rect(x-atlas.border, y-atlas.border, x+atlas.border, y+atlas.border)
		// 		for xx := box.Min.X; xx <= box.Max.X; xx++ {
		// 			for yy := box.Min.Y; yy <= box.Max.Y; yy++ {
		// 				rgba := img.RGBAAt(xx, yy)
		// 				if rgba.A != 0 {
		// 					// Only add a border to transparent pixels
		// 					rgba.R = 0
		// 					rgba.G = 0
		// 					rgba.B = 0
		// 					img.Set(xx, yy, rgba)
		// 				}
		// 			}
		// 		}
		// 	}
		// }
	}

	// // outputFile is a File type which satisfies Writer interface
	// outputFile, err := os.Create("test.png")
	// if err != nil { panic(err) }
	// png.Encode(outputFile, img)
	// outputFile.Close()

	atlas.texture = NewTexture(img, config.Smooth)
	atlas.defaultMaterial = DefaultMaterial(atlas.texture)

	atlas.tmpText = atlas.Text("", 1.0)

	// fmt.Println("TextAtlas: ", atlas.texture.width, atlas.texture.height)
	return atlas
}

func (a *Atlas) Material() *Material {
	return &a.defaultMaterial
}

// func (a *Atlas) GappedLineHeight() float64 {
// 	// TODO - scale?
// 	return (-fixedToFloat(a.ascent) + fixedToFloat(a.descent) - fixedToFloat(a.height)) + float64(2 * a.border)
// }

func (a *Atlas) LineHeight() float64 {
	return a.height
}

func (a *Atlas) UngappedLineHeight() float64 {
	// TODO - scale?
	return (a.ascent + a.descent)
}

func (a *Atlas) getRuneGlyph(r rune) Glyph {
	glyph, ok := a.mapping[r]
	if !ok {
		// Replace rune with '?'
		oldR := r
		r = '?' // TODO - Pick some other rune. TODO - require this rune to be in the atlas!
		glyph, ok = a.mapping[r]
		if !ok {
			panic(fmt.Sprintf("Missing Rune: %v and replacement%v", oldR, r))
		}
	}
	return glyph
}

// Gets the total size of the rune plus advance and kerning
func (a *Atlas) runeSize(r rune, scale float64) (glm.Vec2) {
	// multiplying by texture sizes converts from UV to pixel coords
	scaleX := scale * float64(a.texture.width)
	scaleY := scale * float64(a.texture.height)

	glyph := a.getRuneGlyph(r)

	// UV coordinates of the quad
	u1 := glyph.BoundsUV.Min.X
	u2 := glyph.BoundsUV.Max.X
	v1 := glyph.BoundsUV.Min.Y
	v2 := glyph.BoundsUV.Max.Y

	// Pixel coordinates of the quad (scaled by scale)
	x1 := 0.0
	x2 := x1 + (scaleX * (u2 - u1))

	// Note: Commented out the downard shift here, and I'm doing it in the above func
	y1 := 0.0
	y2 := y1 + (scaleY * (v2 - v1))

	// Also add the advance and kerning
	x2 += (scaleX * glyph.Advance) + (a.defaultKerning * scale) // TODO: Kerning should come from text, not atlas
	return glm.Vec2{
		X: x2 - x1,
		Y: y2 - y1,
	}
}

func (a *Atlas) RuneVerts(mesh *Mesh, r rune, dot Vec2, scale float64, color RGBA) (Vec2, float64) {
	// multiplying by texture sizes converts from UV to pixel coords
	scaleX := scale * float64(a.texture.width)
	scaleY := scale * float64(a.texture.height)

	glyph := a.getRuneGlyph(r)

	// UV coordinates of the quad
	u1 := glyph.BoundsUV.Min.X
	u2 := glyph.BoundsUV.Max.X
	v1 := glyph.BoundsUV.Min.Y
	v2 := glyph.BoundsUV.Max.Y

	// Pixel coordinates of the quad (scaled by scale)
	x1 := dot.X + (scaleX * glyph.Bearing.X)
	x2 := x1 + (scaleX * (u2 - u1))

	// Note: Commented out the downard shift here, and I'm doing it in the above func
	y1 := dot.Y + (scaleY * glyph.Bearing.Y) + (a.descent * scale)
	y2 := y1 + (scaleY * (v2 - v1))

	destRect := glm.R(x1, y1, x2, y2)
	if a.pixelPerfect {
		destRect = glm.R(math.Round(x1), math.Round(y1), math.Round(x2), math.Round(y2))
	}

	if mesh != nil {
		mesh.AppendQuadMesh(destRect, glyph.BoundsUV, color)
		// mesh.AppendQuadMesh(destRect, glm.R(u1, v1, u2, v2), color)
		// mesh := NewQuadMesh(R(x1, y1, x2, y2), R(u1, v1, u2, v2))
	}

	dot.X += (scaleX * glyph.Advance) + (a.defaultKerning * scale) // TODO: Kerning should come from text, not atlas

	return dot, y2
}

func (a *Atlas) NewText(text string, scale float64) textDraw {
	t := textDraw{
		atlas: a,
		text: text,
		scale: scale,
	}
	return t
}

type textDraw struct {
	atlas *Atlas
	text string
	scale float64
}

func (t textDraw) Fill(pool *BufferPool, mat glMat4, mask RGBA) *VertexBuffer {
	// TODO: Adds an additional buffer copy for all of the verts
	tt := t.atlas.tmpText

	tt.SetScale(t.scale)
	tt.Set(t.text)
	tt.currentString = t.text
	tt.Clear()
	tt.bounds = tt.AppendStringVerts(tt.currentString, false)

	return tt.mesh.Fill(pool, mat, mask)
}
func (t textDraw) Bounds() glm.Box {
	return t.atlas.Measure(t.text, t.scale).Box()
}

func (t textDraw) Draw(target BatchTarget, matrix Mat4) {
	t.DrawColorMask(target, matrix, glm.White)
}

func (t textDraw) DrawColorMask(target BatchTarget, matrix Mat4, mask glm.RGBA) {
	geom := GeometryFiller{
		prog: t,
	}
	target.Add(geom, glm4(matrix), mask, t.atlas.defaultMaterial)
}

func (a *Atlas) Text(str string, scale float64) *Text {
	t := &Text{
		currentString: "",
		atlas:         a,
		texture:       a.texture,
		// material: NewSpriteMaterial(a.texture),
		material: a.defaultMaterial,
		scale:    scale,
		// LineHeight: a.UngappedLineHeight(),
		mesh:    NewMesh(),
		tmpMesh: NewMesh(),

		color: RGBA{1, 1, 1, 1},
	}

	t.Set(str)

	return t
}

// TODO: This could be improved by just calling specialized measurement functions
func (a *Atlas) Measure(str string, scale float64) Rect {
	a.tmpText.Clear()

	a.tmpText.currentString = str
	a.tmpText.scale = scale
	return a.tmpText.AppendStringVerts(str, true)

	// fakeText := Text{
	// 	currentString: str,
	// 	atlas:         a,
	// 	scale:         scale,
	// }
	// return fakeText.AppendStringVerts(str, true)
}
func (a *Atlas) MeasureWrapped(str string, scale float64, wrapRect Rect) Rect {
	fakeText := Text{
		currentString: str,
		atlas:         a,
		scale:         scale,
		wordWrap:      true,
		wrapRect:      wrapRect,
	}
	return fakeText.AppendStringVerts(str, true)
}

type Text struct {
	currentString string
	mesh          *Mesh
	tmpMesh       *Mesh // For temporarily buffering data. TODO - would be more efficient just to append the quads directly to the mesh rather than buffering them here
	atlas         *Atlas
	bounds        Rect
	texture       *Texture
	material      Material
	scale         float64
	shadow        Vec2
	wordWrap      bool
	wrapRect      Rect
	// LineHeight float64

	Orig  Vec2 // The baseline starting point from which to draw the text
	Dot   Vec2 // The location of the next rune to draw
	color RGBA // The color with which to draw the next text
}

func (t *Text) Bounds() Rect {
	return t.bounds
}

// func (t *Text) SetMaterial(material Material) {
// 	t.material = material
// }

func (t *Text) Material() *Material {
	return &t.material
}

func (t *Text) MeshBounds() Rect {
	return t.mesh.Bounds().Rect()
}

func (t *Text) SetScale(scale float64) {
	t.scale = scale
}
func (t *Text) SetColor(col RGBA) {
	t.color = col
}

func (t *Text) SetShadow(shadow Vec2) {
	t.shadow = shadow
}

func (t *Text) SetWordWrap(wrap bool, wrapRect Rect) {
	t.wordWrap = wrap
	t.wrapRect = wrapRect

	// if t.wordWrap {
	// 	// lineHeight := t.atlas.UngappedLineHeight() * t.scale
	// 	t.Orig = Vec2{wrapRect.Min.X, wrapRect.Max.Y} // + lineHeight}
	// }
}

func (t *Text) Clear() {
	t.Orig = Vec2{}
	t.Dot = t.Orig
	t.mesh.Clear()
}

// This resets the text and sets it to the passed in string (if the passed in string is different!)
// TODO - I need to deprecate this in favor of a better interface
func (t *Text) Set(str string) {
	// TODO: If wordwrap we also need to regenerate, but technically only if the bounds of the wrapRect have changed
	if t.currentString != str || t.wordWrap {
		t.currentString = str
		t.regenerate()
	}
}

func (t *Text) regenerate() {
	t.Clear()
	t.bounds = t.AppendStringVerts(t.currentString, false)
}

func (t *Text) WriteString(str string) (n int, err error) {
	return t.Write([]byte(str))
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
	newBounds := t.AppendStringVerts(appendedStr, false)
	t.bounds = t.bounds.Union(newBounds)
	return len(p), nil
}

func (t *Text) Draw(target BatchTarget, matrix Mat4) {
	t.DrawColorMask(target, matrix, White)
}

func (t *Text) DrawColorMask(target BatchTarget, matrix Mat4, color RGBA) {
	// mat2 := matrix
	// mat2.Translate(0, -0.5, 0)
	// target.Add(t.mesh, mat2, Black, t.material, false)

	target.Add(t.mesh.g(), glm4(matrix), color, t.material)
}

func (t *Text) RectDraw(target BatchTarget, rect Rect) {
	t.DrawRect(target, rect, White)
}

func (t *Text) DrawRect(target BatchTarget, rect Rect, color RGBA) {
	mat := Mat4Ident

	mat.Scale(1.0, 1.0, 1.0).Translate(rect.Min.X, rect.Min.Y, 0)
	target.Add(t.mesh.g(), glm4(mat), color, t.material)
}

func (t *Text) RectDrawColorMask(target BatchTarget, bounds Rect, mask RGBA) {
	mat := Mat4Ident
	// TODO why shouldn't I be shifting to the middle?
	// mat.Scale(bounds.W() / t.bounds.W(), bounds.H() / t.bounds.H(), 1).Translate(bounds.W()/2 + bounds.Min.X, bounds.H()/2 + bounds.Min.Y, 0)
	// mat.Scale(1.0, 1.0, 1.0).Translate(rect.Min.X, rect.Min.Y, 0)

	// TODO!!! - There's something wrong with this
	mat.Scale(bounds.W()/t.bounds.W(), bounds.H()/t.bounds.H(), 1).Translate(bounds.Min.X, bounds.Min.Y, 0)

	target.Add(t.mesh.g(), glm4(mat), mask, t.material)
}

// If measure is set true, dont add them to the text mesh, just measure the bounds of the string
func (t *Text) AppendStringVerts(text string, measure bool) Rect {
	// maxAscent := float32(0) // Tracks the maximum y point of the text block

	lineHeight := t.atlas.UngappedLineHeight() * t.scale
	initialDot := t.Dot
	maxDotX := t.Dot.X

	numLines := 1.0
	for i, r := range text {
		// On tab, go to the next tabwidth position
		tab := r == '\t'
		if tab {
			spaceSize := t.atlas.runeSize(' ', t.scale)
			tabWidth := 4.0 * spaceSize.X
			currentTabSection := (t.Dot.X - t.Orig.X) / tabWidth
			t.Dot.X = t.Orig.X + math.Floor(currentTabSection + 1.0) * tabWidth
			continue
		}

		// If the rune is a newline, then we need to reset the dot for the next line
		newline := r == '\n'
		// If we wordwraping and are on a space, check to see if we should go to the next line
		if t.wordWrap && r == ' ' {
			nextSpaceIdx := strings.Index(text[i+1:], " ")
			if nextSpaceIdx < 0 {
				// There is no next space so we measure the rest of the line
				nextSpaceIdx = len(text) - i
			}

			nextWord := t.atlas.Measure(text[i:i+nextSpaceIdx], t.scale)

			if (t.Dot.X-initialDot.X)+nextWord.W() > t.wrapRect.W() {
				newline = true
			}
		}
		if newline {
			// t.Dot.Y -= t.atlas.LineHeight()
			t.Dot.Y -= lineHeight
			t.Dot.X = t.Orig.X
			numLines++

			continue
		}

		var dstMesh *Mesh
		if !measure {
			dstMesh = t.mesh
		}
		newDot, _ := t.atlas.RuneVerts(dstMesh, r, t.Dot, t.scale, t.color)

		noShadow := Vec2{}
		if t.shadow != noShadow {
			_, _ = t.atlas.RuneVerts(dstMesh, r, t.Dot.Add(t.shadow), t.scale, Black)
		}

		maxDotX = max(maxDotX, newDot.X)

		t.Dot = newDot

		// if maxAscent < ascent {
		// 	maxAscent = dot.Y + ascent
		// }
	}
	// // return R(meshBounds.Min.X, initialDot.Y, meshBounds.Max.X, initialDot.Y - (numLines * lineHeight)).Norm()

	// // Attempt 4 - use mesh bounds for X and line height for Y
	// meshBounds := t.mesh.Bounds().Rect()

	// // Note: The RuneVerts function corners the glyph into the bounds by applying the descent. So I dont need to track ascent/descent here
	// // top := initialDot.Y + (fixedToFloat(t.atlas.ascent) * t.scale)
	// // bot := top - (numLines * lineHeight)

	// top := initialDot.Y + lineHeight
	// // top := initialDot.Y + t.atlas.descent

	// bot := top - (numLines * lineHeight)
	// return R(meshBounds.Min.X, top, meshBounds.Max.X, bot).Norm()// .MoveMin(Vec2{})

	// Attempt 5 - use dot for X and line height for Y
	top := initialDot.Y + lineHeight
	// top := initialDot.Y + t.atlas.descent

	bot := top - (numLines * lineHeight)
	return glm.R(initialDot.X, top, maxDotX, bot).Norm() // .MoveMin(Vec2{})

	//--------------------------------------------------------------------------------

	// // return mesh, R(initialDot[0], initialDot[1], dot[0], dot[1] + maxAscent)

	// // fmt.Println("-----")
	// // fmt.Println(fixedToFloat(a.ascent))
	// // fmt.Println(fixedToFloat(a.descent))
	// // fmt.Println(fixedToFloat(a.height))
	// // fmt.Println(maxAscent)
	// // fmt.Println(scale)
	// // bounds := R(initialDot[0], initialDot[1], dot[0], dot[1] - a.LineHeight())

	// // bounds := R(initialDot[0],
	// // 	initialDot[1] - (2 * fixedToFloat(a.ascent)),
	// // 	dot[0], // TODO - this is wrong if because this is the length of the last line, we need the length of the longest line
	// // 	dot[1] - (2 * fixedToFloat(a.descent)))

	// // TODO - this used the glyphs to determine bounds, below I use the mesh
	// // // TODO - idk what I'm doing here, but it seems to work. Man text rendering is hard.
	// // bounds := R(initialDot[0],
	// // 	initialDot[1] - (fixedToFloat(t.atlas.ascent)),
	// // 	t.Dot[0], // TODO - this is wrong if because this is the length of the last line, we need the length of the longest line
	// // 	t.Dot[1] - (fixedToFloat(t.atlas.descent))).
	// // 		Norm().
	// // 		Moved(Vec2{0, fixedToFloat(t.atlas.ascent)})
	// // return bounds

	// // Attempt 2 - Use mesh bounds
	// // return t.mesh.Bounds().Rect()

	// // Attempt 3 - use mesh bounds for X and line height for Y
	// meshBounds := t.mesh.Bounds().Rect()
	// return R(meshBounds.Min[0], initialDot[1], meshBounds.Max[0], t.Dot[1] + lineHeight)

	// // fmt.Println(bounds)

	// // bounds := R(initialDot[0],
	// // 	initialDot[1] - (fixedToFloat(a.descent)),
	// // 	dot[0], // TODO - this is wrong if because this is the length of the last line, we need the length of the longest line
	// // 	dot[1] - (fixedToFloat(a.ascent))).Norm()
	// // return mesh, bounds

	// // return mesh, R(initialDot[0], initialDot[1], dot[0], dot[1] + fixedToFloat(a.lineHeight))
	// // return mesh, R(initialDot[0], initialDot[1], dot[0], dot[1] - (fixedToFloat(a.ascent) - fixedToFloat(a.descent)))
	// // return mesh, R(initialDot[0],
	// // 	initialDot[1] - fixedToFloat(a.descent)/1024,
	// // 	dot[0], // TODO - this is wrong if because this is the length of the last line, we need the length of the longest line
	// // 	dot[1] + fixedToFloat(a.ascent)/1024)
}

// //--------------------------------------------------------------------------------
// //--------------------------------------------------------------------------------
// //--------------------------------------------------------------------------------
// //--------------------------------------------------------------------------------
// func NewAtlas(face font.Face, runes []rune, config AtlasConfig) *Atlas {
// 	metrics := face.Metrics()
// 	// fmt.Println("Metrics: ", fixedToFloat(metrics.Height), fixedToFloat(metrics.Ascent), fixedToFloat(metrics.Descent))
// 	atlas := &Atlas{
// 		// face: face,
// 		mapping: make(map[rune]Glyph),
// 		ascent: fixedToFloat(metrics.Ascent),
// 		descent: fixedToFloat(metrics.Descent),
// 		height: fixedToFloat(metrics.Height),
// 		border: int(config.Border),
// 		pixelPerfect: !config.Smooth, // TODO - not sure this is exactly right. You could presumably want a bilinear filtered texture but anti-aliasing turned off on the text.
// 		defaultKerning: config.Kerning,
// 	}

// 	border := int(config.Border)
// 	basePadding := config.Padding
// 	size := config.TextureSize
// 	fixedSize := fixed.I(size)
// 	fSize := float64(size)

// 	blackImg := image.NewRGBA(image.Rect(0, 0, size, size))
// 	draw.Draw(blackImg, blackImg.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)

// 	img := image.NewRGBA(image.Rect(0, 0, size, size))
// 	// draw.Draw(img, img.Bounds(), image.NewUniform(color.Alpha{0}), image.ZP, draw.Src)
// 	// Note: In case you want to see the boundary of each rune, uncomment this
// 	// draw.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)

// 	padding := fixed.I(basePadding + (2 * atlas.border)) // Padding for runes drawn to atlas
// 	startDot := fixed.P(padding.Floor(), (atlas.ascent + padding).Floor()) // Starting point of the dot
// 	dot := startDot
// 	for i, r := range runes {
// 		// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyphterms_2x.png
// 		bounds, mask, maskp, adv, ok := face.Glyph(dot, r)
// 		if !ok { panic("Missing rune!") }
// 		bearingRect, _, _ := face.GlyphBounds(r)

// 		// if r == 'R' {
// 		// 	fmt.Printf("%T\n", mask)
// 		// 	// fmt.Println(mask)
// 		// 	outputFile, err := os.Create("testR.png")
// 		// 	if err != nil { panic(err) }
// 		// 	png.Encode(outputFile, mask)
// 		// 	outputFile.Close()
// 		// }

// 		// Instead of flooring we convert from fixed int to float manually (mult by 10^6 then floor, cast and divide by 10^6). I think this is slightly more accurate but it's hard to tell so I'll leave old code below
// 		//		log.Println("Rune: ", string(r), " - BearingRect: ", bearingRect)
// 		bearingX := float64((bearingRect.Min.X * 1000000).Floor()) / (1000000 * fSize)
// 		bearingY := float64((-bearingRect.Max.Y * 1000000).Floor()) / (1000000 * fSize)

// 		//		advance := float32((adv * 1000000).Floor())/(1000000 * fSize) // TODO - why doesn't this work?
// 		// log.Println("Rune: ", string(r), " - BearingX: ", float32(bearingRect.Min.X.Floor())/fSize)
// 		// log.Println("Rune: ", string(r), " - BearingX: ", bearingX)
// 		// log.Println("Rune: ", string(r), " - BearingY: ", float32(-bearingRect.Max.Y.Floor())/fSize)
// 		// log.Println("Rune: ", string(r), " - BearingY: ", bearingY)

// 		// Before: Single draw which wouldn't have a border
// 		// draw.Draw(img, bounds, mask, maskp, draw.Src)

// 		// After: 9 offset draws in every direction, then a normal draw
// 		// Draw nine slots around
// 		// // x = dist * cos(pi/2)
// 		// diagDist := int(float64(border) * 1.0 / math.Sqrt(2))
// 		// draw.DrawMask(img, bounds.Add(image.Point{border, 0}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 		// draw.DrawMask(img, bounds.Add(image.Point{diagDist, diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 		// draw.DrawMask(img, bounds.Add(image.Point{diagDist, -diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)

// 		// draw.DrawMask(img, bounds.Add(image.Point{-border, 0}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 		// draw.DrawMask(img, bounds.Add(image.Point{-diagDist, diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 		// draw.DrawMask(img, bounds.Add(image.Point{-diagDist, -diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)

// 		// draw.DrawMask(img, bounds.Add(image.Point{0, border}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 		// draw.DrawMask(img, bounds.Add(image.Point{0, -border}), blackImg, image.Point{}, mask, maskp, draw.Over)

// 		// // Draw shadow
// 		// shadow := 1
// 		// draw.DrawMask(img, bounds.Add(image.Point{border + shadow, border + shadow}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 		// // draw.DrawMask(img, bounds.Add(image.Point{0, -border-shadow}), blackImg, image.Point{}, mask, maskp, draw.Over)

// 		if border > 0 && config.Smooth {
// 			// Draw nine slots around
// 			// x = dist * cos(pi/2)
// 			diagDist := int(float64(border) * 1.0 / math.Sqrt(2))
// 			// diagDist := int(float64(border)/2)
// 			draw.DrawMask(img, bounds.Add(image.Point{border, 0}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 			draw.DrawMask(img, bounds.Add(image.Point{diagDist, diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 			draw.DrawMask(img, bounds.Add(image.Point{diagDist, -diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)

// 			draw.DrawMask(img, bounds.Add(image.Point{-border, 0}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 			draw.DrawMask(img, bounds.Add(image.Point{-diagDist, diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 			draw.DrawMask(img, bounds.Add(image.Point{-diagDist, -diagDist}), blackImg, image.Point{}, mask, maskp, draw.Over)

// 			draw.DrawMask(img, bounds.Add(image.Point{0, border}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 			draw.DrawMask(img, bounds.Add(image.Point{0, -border}), blackImg, image.Point{}, mask, maskp, draw.Over)
// 		}

// 		draw.Draw(img, bounds, mask, maskp, draw.Over)
// 		// draw.DrawMask(img, bounds, blackImg, image.Point{}, mask, maskp, draw.Src)

// 		atlas.mapping[r] = Glyph{
// 			// Advance: float64(adv.Floor() + (2*border))/fSize,
// 			Advance: float64(adv.Floor())/fSize,
// 			//			Bearing: Vec2{float32(bearingRect.Min.X.Floor())/fSize, float32((-bearingRect.Max.Y).Floor())/fSize},
// 			//Advance: advance,
// 			Bearing: Vec2{bearingX, bearingY},
// 			BoundsUV: R(
// 				float64(bounds.Min.X - atlas.border)/fSize, float64(bounds.Min.Y - atlas.border)/fSize,
// 				float64(bounds.Max.X + atlas.border)/fSize, float64(bounds.Max.Y + atlas.border)/fSize),
// 		}

// 		// Usual next dot location
// 		nextDotX := dot.X + adv + padding
// 		nextDotY := dot.Y

// 		// Exit if we are at the end
// 		if (i+1) >= len(runes) { break }

// 		// If the rune after this one pushes us too far then loop around
// 		nextAdv, ok := face.GlyphAdvance(runes[i+1])
// 		if !ok { panic("Missing rune!") }
// 		if nextDotX + nextAdv >= fixedSize {
// 			// log.Println("Ascending!")
// 			nextDotX = startDot.X
// 			nextDotY = dot.Y + atlas.ascent + padding
// 		}
// 		// log.Println(nextDotX, nextDotY)
// 		dot = fixed.Point26_6{nextDotX, nextDotY}
// 	}

// 	// This just disables anti-aliasing by snapping pixels to either white or transparent
// 	// if atlas.pixelPerfect {
// 	// 	imgBounds := img.Bounds()
// 	// 	for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
// 	// 		for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
// 	// 			rgba := img.RGBAAt(x, y)
// 	// 			if rgba.A > 0 {
// 	// 				rgba.A = 255
// 	// 				img.Set(x, y, color.White)
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }

// 	// This runs a box filter based on the border side
// 	if atlas.border != 0 && !config.Smooth {
// 		// Only border this way for pixel fonts
// 		// Finds white pixels and draws borders around the edges
// 		imgBounds := img.Bounds()
// 		for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
// 			for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
// 				rgba := img.RGBAAt(x, y)
// 				if (rgba != color.RGBA{255, 255, 255, 255}) {
// 					continue // If the pixel is not white, then it doesnt trigger a border
// 				}

// 				box := image.Rect(x-atlas.border, y-atlas.border, x+atlas.border, y+atlas.border)
// 				for xx := box.Min.X; xx <= box.Max.X; xx++ {
// 					for yy := box.Min.Y; yy <= box.Max.Y; yy++ {
// 						rgba := img.RGBAAt(xx, yy)
// 						if rgba.A == 0 {
// 							// Only add a border to transparent pixels
// 							img.Set(xx, yy, color.Black)
// 						}
// 					}
// 				}
// 			}
// 		}

// 		// Finds transparent pixels and draws borders inward on non-transparent pixels
// 		// imgBounds := img.Bounds()
// 		// for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
// 		// 	for y := imgBounds.Min.Y; y < imgBounds.Max.Y; y++ {
// 		// 		rgba := img.RGBAAt(x, y)
// 		// 		if rgba.A != 0 {
// 		// 			continue // Skip if pixel is not fully transparent
// 		// 		}

// 		// 		box := image.Rect(x-atlas.border, y-atlas.border, x+atlas.border, y+atlas.border)
// 		// 		for xx := box.Min.X; xx <= box.Max.X; xx++ {
// 		// 			for yy := box.Min.Y; yy <= box.Max.Y; yy++ {
// 		// 				rgba := img.RGBAAt(xx, yy)
// 		// 				if rgba.A != 0 {
// 		// 					// Only add a border to transparent pixels
// 		// 					rgba.R = 0
// 		// 					rgba.G = 0
// 		// 					rgba.B = 0
// 		// 					img.Set(xx, yy, rgba)
// 		// 				}
// 		// 			}
// 		// 		}
// 		// 	}
// 		// }
// 	}

// 	// // outputFile is a File type which satisfies Writer interface
// 	// outputFile, err := os.Create("test.png")
// 	// if err != nil { panic(err) }
// 	// png.Encode(outputFile, img)
// 	// outputFile.Close()

// 	atlas.texture = NewTexture(img, config.Smooth)
// 	atlas.defaultMaterial = DefaultMaterial(atlas.texture)

// 	// fmt.Println("TextAtlas: ", atlas.texture.width, atlas.texture.height)
// 	return atlas
// }
