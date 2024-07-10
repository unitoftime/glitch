package main

import (
	"flag"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

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
// func main() {
// 	log.Println("Begin")
// 	glitch.Run(runGame)
// }

func runGame() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
		Vsync:   true,
		Samples: 0,
	})
	if err != nil {
		panic(err)
	}

 	atlasImg, err := assets.LoadImage("atlas-msdf.png")
	if err != nil {
		panic(err)
	}
	atlasJson := glitch.SdfAtlas{}
	err = assets.LoadJson("atlas-msdf.json", &atlasJson)
	if err != nil {
		panic(err)
	}

	atlas, err := glitch.AtlasFromSdf(atlasJson, atlasImg)

	// Text
	// atlas, err := glitch.BasicFontAtlas()
	if err != nil {
		panic(err)
	}

	// text := atlas.Text("Hello World", 1.0)
	text := atlas.Text("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 1.0)
	//"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	screenScale := 1.0 // This is just a weird scaling number

	// A screenspace camera
	camera := glitch.NewCameraOrtho()
	camera.SetOrtho2D(win.Bounds())
	camera.SetView2D(0, 0, screenScale, screenScale)

	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, screenScale, screenScale)
		glitch.SetCamera(camera)

		// mx, my := win.MousePosition()
		// log.Println("Mouse: ", mx, my)

		glitch.Clear(win, glitch.Greyscale(0.5))


		// mat := glitch.Mat4Ident
		// mat.Translate(win.Bounds().Center()[0], win.Bounds().Center()[1], 0)
		// text.Draw(pass, mat)

		lh := atlas.LineHeight()
		y := 0.0
		scale := 0.1
		for i := 0; i < 25; i++ {
			mat := glitch.Mat4Ident
			mat.
				Scale(scale, scale, 1).
				Translate(0, y, 0)
			text.Draw(win, mat)

			y += lh * scale
			scale += 0.1
		}

		// pass.SetUniform("u_threshold", float32(0.5))
		// pass.SetCamera2D(camera)
		// pass.Draw(win)

		win.Update()
	}
}
