package main

import (
	"fmt"
	_ "image/png"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/examples/assets"
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

	// shader, err := glitch.NewShader(shaders.PixelArtShader)
	// check(err)
	// glitch.SetDefaultSpriteShader(shader)

	// buttonImage, err := assets.LoadImage("button.png")
	// check(err)
	// buttonHoverImage, err := assets.LoadImage("button_hover.png")
	// check(err)
	// buttonPressImage, err := assets.LoadImage("button_press.png")
	// check(err)
	// panelImage, err := assets.LoadImage("panel.png")
	// check(err)
	// panelInnerImage, err := assets.LoadImage("panel_inner.png")
	// check(err)


	// scale := 4.0
	// texture := glitch.NewTexture(buttonImage, false)
	// buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds(), glitch.R(1, 1, 1, 1))
	// buttonSprite.Scale = scale

	// texture2 := glitch.NewTexture(buttonPressImage, false)
	// buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds(), glitch.R(1, 1, 1, 1))
	// buttonPressSprite.Scale = scale

	// texture3 := glitch.NewTexture(buttonHoverImage, false)
	// buttonHoverSprite := glitch.NewNinePanelSprite(texture3, texture3.Bounds(), glitch.R(1, 1, 1, 1))
	// buttonHoverSprite.Scale = scale

	// texture4 := glitch.NewTexture(panelImage, false)
	// panelSprite := glitch.NewNinePanelSprite(texture4, texture4.Bounds(), glitch.R(2, 2, 2, 2))
	// panelSprite.Scale = scale

	// panelInnerTex := glitch.NewTexture(panelInnerImage, false)
	// panelInnerSprite := glitch.NewNinePanelSprite(panelInnerTex, panelInnerTex.Bounds(), glitch.R(2, 2, 2, 2))
	// panelInnerSprite.Scale = scale

	// Text
	// atlas, err := glitch.BasicFontAtlas()
	// check(err)
 	atlasImg, err := assets.LoadImage("atlas-msdf.png")
	check(err)
	atlasJson := glitch.SdfAtlas{}
	err = assets.LoadJson("atlas-msdf.json", &atlasJson)
	check(err)
	atlas, err := glitch.AtlasFromSdf(atlasJson, atlasImg, 3)
	check(err)
	atlas.Material().SetUniform("u_threshold", 0.6) // Overwrite the default

	screenScale := 1.0 // This is just a weird scaling number

	// A screenspace camera
	camera := glitch.NewCameraOrtho()
	camera.SetOrtho2D(win.Bounds())
	camera.SetView2D(0, 0, screenScale, screenScale)
	sorter := glitch.NewSorter()
	ui.Initialize(win, camera, atlas, sorter)
	// group.Debug = true

	// // textStyle := ui.NewTextStyle().Scale(4)// .Autofit(true)
	// textStyle := ui.NewTextStyle().Autofit(true).Padding(glitch.R(5, 5, 5, 5))
	// buttonStyle := ui.Style{
	// 	Normal:  ui.NewSpriteStyle(buttonSprite, glitch.White),
	// 	Hovered: ui.NewSpriteStyle(buttonHoverSprite, glitch.White),
	// 	Pressed: ui.NewSpriteStyle(buttonPressSprite, glitch.White),
	// 	Text:    textStyle,
	// }

	textStyle := ui.NewTextStyle().Autofit(true).Padding(glitch.R(5, 5, 5, 5))
	// ui.SetTextStyle(textStyle)
	// ui.SetButtonStyle(buttonStyle)
	// ui.SetPanelStyle(buttonStyle)
	// ui.SetDragSlotStyle(buttonStyle)
	// ui.SetDragItemStyle(buttonStyle)
	// ui.SetCheckboxStyleTrue(buttonStyle)
	// ui.SetCheckboxStyleFalse(buttonStyle)

	ui.SetDragItemLayer(int8(0))

	modes := []string{
		"text", "scroll", "drag", "lists", "grid",
	}
	mode := modes[0]

	dragData := []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "", "",
	}

	scrollIdx := 0

	sliderVal := 0.0
	checkboxVal := false
	inputStr := "Text Input"

	for !win.Closed() {
		if win.Pressed(glitch.KeyEscape) {
			win.Close()
		}

		camera.SetOrtho2D(win.Bounds())
		camera.SetView2D(0, 0, screenScale, screenScale)
		glitch.SetCamera(camera)

		// mx, my := win.MousePosition()
		// log.Println("Mouse: ", mx, my)

		glitch.Clear(win, glitch.Greyscale(0.5))
		ui.Clear()

		// {
		// 	bounds := win.Bounds()
		// 	bounds = bounds.CutTop(100)
		// 	layout := ui.Layout{
		// 		Bounds: bounds,
		// 		Type: ui.CutLeft,
		// 		Padding: glitch.R(5, 5, 5, 5),
		// 		Size: ui.Size{
		// 			TypeX: ui.SizePixels,
		// 			TypeY: ui.SizePixels,
		// 			Value: glitch.Vec2{100, 100},
		// 		},
		// 	}
		// 	ui.PushLayout(layout)

		// 	if ui.Button("Hello") {
		// 		fmt.Println("HELLO")
		// 	}
		// 	if ui.Button("Hello2") {
		// 		fmt.Println("HELLO2")
		// 	}
		// 	if ui.Button("Hello3") {
		// 		fmt.Println("HELLO3")
		// 	}
		// }

		// {
		// 	bounds := win.Bounds()
		// 	bounds = bounds.CutLeft(100)
		// 	layout := ui.Layout{
		// 		Bounds: bounds,
		// 		Type: ui.CutBottom,
		// 		Padding: glitch.R(5, 5, 5, 5),
		// 		// Size: ui.Size{
		// 		// 	TypeX: ui.SizePixels,
		// 		// 	TypeY: ui.SizePixels,
		// 		// 	Value: glitch.Vec2{100, 100},
		// 		// },
		// 		Size: ui.Size{
		// 			TypeX: ui.SizeText,
		// 			TypeY: ui.SizeText,
		// 			Value: glitch.Vec2{1, 1},
		// 		},
		// 	}
		// 	ui.PushLayout(layout)

		// 	if ui.Button("Hello") {
		// 		fmt.Println("HELLO")
		// 	}
		// 	if ui.Button("Hello2") {
		// 		fmt.Println("HELLO2")
		// 	}
		// 	if ui.Button("Hello3") {
		// 		fmt.Println("HELLO3")
		// 	}
		// }

		// {
		// 	bounds := win.Bounds()
		// 	layout := ui.Layout{
		// 		Bounds: bounds,
		// 		Type: ui.Centered,
		// 		Padding: glitch.R(5, 5, 5, 5),
		// 		// Size: ui.Size{
		// 		// 	TypeX: ui.SizePixels,
		// 		// 	TypeY: ui.SizePixels,
		// 		// 	Value: glitch.Vec2{100, 100},
		// 		// },
		// 		Size: ui.Size{
		// 			TypeX: ui.SizeParent,
		// 			TypeY: ui.SizeParent,
		// 			Value: glitch.Vec2{0.5, 0.5},
		// 		},
		// 	}
		// 	ui.PushLayout(layout)
		// 	ui.Panel("##panel")
		// }

		ui.SetLayer(1)

		// Switcher
		{
			bounds := win.Bounds()
			list := ui.VList2(bounds.CutLeft(100), 50)
			for _, m := range modes {
				if ui.Button2(m, list.Next().Unpad(glitch.R(5,5,5,5))) {
					mode = m
				}
			}
		}

		bounds := win.Bounds()
		panelBounds := bounds.SliceVertical(0.5 * bounds.W()).SliceHorizontal(0.5 * bounds.H())

		switch mode {
		case "lists":
			ui.Panel2("##panel", panelBounds)

			// numButtons := 3
			list := ui.VList(panelBounds.Unpad(glitch.R(5, 5, 5, 5)), 7)
			// list := ui.VList2(panelBounds.Unpad(glitch.R(5, 5, 5, 5)), 100)

			ui.TextExt("Title Section", list.Next().Unpad(glitch.R(5, 5, 5, 5)), textStyle)
			ui.TextInput("textinput", &inputStr, list.Next().Unpad(glitch.R(5, 5, 5, 5)), ui.DragItemStyle())

			topButton := ui.HList(list.Next().Unpad(glitch.R(5, 5, 5, 5)), 2)

			ui.Button2("Left", topButton.Next().Unpad(glitch.R(5, 5, 5, 5)))
			// ui.Tooltip("Tooltip:Left", ui.LastRect())
			ui.Tooltip("Tooltip:Left", topButton.Last())
			ui.Button2("Right", topButton.Next().Unpad(glitch.R(5, 5, 5, 5)))
			ui.Button2("Hello2", list.Next().Unpad(glitch.R(5, 5, 5, 5)))
			ui.Button2("Hello3", list.Next().Unpad(glitch.R(5, 5, 5, 5)))
			sliderRect := list.Next().Unpad(glitch.R(5, 5, 5, 5))
			ui.SliderH(&sliderVal, 0.0, 100, 1, sliderRect, sliderRect)

			checkboxRow := list.Next().Unpad(glitch.R(5, 5, 5, 5))
			checkboxRow = checkboxRow.CutLeft(checkboxRow.W()/2)
			checkboxBounds := checkboxRow.CutRight(checkboxRow.H())
			ui.TextExt("Checkbox", checkboxRow, textStyle)
			ui.Checkbox(&checkboxVal, checkboxBounds)
		case "grid":
			ui.Panel2("##panel", panelBounds)

			list := ui.GridList(panelBounds.Unpad(glitch.R(5, 5, 5, 5)), 3, 4)

			for i := 0; i < 12; i++ {
				str := fmt.Sprintf("%d", i)
				if ui.Button2(str, list.Next().Unpad(glitch.R(5, 5, 5, 5))) {
					fmt.Println("PRESSED", str)
				}
			}
		case "drag":
			ui.Panel2("##panel", panelBounds)

			list := ui.GridList(panelBounds.Unpad(glitch.R(5, 5, 5, 5)), 3, 4)

			for i := range dragData {
				str := dragData[i]
				slotRect := list.Next().Unpad(glitch.R(5, 5, 5, 5))
				if ui.DragSlot(str, slotRect, ui.DragSlotStyle()) {
					dragIdx, ok := ui.DragData().(int)
					if ok {
						tmp := dragData[dragIdx]
						dragData[dragIdx] = dragData[i]
						dragData[i] = tmp
					}
				}

				if str == "" { continue } // Skip: No item there

				clicked, hovered, dragging, dropping := ui.DragItem(str, slotRect.Unpad(glitch.R(5, 5, 5, 5)), ui.DragItemStyle())
				if clicked {
					fmt.Println("CLICKED", i)
				}
				if hovered {
					// fmt.Println("HOVERED", i)
				}
				if dragging {
					// fmt.Println("DRAGGING", i)
					// draggedIdx = i
					ui.SetDragData(i)
				}
				if dropping {
					// fmt.Println("DROPPING", i)
					dragIdx, ok := ui.DragData().(int)
					if ok {
						tmp := dragData[dragIdx]
						dragData[dragIdx] = dragData[i]
						dragData[i] = tmp
					}
				}
			}
		// case "smoothscroll": // TODO: Smoothscroll with scissor mask?
		case "scroll":
			ui.Panel2("##panel", panelBounds)
			scrollbarBounds := panelBounds.CutRight(50)
			scrollTotal := 10
			drawTotal := 5
			ui.Scrollbar(&scrollIdx, scrollTotal - drawTotal, scrollbarBounds, panelBounds)

			list := ui.VList2(panelBounds.Unpad(glitch.R(5, 5, 5, 5)), 100)
			for i := scrollIdx; i < scrollIdx + drawTotal; i ++ {
				str := fmt.Sprintf("%d", i)
				if ui.Button2(str, list.Next().Unpad(glitch.R(5, 5, 5, 5))) {
					fmt.Println("Click:", str)
				}
			}
		case "text":
			ui.Panel2("##panel", panelBounds)
			textAreaBounds := panelBounds// .Unpad(glitch.R(10, 10, 10, 10))
			ui.MultiText("Unfinished: Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.", textAreaBounds)
		}

		ui.Update()
		win.Update()
	}
}



