package main

import (
	"fmt"
	"log"
	"embed"
	"image"
	"image/draw"
	_ "image/png"
	"time"
	"math/rand"
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
		Vsync: false,
	})
	if err != nil { panic(err) }

	shader, err := glitch.NewShader(shaders.SpriteShader)
	if err != nil { panic(err) }

	pass := glitch.NewRenderPass(shader)

	manImage, err := loadImage("gopher.png")
	if err != nil {
		panic(err)
	}
	// texture := glitch.NewTexture(160, 200, manImage.Pix)
	texture := glitch.NewTexture(manImage, false)
	// texture := glitch.NewTexture(manImage.Bounds().Dx(), manImage.Bounds().Dy(), manImage.Pix)

	// mesh := glitch.NewQuadMesh()
	x := float32(0)
	y := float32(0)
	manSprite := glitch.NewSprite(texture, glitch.R(x, y, x+160, y+200))

	length := 100000
	man := make([]Man, length)
	for i := range man {
		man[i] = NewMan()
	}

	w := float32(160.0)/4
	h := float32(200.0)/4

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

	camera := glitch.NewCameraOrtho()
	start := time.Now()
	var dt time.Duration
	for !win.ShouldClose() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}
		start = time.Now()
		for i := range man {
			man[i].position[0] += man[i].velocity[0]
			man[i].position[1] += man[i].velocity[1]

			if man[i].position[0] <= 0 || (man[i].position[0]+w) >= float32(1920) {
				man[i].velocity[0] = -man[i].velocity[0]
			}
			if man[i].position[1] <= 0 || (man[i].position[1]+h) >= float32(1080) {
				man[i].velocity[1] = -man[i].velocity[1]
			}
		}

		pass.Clear()

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, 1.0, 1.0)

		for i := range man {
			mat := glitch.Mat4Ident
			// mat.Scale(0.25, 0.25, 1.0).Translate(man[i].position[0], man[i].position[1], -man[i].position[1])
			mat.Scale(0.25, 0.25, 1.0).Translate(man[i].position[0], man[i].position[1], 0)

			// mesh.DrawColorMask(pass, mat, glitch.RGBA{0.5, 1.0, 1.0, 1.0})
			pass.SetLayer(man[i].layer)
			manSprite.DrawColorMask(pass, mat, man[i].color)
			// manSprite.DrawColorMask(pass, mat, glitch.RGBA{1.0, 1.0, 1.0, 1.0})
		}

		mat := glitch.Mat4Ident
		mat.Translate(0, 0, 0)
		pass.SetLayer(0)
		text.Set(fmt.Sprintf("%2.2f ms", 1000*dt.Seconds()))
		text.DrawColorMask(pass, mat, glitch.RGBA{1.0, 1.0, 0.0, 1.0})

		glitch.Clear(win, glitch.RGBA{0.1, 0.2, 0.3, 1.0})

		pass.SetUniform("projection", camera.Projection)
		pass.SetUniform("view", camera.View)
		pass.Draw(win)

		win.Update()

		dt = time.Since(start)
		// fmt.Println(dt.Seconds() * 1000)
	}
}

type Man struct {
	position, velocity glitch.Vec2
	color glitch.RGBA
	layer uint8
}
func NewMan() Man {
	colors := []glitch.RGBA{
		glitch.RGBA{1.0, 0, 0, 1.0},
		glitch.RGBA{0, 1.0, 0, 1.0},
		glitch.RGBA{0, 0, 1.0, 1.0},
	}
	randIndex := rand.Intn(len(colors))
	vScale := 5.0
	return Man{
		// position: mgl32.Vec2{100, 100},
		// position: mgl32.Vec2{float32(float64(width/2) * rand.Float64()),
		// 	float32(float64(height/2) * rand.Float64())},
		position: glitch.Vec2{1920/2, 1080/2},
		velocity: glitch.Vec2{float32(2*vScale * (rand.Float64()-0.5)),
			float32(2*vScale * (rand.Float64()-0.5))},
		color: colors[randIndex],
		layer: uint8(randIndex) + 1,
	}
}
