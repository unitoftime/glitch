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

// // Pads around center
// func (r Rect) Pad(x, y float64) Rect {
// 	if r.W() * r.H() == 0 { return Rect{r.ResizedMin(pixel.V(2*x, 2*y))} }
// 	return Rect{r.Resized(r.Center(), r.Size().Add(pixel.V(2*x, 2*y)))}
// }

// // Scales around center
// func (r Rect) Scale(scale float64) Rect {
// 	if r.W() * r.H() == 0 { return Rect{r.ResizedMin(pixel.V(scale * r.W(), scale * r.H()))} }
// 	return Rect{r.Resized(r.Center(), pixel.V(scale * r.W(), scale * r.H()))}
// }

// func (r Rect) ScaledXY(anchor pixel.Vec, x,y float64) Rect {
// 	if r.W() * r.H() == 0 { return Rect{r.ResizedMin(pixel.V(x * r.W(), y * r.H()))} }
// 	return Rect{r.Resized(anchor, pixel.V(x * r.W(), y * r.H()))}
// }

// func (r Rect) Round() Rect {
// 	return Rect{pixel.R(
// 		math.Round(r.Min.X),
// 		math.Round(r.Min.Y),
// 		math.Round(r.Max.X),
// 		math.Round(r.Max.Y),
// 	)}
// }

// // Maintains the current union of all bounds
// func (c *Context) appendUnionBounds(newBounds Rect) {
// 	if c.unionBounds == nil {
// 		c.unionBounds = &newBounds
// 	} else {
// 		newUnion := c.unionBounds.Union(newBounds)
// 		c.unionBounds = &newUnion
// 	}
// }


// func (c *Context) Place(position, size, pivot pixel.Vec) Rect {
// 	destRect := pixel.R(
// 		position.X - (size.X * pivot.X),
// 		position.Y - (size.Y * pivot.Y),
// 		position.X + (size.X * (1-pivot.X)),
// 		position.Y + (size.Y * (1-pivot.Y)),
// 	)
// 	return Rect{destRect}
// }

// func (c *Context) HoverSlicedSprite(normal, hovered string, bounds Rect) bool {
// 	mousePos := c.win.MousePosition()
// 	if bounds.Contains(mousePos) {
// 		c.SlicedSprite(hovered, bounds)
// 		return true
// 	}

// 	c.SlicedSprite(normal, bounds)
// 	return false
// }

// func (c *Context) ButtonSprite(normal, hovered string, bounds Rect) bool {
// 	isHovered := c.HoverSprite(normal, hovered, bounds)

// 	if isHovered && c.win.JustPressed(pixelgl.MouseButtonLeft) {
// 		return true
// 	}
// 	return false
// }

// func (c *Context) ButtonSlicedSprite(normal, hovered string, bounds Rect) bool {
// 	isHovered := c.HoverSlicedSprite(normal, hovered, bounds)

// 	if isHovered && c.win.JustPressed(pixelgl.MouseButtonLeft) {
// 		return true
// 	}
// 	return false
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

// func (c *Context) Bar(outer, inner string, bounds Rect, value float64, innerColor color.Color) Rect {
// 	_, barInner := c.SlicedSprite(outer, bounds)
// 	c.SpriteColorMask(inner, barInner.ScaledXY(barInner.Anchor(0, 0), value, 1), innerColor)
// 	return bounds
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

// // func (c *Context) ButtonLambda(bounds Rect, ifLambda, elseLambda func()) Rect {
// // 	mousePos := c.win.MousePosition()
// // 	if bounds.Contains(mousePos) && c.win.JustPressed(pixelgl.MouseButtonLeft) {
// // 		ifLambda()
// // 	} else {
// // 		elseLambda()
// // 	}
// // 	return bounds
// // }


// func (c *Context) DebugRect(destRect Rect) {
// 	if !c.Debug { return } // Only draw if debug mode

