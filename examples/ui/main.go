package main

import (
	_ "image/png"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/examples/assets"
	"github.com/unitoftime/glitch/shaders"
	"github.com/unitoftime/glitch/ui"
)

func main() {
	glitch.Run(runGame)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func runGame() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
		Vsync:   true,
	})
	check(err)

	shader, err := glitch.NewShader(shaders.PixelArtShader)
	check(err)
	glitch.SetDefaultSpriteShader(shader)

	buttonImage, err := assets.LoadImage("button.png")
	check(err)
	buttonHoverImage, err := assets.LoadImage("button_hover.png")
	check(err)
	buttonPressImage, err := assets.LoadImage("button_press.png")
	check(err)
	panelImage, err := assets.LoadImage("panel.png")
	check(err)
	panelInnerImage, err := assets.LoadImage("panel_inner.png")
	check(err)


	scale := 4.0
	texture := glitch.NewTexture(buttonImage, false)
	buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds(), glitch.R(1, 1, 1, 1))
	buttonSprite.Scale = scale

	texture2 := glitch.NewTexture(buttonPressImage, false)
	buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds(), glitch.R(1, 1, 1, 1))
	buttonPressSprite.Scale = scale

	texture3 := glitch.NewTexture(buttonHoverImage, false)
	buttonHoverSprite := glitch.NewNinePanelSprite(texture3, texture3.Bounds(), glitch.R(1, 1, 1, 1))
	buttonHoverSprite.Scale = scale

	texture4 := glitch.NewTexture(panelImage, false)
	panelSprite := glitch.NewNinePanelSprite(texture4, texture4.Bounds(), glitch.R(2, 2, 2, 2))
	panelSprite.Scale = scale

	panelInnerTex := glitch.NewTexture(panelInnerImage, false)
	panelInnerSprite := glitch.NewNinePanelSprite(panelInnerTex, panelInnerTex.Bounds(), glitch.R(2, 2, 2, 2))
	panelInnerSprite.Scale = scale

	// Text
	// atlas, err := glitch.BasicFontAtlas()
	// check(err)
 	atlasImg, err := assets.LoadImage("atlas-msdf.png")
	check(err)
	atlasJson := glitch.SdfAtlas{}
	err = assets.LoadJson("atlas-msdf.json", &atlasJson)
	check(err)
	atlas, err := glitch.AtlasFromSdf(atlasJson, atlasImg)
	check(err)
	atlas.Material().SetUniform("u_threshold", 0.6) // Overwrite the default

	screenScale := 1.0 // This is just a weird scaling number

	// A screenspace camera
	camera := glitch.NewCameraOrtho()
	camera.SetOrtho2D(win.Bounds())
	camera.SetView2D(0, 0, screenScale, screenScale)
	sorter := glitch.NewSorter()
	group := ui.NewGroup(win, camera, atlas, sorter)
	group.Debug = true

	textStyle := ui.NewTextStyle().Scale(4).Autofit(true)
	buttonStyle := ui.Style{
		Normal:  ui.NewSpriteStyle(buttonSprite, glitch.White),
		Hovered: ui.NewSpriteStyle(buttonHoverSprite, glitch.White),
		Pressed: ui.NewSpriteStyle(buttonPressSprite, glitch.White),
		Text:    textStyle,
	}

	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, screenScale, screenScale)
		glitch.SetCamera(camera)

		// mx, my := win.MousePosition()
		// log.Println("Mouse: ", mx, my)

		ui.Clear()
		group.Clear()

		menuRect := win.Bounds().SliceHorizontal(500).SliceVertical(500)
		group.Panel(panelSprite, menuRect, glitch.White)

		menuRect.CutLeft(20)
		menuRect.CutRight(20)
		{
			r := menuRect.CutTop(100)
			group.Text("Menu", r, textStyle)
		}
		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			buttonStyle.Text = buttonStyle.Text.Color(glitch.Black)
			if group.Button("-", r.CutLeft(r.W()/2), buttonStyle) {
				screenScale -= 0.1
			}
			if group.Button("+", r, buttonStyle) {
				screenScale += 0.1
			}
		}

		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			buttonStyle.Text = buttonStyle.Text.Color(glitch.Black)
			if group.Button("Button 0", r, buttonStyle) {
				println("Button 0")
			}
		}
		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			buttonStyle.Text = buttonStyle.Text.Color(glitch.White)
			if group.Button("Button 1", r, buttonStyle) {
				println("Button 1")
			}
		}

		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			group.Panel(panelSprite, r, glitch.White)
			group.Panel(panelInnerSprite, r, glitch.White)
		}

		glitch.Clear(win, glitch.Greyscale(0.5))

		// shader.Use()
		// shader.SetUniform("projection", camera.Projection)
		// shader.SetUniform("view", camera.View)

		// tpp := float32(1.0/screenScale)
		// tpp := float32(512.0 / 1920.0) // Texels per screen pixel
		tpp := float32(1.0 / 8.0)
		shader.SetUniform("texelsPerPixel", tpp)

		group.Draw(win)

		win.Update()
	}
}


