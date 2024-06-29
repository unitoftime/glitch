package glitch

import (
	"image"
)

// RBG: msdf-atlas-gen -type mtsdf -emrange 0.2 -dimensions 255 255 -font ~/Library/Fonts/FiraSans-Regular.otf -imageout assets/FiraSans-Regular.png -json assets/FiraSans-Regular.json

// ./msdf-atlas-gen/build/bin/msdf-atlas-gen -font ./Lato-Black.ttf -imageout atlas.png -json atlas.json -pots -size 32 -yorigin top -emrange 0.2 -type mtsdf

// MSDF
// ../msdf-atlas-gen/build/bin/msdf-atlas-gen -font ./Lato-Black.ttf -imageout atlas.png -json atlas.json -pots -size 32 -yorigin top -pxrange 10

// SDF
// ./msdf-atlas-gen/build/bin/msdf-atlas-gen -font ./Lato-Black.ttf -imageout atlas.png -json atlas.json -pots -size 32 -yorigin top -pxrange 10
// PlaneBounds: https://github.com/Chlumsky/msdf-atlas-gen/issues/2
type SdfAtlasPreamble struct {
	Type string
	DistanceRange float64
	DistanceRangeMiddle int
	Size int
	Width int
	Height int
	YOrigin string
}

type SdfMetrics struct {
	EmSize int
	LineHeight float64
	Ascender float64
	Descender float64
	UnderlineY float64
	UnderlineThickness float64
}

type AtlasRect struct {
	Left, Bottom, Right, Top float64
}

type GlyphData struct {
	Unicode int
	Advance float64
	PlaneBounds AtlasRect
	AtlasBounds AtlasRect
}

type SdfAtlas struct {
	Atlas SdfAtlasPreamble
	Metrics SdfMetrics
	Glyphs []GlyphData
}

// Multiply by font size to get in pixels, then divide by texture size
func sdfUnitToFloat(sdfUnit float64, fontSize, texSize int) float64 {
	return (sdfUnit * float64(fontSize)) / float64(texSize)
}

func AtlasFromSdf(sdf SdfAtlas, sdfImg *image.NRGBA) (*Atlas, error) {
	texture := NewTexture(sdfImg, true) // TODO: Smoothing for sdf?

	// height := sdfUnitToFloat(sdf.Metrics.LineHeight, sdf.Atlas.Size, sdf.Atlas.Width);
	height := sdf.Metrics.LineHeight * float64(sdf.Atlas.Size)

	ascent := sdf.Metrics.Ascender * float64(sdf.Atlas.Size)
	descent := sdf.Metrics.Descender * float64(sdf.Atlas.Size)
	atlas := &Atlas{
		mapping: make(map[rune]Glyph),
		ascent: floatToFixed(ascent),//floatToFixed(sdf.Metrics.Ascender),
		descent: floatToFixed(descent), //floatToFixed(sdf.Metrics.Descender),
		height: floatToFixed(height),
		texture: texture,
		pixelPerfect: true,
		defaultKerning: 0,
	}

	for _, g := range sdf.Glyphs {
		// pb := R(
		// 	sdfUnitToFloat(g.PlaneBounds.Left, sdf.Atlas.Size, sdf.Atlas.Width),
		// 	sdfUnitToFloat(g.PlaneBounds.Bottom, sdf.Atlas.Size, sdf.Atlas.Height),
		// 	sdfUnitToFloat(g.PlaneBounds.Right, sdf.Atlas.Size, sdf.Atlas.Width),
		// 	sdfUnitToFloat(g.PlaneBounds.Top, sdf.Atlas.Size, sdf.Atlas.Height),
		// )
		// ww := pb.W() * float64(sdf.Atlas.Width)
		// if ww != g.AtlasBounds.Right - g.AtlasBounds.Left {
		// 	fmt.Println("Failed:", rune(g.Unicode), ww, g.AtlasBounds.Right - g.AtlasBounds.Left)
		// 	panic("aslkfdjsalfd")
		// }

		// bearingRect := image.Rect(
		// 	int(g.PlaneBounds.Left),
		// )
		// bearingX := float64((bearingRect.Min.X * 1000000).Floor()) / (1000000 * fSize)
		// bearingY := float64((-bearingRect.Max.Y * 1000000).Floor()) / (1000000 * fSize)
		bearingX := sdfUnitToFloat(g.PlaneBounds.Left, sdf.Atlas.Size, sdf.Atlas.Width)
		bearingY := sdfUnitToFloat(-g.PlaneBounds.Top, sdf.Atlas.Size, sdf.Atlas.Height)

		glyph := Glyph{
			// Advance: (g.Advance * float64(sdf.Atlas.Size)) / float64(sdf.Atlas.Width),
			Advance: sdfUnitToFloat(g.Advance, sdf.Atlas.Size, sdf.Atlas.Width),
			Bearing: Vec2{bearingX, bearingY},
			BoundsUV: R(
				g.AtlasBounds.Left / float64(sdf.Atlas.Width),
				g.AtlasBounds.Bottom / float64(sdf.Atlas.Height),
				g.AtlasBounds.Right / float64(sdf.Atlas.Width),
				g.AtlasBounds.Top / float64(sdf.Atlas.Height),
			),
		}
		atlas.mapping[rune(g.Unicode)] = glyph
	}

	return atlas, nil
}