// func main() {
// 	glitch.Run(runGame)
// }

// func check(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func runGame() {
// 	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
// 		Vsync:   true,
// 	})
// 	check(err)

// 	shader, err := glitch.NewShader(shaders.PixelArtShader)
// 	check(err)
// 	glitch.SetDefaultSpriteShader(shader)

// 	buttonImage, err := assets.LoadImage("button.png")
// 	check(err)
// 	buttonHoverImage, err := assets.LoadImage("button_hover.png")
// 	check(err)
// 	buttonPressImage, err := assets.LoadImage("button_press.png")
// 	check(err)
// 	panelImage, err := assets.LoadImage("panel.png")
// 	check(err)
// 	panelInnerImage, err := assets.LoadImage("panel_inner.png")
// 	check(err)


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

// 	// Text
// 	// atlas, err := glitch.BasicFontAtlas()
// 	// check(err)
//  	atlasImg, err := assets.LoadImage("atlas-msdf.png")
// 	check(err)
// 	atlasJson := glitch.SdfAtlas{}
// 	err = assets.LoadJson("atlas-msdf.json", &atlasJson)
// 	check(err)
// 	atlas, err := glitch.AtlasFromSdf(atlasJson, atlasImg)
// 	check(err)
// 	atlas.Material().SetUniform("u_threshold", 0.6) // Overwrite the default

