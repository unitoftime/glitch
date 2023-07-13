//go:build js || wasm
package debugui

import (
	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/internal/glfw"
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
}

func NewImgui(win GuiWindow) *Gui {
	return &Gui{}
}

func (g *Gui) NewFrame(win *glitch.Window) {
}

func (g *Gui) Draw(win *glitch.Window) {
}

func (g *Gui) Terminate() {
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Most of this is taken directly from:
// https://github.com/inkyblackness/imgui-go-examples/blob/master/internal/platforms/glfw.go
////////////////////////////////////////////////////////////////////////////////////////////////////
func (g *Gui) setKeyMapping() {
}

func (g *Gui) installCallbacks(win GuiWindow) {
}

func (g *Gui) mouseButtonChange(window *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
}

func (g *Gui) mouseScrollChange(window *glfw.Window, x, y float64) {
}

func (g *Gui) keyChange(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
}

func (g *Gui) charChange(window *glfw.Window, char rune) {
}

// // ClipboardText returns the current clipboard text, if available.
// func (platform *GLFW) ClipboardText() (string, error) {
// 	return platform.window.GetClipboardString()
// }

// // SetClipboardText sets the text as the current clipboard text.
// func (platform *GLFW) SetClipboardText(text string) {
// 	platform.window.SetClipboardString(text)
// }
