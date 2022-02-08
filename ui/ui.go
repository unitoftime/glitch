package ui

import (
	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/shaders"
)

type Drawer interface {
	Bounds() glitch.Rect
	RectDraw(*glitch.RenderPass, glitch.Rect)
}

type Group struct {
	win *glitch.Window
	pass *glitch.RenderPass
	camera *glitch.CameraOrtho
	atlas *glitch.Atlas
	// unionBounds *Rect // A union of all drawn object's bounds
	Debug bool
	geomDraw glitch.GeomDraw
}

func NewGroup(win *glitch.Window, atlas *glitch.Atlas) *Group {
	shader, err := glitch.NewShader(shaders.SpriteShader)
	if err != nil { panic(err) }
	pass := glitch.NewRenderPass(shader)

	return &Group{
		win: win,
		camera: glitch.NewCameraOrtho(),
		pass: pass,
		atlas: atlas,
		// unionBounds: nil,
		Debug: false,
	}
}

// func (g *Group) ContainsMouse() bool {
// 	if g.unionBounds == nil { return false }
// 	return g.unionBounds.Contains(g.win.MousePosition())
// }

func (g *Group) Clear() {
	g.pass.Clear()
	// g.unionBounds = nil
}

func (g *Group) Draw() {
	g.camera.SetOrtho2D(g.win)
	g.camera.SetView2D(0, 0, 1.0, 1.0)

	g.pass.SetUniform("projection", g.camera.Projection)
	g.pass.SetUniform("view", g.camera.View)

	g.pass.Draw(g.win)

	// if g.Debug {
	// 	if g.unionBounds != nil {
	// 		g.debugImd.Color = pixel.RGB(0, 0, 1)
	// 		g.DebugRect(*g.unionBounds)
	// 	}
	// 	g.debugImd.Draw(g.win)
	// }
}

func (g *Group) Panel(sprite Drawer, rect glitch.Rect) {
	sprite.RectDraw(g.pass, rect)
	g.debugRect(rect)
}

func (g *Group) Hover(normal, hovered Drawer, rect glitch.Rect) bool {
	mX, mY := g.win.MousePosition()
	if rect.Contains(mX, mY) {
		g.Panel(hovered, rect)
		return true
	}

	g.Panel(normal, rect)
	return false
}

func (g *Group) Button(normal, hovered, pressed Drawer, rect glitch.Rect) bool {
	mX, mY := g.win.MousePosition()
	if !rect.Contains(mX, mY) {
		g.Panel(normal, rect)
		return false
	}

	// If we are here, then we know we are at least hovering
	if g.win.Pressed(glitch.MouseButtonLeft) {
		g.Panel(pressed, rect)
		return true
	}

	g.Panel(hovered, rect)
	return false
}

func (g *Group) Text(str string, rect glitch.Rect, anchor glitch.Vec2) {
	text := g.atlas.Text(str)
	r := rect.Anchor(text.Bounds(), anchor)
	text.DrawRect(g.pass, r, glitch.RGBA{0, 0, 0, 1.0})
	g.debugRect(rect)
}

func (g *Group) debugRect(rect glitch.Rect) {
	if !g.Debug { return }

	lineWidth := float32(2.0)

	g.pass.SetLayer(126)

	g.geomDraw.SetColor(glitch.RGBA{1.0, 0, 0, 1.0})
	m := g.geomDraw.Rectangle(rect, lineWidth)
	m.Draw(g.pass, glitch.Mat4Ident)
	g.pass.SetLayer(glitch.DefaultLayer)
}

// func (g *Group) Bar(outer, inner Drawer, bounds glitch.Rect, value float64) Rect {
	
// 	_, barInner := c.SlicedSprite(outer, bounds)
// 	c.SpriteColorMask(inner, barInner.ScaledXY(barInner.Anchor(0, 0), value, 1), innerColor)
// 	return bounds
// }

// // TODO - tooltips only seem to work for single lines
// func (c *Context) Tooltip(name string, tip string, startRect Rect, position, anchor pixel.Vec) {
// 	padding := 5.0
// 	c.HoverLambda(startRect,
// 		func() {
// 			textBounds := c.MeasureText(tip, position, anchor)
// 			c.SlicedSprite(name, textBounds.Pad(padding, padding))
// 			c.Text(tip, position, anchor)
// 		},
// 		func() {  })
// }

// func (c *Context) BarVert(outer, inner string, bounds Rect, value float64, innerColor color.Color) Rect {
// 	_, barInner := c.SlicedSprite(outer, bounds)
// 	c.SpriteColorMask(inner, barInner.ScaledXY(barInner.Anchor(0, 0), 1, value), innerColor)
// 	return bounds
// }

// func (c *Context) HoverLambda(bounds Rect, ifLambda, elseLambda func()) Rect {
// 	mousePos := c.win.MousePosition()
// 	if bounds.Contains(mousePos) {
// 		ifLambda()
// 	} else {
// 		elseLambda()
// 	}
// 	return bounds
// }

// // TODO - this might have issues with it's bounding box being slightly off because of the rotation
// func (c *Context) SpriteRotated(name string, destRect Rect, radians float64) {
// 	destRect = destRect.Round()

// 	sprite, err := c.spritesheet.Get(name)
// 	if err != nil { panic(err) }
// 	bounds := ZeroRect(sprite.Frame())

// 	scale := pixel.V(destRect.W() / bounds.W(), destRect.H() / bounds.H())
// 	mat := pixel.IM.ScaledXY(pixel.ZV, scale).Rotated(pixel.ZV, radians)
// 	mat = mat.Moved(destRect.Center())

// 	c.appendUnionBounds(destRect)
// 	sprite.Draw(c.batch, mat)

// 	c.DebugRect(destRect)
// }