// func runGame() {
// 	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
// 		Vsync:   true,
// 		Samples: 0,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	shader, err := glitch.NewShader(shaders.PixelArtShader)
// 	if err != nil {
// 		panic(err)
// 	}
// 	pass := glitch.NewRenderPass(shader)
// 	// pass.SoftwareSort = glitch.SoftwareSortY
// 	// pass.DepthTest = true
// 	// pass.DepthBump = true

// 	buttonImage, err := assets.LoadImage("button.png")
// 	if err != nil {
// 		panic(err)
// 	}
// 	buttonHoverImage, err := assets.LoadImage("button_hover.png")
// 	if err != nil {
// 		panic(err)
// 	}
// 	buttonPressImage, err := assets.LoadImage("button_press.png")
// 	if err != nil {
// 		panic(err)
// 	}
// 	panelImage, err := assets.LoadImage("panel.png")
// 	if err != nil {
// 		panic(err)
// 	}
// 	panelInnerImage, err := assets.LoadImage("panel_inner.png")
// 	if err != nil {
// 		panic(err)
// 	}

// 	scale := 4.0
// 	texture := glitch.NewTexture(buttonImage, false)
// 	buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds(), glitch.R(1, 1, 1, 1))
// 	buttonSprite.Scale = scale

// 	texture2 := glitch.NewTexture(buttonPressImage, false)
// 	buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds(), glitch.R(1, 1, 1, 1))
// 	buttonPressSprite.Scale = scale

// 	texture3 := glitch.NewTexture(buttonHoverImage, false)
// 	buttonHoverSprite := glitch.NewNinePanelSprite(texture3, texture3.Bounds(), glitch.R(1, 1, 1, 1))
// 	buttonHoverSprite.Scale = scale

// 	texture4 := glitch.NewTexture(panelImage, false)
// 	panelSprite := glitch.NewNinePanelSprite(texture4, texture4.Bounds(), glitch.R(2, 2, 2, 2))
// 	panelSprite.Scale = scale

// 	panelInnerTex := glitch.NewTexture(panelInnerImage, false)
// 	panelInnerSprite := glitch.NewNinePanelSprite(panelInnerTex, panelInnerTex.Bounds(), glitch.R(2, 2, 2, 2))
// 	panelInnerSprite.Scale = scale
// 	// panelInnerSprite.Mask = glitch.RGBA{1, 0, 0, 1}

// 	// Text
// 	atlas, err := glitch.BasicFontAtlas()
// 	if err != nil {
// 		panic(err)
// 	}

// 	screenScale := 1.5 // This is just a weird scaling number

// 	// A screenspace camera
// 	camera := glitch.NewCameraOrtho()
// 	camera.SetOrtho2D(win.Bounds())
// 	camera.SetView2D(0, 0, screenScale, screenScale)
// 	group := ui.NewGroup(win, camera, atlas, pass)
// 	// group.Debug = true

// 	textStyle := ui.NewTextStyle().Scale(4)
// 	buttonStyle := ui.Style{
// 		Normal:  ui.NewSpriteStyle(buttonSprite, glitch.White),
// 		Hovered: ui.NewSpriteStyle(buttonHoverSprite, glitch.White),
// 		Pressed: ui.NewSpriteStyle(buttonPressSprite, glitch.White),
// 		Text:    textStyle,
// 	}

