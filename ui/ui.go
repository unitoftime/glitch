package ui

import (
	"fmt"
	"strings"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/graph"
)

type uiGlobals struct {
	mouseCaught bool
}
var global uiGlobals

// Must be called every frame before any UI draws happen
// TODO - This is hard to remember to do
func Clear() {
	global.mouseCaught = false
}

func Contains(point glitch.Vec2) bool {
	return global.mouseCaught
}

// Returns true if the mouse is captured by a group
func MouseCaptured() bool {
	return global.mouseCaught
}

// Returns true if the rect contains the point
func mouseCheck(rect glitch.Rect, point glitch.Vec2) bool {
	// if global.mouseCaught {
	// 	return false
	// }
	if rect.Contains(point[0], point[1]) {
		global.mouseCaught = true
		return true
	}
	return false
}

// TODO do I need an end funtion?
// func End() {
// }

type Drawer interface {
	// Bounds() glitch.Rect
	RectDraw(glitch.BatchTarget, glitch.Rect)
	RectDrawColorMask(glitch.BatchTarget, glitch.Rect, glitch.RGBA)
}

type Group struct {
	win *glitch.Window
	pass *glitch.RenderPass
	camera *glitch.CameraOrtho
	atlas *glitch.Atlas
	unionBoundsSet bool
	unionBounds glitch.Rect // A union of all drawn object's bounds
	allBounds []glitch.Rect
	Debug bool
	OnlyCheckUnion bool
	geomDraw glitch.GeomDraw
	color glitch.RGBA
	textBuffer []*glitch.Text
	currentTextBufferIndex int
	graphBuffer []*graph.Graph
	currentGraphBufferIndex int

	mousePos, mouseDownPos glitch.Vec2

	// New Way
	hotId, activeId eid
	downId eid
	tmpHotId eid

	idCounter eid
	elements map[string]eid // Maps labels to elements
	elementsRev map[eid]string
	dedup map[string]uint32

	// TODO: Element Stylesheet map?
}

type eid uint64 // Element Id
const invalidId eid = 0

func NewGroup(win *glitch.Window, camera *glitch.CameraOrtho, atlas *glitch.Atlas, pass *glitch.RenderPass) *Group {
	return &Group{
		win: win,
		camera: camera,
		pass: pass,
		atlas: atlas,
		unionBoundsSet: false,
		allBounds: make([]glitch.Rect, 0),
		Debug: false,
		OnlyCheckUnion: true,
		color: glitch.RGBA{1, 1, 1, 1},
		textBuffer: make([]*glitch.Text, 0),
		graphBuffer: make([]*graph.Graph, 0),

		elements: make(map[string]eid),
		elementsRev: make(map[eid]string),
		dedup: make(map[string]uint32),
		idCounter: invalidId + 1,
	}
}

func (g *Group) getText(str string) *glitch.Text {
	if g.currentTextBufferIndex >= len(g.textBuffer) {
		g.textBuffer = append(g.textBuffer, g.atlas.Text(str))
	}

	idx := g.currentTextBufferIndex
	g.currentTextBufferIndex++
	// g.textBuffer[idx].Clear()
	g.textBuffer[idx].Set(str)
	return g.textBuffer[idx]
}

func (g *Group) SetLayer(layer int8) {
	g.pass.SetLayer(layer)
}

