package ui

import (
	"fmt"
	"strings"

	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/graph"
)

// TODO:
// 1. Would be nice to have per-state styles that include text rendering style (rather than one global)
// 2. Some default style that "just works" and renders flat geometry
// 3. Mousechecks should happen based on active/hot and not bounds checks
// 4. Better interface for overriding styles: maybe variadic configs? Maybe push/pop? Maybe just really nice builder functions?
// 5. How to do layer swiching? As style override?
// 6. Ability to do WorldSpaceUI (maybe not interactable?)?

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

// type Drawer2 interface {
// 	Draw(*Group, glitch.Rect)
// }

type Drawer interface {
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
	// elementsRev map[eid]string
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
		// color: glitch.RGBA{1, 1, 1, 1},
		textBuffer: make([]*glitch.Text, 0),
		graphBuffer: make([]*graph.Graph, 0),

		elements: make(map[string]eid),
		// elementsRev: make(map[eid]string),
		dedup: make(map[string]uint32),
		idCounter: invalidId + 1,
	}
}

// TODO: cache based on label for more precision?
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

// TODO: cache based on label for more precision?
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

func (g *Group) SetLayer(layer int8) {
	g.pass.SetLayer(layer)
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
	mX, mY := g.mousePosition()
	g.mousePos = glitch.Vec2{mX, mY}

	g.currentTextBufferIndex = 0
	g.currentGraphBufferIndex = 0

	g.unionBoundsSet = false
	g.allBounds = g.allBounds[:0]

	// New
	g.hotId = g.tmpHotId
	g.tmpHotId = invalidId
	// Clearing Optimization: https://go.dev/doc/go1.11#performance-compiler
	for k := range g.dedup {
		delete(g.dedup, k)
	}
}

//--------------------------------------------------------------------------------
// TODO: I kindof like this idea. 1. what I draw doesn't matter 2. toggle between these based on what ui state the element is in
// type TextDraw struct {
// 	str string
// 	style TextStyle
// }
// func (t *TextDraw) Draw(group *Group, rect glitch.Rect) {
// 	text := group.getText(t.str)

// 	rect = rect.Unpad(t.style.padding)
// 	if t.style.autoFit {
// 		rect = rect.FullAnchor(text.Bounds().ScaledToFit(rect), t.style.anchor, t.style.pivot)
// 	} else {
// 		rect = rect.FullAnchor(text.Bounds().Scaled(t.style.scale), t.style.anchor, t.style.pivot)
// 	}

// 	text.RectDrawColorMask(group.pass, rect, t.style.color)
// 	group.appendUnionBounds(rect)
// 	group.debugRect(rect)
// }

// type SpriteDraw struct {
// 	sprite Drawer
// 	color glitch.RGBA
// }
// func (s *SpriteDraw) Draw(group *Group, rect glitch.Rect) {
// 	if s.sprite != nil {
// 		s.sprite.RectDrawColorMask(group.pass, rect, s.color)
// 	}
// 	group.appendUnionBounds(rect)
// 	group.debugRect(rect)
// }

// type WidgetDraw struct {
// 	Sprite Drawer2
// 	Text Drawer2
// }
// func (s *WidgetDraw) Draw(group *Group, rect glitch.Rect) {
// 	s.Sprite.Draw(group, rect)
// 	s.Text.Draw(group, rect)
// }

// type FullWidgetDraw struct {
// 	Normal, Hovered, Pressed WidgetDraw
// }



//--------------------------------------------------------------------------------
func (g *Group) debugRect(rect glitch.Rect) {
	if !g.Debug { return }

	lineWidth := 2.0

	g.geomDraw.SetColor(glitch.RGBA{1.0, 0, 0, 1.0})
	m := g.geomDraw.Rectangle(rect, lineWidth)
	m.Draw(g.pass, glitch.Mat4Ident)
}
//--------------------------------------------------------------------------------
// func (g *Group) getLabel(id eid) string {
// 	l, ok := g.elementsRev[id]
// 	if !ok {
// 		return ""
// 	}
// 	return l
// }

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
		// g.elementsRev[id] = label
	}

	return id
}

func removeDedup(label string) string {
	ret, _, _ := strings.Cut(label, "##")
	return ret
}

func (g *Group) Panel(sprite Drawer, rect glitch.Rect, color glitch.RGBA) {
	g.draw(sprite, rect, color)
}

func (g *Group) draw(sprite Drawer, rect glitch.Rect, color glitch.RGBA) {
	if sprite != nil {
		sprite.RectDrawColorMask(g.pass, rect, color)
	}
	g.appendUnionBounds(rect)
	g.debugRect(rect)
}

