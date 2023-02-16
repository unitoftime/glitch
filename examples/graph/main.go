package main

import (
	"fmt"
	"math"
	"time"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/graph"
	// "github.com/unitoftime/glitch/ui"
	"github.com/unitoftime/glitch/shaders"
)

func check(err error) {
	if err != nil { panic(err) }
}

func main() {
	fmt.Println("Starting")
	glitch.Run(runGame)
}

func runGame() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch - Graph Demo", glitch.WindowConfig{
		Vsync: true,
		Samples: 4,
	})
	check(err)

	shader, err := glitch.NewShader(shaders.SpriteShader)
	if err != nil { panic(err) }

	pass := glitch.NewRenderPass(shader)

	dat := make([]glitch.Vec2, 0)
	for i := 0; i < 1000; i++ {
		dat = append(dat, glitch.Vec2{float64(i)/100.0, float64(math.Sin(float64(i) / 100.0))})
	}

	// lightBlue := glitch.RGBA{0x8a, 0xeb, 0xf1, 0xff}
	// pink := color.NRGBA{0xcd, 0x60, 0x93, 0xff}

	// pad := float32(50)
	// rect := glitch.R(0 + pad, 0 + pad, 1920 - pad, 1080 - pad)
	// rect := glitch.R(0, 0, 1, 1)
	rect := win.Bounds()
	rect = glitch.R(rect.Min[0], rect.Min[1], rect.Min[0] + 500, rect.Min[1] + 500)

	graph := graph.NewGraph(rect)

	camera := glitch.NewCameraOrtho()

	dt := 15 * time.Millisecond
	index := 0
	start := time.Now()
	for !win.Pressed(glitch.KeyBackspace) {

		pass.Clear()
		mat := glitch.Mat4Ident
		// mat = *mat.Scale(100, 100, 100)
		graph.Clear()
		graph.Line(dat)
		graph.DrawColorMask(pass, mat, glitch.RGBA{0, 1, 1, 1})

		glitch.Clear(win, glitch.RGBA{0, 0, 0, 1.0})

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, 1, 1)
		pass.SetUniform("projection", camera.Projection)
		pass.SetUniform("view", camera.View)
		pass.Draw(win)

		win.Update()

		dt = time.Since(start)
		dat[index][1] = float64(dt.Seconds())
		index = (index + 1) % len(dat)
		start = time.Now()
	}
}
