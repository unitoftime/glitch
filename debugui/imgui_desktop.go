//go:build !js

package debugui

import (
	"log"

	"github.com/inkyblackness/imgui-go/v4"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/internal/glfw"
	"github.com/unitoftime/glitch/internal/mainthread"
)

type GuiWindow interface {
	DisplaySize() [2]float32
	FramebufferSize() [2]float32
	AddMouseButtonCallback(glfw.MouseButtonCallback)
	AddScrollCallback(glfw.ScrollCallback)
	AddKeyCallback(glfw.KeyCallback)
	AddCharCallback(glfw.CharCallback)
}

type Gui struct {
	context  *imgui.Context
	io       imgui.IO
	renderer *OpenGL3

	mouseJustPressed [3]bool
	win              *glitch.Window
}

func NewImgui(win *glitch.Window) *Gui {
	g := Gui{
		win: win,
	}
	mainthread.Call(func() {
		g.context = imgui.CreateContext(nil)
		g.io = imgui.CurrentIO()

		renderer, err := NewOpenGL3(g.io)
		if err != nil {
			log.Fatal(err)
		}
		g.renderer = renderer

		g.setKeyMapping()
		g.installCallbacks(win)
	})
	return &g
}

func (g *Gui) NewFrame() {
	mainthread.Call(func() {
		ds := g.win.DisplaySize()
		mouseX, mouseY := g.win.GetMouse()

		g.io.SetDisplaySize(imgui.Vec2{X: ds[0], Y: ds[1]})
		g.io.SetDeltaTime(float32(1 / 60.0))
		// TODO - before setting this check if window is focused
		g.io.SetMousePosition(imgui.Vec2{X: float32(mouseX), Y: float32(mouseY)})

		for i := 0; i < len(g.mouseJustPressed); i++ {
			down := g.mouseJustPressed[i] || (g.win.GetMouseButton(glfwButtonIDByIndex[i]) == glfw.Press)
			g.io.SetMouseButtonDown(i, down)
			g.mouseJustPressed[i] = false
		}

		imgui.NewFrame()
	})
}

func (g *Gui) Captured() bool {
	return g.io.WantCaptureMouse() || g.io.WantCaptureKeyboard()
}

func (g *Gui) Draw() {
	mainthread.Call(func() {
		imgui.Render()

		ds := g.win.DisplaySize()
		fs := g.win.FramebufferSize()
		g.renderer.Render(ds, fs, imgui.RenderedDrawData())
	})
}

func (g *Gui) Terminate() {
	mainthread.Call(func() {
		g.renderer.Dispose()
		g.context.Destroy()
	})
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
// Most of this is taken directly from:
// https://github.com/inkyblackness/imgui-go-examples/blob/master/internal/platforms/glfw.go
// //////////////////////////////////////////////////////////////////////////////////////////////////
func (g *Gui) setKeyMapping() {
	// Keyboard mapping. ImGui will use those indices to peek into the io.KeysDown[] array.
	g.io.KeyMap(imgui.KeyTab, int(glfw.KeyTab))
	g.io.KeyMap(imgui.KeyLeftArrow, int(glfw.KeyLeft))
	g.io.KeyMap(imgui.KeyRightArrow, int(glfw.KeyRight))
	g.io.KeyMap(imgui.KeyUpArrow, int(glfw.KeyUp))
	g.io.KeyMap(imgui.KeyDownArrow, int(glfw.KeyDown))
	g.io.KeyMap(imgui.KeyPageUp, int(glfw.KeyPageUp))
	g.io.KeyMap(imgui.KeyPageDown, int(glfw.KeyPageDown))
	g.io.KeyMap(imgui.KeyHome, int(glfw.KeyHome))
	g.io.KeyMap(imgui.KeyEnd, int(glfw.KeyEnd))
	g.io.KeyMap(imgui.KeyInsert, int(glfw.KeyInsert))
	g.io.KeyMap(imgui.KeyDelete, int(glfw.KeyDelete))
	g.io.KeyMap(imgui.KeyBackspace, int(glfw.KeyBackspace))
	g.io.KeyMap(imgui.KeySpace, int(glfw.KeySpace))
	g.io.KeyMap(imgui.KeyEnter, int(glfw.KeyEnter))
	g.io.KeyMap(imgui.KeyEscape, int(glfw.KeyEscape))
	g.io.KeyMap(imgui.KeyA, int(glfw.KeyA))
	g.io.KeyMap(imgui.KeyC, int(glfw.KeyC))
	g.io.KeyMap(imgui.KeyV, int(glfw.KeyV))
	g.io.KeyMap(imgui.KeyX, int(glfw.KeyX))
	g.io.KeyMap(imgui.KeyY, int(glfw.KeyY))
	g.io.KeyMap(imgui.KeyZ, int(glfw.KeyZ))
}

func (g *Gui) installCallbacks(win GuiWindow) {
	win.AddMouseButtonCallback(g.mouseButtonChange)
	win.AddScrollCallback(g.mouseScrollChange)
	win.AddKeyCallback(g.keyChange)
	win.AddCharCallback(g.charChange)
}

var glfwButtonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: 0,
	glfw.MouseButton2: 1,
	glfw.MouseButton3: 2,
}

var glfwButtonIDByIndex = map[int]glfw.MouseButton{
	0: glfw.MouseButton1,
	1: glfw.MouseButton2,
	2: glfw.MouseButton3,
}

func (g *Gui) mouseButtonChange(window *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	buttonIndex, known := glfwButtonIndexByID[rawButton]

	if known && (action == glfw.Press) {
		g.mouseJustPressed[buttonIndex] = true
	}
}

func (g *Gui) mouseScrollChange(window *glfw.Window, x, y float64) {
	g.io.AddMouseWheelDelta(float32(x), float32(y))
}

func (g *Gui) keyChange(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		g.io.KeyPress(int(key))
	}
	if action == glfw.Release {
		g.io.KeyRelease(int(key))
	}

	// Modifiers are not reliable across systems
	g.io.KeyCtrl(int(glfw.KeyLeftControl), int(glfw.KeyRightControl))
	g.io.KeyShift(int(glfw.KeyLeftShift), int(glfw.KeyRightShift))
	g.io.KeyAlt(int(glfw.KeyLeftAlt), int(glfw.KeyRightAlt))
	g.io.KeySuper(int(glfw.KeyLeftSuper), int(glfw.KeyRightSuper))
}

func (g *Gui) charChange(window *glfw.Window, char rune) {
	g.io.AddInputCharacters(string(char))
}

// // ClipboardText returns the current clipboard text, if available.
// func (platform *GLFW) ClipboardText() (string, error) {
// 	return platform.window.GetClipboardString()
// }

// // SetClipboardText sets the text as the current clipboard text.
// func (platform *GLFW) SetClipboardText(text string) {
// 	platform.window.SetClipboardString(text)
// }
