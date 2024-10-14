package main

// Try: https://www.shadertoy.com/view/csX3RH

import (
	"fmt"
	_ "image/png"
	"math"
	"math/rand"
	"time"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/examples/assets"
)

func main() {
	glitch.Run(run)
}

func run() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch - PixelArt", glitch.WindowConfig{
		Vsync: true,
	})
	if err != nil {
		panic(err)
	}

	// // shader, err := glitch.NewShader(shaders.PixelArtShader)
	// shader, err := glitch.NewShader(shaders.SpriteShader)
	// if err != nil {
	// 	panic(err)
	// }
	// pass := glitch.NewRenderPass(shader)

	zoom := 1.0

	img, err := assets.LoadImage("gopher-small.png")
	if err != nil {
		panic(err)
	}
	texture := glitch.NewTexture(img, false)
	sprite := glitch.NewSprite(texture, texture.Bounds())

	targetBounds := win.Bounds()

	length := 10
	man := make([]Man, length)
	for i := range man {
		man[i] = NewMan(targetBounds.Center())
	}

	// w := sprite.Bounds().W()
	// h := sprite.Bounds().H()

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
	camera.DepthRange = glitch.Vec2{-127, 127}

	start := time.Now()
	var dt time.Duration

	mat := glitch.Mat4Ident
	var t time.Duration
	for !win.Closed() {
		targetBounds = win.Bounds()

		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}

		camera.SetOrtho2D(targetBounds)
		camera.SetView2D(0, 0, zoom, zoom)
		glitch.SetCamera(camera)

		_, sy := win.MouseScroll()
		if sy > 0 {
			zoom += 0.1
		} else if sy < 0 {
			zoom -= 0.1
		}

		start = time.Now()
		t += dt

		counter = (counter + 1) % 60

		radius := 50.0
		man[0].position.X = radius*math.Cos(t.Seconds()) + targetBounds.Center().X
		man[0].position.Y = radius*math.Sin(t.Seconds()) + targetBounds.Center().Y

		man[1].position.X = radius*math.Cos(t.Seconds()) + targetBounds.Center().X
		man[1].position.Y = (2 * radius) + targetBounds.Center().Y

		man[2].position.X = (2 * radius) + targetBounds.Center().X
		man[2].position.Y = radius*math.Sin(t.Seconds()) + targetBounds.Center().Y

		glitch.Clear(win, glitch.White)

		if counter == 0 {
			text.Clear()
			text.Set(fmt.Sprintf("%2.2f (%2.2f, %2.2f) ms",
				1000*dt.Seconds(),
				1000*min.Seconds(),
				1000*max.Seconds()))
			min = 100000000000
			max = 0
		}
		text.DrawColorMask(win, glitch.Mat4Ident, glitch.Black)

		for i := range man {
			mat = glitch.Mat4Ident
			// mat.Translate(math.Round(man[i].position[0]), math.Round(man[i].position[1]), 0)
			mat.Scale(4, 4, 1).Translate(math.Round(man[i].position.X), math.Round(man[i].position.Y), 0)
			// mat.Scale(1, 1, 1).Translate(man[i].position[0], man[i].position[1], 0)
			sprite.DrawColorMask(win, mat, man[i].color)
		}

		// texelsPerPixel := 1.0 / zoom
		// pass.SetUniform("texelsPerPixel", float32(texelsPerPixel))

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

func NewMan(pos glitch.Vec2) Man {
	vScale := 0.1
	return Man{
		position: pos,
		velocity: glitch.Vec2{float64(2 * vScale * (rand.Float64() - 0.5)),
			float64(2 * vScale * (rand.Float64() - 0.5))},
		color: glitch.White,
	}
}

//--------------------------------------------------------------------------------
// func main() {
// 	glitch.Run(run)
// }

// func run() {
// 	win, err := glitch.NewWindow(1920, 1080, "Glitch - PixelArt", glitch.WindowConfig{
// 		Vsync: true,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	pixelShader, err := glitch.NewShader(shaders.PixelArtShader)
// 	if err != nil {
// 		panic(err)
// 	}
// 	pixelPass := glitch.NewRenderPass(pixelShader)

// 	shader, err := glitch.NewShader(shaders.SpriteShader)
// 	if err != nil {
// 		panic(err)
// 	}

// 	pass := glitch.NewRenderPass(shader)
// 	// pass.DepthTest = true
// 	// pass.SoftwareSort = glitch.SoftwareSortY

// 	pixelCam := glitch.NewCameraOrtho()

// 	zoom := 1.0

// 	upscale := 2.0
// 	targetBounds := glitch.R(0, 0, 1920, 1080).Scaled(1 / upscale).Snap()
// 	frame := glitch.NewFrame(targetBounds, false)

// 	img, err := assets.LoadImage("gopher-small.png")
// 	if err != nil {
// 		panic(err)
// 	}
// 	texture := glitch.NewTexture(img, false)
// 	sprite := glitch.NewSprite(texture, texture.Bounds())
// 	// sprite.Translucent = true

