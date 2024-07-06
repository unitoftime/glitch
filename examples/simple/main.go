package main

import (
	_ "image/png"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/examples/assets"
	"github.com/unitoftime/glitch/shaders"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	glitch.Run(run)
}

func run() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch Demo", glitch.WindowConfig{
		Vsync:   true,
	})
	check(err)

	shader, err := glitch.NewShader(shaders.SpriteShader)
	check(err)

	img, err := assets.LoadImage("gopher.png")
	check(err)

	texture := glitch.NewTexture(img, false)
	sprite := glitch.NewSprite(texture, texture.Bounds())

	screenScale := 1.0 // This is just a weird scaling number

	// A screenspace camera
	camera := glitch.NewCameraOrtho()
	camera.SetOrtho2D(win.Bounds())
	camera.SetView2D(0, 0, screenScale, screenScale)

	batcher := glitch.NewBatcher()

	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, screenScale, screenScale)

		glitch.Clear(win, glitch.Greyscale(0.5))

		// you were working on migrating the batcher to be internal to each individual frame buffer thing. so then you just draw directly to one of those and it gets rendered
		// General Plan
		// 1. complex mode: draws -> sorter -> batcher -> opengl
		// 2. simple mode: draws -> batcher -> opengl
		// 3. batcher is internal to the framebuffer we are drawing to
		// 4. bufferpools are managed by each shader
		// 5. user perspective is that everything is immediate mode, but they can sort their draw commands by wrapping the immediate API with a sorter

		win.Bind()
		shader.SetUniform("projection", camera.Projection)
		shader.SetUniform("view", camera.View)

		batcher.SetShader(shader)
		sprite.Draw(batcher, glitch.Mat4Ident)

		batcher.Flush()

		win.Update()
	}
}


// func check(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func main() {
// 	glitch.Run(run)
// }

// func run() {
// 	win, err := glitch.NewWindow(1920, 1080, "Glitch Demo", glitch.WindowConfig{
// 		Vsync:   true,
// 	})
// 	check(err)

// 	// shader, err := glitch.NewShader(shaders.SpriteShader)
// 	// check(err)

// 	MSDFShader, err := glitch.NewShader(shaders.MSDFShader)
// 	check(err)

// 	pass := glitch.NewRenderPass(MSDFShader)

//  	atlasImg, err := assets.LoadImage("atlas-msdf.png")
// 	check(err)
// 	atlasJson := glitch.SdfAtlas{}
// 	err = assets.LoadJson("atlas-msdf.json", &atlasJson)
// 	check(err)

// 	sdfAtlas, err := glitch.AtlasFromSdf(atlasJson, atlasImg)
// 	check(err)

// 	// Text
// 	atlas, err := glitch.BasicFontAtlas()
// 	check(err)

// 	// "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
// 	sdfText := sdfAtlas.Text("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 1.0)
// 	text := atlas.Text("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 1.0)

// 	screenScale := 1.0 // This is just a weird scaling number

// 	// A screenspace camera
// 	camera := glitch.NewCameraOrtho()
// 	camera.SetOrtho2D(win.Bounds())
// 	camera.SetView2D(0, 0, screenScale, screenScale)

// 	for !win.Closed() {
// 		if win.Pressed(glitch.KeyEscape) {
// 			win.Close()
// 		}

// 		camera.SetOrtho2D(win.Bounds())
// 		camera.SetView2D(0, 0, screenScale, screenScale)

// 		glitch.Clear(win, glitch.Greyscale(0.5))
// 		pass.Clear()

// 		mat := glitch.Mat4Ident
// 		mat.
// 			Scale(4, 4, 1).
// 			Translate(win.Bounds().Center().X, win.Bounds().Center().Y, 0)
// 		text.Draw(pass, mat)

// 		lh := sdfAtlas.LineHeight()
// 		y := 0.0
// 		scale := 0.1
// 		for i := 0; i < 25; i++ {
// 			mat := glitch.Mat4Ident
// 			mat.
// 				Scale(scale, scale, 1).
// 				Translate(0, y, 0)
// 			sdfText.Draw(pass, mat)

// 			y += lh * scale
// 			scale += 0.1
// 		}

// 		pass.SetUniform("u_threshold", float32(0.5));
// 		pass.SetCamera2D(camera)
// 		pass.Draw(win)

// 		win.Update()
// 	}
// }