// 	screenScale := 1.0 // This is just a weird scaling number

// 	// A screenspace camera
// 	camera := glitch.NewCameraOrtho()
// 	camera.SetOrtho2D(win.Bounds())
// 	camera.SetView2D(0, 0, screenScale, screenScale)
// 	sorter := glitch.NewSorter()
// 	group := ui.NewGroup(win, camera, atlas, sorter)
// 	group.Debug = true

// 	textStyle := ui.NewTextStyle().Scale(4)// .Autofit(true)
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
// 		glitch.SetCamera(camera)

// 		// mx, my := win.MousePosition()
// 		// log.Println("Mouse: ", mx, my)

// 		ui.Clear()
// 		group.Clear()

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

// 		glitch.Clear(win, glitch.Greyscale(0.5))

// 		// shader.Use()
// 		// shader.SetUniform("projection", camera.Projection)
// 		// shader.SetUniform("view", camera.View)

// 		// tpp := float32(1.0/screenScale)
// 		// tpp := float32(512.0 / 1920.0) // Texels per screen pixel
// 		tpp := float32(1.0 / 8.0)
// 		shader.SetUniform("texelsPerPixel", tpp)

// 		group.Draw(win)

// 		win.Update()
// 	}
// }


// // func runGame() {
// // 	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
// // 		Vsync:   true,
// // 		Samples: 0,
// // 	})
// // 	if err != nil {
// // 		panic(err)
// // 	}

