package main

import (
	"log"
	"embed"
	"image"
	"image/draw"
	_ "image/png"

	"unicode"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/ui"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
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


	texture := glitch.NewTexture(buttonImage)
	buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds(), glitch.R(1, 1, 1, 1))
	buttonSprite.Scale = 10

	texture2 := glitch.NewTexture(buttonPressImage)
	buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds(), glitch.R(1, 1, 1, 1))
	buttonPressSprite.Scale = 10

	texture3 := glitch.NewTexture(buttonHoverImage)
	buttonHoverSprite := glitch.NewNinePanelSprite(texture3, texture3.Bounds(), glitch.R(1, 1, 1, 1))
	buttonHoverSprite.Scale = 10

	texture4 := glitch.NewTexture(panelImage)
	panelSprite := glitch.NewNinePanelSprite(texture4, texture4.Bounds(), glitch.R(2, 2, 2, 2))
	panelSprite.Scale = 10

	panelInnerTex := glitch.NewTexture(panelInnerImage)
	panelInnerSprite := glitch.NewNinePanelSprite(panelInnerTex, panelInnerTex.Bounds(), glitch.R(2, 2, 2, 2))
	panelInnerSprite.Scale = 10
	panelInnerSprite.Mask = glitch.RGBA{1, 0, 0, 1}

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

	group := ui.NewGroup(win, atlas)
	// group.Debug = true

	for !win.ShouldClose() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}

		mx, my := win.MousePosition()
		log.Println("Mouse: ", mx, my)

		glitch.Clear(glitch.Black)

		group.Clear()
		menuRect := win.Bounds().SliceHorizontal(500).SliceVertical(500)
		group.Panel(panelSprite, menuRect)

		// basicHover := ui.BasicHover{buttonSprite, buttonPressSprite}

		menuRect.CutLeft(20)
		menuRect.CutRight(20)
		{
			r := menuRect.CutTop(100)
			group.Text("Menu", r, glitch.Vec2{0.5, 0.5})
		}
		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			group.Button(buttonSprite, buttonHoverSprite, buttonPressSprite, r)
			group.Text("Button 0", r, glitch.Vec2{0.5, 0.5})
		}
		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			group.Button(buttonSprite, buttonHoverSprite, buttonPressSprite, r)
			group.Text("Button 1", r, glitch.Vec2{0.5, 0.5})
		}
		// group.Sprite(buttonSprite, glitch.R(0, 0, 200, 75))

		menuRect.CutTop(10) // Padding
		{
			r := menuRect.CutTop(100)
			group.Panel(panelSprite, r)
			group.Panel(panelInnerSprite, r)
		}

		group.Draw()

		win.Update()
	}
}

