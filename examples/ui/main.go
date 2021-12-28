package main

import (
	"log"
	"embed"
	"image"
	"image/draw"
	_ "image/png"

	"unicode"

	"github.com/unitoftime/glitch"
	// "github.com/unitoftime/glitch/shaders"
	"github.com/unitoftime/glitch/ui"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

//go:embed button.png button_press.png
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
	win, err := glitch.NewWindow(1920, 1080, "Glitch", glitch.WindowConfig{
		Vsync: false,
	})
	if err != nil { panic(err) }

	// shader, err := glitch.NewShader(shaders.SpriteShader)
	// if err != nil { panic(err) }

	// pass := glitch.NewRenderPass(shader)

	buttonImage, err := loadImage("button.png")
	if err != nil {
		panic(err)
	}
	buttonPressImage, err := loadImage("button_press.png")
	if err != nil {
		panic(err)
	}

	texture := glitch.NewTexture(buttonImage)
	buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds())
	texture2 := glitch.NewTexture(buttonPressImage)
	buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds())

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

	// text := atlas.Text("hello world")
	group := ui.NewGroup(win, atlas)

	// camera := glitch.NewCameraOrtho()

	for !win.ShouldClose() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}

		mx, my := win.MousePosition()
		log.Println("Mouse: ", mx, my)
		// pass.Clear()

		// camera.SetOrtho2D(win)
		// camera.SetView2D(0, 0, 1.0, 1.0)

		// mat := glitch.Mat4Ident
		// mat.Translate(0, 0, 0)
		// text.Set(fmt.Sprintf("%2.2f ms", 1000*dt.Seconds()))
		// text.DrawColorMask(pass, mat, glitch.RGBA{1.0, 1.0, 0.0, 1.0})

		glitch.Clear(glitch.RGBA{0.1, 0.2, 0.3, 1.0})

		// pass.SetUniform("projection", camera.Projection)
		// pass.SetUniform("view", camera.View)
		// pass.Draw(win)

		group.Clear()
		menuRect := win.Bounds().SliceHorizontal(500).SliceVertical(500)
		group.Sprite(buttonPressSprite, menuRect)

		// basicHover := ui.BasicHover{buttonSprite, buttonPressSprite}

		r := menuRect.CutTop(100)
		r = menuRect.CutTop(200)
		group.HoveredSprite(buttonSprite, buttonPressSprite, r)
		group.Text("Hello World", r, glitch.Vec2{0.5, 0.5})
		// group.Sprite(buttonSprite, glitch.R(0, 0, 200, 75))
		group.Draw()

		win.Update()
	}
}

