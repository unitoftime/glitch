package glitch

import (
	"testing"

	"github.com/unitoftime/flow/glm"
)

// These benchmarks measure the CPU side of draw submission: building the GeometryFiller for each draw,
// adding it to a Sorter, sorting, and draining the commands into a BatchTarget.
// They don't need a window or GL context, because nothing is filled or drawn on the GPU.
// Each benchmark op is one simulated frame of drawsPerFrame draws.

const drawsPerFrame = 1000

// A BatchTarget that only counts draws, so the benchmark captures the cost of passing commands through Add
type nopTarget struct {
	count int
}

func (t *nopTarget) Add(filler GeometryFiller, mat glMat4, mask RGBA, material Material) {
	t.count++
}

// Builds a texture/material without touching GL
func benchMaterial() Material {
	return Material{
		texture: &Texture{width: 512, height: 512},
		blend:   BlendModeNormal,
	}
}

func benchSprite() *Sprite {
	frame := glm.R(0, 0, 32, 32)
	return &Sprite{
		mesh:     NewSpriteMesh(frame.W(), frame.H(), glm.R(0, 0, 32.0/512, 32.0/512)),
		bounds:   frame.Moved(frame.Center().Scaled(-1)),
		frame:    frame,
		material: benchMaterial(),
	}
}

func benchQuad() Quad {
	return Quad{
		Frame:    glm.R(0, 0, 32, 32),
		material: benchMaterial(),
	}
}

func benchText() textDraw {
	return textDraw{
		atlas: &Atlas{defaultMaterial: benchMaterial()},
		text:  "Hello",
		scale: 1.0,
	}
}

// Deterministic, spread out positions so that the sorter does real work
func benchMatrix(i int) Mat4 {
	x := float64((i * 7919) % 1920)
	y := float64((i * 104729) % 1080)
	mat := Mat4Ident
	mat.Translate(x, y, 0)
	return mat
}

func newBenchSorter() *Sorter {
	sorter := NewSorter()
	sorter.SoftwareSort = SoftwareSortY
	return sorter
}

func BenchmarkFrameSprites(b *testing.B) {
	sorter := newBenchSorter()
	target := &nopTarget{}
	sprite := benchSprite()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < drawsPerFrame; i++ {
			sprite.DrawColorMask(sorter, benchMatrix(i), White)
		}
		sorter.Draw(target)
	}
}

func BenchmarkFrameQuads(b *testing.B) {
	sorter := newBenchSorter()
	target := &nopTarget{}
	quad := benchQuad()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < drawsPerFrame; i++ {
			quad.DrawColorMask(sorter, benchMatrix(i), White)
		}
		sorter.Draw(target)
	}
}

func BenchmarkFrameText(b *testing.B) {
	sorter := newBenchSorter()
	target := &nopTarget{}
	text := benchText()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < drawsPerFrame; i++ {
			text.DrawColorMask(sorter, benchMatrix(i), White)
		}
		sorter.Draw(target)
	}
}

// Roughly mog's mix: mostly sprites, with some quads (armor/boots/swimming) and some text (dialogue)
func BenchmarkFrameMixed(b *testing.B) {
	sorter := newBenchSorter()
	target := &nopTarget{}
	sprite := benchSprite()
	quad := benchQuad()
	text := benchText()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < drawsPerFrame; i++ {
			mat := benchMatrix(i)
			switch {
			case i%10 < 7:
				sprite.DrawColorMask(sorter, mat, White)
			case i%10 < 9:
				quad.DrawColorMask(sorter, mat, White)
			default:
				text.DrawColorMask(sorter, mat, White)
			}
		}
		sorter.Draw(target)
	}
}
