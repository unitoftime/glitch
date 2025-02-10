package ui

import (
	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/graph"
)

type Drawer interface {
	RectDraw(glitch.BatchTarget, glitch.Rect)
	RectDrawColorMask(glitch.BatchTarget, glitch.Rect, glitch.RGBA)
}

func Bounds() glitch.Rect {
	bounds := global.camera.Bounds()
	bounds.Min = global.camera.Unproject(bounds.Min.Vec3()).Vec2()
	bounds.Max = global.camera.Unproject(bounds.Max.Vec3()).Vec2()
	return bounds
}

// TODO:
// 1. Would be nice to have per-state styles that include text rendering style (rather than one global)
// 2. Some default style that "just works" and renders flat geometry
// 3. Mousechecks should happen based on active/hot and not bounds checks
// 4. Better interface for overriding styles: maybe variadic configs? Maybe push/pop? Maybe just really nice builder functions?
// 5. How to do layer swiching? As style override?
// 6. Ability to do WorldSpaceUI (maybe not interactable?)?
func Initialize(win *glitch.Window, camera *glitch.CameraOrtho, atlas *glitch.Atlas, sorter *glitch.Sorter) {
	global.win = win
	global.camera = camera
	global.atlas = atlas
	global.sorter = sorter
	global.geomDraw = glitch.NewGeomDraw()
	global.geomDraw.SetColor(glitch.RGBA{1.0, 0, 0, 1.0})
	global.debugMesh = glitch.NewMesh()

	global.layout = Layout{
		Type:   CutTop,
		Bounds: win.Bounds(),
	}

	// // Setup Default styles
	// gStyle.textStyle = NewTextStyle().Autofit(true).Padding(glitch.R(5, 5, 5, 5))
	defaultTextStyle := NewTextStyle().Autofit(true).Padding(glm.R(5, 5, 5, 5))

	whiteTexture := glitch.WhiteTexture()
	whiteSquare := glitch.NewSprite(whiteTexture, whiteTexture.Bounds())

	grey := make([]glitch.RGBA, 16)
	for i := range grey {
		val := uint8(i * 0x11)
		grey[i] = glm.FromUint8(val, val, val, 0xff)
	}

	gStyle.buttonStyle = Style{
		Normal:  NewSpriteStyle(whiteSquare, grey[6]),
		Hovered: NewSpriteStyle(whiteSquare, grey[7]),
		Pressed: NewSpriteStyle(whiteSquare, grey[8]),
		Text:    defaultTextStyle,
	}

	gStyle.panelStyle = Style{
		Normal:  NewSpriteStyle(whiteSquare, grey[3]),
		Hovered: NewSpriteStyle(whiteSquare, grey[3]),
		Pressed: NewSpriteStyle(whiteSquare, grey[3]),
		Text:    defaultTextStyle,
	}

	gStyle.dragSlotStyle = Style{
		Normal:  NewSpriteStyle(whiteSquare, grey[2]),
		Hovered: NewSpriteStyle(whiteSquare, grey[2]),
		Pressed: NewSpriteStyle(whiteSquare, grey[2]),
		Text:    defaultTextStyle,
	}

	gStyle.dragItemStyle = Style{
		Normal:  NewSpriteStyle(whiteSquare, grey[9]),
		Hovered: NewSpriteStyle(whiteSquare, grey[9]),
		Pressed: NewSpriteStyle(whiteSquare, grey[9]),
		Text:    defaultTextStyle,
	}

	gStyle.scrollbarBotStyle = gStyle.buttonStyle
	gStyle.scrollbarTopStyle = gStyle.buttonStyle
	gStyle.scrollbarHandleStyle = Style{
		Normal:  NewSpriteStyle(whiteSquare, grey[7]),
		Hovered: NewSpriteStyle(whiteSquare, grey[8]),
		Pressed: NewSpriteStyle(whiteSquare, grey[9]),
		Text:    defaultTextStyle,
	}
	gStyle.scrollbarBgStyle = Style{
		Normal:  NewSpriteStyle(whiteSquare, grey[2]),
		Hovered: NewSpriteStyle(whiteSquare, grey[2]),
		Pressed: NewSpriteStyle(whiteSquare, grey[2]),
		Text:    defaultTextStyle,
	}

	gStyle.checkboxStyleTrue = gStyle.buttonStyle
	gStyle.checkboxStyleFalse = gStyle.buttonStyle

	gStyle.textInputPanelStyle = Style{
		Normal:  NewSpriteStyle(whiteSquare, grey[2]),
		Hovered: NewSpriteStyle(whiteSquare, grey[2]),
		Pressed: NewSpriteStyle(whiteSquare, grey[2]),
		Text:    defaultTextStyle,
	}
	gStyle.textCursorStyle = Style{
		Normal:  NewSpriteStyle(whiteSquare, grey[15]),
		Hovered: NewSpriteStyle(whiteSquare, grey[15]),
		Pressed: NewSpriteStyle(whiteSquare, grey[15]),
		Text:    defaultTextStyle,
	}

	gStyle.tooltipStyle = Style{
		// Draw no backgrounds
		// Normal:  NewSpriteStyle(whiteSquare, grey[15]),
		// Hovered: NewSpriteStyle(whiteSquare, grey[15]),
		// Pressed: NewSpriteStyle(whiteSquare, grey[15]),
		Text: NewTextStyle().Autofit(false),
	}

}