func (g *Group) drawSprite(rect glitch.Rect, style SpriteStyle) {
	if style.sprite != nil {
		style.sprite.RectDrawColorMask(g.pass, rect, style.color)
	}
	g.appendUnionBounds(rect)
	g.debugRect(rect)
}

type SpriteStyle struct {
	sprite Drawer
	color glitch.RGBA
}
func NewSpriteStyle(sprite Drawer, color glitch.RGBA) SpriteStyle {
	return SpriteStyle{
		sprite, color,
	}
}
func (s SpriteStyle) Sprite(v Drawer) SpriteStyle {
	s.sprite = v
	return s
}
func (s SpriteStyle) Color(v glitch.RGBA) SpriteStyle {
	s.color = v
	return s
}

type Style struct {
	Normal, Hovered, Pressed SpriteStyle // These are kind of like button states
	Text TextStyle
}
func ButtonStyle(normal, hovered, pressed Drawer) Style {
	return Style{
		Normal: NewSpriteStyle(normal, glitch.White),
		Hovered: NewSpriteStyle(hovered, glitch.White),
		Pressed: NewSpriteStyle(pressed, glitch.White),
		Text: NewTextStyle(),
	}
}
// func (s Style) Normal(v Drawer, c glitch.RGBA) Style {
// 	s.normal = SpriteStyle{v, c}
// 	return s
// }
// func (s Style) Hovered(v Drawer, c glitch.RGBA) Style {
// 	s.hovered = SpriteStyle{v, c}
// 	return s
// }
// func (s Style) Pressed(v Drawer, c glitch.RGBA) Style {
// 	s.hovered = SpriteStyle{v, c}
// 	return s
// }
// func (s Style) Text(v TextStyle) Style {
// 	s.Text = v
// 	return s
// }

// func (s *Style) DrawNormal(group *Group, rect glitch.Rect, color glitch.RGBA) {
// 	group.draw(s.Normal, rect, color)
// }
// func (s *Style) DrawHot(group *Group, rect glitch.Rect, color glitch.RGBA) {
// 	group.draw(s.Hovered, rect, color)
// }
// func (s *Style) DrawActive(group *Group, rect glitch.Rect, color glitch.RGBA) {
// 	group.draw(s.Pressed, rect, color)
// }

type TextStyle struct {
	// TODO: atlas/fontface

	anchor, pivot glitch.Vec2
	padding glitch.Rect
	color glitch.RGBA
	scale float64
	autoFit bool
}

// TODO: I kind of feel like the string needs to be in here, I'm not sure though
func NewTextStyle() TextStyle {
	return TextStyle{
		anchor: glitch.Vec2{0.5, 0.5},
		pivot: glitch.Vec2{0.5, 0.5},
		padding: glitch.R(0, 0, 0, 0),
		color: glitch.White,
		scale: 1.0,
		autoFit: false,
	}
}

func (s TextStyle) Anchor(v glitch.Vec2) TextStyle {
	s.anchor = v
	s.pivot = v
	return s
}

func (s TextStyle) Scale(v float64) TextStyle {
	s.scale = v
	return s
}

func (s TextStyle) Padding(v glitch.Rect) TextStyle {
	s.padding = v
	return s
}
func (s TextStyle) Color(v glitch.RGBA) TextStyle {
	s.color = v
	return s
}



func (group *Group) Text(str string, rect glitch.Rect, t TextStyle) {
	text := group.getText(str)

	rect = rect.Unpad(t.padding)
	if t.autoFit {
		rect = rect.FullAnchor(text.Bounds().ScaledToFit(rect), t.anchor, t.pivot)
	} else {
		rect = rect.FullAnchor(text.Bounds().Scaled(t.scale), t.anchor, t.pivot)
	}
	text.RectDrawColorMask(group.pass, rect, t.color)
	group.appendUnionBounds(rect)
	group.debugRect(rect)
}


func (g *Group) trackHover(id eid, rect glitch.Rect) {
	if mouseCheck(rect, g.mousePos) {
		g.tmpHotId = id
	}
}

// Same thing as a button but returns true for the duration that the button is pressed
// TODO: Rename HeldButton?
func (g *Group) PressedButton(label string, rect glitch.Rect, style Style) bool {
	_, held, _ := g.button(label, rect, style)
	return held
}

func (g *Group) Button(label string, rect glitch.Rect, style Style) bool {
	_, _, released := g.button(label, rect, style)
	return released
}