// // 	shader, err := glitch.NewShader(shaders.PixelArtShader)
// // 	if err != nil {
// // 		panic(err)
// // 	}
// // 	pass := glitch.NewRenderPass(shader)
// // 	// pass.SoftwareSort = glitch.SoftwareSortY
// // 	// pass.DepthTest = true
// // 	// pass.DepthBump = true

// // 	buttonImage, err := assets.LoadImage("button.png")
// // 	if err != nil {
// // 		panic(err)
// // 	}
// // 	buttonHoverImage, err := assets.LoadImage("button_hover.png")
// // 	if err != nil {
// // 		panic(err)
// // 	}
// // 	buttonPressImage, err := assets.LoadImage("button_press.png")
// // 	if err != nil {
// // 		panic(err)
// // 	}
// // 	panelImage, err := assets.LoadImage("panel.png")
// // 	if err != nil {
// // 		panic(err)
// // 	}
// // 	panelInnerImage, err := assets.LoadImage("panel_inner.png")
// // 	if err != nil {
// // 		panic(err)
// // 	}

// // 	scale := 4.0
// // 	texture := glitch.NewTexture(buttonImage, false)
// // 	buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds(), glitch.R(1, 1, 1, 1))
// // 	buttonSprite.Scale = scale

// // 	texture2 := glitch.NewTexture(buttonPressImage, false)
// // 	buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds(), glitch.R(1, 1, 1, 1))
// // 	buttonPressSprite.Scale = scale

// // 	texture3 := glitch.NewTexture(buttonHoverImage, false)
// // 	buttonHoverSprite := glitch.NewNinePanelSprite(texture3, texture3.Bounds(), glitch.R(1, 1, 1, 1))
// // 	buttonHoverSprite.Scale = scale

// // 	texture4 := glitch.NewTexture(panelImage, false)
// // 	panelSprite := glitch.NewNinePanelSprite(texture4, texture4.Bounds(), glitch.R(2, 2, 2, 2))
// // 	panelSprite.Scale = scale

