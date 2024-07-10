package main

import (
	"flag"
	"fmt"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/shaders"
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
			time.Sleep(10 * time.Second)
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
	win, err := glitch.NewWindow(1920, 1080, "Glitch", glitch.WindowConfig{
		Vsync:   true,
		Samples: 8,
	})
	if err != nil {
		panic(err)
	}

	diffuseShader, err := glitch.NewShader(shaders.DiffuseShader)
	if err != nil {
		panic(err)
	}

	// Text
	atlas, err := glitch.DefaultAtlas()
	if err != nil {
		panic(err)
	}

	text := atlas.Text("hello world", 1)

	diffuseMaterial := glitch.NewMaterial(diffuseShader)
	diffuseMaterial.SetDepthMode(glitch.DepthModeLess)
	diffuseMaterial.SetCullMode(glitch.CullModeNormal)

	diffuseMaterial.SetUniform("material.ambient", glitch.Vec3{1, 0.5, 0.31})
	diffuseMaterial.SetUniform("material.diffuse", glitch.Vec3{1, 0.5, 0.31})
	diffuseMaterial.SetUniform("material.specular", glitch.Vec3{1, 0.5, 0.31})
	diffuseMaterial.SetUniform("material.shininess", float32(32.0))

	diffuseMaterial.SetUniform("dirLight.direction", glitch.Vec3{0, 1, 0})
	diffuseMaterial.SetUniform("dirLight.ambient", glitch.Vec3{0.5, 0.5, 0.5})
	diffuseMaterial.SetUniform("dirLight.diffuse", glitch.Vec3{0.5, 0.5, 0.5})
	diffuseMaterial.SetUniform("dirLight.specular", glitch.Vec3{0.5, 0.5, 0.5})

	cube := glitch.NewModel(glitch.NewCubeMesh(50), diffuseMaterial)

	camera := glitch.NewCameraOrtho()
	pCam := glitch.NewCamera()
	start := time.Now()

	tt := 0.0
	var dt time.Duration
	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}
		start = time.Now()

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, 1.0, 1.0)

		tt += dt.Seconds()
		// pCam.Position = glitch.Vec3{float32(100 * math.Cos(tt)), float32(100 * math.Sin(tt)), 50}
		pCam.Position = glitch.Vec3{100 * math.Cos(0), 100 * math.Sin(0), 50}
		pCam.Target = glitch.Vec3{0, 0, 0}

		pCam.SetPerspective(win)
		pCam.SetViewLookAt(win)

		glitch.Clear(win, glitch.RGBA{R: 0.1, G: 0.2, B: 0.3, A: 1.0})

		glitch.SetCameraMaterial(pCam.Material())
		diffuseShader.SetUniform("viewPos", pCam.Position) // TODO: This needs to be better encapsulated somehow?
		{
			mat := glitch.Mat4Ident
			mat.Scale(0.25, 0.25, 1.0).Translate(100, 100, 0)

			cubeMat := glitch.Mat4Ident
			cubeMat = *cubeMat.Translate(0, 0, 0).Rotate(float64(tt), glitch.Vec3{0, 0, 1})
			cube.Draw(win, cubeMat)
		}

		glitch.SetCamera(camera)
		{
			mat := glitch.Mat4Ident
			mat.Translate(0, 0, 0)
			text.Set(fmt.Sprintf("%2.2f ms", 1000*dt.Seconds()))
			text.DrawColorMask(win, mat, glitch.RGBA{R: 1.0, G: 1.0, B: 0.0, A: 1.0})
		}



		win.Update()

		dt = time.Since(start)
		// fmt.Println(dt.Seconds() * 1000)
	}
}