// pressed: frame that it was pressed on
// held: frame(s) that it was held for (after pressing, but before releasing)
// released: frame that it was released on
func (g *Group) button(label string, rect glitch.Rect, style Style) (pressed, held, released bool) {
	id := g.getId(label)
	text := removeDedup(label)


	if g.activeId == id {
		held = true
		if g.win.JustReleased(glitch.MouseButtonLeft) {
			g.activeId = invalidId
			if g.hotId == id {
				released = true
			}
		}
	} else if g.hotId == id {
		if g.win.JustPressed(glitch.MouseButtonLeft) {
			pressed = true
			g.activeId = id
		}
	}

	g.trackHover(id, rect)

	if g.activeId == id {
		g.drawSprite(rect, style.Pressed)
	} else if g.hotId == id {
		g.drawSprite(rect, style.Hovered)
	} else {
		g.drawSprite(rect, style.Normal)
	}

	if text != "" {
		g.Text(text, rect, style.Text)
	}

	g.appendUnionBounds(rect)
	g.debugRect(rect)

	return
}

// TODO: label
func (g *Group) TextInput(prefix, postfix string, str *string, rect glitch.Rect, style Style) {
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

	*str = tStr

	// TODO: Change sprite depending on state
	g.drawSprite(rect, style.Normal)

	g.Text(prefix + *str + postfix, rect, style.Text)
}

// TODO - tooltips only seem to work for single lines
// TODO: Configurable padding
func (g *Group) Tooltip(tip string, rect glitch.Rect, style Style) {
	if !mouseCheck(rect, g.mousePos) {
		return // If mouse not contained by rect, then don't draw
	}

	padding := 10.0
	quadrant := g.win.Bounds().Center().Sub(g.mousePos).Unit()

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
	tipRect = tipRect.WithCenter(g.mousePos)
	tipRect = tipRect.
		Moved(glitch.Vec2{
		(padding + (tipRect.W() / 2)) * movement[0],
		(padding + (tipRect.H() / 2)) * movement[1],
	})

	g.drawSprite(tipRect, style.Normal)

	text.DrawRect(g.pass, tipRect, style.Text.color)
	g.appendUnionBounds(tipRect)
	g.debugRect(tipRect)
}

// Returns true if we successfully completed a drag and drop ending on this element
func (g *Group) DragAndDropSlot(label string, style Style, rect glitch.Rect) bool {
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
		g.trackHover(id, rect)

	}

	g.drawSprite(rect, style.Normal)

	// Note: This is only so that the mouseCaught boolean is tracked for this rect
	mouseCheck(rect, g.mousePos)
	return false
}

// returns (Clicked, hovered, isdragging, dropSlot)
func (g *Group) DragAndDropItem(label string, style Style, rect glitch.Rect) (bool, bool, bool, bool) {
	id := g.getId(label)

	buttonClick := false
	buttonHover := false
	if g.activeId == id {
		global.mouseCaught = true // Because we are actively dragging, the mouse should be captured
		g.drawSprite(rect.WithCenter(g.mousePos), style.Normal.Color(glitch.RGBA{0.5, 0.5, 0.5, 0.5}))
	} else if g.downId == id {
		g.drawSprite(rect, style.Normal.Color(glitch.RGBA{0.5, 0.5, 0.5, 0.5})) //TODO Push outward
	} else if g.hotId == id {
		g.drawSprite(rect, style.Normal)
	} else {
		g.drawSprite(rect, style.Normal)
	}

	// Make it so we can't hover ourself if we are currently being dragged
	if g.activeId != id {
		g.trackHover(id, rect)
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

		g.trackHover(id, rect)
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

func (g *Group) LineGraph(rect glitch.Rect, series []glitch.Vec2, style TextStyle) {
	line := g.getGraph(rect)

	line.Line(series)
	line.Axes()
	line.DrawColorMask(g.pass, glitch.Mat4Ident, style.color)

	g.appendUnionBounds(rect)
	g.debugRect(rect)

	// Draw text around axes
	axes := line.GetAxes()

	style.anchor = glitch.Vec2{0, 0}
	style.pivot = glitch.Vec2{1, 0.5}
	g.Text(fmt.Sprintf("%.2f ms", axes.Min[1]), rect, style)

	style.anchor = glitch.Vec2{0, 1}
	style.pivot = glitch.Vec2{1, 0.5}
	g.Text(fmt.Sprintf("%.2f ms", axes.Max[1]), rect, style)
}
