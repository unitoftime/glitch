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
	unionBounds *glitch.Rect // A union of all drawn object's bounds
	allBounds []glitch.Rect
	Debug bool
	OnlyCheckUnion bool
	geomDraw glitch.GeomDraw
	color glitch.RGBA
}

func NewGroup(win *glitch.Window, camera *glitch.CameraOrtho, atlas *glitch.Atlas) *Group {
	shader, err := glitch.NewShader(shaders.SpriteShader)
	if err != nil { panic(err) }
	pass := glitch.NewRenderPass(shader)

	return &Group{
		win: win,
		camera: camera,
		pass: pass,
		atlas: atlas,
		unionBounds: nil,
		allBounds: make([]glitch.Rect, 0),
		Debug: false,
		OnlyCheckUnion: true,
		color: glitch.RGBA{1, 1, 1, 1},
	}
}

// onlyCheckUnion is an optimization if the ui elements are tightly packed (it doesn't loop through each rect
func (g *Group) ContainsMouse() bool {
	if g.OnlyCheckUnion {
		if g.unionBounds == nil { return false }
		return g.unionBounds.Contains(g.mousePosition())
	} else {
		x, y := g.mousePosition()
		for i := range g.allBounds {
			if g.allBounds[i].Contains(x, y) {
				return true
			}
		}
	}
	return false
}

func (g *Group) mousePosition() (float32, float32) {
	x, y := g.win.MousePosition()
	worldSpaceMouse := g.camera.Unproject(glitch.Vec3{x, y, 0})
	return worldSpaceMouse[0], worldSpaceMouse[1]
}

// TODO - Should this be a list of rects that we loop through?
func (g *Group) appendUnionBounds(newBounds glitch.Rect) {
	g.allBounds = append(g.allBounds, newBounds)

	if g.unionBounds == nil {
		g.unionBounds = &newBounds
	} else {
		newUnion := g.unionBounds.Union(newBounds)
		g.unionBounds = &newUnion
	}
}

func (g *Group) Clear() {
	g.pass.Clear()
	g.unionBounds = nil
	g.allBounds = g.allBounds[:0]
}

// Performs a draw of the UI Group
func (g *Group) Draw() {
	g.pass.SetUniform("projection", g.camera.Projection)
	g.pass.SetUniform("view", g.camera.View)

	// Draw the union rect
	if g.Debug {
		if g.unionBounds != nil {
			g.debugRect(*g.unionBounds)
		}
	}

	g.pass.Draw(g.win)
}

func (g *Group) SetColor(color glitch.RGBA) {
	g.color = color
}

func (g *Group) Panel(sprite Drawer, rect glitch.Rect) {
	sprite.RectDraw(g.pass, rect)
	g.appendUnionBounds(rect)
	g.debugRect(rect)
}

func (g *Group) Hover(normal, hovered Drawer, rect glitch.Rect) bool {
	mX, mY := g.mousePosition()
	if rect.Contains(mX, mY) {
		g.Panel(hovered, rect)
		return true
	}

	g.Panel(normal, rect)
	return false
}

func (g *Group) Button(normal, hovered, pressed Drawer, rect glitch.Rect) bool {
	mX, mY := g.mousePosition()
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
	text.DrawRect(g.pass, r, g.color)
	g.appendUnionBounds(rect)
	g.debugRect(rect)
}

// TODO - tooltips only seem to work for single lines
func (g *Group) Tooltip(panel Drawer, tip string, rect glitch.Rect, anchor glitch.Vec2) {
	mX, mY := g.mousePosition()
	if !rect.Contains(mX, mY) {
		return // If mouse not contained by rect, then don't draw
	}

	text := g.atlas.Text(tip)
	tipRect := rect.Anchor(text.Bounds(), anchor)

	g.Panel(panel, tipRect)

	text.DrawRect(g.pass, tipRect, g.color)
	g.appendUnionBounds(tipRect)
	g.debugRect(tipRect)
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
	// mX, mY := g.mousePosition()
// // 	mousePos := c.win.MousePosition()
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