// 	c.debugImd.Push(destRect.Min)
// 	c.debugImd.Push(destRect.Max)
// 	c.debugImd.Rectangle(1)
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

// // func (c *Context) SpriteFull(name string, size, position, pivot pixel.Vec) {
// // 	// TODO - Round?
// // 	// This is the exact bounds of the rectangle we will draw to
// // 	destRect := pixel.R(
// // 		position.X - (size.X * pivot.X),
// // 		position.Y - (size.Y * pivot.Y),
// // 		position.X + (size.X * (1-pivot.X)),
// // 		position.Y + (size.Y * (1-pivot.Y)),
// // 	)

// // 	c.appendUnionBounds(Rect{destRect})
// // 	c.Sprite(name, Rect{destRect})
// // }

// func (c *Context) SlicedSprite(name string, destRect Rect) (Rect, Rect) {
// 	destRect = destRect.Round()
// 	// sprites, mats, finalBounds, borderSize := getNinePanelDraw(name, destRect)
// 	sprites, mats, finalBounds, innerBounds := c.getNinePanelDraw(name, destRect)
// 	for i := range sprites {
// 		sprites[i].Draw(c.batch, mats[i])
// 	}

// 	c.DebugRect(destRect)

// 	c.appendUnionBounds(Rect{finalBounds})
// 	return Rect{finalBounds}, Rect{innerBounds}
// }

// // func (c *Context) PanelFull(name string, size, position, pivot pixel.Vec) (Rect, Rect) {
// // 	destRect := pixel.R(
// // 		position.X - (size.X * pivot.X),
// // 		position.Y - (size.Y * pivot.Y),
// // 		position.X + (size.X * (1-pivot.X)),
// // 		position.Y + (size.Y * (1-pivot.Y)),
// // 	)
// // 	return c.SlicedSprite(name, Rect{destRect})
// // }

// // Generates a set of sprites and mat draws for a 9 panel
// // Assumes 16x16 pixel in the middle of scalable sprite, then everything else is inferred as a corner or edge
// // TODO - things have to be perfectly aligned... Not super flexible for non-symmetric panels
// func (ctx *Context) getNinePanelDraw(name string, destRect Rect) ([]*pixel.Sprite, []pixel.Matrix, pixel.Rect, pixel.Rect) {
// 	scale := 2.0 // Scale to use for borders

// 	mats := make([]pixel.Matrix, 0)

// 	sprites, rects := GetNinePanel(name, ctx.spritesheet)
// 	sprite, err := ctx.spritesheet.Get(name)
// 	if err != nil { panic(err) }
// 	bounds := ZeroRect(sprite.Frame())

// 	centerSize := float64(16) // TODO - Assume center chunk is 16x16 pixels

// 	borderWidth := scale * (bounds.W() - centerSize) / 2 // This is the width of each vertical border in the destination space
// 	borderHeight := scale * (bounds.H() - centerSize) / 2 // This is the height of each horizontal border in the destination space
// 	destSizeInner := pixel.V(destRect.W() - (2 * borderWidth), destRect.H() - (2 * borderHeight)) // This is the size of the destination inner rectangle (with borders removed)
// 	destRectInner := destRect.Pad(-borderWidth, -borderHeight)