type uiGlobals struct {
	mouseCaught bool
	hudScale    float64
	fontScale   float64

	// For global interface
	win    *glitch.Window
	camera *glitch.CameraOrtho
	atlas  *glitch.Atlas
	sorter *glitch.Sorter

	debug bool

	layout Layout

	geomDraw  *glitch.GeomDraw
	debugMesh *glitch.Mesh

	allBounds               []glitch.Rect
	unionBoundsSet          bool
	unionBounds             glitch.Rect // A union of all drawn object's bounds
	mousePos, mouseDownPos  glitch.Vec2
	textBuffer              []*glitch.Text
	currentTextBufferIndex  int
	graphBuffer             []*graph.Graph
	currentGraphBufferIndex int

	drawDraggedDrawers bool // Indicates that we should internally draw dragged objects on the mouse
	dragItemLayer int8

	// lastRect glitch.Rect

	// New Way
	hotId     eid // The element id that you are hovering over or about to interact with
	downId    eid // The element id that you have selected or are holding down on
	activeId  eid // The element id that is active?
	lastHotId eid // HotId from the last frame

	// Note: This is for strictly hover zones: Things like tooltips that you dont want blocking actual interactive things like buttons
	hoverOnlyId     eid // The element that you are currently hovering
	lastHoverOnlyId eid // HoverId from the last frame

	stopDragging bool
	dragData     any

	cursorPos int

	idCounter eid
	elements  map[uint64]eid // Maps labels to elements
	// elementsRev map[eid]string
	dedup map[uint64]uint32

	stackIndex int // Indicates the next stack index for pushing
	idStack    [][]byte
}

var global = uiGlobals{
	hudScale:  1.0,
	fontScale: 1.0,

	drawDraggedDrawers: true,

	unionBoundsSet: false,
	allBounds:      make([]glitch.Rect, 0),
	debug:          false,
	// color: glitch.RGBA{1, 1, 1, 1},
	textBuffer:  make([]*glitch.Text, 0),
	graphBuffer: make([]*graph.Graph, 0),

	// For global interface
	idCounter: invalidId + 1,
	elements:  make(map[uint64]eid),
	// elementsRev: make(map[eid]string),
	dedup: make(map[uint64]uint32),
}

func SetHudScale(scale float64) {
	global.hudScale = scale
}

func SetFontScale(scale float64) {
	global.fontScale = scale
}

func DrawDraggedObjects(val bool) {
	global.drawDraggedDrawers = val
}

func SetDragItemLayer(layer int8) {
	global.dragItemLayer = layer
}

func SetDebug(val bool) {
	global.debug = val
}

func Debug() bool {
	return global.debug
}

func SetLayer(layer int8) {
	global.sorter.SetLayer(layer)
}

func Layer() int8 {
	return global.sorter.Layer()
}

func SetDragData(data any) {
	global.dragData = data
}
func DragData() any {
	return global.dragData
}

func SetCursorPos(pos int) {
	global.cursorPos = pos
}

func CursorPos() int {
	return global.cursorPos
}

// func LastRect() glitch.Rect {
// 	return global.lastRect
// }

// Must be called every frame before any UI draws happen
// TODO - This is hard to remember to do
func Clear() {
	global.mouseCaught = false

	mX, mY := MousePosition()
	global.mousePos = glitch.Vec2{mX, mY}

	global.debugMesh.Draw(global.sorter, glitch.Mat4Ident) // TODO: Its a little weird to draw this on the clear. but it should technically work
	global.debugMesh.Clear()

	global.currentTextBufferIndex = 0
	global.currentGraphBufferIndex = 0

	global.unionBoundsSet = false
	global.allBounds = global.allBounds[:0]

	// New
	global.hotId = global.lastHotId
	global.lastHotId = invalidId

	global.hoverOnlyId = global.lastHoverOnlyId
	global.lastHoverOnlyId = invalidId

	if global.stopDragging {
		global.stopDragging = false
		global.activeId = invalidId
		global.dragData = nil
	}

	clear(global.dedup)
}

