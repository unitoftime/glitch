package main

// Try: https://www.shadertoy.com/view/csX3RH

import (
	"flag"
	"fmt"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/examples/assets"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		go func() {
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
		}()
		defer pprof.StopCPUProfile()
	}

	glitch.Run(runGame)

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func runGame() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch - Gophermark", glitch.WindowConfig{
		Vsync: true,
	})
	if err != nil {
		panic(err)
	}

	// shader, err := glitch.NewShader(shaders.SpriteShader)
	// if err != nil {
	// 	panic(err)
	// }

	// pass := glitch.NewRenderPass(shader)
	// pass.DepthTest = true

	manImage, err := assets.LoadImage("gopher.png")
	if err != nil {
		panic(err)
	}
	texture := glitch.NewTexture(manImage, false)
	manSprite := glitch.NewSprite(texture, texture.Bounds())

	length := 2000
	man := make([]Man, length)
	for i := range man {
		man[i] = NewMan()
	}

	w := float64(160.0) / 4
	h := float64(200.0) / 4

	// Text
	atlas, err := glitch.DefaultAtlas()
	if err != nil {
		panic(err)
	}

	text := atlas.Text("", 1)

	min := time.Duration(0)
	max := time.Duration(0)

	counter := 0
	camera := glitch.NewCameraOrtho()
	// camera.DepthRange = glitch.Vec2{-127, 127}

	start := time.Now()
	var dt time.Duration

	// geom := glitch.NewGeomDraw()
	// geomRect := glitch.R(-16, -16, 16, 16)
	// geomMesh := glitch.NewQuadMesh(geomRect, geomRect)

	mat := glitch.Mat4Ident
	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}
		start = time.Now()

		counter = (counter + 1) % 60

		for i := range man {
			man[i].position.X += man[i].velocity.X
			man[i].position.Y += man[i].velocity.Y

			if man[i].position.X <= 0 || (man[i].position.X+w) >= float64(1920) {
				man[i].velocity.X = -man[i].velocity.X
			}
			if man[i].position.Y <= 0 || (man[i].position.Y+h) >= float64(1080) {
				man[i].velocity.Y = -man[i].velocity.Y
			}
		}

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, 1.0, 1.0)
		glitch.SetCamera(camera)

		glitch.Clear(win, glitch.RGBA{R: 0.1, G: 0.2, B: 0.3, A: 1.0})

		// geom.Clear()
		for i := range man {
			mat = glitch.Mat4Ident
			mat.Scale(0.25, 0.25, 1.0).Translate(man[i].position.X, man[i].position.Y, 0)
			manSprite.DrawColorMask(win, mat, man[i].color)
			// geom.DrawRect(pass, geomRect, mat, man[i].color)
			// geomMesh.DrawColorMask(pass, mat, man[i].color)
			// geom.DrawRect2(geomRect, mat, man[i].color)
		}
		// geom.Draw(pass, glitch.Mat4Ident)

		if counter == 0 {
			text.Clear()
			text.Set(fmt.Sprintf("%2.2f (%2.2f, %2.2f) ms",
				1000*dt.Seconds(),
				1000*min.Seconds(),
				1000*max.Seconds()))
			min = 100000000000
			max = 0

			metrics := glitch.GetMetrics()
			fmt.Printf("%+v\n", metrics)
		}
		text.DrawColorMask(win, glitch.Mat4Ident, glitch.White)

		win.Update()

		dt = time.Since(start)

		// dt = time.Since(start)
		if dt > max {
			max = dt
		}
		if dt < min {
			min = dt
		}
		// fmt.Println(dt.Seconds() * 1000)
	}
}

type Man struct {
	position, velocity glitch.Vec2
	color              glitch.RGBA
	layer              uint8
}

func NewMan() Man {
	colors := []glitch.RGBA{
		glitch.RGBA{R: 1.0, G: 0, B: 0, A: 1.0},
		glitch.RGBA{R: 0, G: 1.0, B: 0, A: 1.0},
		glitch.RGBA{R: 0, G: 0, B: 1.0, A: 1.0},
	}
	randIndex := rand.Intn(len(colors))
	vScale := 5.0
	return Man{
		// position: mgl32.Vec2{100, 100},
		// position: mgl32.Vec2{float32(float64(width/2) * rand.Float64()),
		// 	float32(float64(height/2) * rand.Float64())},
		position: glitch.Vec2{1920 / 2, 1080 / 2},
		velocity: glitch.Vec2{float64(2 * vScale * (rand.Float64() - 0.5)),
			float64(2 * vScale * (rand.Float64() - 0.5))},
		color: colors[randIndex],
		layer: uint8(randIndex) + 1,
	}
}
