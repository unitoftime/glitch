package ui

// onlyCheckUnion is an optimization if the ui elements are tightly packed (it doesn't loop through each rect
// func (g *Group) ContainsMouse() bool {
// 	if g.OnlyCheckUnion {
// 		if !g.unionBoundsSet { return false }
// 		return g.unionBounds.Contains(g.mousePosition())
// 	} else {
// 		x, y := g.mousePosition()
// 		for i := range g.allBounds {
// 			if g.allBounds[i].Contains(x, y) {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func (g *Group) SetColor(color glitch.RGBA) {
// 	g.color = color
// }

// func (g *Group) PanelColorMask(sprite Drawer, rect glitch.Rect, color glitch.RGBA) {
// 	sprite.RectDrawColorMask(g.pass, rect, color)
// 	g.appendUnionBounds(rect)
// 	g.debugRect(rect)
// }

// func (g *Group) Panel(sprite Drawer, rect glitch.Rect) {
// 	if sprite != nil {
// 		sprite.RectDrawColorMask(g.pass, rect, g.color)
// 	}
// 	g.appendUnionBounds(rect)
// 	g.debugRect(rect)
// }

// // Adds a panel with padding to the current bounds of the group
// func (g *Group) PanelizeBounds(sprite Drawer, padding glitch.Rect) {
// 	if !g.unionBoundsSet { return }
// 	rect := g.unionBounds
// 	rect = rect.Pad(padding)
// 	g.Panel(sprite, rect)
// }

// func (g *Group) Hover(normal, hovered Drawer, rect glitch.Rect) bool {
// 	mX, mY := g.mousePosition()
// 	if mouseCheck(rect, glitch.Vec2{mX, mY}) {
// 		g.Panel(hovered, rect)
// 		return true
// 	}

// 	g.Panel(normal, rect)
// 	return false
// }

// func (g *Group) Button(normal, hovered, pressed Drawer, rect glitch.Rect) bool {
// 	mX, mY := g.mousePosition()

// 	if !mouseCheck(rect, glitch.Vec2{mX, mY}) {
// 		g.Panel(normal, rect)
// 		return false
// 	}

// 	// If we are here, then we know we are at least hovering
// 	if g.win.JustPressed(glitch.MouseButtonLeft) {
// 		g.Panel(pressed, rect)
// 		return true
// 	}

// 	g.Panel(hovered, rect)
// 	return false
// }

// // Same thing as a button but returns true for the duration that the button is pressed
// func (g *Group) PressedButton(normal, hovered, pressed Drawer, rect glitch.Rect) bool {
// 	mX, mY := g.mousePosition()

// 	if !mouseCheck(rect, glitch.Vec2{mX, mY}) {
// 		g.Panel(normal, rect)
// 		return false
// 	}

// 	// If we are here, then we know we are at least hovering
// 	if g.win.Pressed(glitch.MouseButtonLeft) {
// 		g.Panel(pressed, rect)
// 		return true
// 	}

// 	g.Panel(hovered, rect)
// 	return false
// }

// // TODO! - text masking around rect?
// func (g *Group) Text(str string, rect glitch.Rect, anchor glitch.Vec2) {
// 	text := g.getText(str)
// 	r := rect.Anchor(text.Bounds().ScaledToFit(rect), anchor)
// 	text.RectDrawColorMask(g.pass, r, g.color)
// 	g.appendUnionBounds(r)
// 	g.debugRect(r)
// }

// // Text, but doesn't automatically scale to fill the rect
// // TODO maybe I should call the other text "AutoText"? or something
// func (g *Group) FixedText(str string, rect glitch.Rect, anchor glitch.Vec2, scale float64) {
// 	text := g.getText(str)
// 	r := rect.Anchor(text.Bounds().Scaled(scale), anchor)
// 	text.RectDrawColorMask(g.pass, r, g.color)
// 	g.appendUnionBounds(r)
// 	g.debugRect(r)
// }

// // TODO - combine with fixedtext
// func (g *Group) FullFixedText(str string, rect glitch.Rect, anchor, anchor2 glitch.Vec2, scale float64) {
// 	text := g.getText(str)
// 	r := rect.FullAnchor(text.Bounds().Scaled(scale), anchor, anchor2)
// 	text.RectDrawColorMask(g.pass, r, g.color)
// 	g.appendUnionBounds(r)
// 	g.debugRect(r)
// }

// func (g *Group) TextInput(prefix, postfix string, str *string, rect glitch.Rect, anchor glitch.Vec2, scale float64) {
// 	if str == nil { return }

// 	runes := g.win.Typed()
// 	*str = *str + string(runes)

// 	tStr := *str
// 	if g.win.JustPressed(glitch.KeyBackspace) {
// 		if g.win.Pressed(glitch.KeyLeftControl) || g.win.Pressed(glitch.KeyRightControl) {
// 			// Delete whole word
// 			lastIndex := strings.LastIndex(strings.TrimRight(tStr, " "), " ")
// 			if lastIndex < 0 {
// 				// Means there were no spaces, delete everything
// 				lastIndex = 0
// 			}

// 			tStr = tStr[:lastIndex]
// 		} else {
// 			if len(tStr) > 0 {
// 				tStr = tStr[:len(tStr)-1]
// 			}
// 		}
// 	} else if g.win.Repeated(glitch.KeyBackspace) {
// 		if len(tStr) > 0 {
// 			tStr = tStr[:len(tStr)-1]
// 		}
// 	}

// 	// ret := false
// 	// if g.win.JustPressed(glitch.KeyEnter) {
// 	// 	ret = true
// 	// }
// 	*str = tStr