// // 	panelInnerTex := glitch.NewTexture(panelInnerImage, false)
// // 	panelInnerSprite := glitch.NewNinePanelSprite(panelInnerTex, panelInnerTex.Bounds(), glitch.R(2, 2, 2, 2))
// // 	panelInnerSprite.Scale = scale
// // 	// panelInnerSprite.Mask = glitch.RGBA{1, 0, 0, 1}

// // 	// Text
// // 	atlas, err := glitch.BasicFontAtlas()
// // 	if err != nil {
// // 		panic(err)
// // 	}

// // 	screenScale := 1.5 // This is just a weird scaling number

// // 	// A screenspace camera
// // 	camera := glitch.NewCameraOrtho()
// // 	camera.SetOrtho2D(win.Bounds())
// // 	camera.SetView2D(0, 0, screenScale, screenScale)
// // 	group := ui.NewGroup(win, camera, atlas, pass)
// // 	// group.Debug = true

// // 	textStyle := ui.NewTextStyle().Scale(4)
// // 	buttonStyle := ui.Style{
// // 		Normal:  ui.NewSpriteStyle(buttonSprite, glitch.White),
// // 		Hovered: ui.NewSpriteStyle(buttonHoverSprite, glitch.White),
// // 		Pressed: ui.NewSpriteStyle(buttonPressSprite, glitch.White),
// // 		Text:    textStyle,
// // 	}

// // 	for !win.Closed() {
// // 		if win.Pressed(glitch.KeyEscape) {
// // 			win.Close()
// // 		}

// // 		camera.SetOrtho2D(win.Bounds())
// // 		camera.SetView2D(0, 0, screenScale, screenScale)

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
// // 			if group.Button("-", r.CutLeft(r.W()/2), buttonStyle) {
// // 				screenScale -= 0.1
// // 			}
// // 			if group.Button("+", r, buttonStyle) {
// // 				screenScale += 0.1
// // 			}
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
// // 		// tpp := float32(1.0/screenScale)
// // 		// tpp := float32(512.0 / 1920.0) // Texels per screen pixel
// // 		tpp := float32(1.0 / 8.0)
// // 		pass.SetUniform("texelsPerPixel", tpp)
// // 		pass.Draw(win)

// // 		win.Update()
// // 	}
// // }

// // // package main

// // // import (
// // // 	"embed"
// // // 	"image"
// // // 	"image/draw"
// // // 	_ "image/png"
// // // 	"log"

// // // 	"github.com/unitoftime/glitch"
// // // 	"github.com/unitoftime/glitch/shaders"
// // // 	"github.com/unitoftime/glitch/ui"
// // // )

// // // //go:embed button.png button_hover.png button_press.png panel.png panel_inner.png
// // // var f embed.FS

// // // func loadImage(path string) (*image.NRGBA, error) {
// // // 	file, err := f.Open(path)
// // // 	if err != nil {
// // // 		return nil, err
// // // 	}
// // // 	img, _, err := image.Decode(file)
// // // 	if err != nil {
// // // 		return nil, err
// // // 	}
// // // 	bounds := img.Bounds()
// // // 	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
// // // 	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
// // // 	return nrgba, nil
// // // }

// // // func main() {
// // // 	log.Println("Begin")
// // // 	glitch.Run(runGame)
// // // }

// // // func runGame() {
// // // 	win, err := glitch.NewWindow(1920, 1080, "Glitch UI Demo", glitch.WindowConfig{
// // // 		Vsync: false,
// // // 		Samples: 0,
// // // 	})
// // // 	if err != nil { panic(err) }

// // // 	shader, err := glitch.NewShader(shaders.SpriteShader)
// // // 	if err != nil { panic(err) }
// // // 	pass := glitch.NewRenderPass(shader)
// // // 	pass.SoftwareSort = glitch.SoftwareSortY
// // // 	pass.DepthTest = true
// // // 	pass.DepthBump = true

// // // 	buttonImage, err := loadImage("button.png")
// // // 	if err != nil { panic(err) }
// // // 	buttonHoverImage, err := loadImage("button_hover.png")
// // // 	if err != nil { panic(err) }
// // // 	buttonPressImage, err := loadImage("button_press.png")
// // // 	if err != nil { panic(err) }
// // // 	panelImage, err := loadImage("panel.png")
// // // 	if err != nil { panic(err) }
// // // 	panelInnerImage, err := loadImage("panel_inner.png")
// // // 	if err != nil { panic(err) }

