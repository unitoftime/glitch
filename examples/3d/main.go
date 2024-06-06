package main

import (
	"embed"
	"flag"
	"fmt"
	"image"
	"image/draw"
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

//go:embed gopher.png
var f embed.FS

func loadImage(path string) (*image.NRGBA, error) {
	file, err := f.Open(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
	return nrgba, nil
}

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

	shader, err := glitch.NewShader(shaders.SpriteShader)
	if err != nil {
		panic(err)
	}

	pass := glitch.NewRenderPass(shader)

	diffuseShader, err := glitch.NewShader(shaders.DiffuseShader)
	if err != nil {
		panic(err)
	}
	diffusePass := glitch.NewRenderPass(diffuseShader)
	diffusePass.DepthTest = true

	manImage, err := loadImage("gopher.png")
	if err != nil {
		panic(err)
	}
	// texture := glitch.NewTexture(160, 200, manImage.Pix)
	texture := glitch.NewTexture(manImage, false)
	// texture := glitch.NewTexture(manImage.Bounds().Dx(), manImage.Bounds().Dy(), manImage.Pix)

	// mesh := glitch.NewQuadMesh()
	x := 0.0
	y := 0.0
	manSprite := glitch.NewSprite(texture, glitch.R(x, y, x+160, y+200))

	// Text
	atlas, err := glitch.DefaultAtlas()
	if err != nil {
		panic(err)
	}

	text := atlas.Text("hello world", 1)

	cube := glitch.NewModel(glitch.NewCubeMesh(50), glitch.DefaultMaterial())

	camera := glitch.NewCameraOrtho()
	pCam := glitch.NewCamera()
	start := time.Now()

	geom := glitch.NewGeomDraw()
	quad := geom.FillRect(glitch.R(0, 0, 100, 100))
	quadModel := glitch.NewModel(quad, glitch.NewSpriteMaterial(texture))

	tt := 0.0
	var dt time.Duration
	for !win.Closed() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}
		start = time.Now()

		pass.Clear()
		diffusePass.Clear()

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, 1.0, 1.0)

		tt += dt.Seconds()
		// pCam.Position = glitch.Vec3{float32(100 * math.Cos(tt)), float32(100 * math.Sin(tt)), 50}
		pCam.Position = glitch.Vec3{100 * math.Cos(0), 100 * math.Sin(0), 50}
		pCam.Target = glitch.Vec3{0, 0, 0}

		pCam.SetPerspective(win)
		pCam.SetViewLookAt(win)

		mat := glitch.Mat4Ident
		mat.Scale(0.25, 0.25, 1.0).Translate(100, 100, 0)

		pass.SetLayer(0)
		manSprite.DrawColorMask(pass, mat, glitch.RGBA{R: 1, G: 1, B: 1, A: 1})
		quadModel.Draw(pass, glitch.Mat4Ident)

		cubeMat := glitch.Mat4Ident
		cubeMat = *cubeMat.Translate(0, 0, 0).Rotate(float64(tt), glitch.Vec3{0, 0, 1})

		cube.Draw(diffusePass, cubeMat)

		mat = glitch.Mat4Ident
		mat.Translate(0, 0, 0)
		text.Set(fmt.Sprintf("%2.2f ms", 1000*dt.Seconds()))
		text.DrawColorMask(pass, mat, glitch.RGBA{R: 1.0, G: 1.0, B: 0.0, A: 1.0})

		glitch.Clear(win, glitch.RGBA{R: 0.1, G: 0.2, B: 0.3, A: 1.0})

		pass.SetUniform("projection", camera.Projection)
		pass.SetUniform("view", camera.View)
		pass.Draw(win)

		diffusePass.SetUniform("projection", pCam.Projection)
		diffusePass.SetUniform("view", pCam.View)
		// diffusePass.SetUniform("model", cubeMat)

		diffusePass.SetUniform("viewPos", pCam.Position)

		diffusePass.SetUniform("material.ambient", glitch.Vec3{1, 0.5, 0.31})
		diffusePass.SetUniform("material.diffuse", glitch.Vec3{1, 0.5, 0.31})
		diffusePass.SetUniform("material.specular", glitch.Vec3{1, 0.5, 0.31})
		diffusePass.SetUniform("material.shininess", float32(32.0))

		diffusePass.SetUniform("dirLight.direction", glitch.Vec3{0, 1, 0})
		diffusePass.SetUniform("dirLight.ambient", glitch.Vec3{0.5, 0.5, 0.5})
		diffusePass.SetUniform("dirLight.diffuse", glitch.Vec3{0.5, 0.5, 0.5})
		diffusePass.SetUniform("dirLight.specular", glitch.Vec3{0.5, 0.5, 0.5})
		diffusePass.Draw(win)

		win.Update()

		dt = time.Since(start)
		// fmt.Println(dt.Seconds() * 1000)
	}
}