// 	length := 10
// 	man := make([]Man, length)
// 	for i := range man {
// 		man[i] = NewMan(targetBounds.Center())
// 	}

// 	// w := sprite.Bounds().W()
// 	// h := sprite.Bounds().H()

// 	// Text
// 	atlas, err := glitch.DefaultAtlas()
// 	if err != nil {
// 		panic(err)
// 	}

// 	text := atlas.Text("", 1)

// 	min := time.Duration(0)
// 	max := time.Duration(0)

// 	counter := 0
// 	camera := glitch.NewCameraOrtho()
// 	camera.DepthRange = glitch.Vec2{-127, 127}

// 	start := time.Now()
// 	var dt time.Duration

// 	mat := glitch.Mat4Ident
// 	var t time.Duration
// 	for !win.Closed() {
// 		if win.Pressed(glitch.KeyEscape) {
// 			win.Close()
// 		}
// 		_, sy := win.MouseScroll()
// 		if sy > 0 {
// 			zoom += 0.1
// 		} else if sy < 0 {
// 			zoom -= 0.1
// 		}

// 		start = time.Now()
// 		t += dt

// 		counter = (counter + 1) % 60

// 		radius := 50.0
// 		man[0].position[0] = radius * math.Cos(t.Seconds()) + targetBounds.Center()[0]
// 		man[0].position[1] = radius * math.Sin(t.Seconds()) + targetBounds.Center()[1]

// 		man[1].position[0] = radius * math.Cos(t.Seconds()) + targetBounds.Center()[0]
// 		man[1].position[1] = (2 * radius) + targetBounds.Center()[1]

// 		man[2].position[0] = (2 * radius) + targetBounds.Center()[0]
// 		man[2].position[1] = radius * math.Sin(t.Seconds()) + targetBounds.Center()[1]

// 		// for i := range man {
// 		// 	man[i].position[0] += man[i].velocity[0]
// 		// 	man[i].position[1] += man[i].velocity[1]

// 		// 	if man[i].position[0] <= 0 || (man[i].position[0]+w) >= float64(1920) {
// 		// 		man[i].velocity[0] = -man[i].velocity[0]
// 		// 	}
// 		// 	if man[i].position[1] <= 0 || (man[i].position[1]+h) >= float64(1080) {
// 		// 		man[i].velocity[1] = -man[i].velocity[1]
// 		// 	}
// 		// }

// 		pass.Clear()
// 		pixelPass.Clear()

// 		camera.SetOrtho2D(frame.Bounds())
// 		camera.SetView2D(0, 0, 1, 1)

// 		pixelCam.SetOrtho2D(win.Bounds())
// 		pixelCam.SetView2D(0, 0, zoom, zoom)

// 		pass.SetLayer(0)
// 		if counter == 0 {
// 			text.Clear()
// 			text.Set(fmt.Sprintf("%2.2f (%2.2f, %2.2f) ms",
// 				1000*dt.Seconds(),
// 				1000*min.Seconds(),
// 				1000*max.Seconds()))
// 			min = 100000000000
// 			max = 0
// 		}
// 		text.DrawColorMask(pass, glitch.Mat4Ident, glitch.Black)

// 		pass.SetLayer(1)
// 		for i := range man {
// 			mat = glitch.Mat4Ident
// 			// mat.Translate(math.Round(man[i].position[0]), math.Round(man[i].position[1]), 0)
// 			mat.Translate(man[i].position[0], man[i].position[1], 0)
// 			sprite.DrawColorMask(pass, mat, man[i].color)
// 		}

// 		glitch.Clear(win, glitch.White)
// 		glitch.Clear(frame, glitch.White)

// 		pass.SetCamera2D(camera)
// 		pass.Draw(frame)

// 		// frame.Draw(pixelPass, glitch.Mat4Ident)
// 		frame.RectDraw(pixelPass, win.Bounds())

// 		pixelPass.SetCamera2D(pixelCam)
// 		texelsPerPixel := 1.0 / upscale
// 		pixelPass.SetUniform("texelsPerPixel", float32(texelsPerPixel))
// 		pixelPass.Draw(win)

// 		win.Update()

// 		dt = time.Since(start)

// 		// dt = time.Since(start)
// 		if dt > max {
// 			max = dt
// 		}
// 		if dt < min {
// 			min = dt
// 		}
// 		// fmt.Println(dt.Seconds() * 1000)
// 	}
// }

// type Man struct {
// 	position, velocity glitch.Vec2
// 	color              glitch.RGBA
// 	layer              uint8
// }

// func NewMan(pos glitch.Vec2) Man {
// 	vScale := 0.1
// 	return Man{
// 		position: pos,
// 		velocity: glitch.Vec2{float64(2 * vScale * (rand.Float64() - 0.5)),
// 			float64(2 * vScale * (rand.Float64() - 0.5))},
// 		color: glitch.White,
// 	}
// }