// // // 	texture := glitch.NewTexture(buttonImage, false)
// // // 	buttonSprite := glitch.NewNinePanelSprite(texture, texture.Bounds(), glitch.R(1, 1, 1, 1))
// // // 	buttonSprite.Scale = 1

// // // 	texture2 := glitch.NewTexture(buttonPressImage, false)
// // // 	buttonPressSprite := glitch.NewNinePanelSprite(texture2, texture2.Bounds(), glitch.R(1, 1, 1, 1))
// // // 	buttonPressSprite.Scale = 1

// // // 	texture3 := glitch.NewTexture(buttonHoverImage, false)
// // // 	buttonHoverSprite := glitch.NewNinePanelSprite(texture3, texture3.Bounds(), glitch.R(1, 1, 1, 1))
// // // 	buttonHoverSprite.Scale = 1

// // // 	texture4 := glitch.NewTexture(panelImage, false)
// // // 	panelSprite := glitch.NewNinePanelSprite(texture4, texture4.Bounds(), glitch.R(2, 2, 2, 2))
// // // 	panelSprite.Scale = 1

// // // 	panelInnerTex := glitch.NewTexture(panelInnerImage, false)
// // // 	panelInnerSprite := glitch.NewNinePanelSprite(panelInnerTex, panelInnerTex.Bounds(), glitch.R(2, 2, 2, 2))
// // // 	panelInnerSprite.Scale = 1
// // // 	// panelInnerSprite.Mask = glitch.RGBA{1, 0, 0, 1}

// // // 	// Text
// // // 	atlas, err := glitch.DefaultAtlas()
// // // 	if err != nil { panic(err) }

// // // 	// A screenspace camera
// // // 	camera := glitch.NewCameraOrtho()
// // // 	camera.SetOrtho2D(win.Bounds())
// // // 	camera.SetView2D(0, 0, 1.0, 1.0)
// // // 	group := ui.NewGroup(win, camera, atlas, pass)
// // // 	// group.Debug = true

// // // 	textStyle := ui.NewTextStyle().Scale(1)
// // // 	buttonStyle := ui.Style{
// // // 		Normal: ui.NewSpriteStyle(buttonSprite, glitch.White),
// // // 		Hovered: ui.NewSpriteStyle(buttonHoverSprite, glitch.White),
// // // 		Pressed: ui.NewSpriteStyle(buttonPressSprite, glitch.White),
// // // 		Text: textStyle,
// // // 	}

// // // 	for !win.Closed() {
// // // 		if win.Pressed(glitch.KeyEscape) {
// // // 			win.Close()
// // // 		}

// // // 		// mx, my := win.MousePosition()
// // // 		// log.Println("Mouse: ", mx, my)

// // // 		glitch.Clear(win, glitch.Black)

// // // 		ui.Clear()
// // // 		group.Clear()
// // // 		pass.Clear()

// // // 		menuRect := win.Bounds().SliceHorizontal(500).SliceVertical(500)
// // // 		group.Panel(panelSprite, menuRect, glitch.White)

// // // 		menuRect.CutLeft(20)
// // // 		menuRect.CutRight(20)
// // // 		{
// // // 			r := menuRect.CutTop(100)
// // // 			group.Text("Menu", r, textStyle)
// // // 		}
// // // 		menuRect.CutTop(10) // Padding
// // // 		{
// // // 			r := menuRect.CutTop(100)
// // // 			buttonStyle.Text = buttonStyle.Text.Color(glitch.Black)
// // // 			if group.Button("Button 0", r, buttonStyle) {
// // // 				println("Button 0")
// // // 			}
// // // 		}
// // // 		menuRect.CutTop(10) // Padding
// // // 		{
// // // 			r := menuRect.CutTop(100)
// // // 			buttonStyle.Text = buttonStyle.Text.Color(glitch.White)
// // // 			if group.Button("Button 1", r, buttonStyle) {
// // // 				println("Button 1")
// // // 			}
// // // 		}

// // // 		menuRect.CutTop(10) // Padding
// // // 		{
// // // 			r := menuRect.CutTop(100)
// // // 			group.Panel(panelSprite, r, glitch.White)
// // // 			group.Panel(panelInnerSprite, r, glitch.White)
// // // 		}

// // // 		pass.SetCamera2D(camera)
// // // 		pass.Draw(win)

// // // 		win.Update()
// // // 	}
// // // }