// 	for !win.Closed() {
// 		if win.Pressed(glitch.KeyEscape) {
// 			win.Close()
// 		}

// 		camera.SetOrtho2D(win.Bounds())
// 		camera.SetView2D(0, 0, screenScale, screenScale)

// 		// mx, my := win.MousePosition()
// 		// log.Println("Mouse: ", mx, my)

// 		glitch.Clear(win, glitch.Black)

// 		ui.Clear()
// 		group.Clear()
// 		pass.Clear()

// 		menuRect := win.Bounds().SliceHorizontal(500).SliceVertical(500)
// 		group.Panel(panelSprite, menuRect, glitch.White)

// 		menuRect.CutLeft(20)
// 		menuRect.CutRight(20)
// 		{
// 			r := menuRect.CutTop(100)
// 			group.Text("Menu", r, textStyle)
// 		}
// 		menuRect.CutTop(10) // Padding
// 		{
// 			r := menuRect.CutTop(100)
// 			buttonStyle.Text = buttonStyle.Text.Color(glitch.Black)
// 			if group.Button("-", r.CutLeft(r.W()/2), buttonStyle) {
// 				screenScale -= 0.1
// 			}
// 			if group.Button("+", r, buttonStyle) {
// 				screenScale += 0.1
// 			}
// 		}

// 		menuRect.CutTop(10) // Padding
// 		{
// 			r := menuRect.CutTop(100)
// 			buttonStyle.Text = buttonStyle.Text.Color(glitch.Black)
// 			if group.Button("Button 0", r, buttonStyle) {
// 				println("Button 0")
// 			}
// 		}
// 		menuRect.CutTop(10) // Padding
// 		{
// 			r := menuRect.CutTop(100)
// 			buttonStyle.Text = buttonStyle.Text.Color(glitch.White)
// 			if group.Button("Button 1", r, buttonStyle) {
// 				println("Button 1")
// 			}
// 		}

// 		menuRect.CutTop(10) // Padding
// 		{
// 			r := menuRect.CutTop(100)
// 			group.Panel(panelSprite, r, glitch.White)
// 			group.Panel(panelInnerSprite, r, glitch.White)
// 		}

// 		pass.SetCamera2D(camera)
// 		// tpp := float32(1.0/screenScale)
// 		// tpp := float32(512.0 / 1920.0) // Texels per screen pixel
// 		tpp := float32(1.0 / 8.0)
// 		pass.SetUniform("texelsPerPixel", tpp)
// 		pass.Draw(win)

// 		win.Update()
// 	}
// }

// // package main

// // import (
// // 	"embed"
// // 	"image"
// // 	"image/draw"
// // 	_ "image/png"
// // 	"log"

// // 	"github.com/unitoftime/glitch"
// // 	"github.com/unitoftime/glitch/shaders"
// // 	"github.com/unitoftime/glitch/ui"
// // )

// // //go:embed button.png button_hover.png button_press.png panel.png panel_inner.png
// // var f embed.FS

// // func loadImage(path string) (*image.NRGBA, error) {
// // 	file, err := f.Open(path)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	img, _, err := image.Decode(file)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	bounds := img.Bounds()
// // 	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
// // 	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
// // 	return nrgba, nil
// // }

// // func main() {
// // 	log.Println("Begin")
// // 	glitch.Run(runGame)
// // }

// // func runGame() {
// // 	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
// // 		Vsync: false,
// // 		Samples: 0,
// // 	})
// // 	if err != nil { panic(err) }

// // 	shader, err := glitch.NewShader(shaders.SpriteShader)
// // 	if err != nil { panic(err) }
// // 	pass := glitch.NewRenderPass(shader)
// // 	pass.SoftwareSort = glitch.SoftwareSortY
// // 	pass.DepthTest = true
// // 	pass.DepthBump = true

// // 	buttonImage, err := loadImage("button.png")
// // 	if err != nil { panic(err) }
// // 	buttonHoverImage, err := loadImage("button_hover.png")
// // 	if err != nil { panic(err) }
// // 	buttonPressImage, err := loadImage("button_press.png")
// // 	if err != nil { panic(err) }
// // 	panelImage, err := loadImage("panel.png")
// // 	if err != nil { panic(err) }
// // 	panelInnerImage, err := loadImage("panel_inner.png")
// // 	if err != nil { panic(err) }

