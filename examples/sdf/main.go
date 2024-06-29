package main

import (
	_ "image/png"
	"log"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/examples/assets"
	"github.com/unitoftime/glitch/shaders"
)

func main() {
	log.Println("Begin")
	glitch.Run(runGame)
}

func runGame() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
		Vsync:   true,
		Samples: 0,
	})
	if err != nil {
		panic(err)
	}

	shader, err := glitch.NewShader(shaders.MSDFShader)
	if err != nil {
		panic(err)
	}
	pass := glitch.NewRenderPass(shader)
	// pass.SoftwareSort = glitch.SoftwareSortY
	// pass.DepthTest = true
	// pass.DepthBump = true

 	atlasImg, err := assets.LoadImage("atlas-msdf.png")
	if err != nil {
		panic(err)
	}
	atlasJson := glitch.SdfAtlas{}
	err = assets.LoadJson("atlas-msdf.json", &atlasJson)
	if err != nil {
		panic(err)
	}

	// fmt.Println(&atlasImg)
	// fmt.Println(&atlasJson)
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

		// mx, my := win.MousePosition()
		// log.Println("Mouse: ", mx, my)

		glitch.Clear(win, glitch.Greyscale(0.5))

		pass.Clear()

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
			text.Draw(pass, mat)

			y += lh * scale
			scale += 0.1
		}

		pass.SetUniform("u_threshold", float32(0.5));
		pass.SetCamera2D(camera)
		pass.Draw(win)

		win.Update()
	}
}