// 	g.FixedText(prefix + *str + postfix, rect, anchor, scale)
// 	// return ret
// }

// // TODO - tooltips only seem to work for single lines
// // TODO: Configurable padding
// func (g *Group) Tooltip(panel Drawer, tip string, rect glitch.Rect) {
// 	mX, mY := g.mousePosition()
// 	mousePos := glitch.Vec2{mX, mY}
// 	if !mouseCheck(rect, mousePos) {
// 		return // If mouse not contained by rect, then don't draw
// 	}

// 	padding := 10.0
// 	quadrant := g.win.Bounds().Center().Sub(mousePos).Unit()

// 	var movement glitch.Vec2
// 	if quadrant[0] < 0 {
// 		movement[0] = -1
// 	} else {
// 		movement[0] = 1
// 	}

// 	if quadrant[1] < 0 {
// 		movement[1] = -1
// 	} else {
// 		movement[1] = 1
// 	}

// 	text := g.getText(tip)
// 	// tipRect := rect.Anchor(text.Bounds(), anchor)
// 	tipRect := text.Bounds()
// 	tipRect = tipRect.WithCenter(mousePos)
// 	tipRect = tipRect.
// 		Moved(glitch.Vec2{
// 		(padding + (tipRect.W() / 2)) * movement[0],
// 		(padding + (tipRect.H() / 2)) * movement[1],
// 	})

// 	g.Panel(panel, tipRect)

// 	text.DrawRect(g.pass, tipRect, g.color)
// 	g.appendUnionBounds(tipRect)
// 	g.debugRect(tipRect)
// }

// func (g *Group) getGraph(bounds glitch.Rect) *graph.Graph {
// 	if g.currentGraphBufferIndex >= len(g.graphBuffer) {
// 		g.graphBuffer = append(g.graphBuffer, graph.NewGraph(bounds))
// 	}

// 	idx := g.currentGraphBufferIndex
// 	g.currentGraphBufferIndex++
// 	g.graphBuffer[idx].Clear()
// 	g.graphBuffer[idx].SetBounds(bounds)
// 	return g.graphBuffer[idx]
// }

// func (g *Group) LineGraph(rect glitch.Rect, series []glitch.Vec2) {
// 	line := g.getGraph(rect)

// 	// line := graph.NewGraph(rect)
// 	line.Line(series)
// 	line.Axes()
// 	line.DrawColorMask(g.pass, glitch.Mat4Ident, g.color)

// 	g.appendUnionBounds(rect)
// 	g.debugRect(rect)

// 	// Draw text around axes
// 	axes := line.GetAxes()
// 	g.FullFixedText(fmt.Sprintf("%.2f ms", axes.Min[1]), rect, glitch.Vec2{0, 0}, glitch.Vec2{1, 0.5}, 0.25)
// 	g.FullFixedText(fmt.Sprintf("%.2f ms", axes.Max[1]), rect, glitch.Vec2{0, 1}, glitch.Vec2{1, 0.5}, 0.25)
// }

// func (g *Group) debugRect(rect glitch.Rect) {
// 	if !g.Debug { return }

// 	lineWidth := 2.0

// 	g.geomDraw.SetColor(glitch.RGBA{1.0, 0, 0, 1.0})
// 	m := g.geomDraw.Rectangle(rect, lineWidth)
// 	m.Draw(g.pass, glitch.Mat4Ident)
// }

// // func (g *Group) Bar(outer, inner Drawer, bounds glitch.Rect, value float64) Rect {
// // 	_, barInner := c.SlicedSprite(outer, bounds)
// // 	c.SpriteColorMask(inner, barInner.ScaledXY(barInner.Anchor(0, 0), value, 1), innerColor)
// // 	return bounds
// // }

// // func (g *Group) Tooltip(name string, tip string, startRect Rect, position, anchor pixel.Vec) {
// // 	padding := 5.0
// // 	c.HoverLambda(startRect,
// // 		func() {
// // 			textBounds := c.MeasureText(tip, position, anchor)
// // 			c.SlicedSprite(name, textBounds.Pad(padding, padding))
// // 			c.Text(tip, position, anchor)
// // 		},
// // 		func() {  })
// // }

// // func (c *Context) BarVert(outer, inner string, bounds Rect, value float64, innerColor color.Color) Rect {
// // 	_, barInner := c.SlicedSprite(outer, bounds)
// // 	c.SpriteColorMask(inner, barInner.ScaledXY(barInner.Anchor(0, 0), 1, value), innerColor)
// // 	return bounds
// // }

// // func (c *Context) HoverLambda(bounds Rect, ifLambda, elseLambda func()) Rect {
// 	// mX, mY := g.mousePosition()
// // // 	mousePos := c.win.MousePosition()
// // 	if bounds.Contains(mousePos) {
// // 		ifLambda()
// // 	} else {
// // 		elseLambda()
// // 	}
// // 	return bounds
// // }

// // // TODO - this might have issues with it's bounding box being slightly off because of the rotation
// // func (c *Context) SpriteRotated(name string, destRect Rect, radians float64) {
// // 	destRect = destRect.Round()

// // 	sprite, err := c.spritesheet.Get(name)
// // 	if err != nil { panic(err) }
// // 	bounds := ZeroRect(sprite.Frame())

// // 	scale := pixel.V(destRect.W() / bounds.W(), destRect.H() / bounds.H())
// // 	mat := pixel.IM.ScaledXY(pixel.ZV, scale).Rotated(pixel.ZV, radians)
// // 	mat = mat.Moved(destRect.Center())

// // 	c.appendUnionBounds(destRect)
// // 	sprite.Draw(c.batch, mat)

// // 	c.DebugRect(destRect)
// // }