// // 	texture := glitch.NewTexture(buttonImage, false)
// // 	buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds(), glitch.R(1, 1, 1, 1))
// // 	buttonSprite.Scale = 1

// // 	texture2 := glitch.NewTexture(buttonPressImage, false)
// // 	buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds(), glitch.R(1, 1, 1, 1))
// // 	buttonPressSprite.Scale = 1

// // 	texture3 := glitch.NewTexture(buttonHoverImage, false)
// // 	buttonHoverSprite := glitch.NewNinePanelSprite(texture3, texture3.Bounds(), glitch.R(1, 1, 1, 1))
// // 	buttonHoverSprite.Scale = 1

// // 	texture4 := glitch.NewTexture(panelImage, false)
// // 	panelSprite := glitch.NewNinePanelSprite(texture4, texture4.Bounds(), glitch.R(2, 2, 2, 2))
// // 	panelSprite.Scale = 1

// // 	panelInnerTex := glitch.NewTexture(panelInnerImage, false)
// // 	panelInnerSprite := glitch.NewNinePanelSprite(panelInnerTex, panelInnerTex.Bounds(), glitch.R(2, 2, 2, 2))
// // 	panelInnerSprite.Scale = 1
// // 	// panelInnerSprite.Mask = glitch.RGBA{1, 0, 0, 1}

// // 	// Text
// // 	atlas, err := glitch.DefaultAtlas()
// // 	if err != nil { panic(err) }

// // 	// A screenspace camera
// // 	camera := glitch.NewCameraOrtho()
// // 	camera.SetOrtho2D(win.Bounds())
// // 	camera.SetView2D(0, 0, 1.0, 1.0)
// // 	group := ui.NewGroup(win, camera, atlas, pass)
// // 	// group.Debug = true

// // 	textStyle := ui.NewTextStyle().Scale(1)
// // 	buttonStyle := ui.Style{
// // 		Normal: ui.NewSpriteStyle(buttonSprite, glitch.White),
// // 		Hovered: ui.NewSpriteStyle(buttonHoverSprite, glitch.White),
// // 		Pressed: ui.NewSpriteStyle(buttonPressSprite, glitch.White),
// // 		Text: textStyle,
// // 	}

// // 	for !win.Closed() {
// // 		if win.Pressed(glitch.KeyEscape) {
// // 			win.Close()
// // 		}

// // 		// mx, my := win.MousePosition()
// // 		// log.Println("Mouse: ", mx, my)

// // 		glitch.Clear(win, glitch.Black)

// // 		ui.Clear()
// // 		group.Clear()
// // 		pass.Clear()

// // 		menuRect := win.Bounds().SliceHorizontal(500).SliceVertical(500)
// // 		group.Panel(panelSprite, menuRect, glitch.White)

// // 		menuRect.CutLeft(20)
// // 		menuRect.CutRight(20)
// // 		{
// // 			r := menuRect.CutTop(100)
// // 			group.Text("Menu", r, textStyle)
// // 		}
// // 		menuRect.CutTop(10) // Padding
// // 		{
// // 			r := menuRect.CutTop(100)
// // 			buttonStyle.Text = buttonStyle.Text.Color(glitch.Black)
// // 			if group.Button("Button 0", r, buttonStyle) {
// // 				println("Button 0")
// // 			}
// // 		}
// // 		menuRect.CutTop(10) // Padding
// // 		{
// // 			r := menuRect.CutTop(100)
// // 			buttonStyle.Text = buttonStyle.Text.Color(glitch.White)
// // 			if group.Button("Button 1", r, buttonStyle) {
// // 				println("Button 1")
// // 			}
// // 		}

// // 		menuRect.CutTop(10) // Padding
// // 		{
// // 			r := menuRect.CutTop(100)
// // 			group.Panel(panelSprite, r, glitch.White)
// // 			group.Panel(panelInnerSprite, r, glitch.White)
// // 		}

// // 		pass.SetCamera2D(camera)
// // 		pass.Draw(win)

// // 		win.Update()
// // 	}
// // }
