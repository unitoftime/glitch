package main

import (
	"log"
	"embed"
	"image"
	"image/draw"
	_ "image/png"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/shaders"
	"github.com/unitoftime/glitch/ui"
)

//go:embed button.png button_hover.png button_press.png panel.png panel_inner.png
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
	log.Println("Begin")
	glitch.Run(runGame)
}

func runGame() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
		Vsync: false,
		Samples: 0,
	})
	if err != nil { panic(err) }

	shader, err := glitch.NewShader(shaders.SpriteShader)
	if err != nil { panic(err) }
	pass := glitch.NewRenderPass(shader)
	pass.SoftwareSort = glitch.SoftwareSortY
	pass.DepthTest = true

	buttonImage, err := loadImage("button.png")
	if err != nil { panic(err) }
	buttonHoverImage, err := loadImage("button_hover.png")
	if err != nil { panic(err) }
	buttonPressImage, err := loadImage("button_press.png")
	if err != nil { panic(err) }
	panelImage, err := loadImage("panel.png")
	if err != nil { panic(err) }
	panelInnerImage, err := loadImage("panel_inner.png")
	if err != nil { panic(err) }


	texture := glitch.NewTexture(buttonImage, false)
	buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds(), glitch.R(1, 1, 1, 1))
	buttonSprite.Scale = 10

	texture2 := glitch.NewTexture(buttonPressImage, false)
	buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds(), glitch.R(1, 1, 1, 1))
	buttonPressSprite.Scale = 10

	texture3 := glitch.NewTexture(buttonHoverImage, false)
	buttonHoverSprite := glitch.NewNinePanelSprite(texture3, texture3.Bounds(), glitch.R(1, 1, 1, 1))
	buttonHoverSprite.Scale = 10

	texture4 := glitch.NewTexture(panelImage, false)
	panelSprite := glitch.NewNinePanelSprite(texture4, texture4.Bounds(), glitch.R(2, 2, 2, 2))
	panelSprite.Scale = 10

	panelInnerTex := glitch.NewTexture(panelInnerImage, false)
	panelInnerSprite := glitch.NewNinePanelSprite(panelInnerTex, panelInnerTex.Bounds(), glitch.R(2, 2, 2, 2))
	panelInnerSprite.Scale = 10
	panelInnerSprite.Mask = glitch.RGBA{1, 0, 0, 1}

	// Text
	atlas, err := glitch.DefaultAtlas()
	if err != nil { panic(err) }

	// A screenspace camera
	camera := glitch.NewCameraOrtho()
	camera.SetOrtho2D(win.Bounds())
	camera.SetView2D(0, 0, 1.0, 1.0)
	group := ui.NewGroup(win, camera, atlas, pass)
	// group.Debug = true

	textStyle := group.NewTextStyle()
	buttonStyle := ui.Style{
		Normal: buttonSprite,
		Hovered: buttonHoverSprite,
		Pressed: buttonPressSprite,
		Text: textStyle,
	}

	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}

		// mx, my := win.MousePosition()
		// log.Println("Mouse: ", mx, my)

		glitch.Clear(win, glitch.Black)

		ui.Clear()
		group.Clear()
		pass.Clear()

		menuRect := win.Bounds().SliceHorizontal(500).SliceVertical(500)
		group.Panel(panelSprite, menuRect)

		menuRect.CutLeft(20)
		menuRect.CutRight(20)
		{
			r := menuRect.CutTop(100)
			group.Text("Menu", r, glitch.Vec2{0.5, 0.5})
		}
		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			buttonStyle.Text.Color = glitch.Black
			if group.ButtonE("Button 0", buttonStyle, r) {
				println("Button 0")
			}
		}
		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			buttonStyle.Text.Color = glitch.RGBA{1, 0, 0, 1}
			if group.ButtonE("Button 1", buttonStyle, r) {
				println("Button 1")
			}
		}

		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			group.Panel(panelSprite, r)
			group.Panel(panelInnerSprite, r)
		}

		pass.SetUniform("projection", camera.Projection)
		pass.SetUniform("view", camera.View)
		pass.Draw(win)

		win.Update()
	}
}