// 	destRects := []pixel.Rect{
// 		pixel.R(-destSizeInner.X/2, -destSizeInner.Y/2, destSizeInner.X/2, destSizeInner.Y/2).Moved(destRect.Center()), // Center
// 		pixel.R(-destSizeInner.X/2, destSizeInner.Y/2, destSizeInner.X/2, destRect.H()/2).Moved(destRect.Center()), // Top
// 		pixel.R(-destSizeInner.X/2, -destRect.H()/2, destSizeInner.X/2, -destSizeInner.Y/2).Moved(destRect.Center()), // Bottom
// 		pixel.R(-destRect.W()/2, -destSizeInner.Y/2, -destSizeInner.X/2, destSizeInner.Y/2).Moved(destRect.Center()), // Left
// 		pixel.R(destSizeInner.X/2, -destSizeInner.Y/2, destRect.W()/2, destSizeInner.Y/2).Moved(destRect.Center()), // Right
// 		pixel.R(-destRect.W()/2, destSizeInner.Y/2, -destSizeInner.X/2, destRect.H()/2).Moved(destRect.Center()), // TL
// 		pixel.R(destSizeInner.X/2, destSizeInner.Y/2, destRect.W()/2, destRect.H()/2).Moved(destRect.Center()), // TR
// 		pixel.R(-destRect.W()/2, -destRect.H()/2, -destSizeInner.X/2, -destSizeInner.Y/2).Moved(destRect.Center()), // BL
// 		pixel.R(destSizeInner.X/2, -destRect.H()/2, destRect.W()/2, -destSizeInner.Y/2).Moved(destRect.Center()), // BR
// 	}

// 	finalBounds := pixel.Rect{}
// 	for i := range rects {
// 		bounds = ZeroRect(sprites[i].Frame())

// 		scale := pixel.V(destRects[i].W() / rects[i].W(), destRects[i].H() / rects[i].H())
// 		mat := pixel.IM.ScaledXY(pixel.ZV, scale)
// 		mat = mat.Moved(destRects[i].Center())
// 		mats = append(mats, mat)

// 		if i == 0 {
// 			finalBounds = pixel.Rect{mat.Project(bounds.Min), mat.Project(bounds.Max)}
// 		} else {
// 			finalBounds = finalBounds.Union(pixel.Rect{mat.Project(bounds.Min), mat.Project(bounds.Max)})
// 		}
// 	}

// 	return sprites, mats, finalBounds, destRectInner.Rect
// }


// func GetNinePanel(name string, s *asset.Spritesheet) ([]*pixel.Sprite, []pixel.Rect) {
// 	sprite, err := s.Get(name)
// 	if err != nil { panic(err) }
// 	sprites := make([]*pixel.Sprite, 0)

// 	centerSize := float64(16) // TODO - Assume center chunk is 16x16 pixels
// 	c := sprite.Frame().Center()

// 	rects := []pixel.Rect{
// 		pixel.R(c.X - centerSize/2, c.Y - centerSize/2, c.X + centerSize/2, c.Y + centerSize/2), // Center
// 		pixel.R(c.X - centerSize/2, c.Y + centerSize/2, c.X + centerSize/2, c.Y + sprite.Frame().H()/2), // Top
// 		pixel.R(c.X - centerSize/2, c.Y - sprite.Frame().H()/2, c.X + centerSize/2, c.Y - centerSize/2), // Bottom
// 		pixel.R(c.X - sprite.Frame().W()/2, c.Y - centerSize/2, c.X - centerSize/2, c.Y + centerSize/2), // Left
// 		pixel.R(c.X + centerSize/2, c.Y - centerSize/2, c.X + sprite.Frame().W()/2, c.Y + centerSize/2), // Right
// 		pixel.R(c.X - sprite.Frame().W()/2, c.Y + centerSize/2, c.X - centerSize/2, c.Y + sprite.Frame().H()/2), // TL
// 		pixel.R(c.X + centerSize/2, c.Y + centerSize/2, c.X + sprite.Frame().W()/2, c.Y + sprite.Frame().H()/2), // TR
// 		pixel.R(c.X - sprite.Frame().W()/2, c.Y - sprite.Frame().H()/2, c.X - centerSize/2, c.Y - centerSize/2), // BL
// 		pixel.R(c.X + centerSize/2, c.Y - sprite.Frame().H()/2, c.X + sprite.Frame().W()/2, c.Y - centerSize/2), // BR
// 	}

// 	for i := range rects {
// 		sprites = append(sprites, pixel.NewSprite(sprite.Picture(), rects[i]))
// 	}
// 	return sprites, rects
// }
