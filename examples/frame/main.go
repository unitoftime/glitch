package main

import (
	"fmt"
	_ "image/png"
	"math/rand"
	"time"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/examples/assets"
)

func main() {
	glitch.Run(runGame)
}

func runGame() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch - Framebuffer", glitch.WindowConfig{
		Vsync: false,
	})
	if err != nil {
		panic(err)
	}

	// shader, err := glitch.NewShader(shaders.SpriteShader)
	// if err != nil {
	// 	panic(err)
	// }

	manImage, err := assets.LoadImage("gopher.png")
	if err != nil {
		panic(err)
	}
	texture := glitch.NewTexture(manImage, false)

	x := 0.0
	y := 0.0
	manSprite := glitch.NewSprite(texture, glm.R(x, y, x+160, y+200))

	length := 100000
	man := make([]Man, length)
	for i := range man {
		man[i] = NewMan()
	}

	w := 160.0 / 4.0
	h := 200.0 / 4.0

	// Text
	atlas, err := glitch.DefaultAtlas()
	if err != nil {
		panic(err)
	}

	text := atlas.Text("hello world", 1)

	fmt.Println(win.Bounds())
	frame := glitch.NewFrame(win.Bounds(), false)

	camera := glitch.NewCameraOrtho()
	start := time.Now()
	var dt time.Duration
	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}
		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, 1, 1)
		glitch.SetCamera(camera)

		start = time.Now()
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

		glitch.Clear(frame, glm.Greyscale(0.5))

		for i := range man {
			mat := glitch.Mat4Ident
			// mat.Scale(0.25, 0.25, 1.0).Translate(man[i].position.X, man[i].position.Y, -man[i].position.Y)
			mat.Scale(0.25, 0.25, 1.0).Translate(man[i].position.X, man[i].position.Y, 0)

			// mesh.DrawColorMask(pass, mat, glitch.RGBA{0.5, 1.0, 1.0, 1.0})
			// pass.SetLayer(man[i].layer)
			manSprite.DrawColorMask(frame, mat, man[i].color)
			// manSprite.DrawColorMask(pass, mat, glitch.RGBA{1.0, 1.0, 1.0, 1.0})
		}

		mat := glitch.Mat4Ident

		text.Set(fmt.Sprintf("%2.2f ms", 1000*dt.Seconds()))
		text.DrawColorMask(frame, mat, glitch.Black)

		glitch.Clear(win, glm.White)
		frame.Draw(win, glitch.Mat4Ident)
		win.Update()

		dt = time.Since(start)
	}
}

type Man struct {
	position, velocity glitch.Vec2
	color              glitch.RGBA
	layer              int8
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
		layer: int8(randIndex) + 1,
	}
}
