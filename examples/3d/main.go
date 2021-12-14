package main

import (
	"fmt"
	"log"
	"embed"
	"image"
	"image/draw"
	_ "image/png"
	"time"
	"runtime"
	"runtime/pprof"
	"flag"
	"os"

	"unicode"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/shaders"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
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
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func runGame() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch", glitch.WindowConfig{
		Vsync: true,
	})
	if err != nil { panic(err) }

	shader, err := glitch.NewShader(shaders.SpriteShader)
	if err != nil { panic(err) }

	pass := glitch.NewRenderPass(shader)

	diffuseShader, err := glitch.NewShader(shaders.DiffuseShader)
	if err != nil { panic(err) }
	diffusePass := glitch.NewRenderPass(diffuseShader)

	manImage, err := loadImage("gopher.png")
	if err != nil {
		panic(err)
	}
	// texture := glitch.NewTexture(160, 200, manImage.Pix)
	texture := glitch.NewTexture(manImage)
	// texture := glitch.NewTexture(manImage.Bounds().Dx(), manImage.Bounds().Dy(), manImage.Pix)

	// mesh := glitch.NewQuadMesh()
	x := float32(0)
	y := float32(0)
	manSprite := glitch.NewSprite(texture, glitch.R(x, y, x+160, y+200))

	// Text
	// TODO - use this instead of hardcoding
	runes := make([]rune, unicode.MaxASCII - 32)
	for i := range runes {
		runes[i] = rune(32 + i)
	}
	font, err := truetype.Parse(goregular.TTF)
	atlas := glitch.NewAtlas(
		truetype.NewFace(font, &truetype.Options{
			Size: 64,
			GlyphCacheEntries: 1,
		}),
		runes)

	text := atlas.Text("hello world")

	cube := glitch.NewModel(glitch.NewCubeMesh(100), nil)

	camera := glitch.NewCameraOrtho()
	pCam := glitch.NewCamera()
	start := time.Now()
	var dt time.Duration
	for !win.ShouldClose() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}
		start = time.Now()

		pass.Clear()
		diffusePass.Clear()

		camera.SetOrtho2D(win)
		camera.SetView2D(0, 0, 1.0, 1.0)

		pCam.SetPerspective(win)
		pCam.SetViewLookAt(win)

		mat := glitch.Mat4Ident
		mat.Scale(0.25, 0.25, 1.0).Translate(100, 100, 0)

		pass.SetLayer(glitch.DefaultLayer)
		manSprite.DrawColorMask(pass, mat, glitch.RGBA{1, 1, 1, 1})

		mat = glitch.Mat4Ident
		mat.Translate(200, 200, 0)
		cube.Draw(diffusePass, mat)

		mat = glitch.Mat4Ident
		mat.Translate(0, 0, 0)
		text.Set(fmt.Sprintf("%2.2f ms", 1000*dt.Seconds()))
		text.DrawColorMask(pass, mat, glitch.RGBA{1.0, 1.0, 0.0, 1.0})

		glitch.Clear(glitch.RGBA{0.1, 0.2, 0.3, 1.0})

		pass.SetUniform("projection", camera.Projection)
		pass.SetUniform("view", camera.View)
		pass.Draw(win)

		diffusePass.SetUniform("projection", pCam.Projection)
		diffusePass.SetUniform("view", pCam.View)
		diffusePass.Draw(win)

		win.Update()

		dt = time.Since(start)
		fmt.Println(dt.Seconds() * 1000)
	}
}
