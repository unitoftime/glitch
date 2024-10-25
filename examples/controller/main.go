package main

import (
	"fmt"
	_ "image/png"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/examples/assets"
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
		Vsync: true,
	})
	check(err)

	img, err := assets.LoadImage("gopher.png")
	check(err)

	texture := glitch.NewTexture(img, false)
	sprite := glitch.NewSprite(texture, texture.Bounds())

	// atlasImg, err := assets.LoadImage("atlas-msdf.png")
	// check(err)
	// atlasJson := glitch.SdfAtlas{}
	// err = assets.LoadJson("atlas-msdf.json", &atlasJson)
	// check(err)
	// atlas, err := glitch.AtlasFromSdf(atlasJson, atlasImg, 1.0)
	// check(err)

	// text := atlas.Text("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 1.0)
	// text.Material().SetUniform("u_threshold", 0.6) // TODO: Should mostly come from default

	// A screenspace camera
	camera := glitch.NewCameraOrtho()

	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}

		primaryGamepad := win.GetPrimaryGamepad()
		for b := glitch.ButtonFirst; b <= glitch.ButtonLast; b++ {
			if win.GetGamepadJustPressed(primaryGamepad, b) {
				fmt.Println("Just Pressed:", b)
			}

			if win.GetGamepadPressed(primaryGamepad, b) {
				fmt.Println("Pressed:", b)
			}
		}


		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, 1, 1)
		glitch.SetCamera(camera)

		glitch.Clear(win, glm.Greyscale(0.5))

		// mat := glitch.Mat4Ident
		// text.Draw(win, *mat.Translate(100, 100, 0))

		center := win.Bounds().Center()

		{
			leftX := win.GetGamepadAxis(primaryGamepad, glitch.AxisLeftX)
			leftY := -win.GetGamepadAxis(primaryGamepad, glitch.AxisLeftY)

			axisLeft := glm.Vec2{leftX, leftY}.Scaled(100)

			mat := glitch.Mat4Ident
			mat.Translate(center.X-200, center.Y, 0)
			mat.Translate(axisLeft.X, axisLeft.Y, 0)
			sprite.Draw(win, mat)
		}

		{
			rightX := win.GetGamepadAxis(primaryGamepad, glitch.AxisRightX)
			rightY := -win.GetGamepadAxis(primaryGamepad, glitch.AxisRightY)

			axisRight := glm.Vec2{rightX, rightY}.Scaled(100)

			mat := glitch.Mat4Ident
			mat.Translate(center.X+200, center.Y, 0)
			mat.Translate(axisRight.X, axisRight.Y, 0)
			sprite.Draw(win, mat)
		}

		// Left Trigger
		{
			leftTrigger := win.GetGamepadAxis(primaryGamepad, glitch.AxisLeftTrigger)

			mat := glitch.Mat4Ident
			mat.Translate(center.X - 400, center.Y, 0)
			mat.Translate(0, leftTrigger * 100, 0)
			sprite.Draw(win, mat)
		}

		// Right Trigger
		{
			rightTrigger := win.GetGamepadAxis(primaryGamepad, glitch.AxisRightTrigger)

			mat := glitch.Mat4Ident
			mat.Translate(center.X + 400, center.Y, 0)
			mat.Translate(0, rightTrigger * 100, 0)
			sprite.Draw(win, mat)
		}

		win.Update()
	}
}