func Update() {
	global.sorter.Draw(global.win) // TODO: could have a SetTarget instead of using win.
	// Clear()

	// global.layout.Reset()
}

// Returns true if the rect contains the point
func mouseCheck(rect glitch.Rect, point glitch.Vec2) bool {
	// if global.mouseCaught {
	// 	return false
	// }
	if rect.Contains(point) {
		global.mouseCaught = true
		return true
	}
	return false
}

func Contains(point glitch.Vec2) bool {
	return global.mouseCaught
}

// Returns true if the mouse is captured by a group
func MouseCaptured() bool {
	// TODO: Would this be better if I just check the hot/active eid's and see if they are set? if they are set then clearly we are interacting or hovering something?
	return global.mouseCaught
}

func (g *uiGlobals) trackHoverOnly(id eid, rect glitch.Rect) {
	if mouseCheck(rect, g.mousePos) {
		g.lastHoverOnlyId = id
	}
}

func (g *uiGlobals) trackHover(id eid, rect glitch.Rect) {
	if mouseCheck(rect, g.mousePos) {
		g.lastHotId = id
		g.lastHoverOnlyId = id // TODO: Is this always right? when would it not be?
	}
}

// Returns the mouse position with respect to the group camera
func MousePosition() (float64, float64) {
	// // x, y := g.win.MousePosition()
	// // worldSpaceMouse := g.camera.Unproject(glitch.Vec3{x, y, 0})
	// // return worldSpaceMouse[0], worldSpaceMouse[1]

	// // 1. Get mouse position in window bounds
	// // 2. Convert to (0, 1) ratios
	// // 3. Convert back to group camera bounds
	// x, y := g.win.MousePosition()
	// // winSpacePos := g.camera.Unproject(glitch.Vec3{x, y, 0}) // TODO: Is this right? Or does it just not matter because my camera is identity?
	// winSpacePos := glitch.Vec2{x, y}
	// winBounds := g.win.Bounds()
	// normBoundsX := winSpacePos.X / winBounds.W()
	// normBoundsY := winSpacePos.Y / winBounds.H()

	// uiBounds := g.Bounds()
	// uiPosX := normBoundsX * uiBounds.W()
	// uiPosY := normBoundsY * uiBounds.H()

	// return uiPosX, uiPosY

	// // TODO: I think I need to do this if I ever have a scaling camera
	// // unprojPos := g.camera.Project(glitch.Vec3{uiPosX, uiPosY})
	// // return unprojPos.X, unprojPos.Y

	x, y := global.win.MousePosition()
	worldSpaceMouse := global.camera.Unproject(glitch.Vec3{x, y, 0})
	return worldSpaceMouse.X, worldSpaceMouse.Y
}

// TODO: cache based on label for more precision?
// TODO: I used to pass the style scale into the text here, but it never quite lined up right. so I just scale later now
func (g *uiGlobals) getText(str string, style TextStyle, bounds glitch.Rect) *glitch.Text {
	if g.currentTextBufferIndex >= len(g.textBuffer) {
		text := g.atlas.Text("", g.fontScale)
		text.SetShadow(style.shadow)
		g.textBuffer = append(g.textBuffer, text)
	}

	idx := g.currentTextBufferIndex
	g.currentTextBufferIndex++
	// g.textBuffer[idx].Clear()
	// g.textBuffer[idx].SetScale(style.scale)
	g.textBuffer[idx].SetShadow(style.shadow)
	wrapBounds := bounds.Scaled(1.0 / (style.scale)) // TODO: a bit hacky
	g.textBuffer[idx].SetWordWrap(style.wordWrap, wrapBounds)
	g.textBuffer[idx].SetScale(g.fontScale)
	g.textBuffer[idx].Set(str)
	return g.textBuffer[idx]
}

// TODO: cache based on label for more precision?
func (g *uiGlobals) getGraph(bounds glitch.Rect) *graph.Graph {
	if g.currentGraphBufferIndex >= len(g.graphBuffer) {
		g.graphBuffer = append(g.graphBuffer, graph.NewGraph(bounds))
	}

	idx := g.currentGraphBufferIndex
	g.currentGraphBufferIndex++
	g.graphBuffer[idx].Clear()
	g.graphBuffer[idx].SetBounds(bounds)
	return g.graphBuffer[idx]
}