// onlyCheckUnion is an optimization if the ui elements are tightly packed (it doesn't loop through each rect
func (g *Group) ContainsMouse() bool {
	if g.OnlyCheckUnion {
		if !g.unionBoundsSet { return false }
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

func (g *Group) mousePosition() (float64, float64) {
	x, y := g.win.MousePosition()
	worldSpaceMouse := g.camera.Unproject(glitch.Vec3{x, y, 0})
	return worldSpaceMouse[0], worldSpaceMouse[1]
}

// TODO - Should this be a list of rects that we loop through?
func (g *Group) appendUnionBounds(newBounds glitch.Rect) {
	g.allBounds = append(g.allBounds, newBounds)

	if !g.unionBoundsSet {
		g.unionBounds = newBounds
	} else {
		g.unionBounds = g.unionBounds.Union(newBounds)
	}
}

func (g *Group) Clear() {
	// fmt.Println("State: ", g.hover, g.down, g.active)

	mX, mY := g.mousePosition()
	g.mousePos = glitch.Vec2{mX, mY}

	g.currentTextBufferIndex = 0
	g.currentGraphBufferIndex = 0

	// g.pass.Clear()
	g.unionBoundsSet = false
	g.allBounds = g.allBounds[:0]


	// New
	// Clearing Optimization: https://go.dev/doc/go1.11#performance-compiler
	g.hotId = g.tmpHotId
	g.tmpHotId = invalidId
	for k := range g.dedup {
		delete(g.dedup, k)
	}
}

func (g *Group) SetColor(color glitch.RGBA) {
	g.color = color
}

func (g *Group) PanelColorMask(sprite Drawer, rect glitch.Rect, color glitch.RGBA) {
	sprite.RectDrawColorMask(g.pass, rect, color)
	g.appendUnionBounds(rect)
	g.debugRect(rect)
}

func (g *Group) Panel(sprite Drawer, rect glitch.Rect) {
	if sprite != nil {
		sprite.RectDrawColorMask(g.pass, rect, g.color)
	}
	g.appendUnionBounds(rect)
	g.debugRect(rect)
}

// Adds a panel with padding to the current bounds of the group
func (g *Group) PanelizeBounds(sprite Drawer, padding glitch.Rect) {
	if !g.unionBoundsSet { return }
	rect := g.unionBounds
	rect = rect.Pad(padding)
	g.Panel(sprite, rect)
}

func (g *Group) Hover(normal, hovered Drawer, rect glitch.Rect) bool {
	mX, mY := g.mousePosition()
	if mouseCheck(rect, glitch.Vec2{mX, mY}) {
		g.Panel(hovered, rect)
		return true
	}

	g.Panel(normal, rect)
	return false
}

func (g *Group) Button(normal, hovered, pressed Drawer, rect glitch.Rect) bool {
	mX, mY := g.mousePosition()

	if !mouseCheck(rect, glitch.Vec2{mX, mY}) {
		g.Panel(normal, rect)
		return false
	}

	// If we are here, then we know we are at least hovering
	if g.win.JustPressed(glitch.MouseButtonLeft) {
		g.Panel(pressed, rect)
		return true
	}

	g.Panel(hovered, rect)
	return false
}

// Same thing as a button but returns true for the duration that the button is pressed
func (g *Group) PressedButton(normal, hovered, pressed Drawer, rect glitch.Rect) bool {
	mX, mY := g.mousePosition()

	if !mouseCheck(rect, glitch.Vec2{mX, mY}) {
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

// TODO! - text masking around rect?
func (g *Group) Text(str string, rect glitch.Rect, anchor glitch.Vec2) {
	text := g.getText(str)
	r := rect.Anchor(text.Bounds().ScaledToFit(rect), anchor)
	text.RectDrawColorMask(g.pass, r, g.color)
	g.appendUnionBounds(r)
	g.debugRect(r)
}

// Text, but doesn't automatically scale to fill the rect
// TODO maybe I should call the other text "AutoText"? or something
func (g *Group) FixedText(str string, rect glitch.Rect, anchor glitch.Vec2, scale float64) {
	text := g.getText(str)
	r := rect.Anchor(text.Bounds().Scaled(scale), anchor)
	text.RectDrawColorMask(g.pass, r, g.color)
	g.appendUnionBounds(r)
	g.debugRect(r)
}

// TODO - combine with fixedtext
func (g *Group) FullFixedText(str string, rect glitch.Rect, anchor, anchor2 glitch.Vec2, scale float64) {
	text := g.getText(str)
	r := rect.FullAnchor(text.Bounds().Scaled(scale), anchor, anchor2)
	text.RectDrawColorMask(g.pass, r, g.color)
	g.appendUnionBounds(r)
	g.debugRect(r)
}

func (g *Group) TextInput(prefix, postfix string, str *string, rect glitch.Rect, anchor glitch.Vec2, scale float64) {
	if str == nil { return }

	runes := g.win.Typed()
	*str = *str + string(runes)

	tStr := *str
	if g.win.JustPressed(glitch.KeyBackspace) {
		if g.win.Pressed(glitch.KeyLeftControl) || g.win.Pressed(glitch.KeyRightControl) {
			// Delete whole word
			lastIndex := strings.LastIndex(strings.TrimRight(tStr, " "), " ")
			if lastIndex < 0 {
				// Means there were no spaces, delete everything
				lastIndex = 0
			}

			tStr = tStr[:lastIndex]
		} else {
			if len(tStr) > 0 {
				tStr = tStr[:len(tStr)-1]
			}
		}
	} else if g.win.Repeated(glitch.KeyBackspace) {
		if len(tStr) > 0 {
			tStr = tStr[:len(tStr)-1]
		}
	}

	// ret := false
	// if g.win.JustPressed(glitch.KeyEnter) {
	// 	ret = true
	// }
	*str = tStr

	g.FixedText(prefix + *str + postfix, rect, anchor, scale)
	// return ret
}

// TODO - tooltips only seem to work for single lines
// TODO: Configurable padding
func (g *Group) Tooltip(panel Drawer, tip string, rect glitch.Rect) {
	mX, mY := g.mousePosition()
	mousePos := glitch.Vec2{mX, mY}
	if !mouseCheck(rect, mousePos) {
		return // If mouse not contained by rect, then don't draw
	}

	padding := 10.0
	quadrant := g.win.Bounds().Center().Sub(mousePos).Unit()

	var movement glitch.Vec2
	if quadrant[0] < 0 {
		movement[0] = -1
	} else {
		movement[0] = 1
	}

	if quadrant[1] < 0 {
		movement[1] = -1
	} else {
		movement[1] = 1
	}

	text := g.getText(tip)
	// tipRect := rect.Anchor(text.Bounds(), anchor)
	tipRect := text.Bounds()
	tipRect = tipRect.WithCenter(mousePos)
	tipRect = tipRect.
		Moved(glitch.Vec2{
		(padding + (tipRect.W() / 2)) * movement[0],
		(padding + (tipRect.H() / 2)) * movement[1],
	})

	g.Panel(panel, tipRect)

	text.DrawRect(g.pass, tipRect, g.color)
	g.appendUnionBounds(tipRect)
	g.debugRect(tipRect)
}

func (g *Group) getGraph(bounds glitch.Rect) *graph.Graph {
	if g.currentGraphBufferIndex >= len(g.graphBuffer) {
		g.graphBuffer = append(g.graphBuffer, graph.NewGraph(bounds))
	}

	idx := g.currentGraphBufferIndex
	g.currentGraphBufferIndex++
	g.graphBuffer[idx].Clear()
	g.graphBuffer[idx].SetBounds(bounds)
	return g.graphBuffer[idx]
}

func (g *Group) LineGraph(rect glitch.Rect, series []glitch.Vec2) {
	line := g.getGraph(rect)

	// line := graph.NewGraph(rect)
	line.Line(series)
	line.Axes()
	line.DrawColorMask(g.pass, glitch.Mat4Ident, g.color)

	g.appendUnionBounds(rect)
	g.debugRect(rect)

	// Draw text around axes
	axes := line.GetAxes()
	g.FullFixedText(fmt.Sprintf("%.2f ms", axes.Min[1]), rect, glitch.Vec2{0, 0}, glitch.Vec2{1, 0.5}, 0.25)
	g.FullFixedText(fmt.Sprintf("%.2f ms", axes.Max[1]), rect, glitch.Vec2{0, 1}, glitch.Vec2{1, 0.5}, 0.25)
}

func (g *Group) debugRect(rect glitch.Rect) {
	if !g.Debug { return }

	lineWidth := 2.0

	g.geomDraw.SetColor(glitch.RGBA{1.0, 0, 0, 1.0})
	m := g.geomDraw.Rectangle(rect, lineWidth)
	m.Draw(g.pass, glitch.Mat4Ident)
}

// func (g *Group) Bar(outer, inner Drawer, bounds glitch.Rect, value float64) Rect {
// 	_, barInner := c.SlicedSprite(outer, bounds)
// 	c.SpriteColorMask(inner, barInner.ScaledXY(barInner.Anchor(0, 0), value, 1), innerColor)
// 	return bounds
// }

// func (g *Group) Tooltip(name string, tip string, startRect Rect, position, anchor pixel.Vec) {
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

//--------------------------------------------------------------------------------
func (g *Group) getLabel(id eid) string {
	l, ok := g.elementsRev[id]
	if !ok {
		return ""
	}
	return l
}

func (g *Group) getId(label string) eid {
	bump, alreadyFetched := g.dedup[label]
	if alreadyFetched {
		g.dedup[label] = bump + 1
		label = fmt.Sprintf("%s##%d", label, bump)
		// fmt.Printf("duplicate label, using bump: %s\n", label)
		// panic(fmt.Sprintf("duplicate label found: %s", label))
	} else {
		g.dedup[label] = 0
	}

	id, ok := g.elements[label]
	if !ok {
		id = g.idCounter
		g.idCounter++
		g.elements[label] = id
		g.elementsRev[id] = label
	}

	return id
}

func removeDedup(label string) string {
	ret, _, _ := strings.Cut(label, "##")
	return ret
}

type Style struct {
	Normal, Hovered, Pressed Drawer
	Text *TextStyle
}

type TextStyle struct {
	group *Group
	Anchor glitch.Vec2
	Padding glitch.Rect
	Color glitch.RGBA
	// TODO: atlas/fontface
	// TODO: color, fixedsize, etc
}

// TODO: Remove group dependence
func (g *Group) NewTextStyle() *TextStyle {
	return &TextStyle{
		group: g,
		Anchor: glitch.Vec2{0.5, 0.5},
		Padding: glitch.R(5, 5, 5, 5),
		Color: glitch.White,
	}
}
func (t *TextStyle) Draw(str string, rect glitch.Rect) {
	rect = rect.Unpad(t.Padding)

	// TODO: pass color into new text function
	oldCol := t.group.color
	t.group.color = t.Color
	t.group.Text(str, rect, t.Anchor)
	t.group.color = oldCol
}


func (g *Group) trackHover2(id eid, rect glitch.Rect) {
	// TODO - centralize in group
	// mX, mY := g.mousePosition()
	// if rect.Contains(mX, mY) {
	// 	g.tmpHover = elem
	// }

	mX, mY := g.mousePosition()
	if mouseCheck(rect, glitch.Vec2{mX, mY}) {
		g.tmpHotId = id
	}
}

func (g *Group) ButtonE(label string, style Style, rect glitch.Rect) bool {
	id := g.getId(label)
	text := removeDedup(label)

	ret := false
	if g.activeId == id {
		if g.win.JustReleased(glitch.MouseButtonLeft) {
			g.activeId = invalidId
			if g.hotId == id {
				ret = true
			}
		}
	} else if g.hotId == id {
		if g.win.JustPressed(glitch.MouseButtonLeft) {
			g.activeId = id
		}
	}

	g.trackHover2(id, rect)

	if g.activeId == id {
		g.Panel(style.Pressed, rect)
	} else if g.hotId == id {
		g.Panel(style.Hovered, rect)
	} else {
		g.Panel(style.Normal, rect)
	}

	if style.Text != nil && text != "" {
		style.Text.Draw(text, rect)
	}

	g.appendUnionBounds(rect)
	g.debugRect(rect)

	return ret
}

// Returns true if we successfully completed a drag and drop ending on this element
func (g *Group) DragAndDropSlotE(label string, style Style, rect glitch.Rect) bool {
	id := g.getId(label)

	if g.hotId == id {
		if g.win.JustReleased(glitch.MouseButtonLeft) {
			g.activeId = id
			return true
		}
	}

	// We can only interact with this if we currently have a drag and drop item active
	// TODO: restrict to only drag and drop items?
	if g.activeId != invalidId {
		g.trackHover2(id, rect)

	}

	g.PanelColorMask(style.Normal, rect, glitch.White)

	// Note: This is only so that the mouseCaught boolean is tracked for this rect
	mX, mY := g.mousePosition()
	mouseCheck(rect, glitch.Vec2{mX, mY})
	return false
}

// returns (Clicked, hovered, isdragging, dropSlot)
func (g *Group) DragAndDropItemE(label string, style Style, rect glitch.Rect) (bool, bool, bool, bool) {
	id := g.getId(label)

	buttonClick := false
	buttonHover := false
	if g.activeId == id {
		mX, mY := g.mousePosition()
		global.mouseCaught = true // Because we are actively dragging, the mouse should be captured
		g.PanelColorMask(style.Normal, rect.WithCenter(glitch.Vec2{mX, mY}), glitch.RGBA{0.5, 0.5, 0.5, 0.5})
	} else if g.downId == id {
		g.PanelColorMask(style.Normal, rect, glitch.RGBA{0.5, 0.5, 0.5, 0.5})
	} else if g.hotId == id {
		g.PanelColorMask(style.Normal, rect, glitch.White)
	} else {
		g.PanelColorMask(style.Normal, rect, glitch.White)
	}

	// Make it so we can't hover ourself if we are currently being dragged
	if g.activeId != id {
		g.trackHover2(id, rect)
	}

	dropSlot := false
	if g.activeId == id {
		if g.win.JustReleased(glitch.MouseButtonLeft) {
			g.activeId = invalidId
		}
	} else if g.downId == id {
		buttonHover = true
		if g.mousePos.Sub(g.mouseDownPos).Len() > 5.0 { // TODO - arbitrary
			// fmt.Println("Drag:", elem)
			g.activeId = id
			g.downId = invalidId
		} else if g.win.JustReleased(glitch.MouseButtonLeft) {
			// fmt.Println("Click:", elem)
			buttonClick = true
			g.downId = invalidId
		}

		g.trackHover2(id, rect)
	} else if g.hotId == id {
		if g.win.JustReleased(glitch.MouseButtonLeft) {
			dropSlot = true
		}

		buttonHover = true
		if g.win.JustPressed(glitch.MouseButtonLeft) {
			// fmt.Println("Down:", elem)
			g.downId = id
			g.mouseDownPos = g.mousePos
		}
	}

	// This item is currently dragging if the active element is itself
	currentlyDragging := (g.activeId == id)

	return buttonClick, buttonHover, currentlyDragging, dropSlot
}
