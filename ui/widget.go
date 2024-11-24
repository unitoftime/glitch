package ui

import (
	"cmp"
	"fmt"
	"math"
	"strings"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
)

func SetActive(label string) {
	global.activeId = getIdNoBump(label)
}

func ClearActive() {
	global.activeId = invalidId
}

// --------------------------------------------------------------------------------
// - Widgets
// --------------------------------------------------------------------------------
type widget struct {
	id eid

	// Recompute each frame
	relPos       glitch.Vec2 // Position relative to parent position
	computedSize glitch.Vec2
	rect         glitch.Rect // The actual bounding rect for the draw
}

type widgetMask uint64

const (
	wmHoverable widgetMask = 1 << iota
	wmClickable
	wmDrawPanel
	wmDrawText
	wmDropSlot
	wmDragItem
	// wmDragY
)

func (m widgetMask) Get(m2 widgetMask) bool {
	return (m & m2) != 0
}

type WidgetResp struct {
	id        eid
	droppedId eid
	textRect  glitch.Rect
	mousePos  glitch.Vec2
	// mouseDragDelta glitch.Vec2
	Pressed  bool
	Repeated bool
	Held     bool
	Released bool
	Dragging bool
}

func (resp *WidgetResp) clickable(id eid) {
	if global.activeId == id {
		resp.Held = true
		if !global.win.Pressed(glitch.MouseButtonLeft) {
			global.activeId = invalidId
			if global.hotId == id {
				resp.Released = true
			}
		}
	} else if global.hotId == id {
		if global.win.JustPressed(glitch.MouseButtonLeft) {
			resp.Pressed = true
			global.activeId = id // TODO: should maybe do tmpActiveId and flip it so that the interaction ordering works?
		}
	}

	if global.activeId == id {
		if global.win.Repeated(glitch.MouseButtonLeft) {
			resp.Repeated = true
			global.activeId = id // TODO: should maybe do tmpActiveId and flip it so that the interaction ordering works?
		}
	}
}

func (resp *WidgetResp) hoverOnly(id eid, rect glitch.Rect) {
	global.trackHoverOnly(id, rect)
}

func (resp *WidgetResp) hoverable(id eid, rect glitch.Rect) {
	global.trackHover(id, rect)
}

func clamp[T cmp.Ordered](start, end T, val T) T {
	return min(max(val, start), end)
}

// Returns true if currently dragging
func (resp *WidgetResp) smoothDrag(rect *glitch.Rect, bounds glitch.Rect, step glitch.Vec2) {
	if resp.Held {
		halfWidth := rect.W() / 2.0
		halfHeight := rect.H() / 2.0
		pos := global.mousePos
		// snap based on step
		// Note: - 0.5 because we draw based on the center point of the slider
		// TODO: There's a bug where it seems to be 1px off the snap point in some cases.
		if step.X > 0 {
			pos.X = (math.Round(pos.X / step.X)) * step.X
		}
		if step.Y > 0 {
			pos.Y = (math.Round(pos.Y / step.Y)) * step.Y
		}

		// pos.X = min(max(pos.X, bounds.Min.X), bounds.Max.X)
		// pos.Y = min(max(pos.Y, bounds.Min.Y), bounds.Max.Y)
		pos.X = clamp(bounds.Min.X+halfWidth, bounds.Max.X-halfWidth, pos.X)
		pos.Y = clamp(bounds.Min.Y+halfHeight, bounds.Max.Y-halfHeight, pos.Y)

		*rect = rect.WithCenter(pos)
	}
}

// Lets the widget be selectable if clicked
// Returns true if we just selected this widget
func (resp *WidgetResp) selectableOnClick(id eid, rect glitch.Rect) bool {
	ret := false
	if global.activeId == id {
		if global.win.JustPressed(glitch.MouseButtonLeft) {
			if global.lastHotId != id {
				// If we are active, but not hot this frame and we click, then the user has clicked off the input field
				global.activeId = invalidId
			}
		}
	} else if global.hotId == id {
		if global.win.JustPressed(glitch.MouseButtonLeft) {
			ret = true
			global.activeId = id
		}
	}
	return ret
}

func findWordIndexLeft(str string, cursorPos int) int {
	lastIndex := strings.LastIndex(strings.TrimRight(str[:cursorPos], " "), " ")
	if lastIndex < 0 {
		// Means there were no spaces, find start
		return 0
	}
	return lastIndex + 1 // To give back the " "
}

func findWordIndexRight(str string, cursorPos int) int {
	trimmed := strings.TrimLeft(str[cursorPos:], " ")
	trimAmount := len(str[cursorPos:]) - len(trimmed)
	lastIndex := strings.Index(trimmed, " ")
	if lastIndex < 0 {
		// Means there were no spaces, find end
		return len(str)
	}
	return cursorPos + lastIndex + trimAmount
}

func (resp *WidgetResp) recordTyped(str *string, cursorPosRet *int) {
	if str == nil {
		return
	}

	runes := global.win.Typed()

	tStr := *str
	cursorPos := clamp(0, len(tStr), *cursorPosRet)

	// *str = *str + string(runes)
	tStr = tStr[:cursorPos] + string(runes) + tStr[cursorPos:]
	cursorPos += len(runes)

	controlPressed := global.win.Pressed(glitch.KeyLeftControl) || global.win.Pressed(glitch.KeyRightControl)

	if global.win.JustPressed(glitch.KeyBackspace) {
		if controlPressed {
			// // Delete whole word
			// lastIndex := strings.LastIndex(strings.TrimRight(tStr, " "), " ")
			// if lastIndex < 0 {
			// 	// Means there were no spaces, delete everything
			// 	lastIndex = 0
			// }

			// tStr = tStr[:lastIndex]

			// lastIndex := strings.LastIndex(strings.TrimRight(tStr[:cursorPos], " "), " ")
			// if lastIndex < 0 {
			// 	// Means there were no spaces, delete everything
			// 	lastIndex = 0
			// }
			lastIndex := findWordIndexLeft(tStr, cursorPos)
			tStr = tStr[:lastIndex] + tStr[cursorPos:]
			cursorPos -= cursorPos - lastIndex
		} else {
			if len(tStr) > 0 {
				// tStr = tStr[:len(tStr)-1]
				if cursorPos > 0 {
					tStr = tStr[:cursorPos-1] + tStr[cursorPos:]
					cursorPos--
				}
			}
		}
	} else if global.win.Repeated(glitch.KeyBackspace) {
		if len(tStr) > 0 {
			// tStr = tStr[:len(tStr)-1]
			if cursorPos > 0 {
				tStr = tStr[:cursorPos-1] + tStr[cursorPos:]
				cursorPos--
			}
		}
	}

	if global.win.JustPressed(glitch.KeyLeft) {
		if controlPressed {
			lastIndex := findWordIndexLeft(tStr, cursorPos)
			cursorPos = lastIndex
		} else {
			cursorPos -= 1
		}
	}
	if global.win.JustPressed(glitch.KeyRight) {
		if controlPressed {
			lastIndex := findWordIndexRight(tStr, cursorPos)
			cursorPos = lastIndex
		} else {
			cursorPos += 1
		}
	}

	*cursorPosRet = min(max(cursorPos, 0), len(tStr))

	*str = tStr
}

func doWidget(id eid, text string, mask widgetMask, style Style, rect glitch.Rect) WidgetResp {
	// style.Text = gStyle.textStyle

	resp := WidgetResp{}

	// -- Clicking and Releasing ---
	if mask.Get(wmClickable) {
		resp.clickable(id)
	}

	// // -- DragY --
	// if mask.Get(wmDragY) {
	// 	if resp.held {
	// 		rect = rect.WithCenter(glitch.Vec2{rect.Center().X, global.mousePos.Y})
	// 	}
	// }

	// -- Drop Slot --
	if mask.Get(wmDropSlot) {
		if global.hotId == id {
			// if global.win.JustReleased(glitch.MouseButtonLeft) {
			if !global.win.Pressed(glitch.MouseButtonLeft) {
				if global.activeId != invalidId {
					global.stopDragging = true
					resp.droppedId = global.activeId
					global.activeId = invalidId
				}
			}
		}
	}

	// -- Drag Item --
	if mask.Get(wmDragItem) {
		if global.activeId == id {
			// If we are currently dragging this item
			resp.Dragging = true
			// TODO: Do I need this?
			// global.mouseCaught = true // Because we are actively dragging, the mouse should be captured

			if !global.win.Pressed(glitch.MouseButtonLeft) {
				// Indicates that we want to stop dragging at the end of the frame
				global.stopDragging = true
			}

		} else if global.downId == id {
			// buttonHover = true
			if global.mousePos.Sub(global.mouseDownPos).Len() > 5.0 { // TODO - arbitrary
				// fmt.Println("Drag:", elem)
				global.activeId = id
				global.downId = invalidId
				// } else if global.win.JustReleased(glitch.MouseButtonLeft) {
			} else if !global.win.Pressed(glitch.MouseButtonLeft) {
				// fmt.Println("Click:", elem)
				resp.Released = true // buttonClick = true
				global.downId = invalidId
			}

			// global.trackHover(id, rect)
		} else if global.hotId == id {
			// if global.win.JustReleased(glitch.MouseButtonLeft) {
			// 	dropSlot = true
			// }

			// buttonHover = true
			if global.win.JustPressed(glitch.MouseButtonLeft) {
				// fmt.Println("Down:", elem)
				global.downId = id
				global.mouseDownPos = global.mousePos
			}
		}
	}

	// -- Hovering ---
	if !resp.Dragging { // Note: You cant hover an item that you are dragging
		if mask.Get(wmHoverable) {
			resp.hoverable(id, rect)
		}
	}

	// -- Drawing Panels ---
	if resp.Dragging {
		rect = rect.WithCenter(global.mousePos)
		lastLayer := global.sorter.Layer()
		global.sorter.SetLayer(global.dragItemLayer)
		defer global.sorter.SetLayer(lastLayer)
	}

	if mask.Get(wmDrawPanel) {
		if global.activeId == id {
			if resp.Dragging {
				drawSprite(rect, style.Normal.Color(glitch.RGBA{0.5, 0.5, 0.5, 0.5}))
			} else {
				drawSprite(rect, style.Pressed)
			}

			// TODO: Do I need to differentiate for dragged item? Or just use pressed style
			// drawSprite(rect, style.Pressed)
		} else if global.downId == id {
			drawSprite(rect, style.Pressed) // Note: This is for wmDragItem
		} else if global.hotId == id {
			drawSprite(rect, style.Hovered)
		} else {
			drawSprite(rect, style.Normal)
		}
	}

	// -- Drawing Text ---
	if mask.Get(wmDrawText) {
		resp.textRect = drawText(text, rect, style.Text)
	}

	return resp
}

func drawSprite(rect glitch.Rect, style SpriteStyle) {
	if style.sprite == nil {
		return
	}
	style.sprite.RectDrawColorMask(global.sorter, rect, style.color)

	// TODO: add back
	// g.appendUnionBounds(rect)
	drawDebug(rect)
}

// Returns the rectangular bounds of the drawn text
func drawText(str string, rect glitch.Rect, t TextStyle) glitch.Rect {
	if str == "" {
		return rect
	} // TODO: Return empty?

	text := global.getText(str, t, rect)

	rect = rect.Unpad(t.padding)
	// if t.autoFit {
	// 	rect = rect.FullAnchor(text.Bounds(), t.anchor, t.pivot)
	// } else {
	// 	rect = rect.FullAnchor(text.Bounds(), t.anchor, t.pivot)
	// }
	if t.autoFit {
		if t.fitInteger {
			intFitScale := math.Floor(text.Bounds().FitScale(rect))
			rect = rect.FullAnchor(text.Bounds().Scaled(intFitScale), t.anchor, t.pivot)
		} else {
			rect = rect.FullAnchor(text.Bounds().ScaledToFit(rect), t.anchor, t.pivot)
		}
	} else {
		// rect = rect.FullAnchor(text.Bounds().Scaled(global.fontScale * t.scale), t.anchor, t.pivot)
		rect = rect.FullAnchor(text.Bounds().Scaled(t.scale), t.anchor, t.pivot)
		// rect = rect.FullAnchor(text.Bounds(), t.anchor, t.pivot)
	}

	// rect = rectSnap(rect)
	text.RectDrawColorMask(global.sorter, rect, t.color)

	// global.appendUnionBounds(rect)
	drawDebug(rect)
	return rect
}

func MeasureTextSize(str string, t TextStyle) glm.Vec2 {
		if str == "" {
		return glm.Vec2{}
	}

	textBounds := global.atlas.Measure(str, 1.0)
	textBounds = textBounds.Scaled(global.fontScale * t.scale)
	textSize := glm.Vec2{textBounds.W(), textBounds.H()}

	padSize := glm.Vec2{
		t.padding.Min.X + t.padding.Max.X,
		t.padding.Min.Y + t.padding.Max.Y,
	}

	return textSize.Add(padSize)
}

// Returns the rectangular bounds of the drawn text
func MeasureText(str string, rect glitch.Rect, t TextStyle) glitch.Rect {
	if str == "" {
		return glitch.Rect{}
	}

	// text := global.getText(str, t)
	textBounds := global.atlas.Measure(str, 1.0)

	rect = rect.Unpad(t.padding)
	if t.autoFit {
		if t.fitInteger {
			intFitScale := math.Floor(textBounds.FitScale(rect))
			rect = rect.FullAnchor(textBounds.Scaled(intFitScale), t.anchor, t.pivot)
		} else {
			rect = rect.FullAnchor(textBounds.ScaledToFit(rect), t.anchor, t.pivot)
		}
	} else {
		rect = rect.FullAnchor(textBounds.Scaled(global.fontScale*t.scale), t.anchor, t.pivot)
		// rect = rect.FullAnchor(text.Bounds(), t.anchor, t.pivot)
	}

	return rect
}

func drawDebug(rect glitch.Rect) {
	if !global.debug {
		return
	}

	lineWidth := 0.4
	global.geomDraw.Rectangle2(global.debugMesh, rect, lineWidth)

	// m := global.geomDraw.Rectangle(rect, lineWidth)
	// m.Draw(global.sorter, glitch.Mat4Ident)
}

//--------------------------------------------------------------------------------
// - Widgets (Actual)
//--------------------------------------------------------------------------------
// func Button(label string) bool {
// 	style := buttonStyle
// 	mask := wmHoverable | wmClickable | wmDrawPanel | wmDrawText
// 	id := getId(label)
// 	text := removeDedup(label)

// 	// TODO: Obviously very inefficient. Some kind of measure function
// 	txt := global.getText(text, style.Text)

// 	rect := global.layout.Next(txt.Bounds())

// 	resp := doWidget(id, text, mask, style, rect)

// 	return resp.released
// }

// func Panel(label string) bool {
// 	style := panelStyle
// 	mask := wmHoverable | wmDrawPanel

// 	id := getId(label)
// 	text := removeDedup(label)

// 	// TODO: Obviously very inefficient. Some kind of measure function
// 	txt := global.getText(text, style.Text)

// 	rect := global.layout.Next(txt.Bounds())

// 	resp := doWidget(id, text, mask, style, rect)

// 	return resp.released
// }

// func simple(label string, rect glitch.Rect, mask widgetMask, style Style) widgetResp {
// 	id := getId(label)
// 	text := removeDedup(label)
// 	resp := doWidget(id, text, mask, style, rect)
// 	return resp
// }

// --------------------------------------------------------------------------------
func SmoothDragButton(label string, rect *glitch.Rect, bounds glitch.Rect, step glitch.Vec2, style Style) bool {
	// mask := wmHoverable | wmClickable | wmDrawPanel | wmDrawText // | wmDragY
	mask := wmDrawPanel
	id := getId(label)
	text := removeDedup(label)

	resp := WidgetResp{}
	resp.clickable(id)
	resp.smoothDrag(rect, bounds, step)
	resp.hoverable(id, *rect)

	doWidget(id, text, mask, style, *rect)

	return resp.Held
}

func Hovered(label string, rect glitch.Rect) bool {
	// return mouseCheck(rect, global.mousePos)
	// label := "##__h"

	id := getId(label)

	resp := WidgetResp{}
	resp.hoverOnly(id, rect)

	isHovering := global.hoverOnlyId == id

	return isHovering
}

func HoveredNoBlock(rect glitch.Rect) bool {
	return mouseCheck(rect, global.mousePos)
}

// func Text(label string, rect glitch.Rect) glitch.Rect {
// 	return TextExt(label, rect, gStyle.textStyle)
// }

func TextExt(label string, rect glitch.Rect, textStyle TextStyle) glitch.Rect {
	style := Style{
		Text: textStyle,
	}

	mask := wmDrawText
	id := getId(label)
	text := removeDedup(label)
	resp := doWidget(id, text, mask, style, rect)
	return resp.textRect
}

func ButtonFull(label string, rect glitch.Rect, style Style) WidgetResp {
	mask := wmHoverable | wmClickable | wmDrawPanel | wmDrawText
	id := getId(label)
	text := removeDedup(label)
	resp := doWidget(id, text, mask, style, rect)

	return resp
}

func ButtonExt(label string, rect glitch.Rect, style Style) bool {
	resp := ButtonFull(label, rect, style)
	return resp.Released
}

func CheckboxExt(val *bool, rect glitch.Rect, styleTrue, styleFalse Style) bool {
	style := styleTrue
	str := "##checkbox"
	// str := "X##checkbox"
	if !(*val) {
		style = styleFalse
		// str = "##checkbox"
	}

	toggle := ButtonExt(str, rect, style)
	if toggle {
		*val = !(*val)
	}
	return toggle
}

// Returns true if the value has changed
func Checkbox(val *bool, rect glitch.Rect) bool {
	return CheckboxExt(val, rect, gStyle.checkboxStyleTrue, gStyle.checkboxStyleFalse)
}

func Button(label string, rect glitch.Rect) bool {
	style := gStyle.buttonStyle
	return ButtonExt(label, rect, style)
}

func Button2(label string, rect glitch.Rect) bool {
	style := gStyle.buttonStyle
	return ButtonExt(label, rect, style)
}

func PanelExt(label string, rect glitch.Rect, style Style) bool {
	mask := wmHoverable | wmDrawPanel

	id := getId(label)
	text := removeDedup(label)

	resp := doWidget(id, text, mask, style, rect)

	return resp.Released
}
func Panel2(label string, rect glitch.Rect) bool {
	return PanelExt(label, rect, gStyle.panelStyle)
}
func SpritePanel(sprite Drawer, rect glitch.Rect, color glitch.RGBA) bool {
	label := "##__p"
	style := NewStyle(sprite, color)
	return PanelExt(label, rect, style)
}

func Sprite(sprite Drawer, rect glitch.Rect, color glitch.RGBA) {
	// style := Style{
	// 	Normal: NewSpriteStyle(sprite, color),
	// 	Hovered: NewSpriteStyle(sprite, color),
	// 	Pressed: NewSpriteStyle(sprite, color),
	// }
	// PanelExt("##_panelsprite", rect, style)
	drawSprite(rect, SpriteStyle{sprite, color})
	Hovered("##_h", rect) // Just a discarded hover
}

// returns (Clicked, hovered, isdragging, dropSlot)
func DragItem(label string, rect glitch.Rect, style Style) (bool, bool, bool, bool) {
	// style := gStyle.dragItemStyle
	mask := wmHoverable | wmDrawPanel | wmDropSlot | wmDragItem | wmDrawText

	id := getId(label)
	text := removeDedup(label)

	resp := doWidget(id, text, mask, style, rect)

	clicked := resp.Released
	hovered := global.hoverOnlyId == id
	dragging := resp.Dragging
	dropping := (resp.droppedId != invalidId)
	return clicked, hovered, dragging, dropping
}

// Returns true if we dropped the drag Item to this drag slot location
func DragSlot(label string, rect glitch.Rect, style Style) bool {
	// style := gStyle.dragSlotStyle
	mask := wmHoverable | wmDrawPanel | wmDropSlot

	id := getId(label)
	text := removeDedup(label)

	resp := doWidget(id, text, mask, style, rect)

	return (resp.droppedId != invalidId)
}

func Scrollbar(idx *int, total int, rect, hoverRect glitch.Rect) {
	val := float64(*idx)
	SliderV(&val, 0, float64(total), 1, rect, hoverRect)
	*idx = int(math.Round(val))
}

func SliderV(val *float64, min, max, step float64, rect, hoverRect glitch.Rect) {
	square := rect.SubSquare()

	buttonTop := rect.CutTop(square.H())
	buttonBot := rect.CutBottom(square.H())

	delta := max - min
	ratio := (*val - min) / delta
	height := rect.H() - square.H()

	// deltaY := rect.Min.Y - rect.Max.Y + square.H()
	// yPos := (ratio * deltaY) + rect.Max.Y - (square.H() / 2)

	yPos := -(ratio * height) + rect.Max.Y - (square.H() / 2)
	PanelExt("##vscrollbarbg", rect, gStyle.scrollbarBgStyle)

	numSteps := (delta / step)
	floatStep := height / numSteps
	draggerRect := square.WithCenter(glitch.Vec2{rect.Center().X, yPos})
	if SmoothDragButton("##vscrollbarfg", &draggerRect, rect, glitch.Vec2{floatStep, floatStep}, gStyle.scrollbarHandleStyle) {
		y := rect.Max.Y - draggerRect.Center().Y - (square.H() / 2)

		ratio := y / height
		*val = ratio*float64(max) + min
	}

	topResp := ButtonFull("##vscrolltop", buttonTop, gStyle.scrollbarTopStyle)
	if topResp.Released || topResp.Repeated {
		*val -= step
	}
	botResp := ButtonFull("##vscrollbot", buttonBot, gStyle.scrollbarBotStyle)
	if botResp.Released || botResp.Repeated {
		*val += step
	}

	if HoveredNoBlock(hoverRect) || HoveredNoBlock(rect) {
		_, scrollY := global.win.MouseScroll()
		*val -= (scrollY * step)
	}

	*val = clamp(min, max, *val)
}

func SliderH(val *float64, min, max, step float64, rect, hoverRect glitch.Rect) {
	square := rect.SubSquare()

	buttonLeft := rect.CutLeft(square.W())
	buttonRight := rect.CutRight(square.W())

	delta := max - min
	ratio := (*val - min) / delta
	width := rect.W() - square.W()

	xPos := (ratio * width) + rect.Min.X + (square.W() / 2)
	PanelExt("##hscrollbarbg", rect, gStyle.scrollbarBgStyle)

	numSteps := (delta / step)
	floatStep := width / numSteps
	draggerRect := square.WithCenter(glitch.Vec2{xPos, rect.Center().Y})
	if SmoothDragButton("##hscrollbarfg", &draggerRect, rect, glitch.Vec2{floatStep, floatStep}, gStyle.scrollbarHandleStyle) {
		x := -rect.Min.X + draggerRect.Center().X - (square.W() / 2)

		ratio := x / width
		*val = (ratio * float64(delta)) + min
	}

	leftResp := ButtonFull("##hscrollup", buttonLeft, gStyle.scrollbarTopStyle)
	if leftResp.Released || leftResp.Repeated {
		*val -= step
	}

	rightResp := ButtonFull("##hscrolldown", buttonRight, gStyle.scrollbarBotStyle)
	if rightResp.Released || rightResp.Repeated {
		*val += step
	}

	// TODO: I disabled this for horizontal sliders  because I dont have a good way to detect if a scroll has been captured. For example. when you have a horizontal slider inside of a scrollbox (which is a vertical slider). the scroll from one should block the scroll from the other. I think I could fix by having the scroll get constructed and consumed each frame. So for example, you'd have it get consumed by the sliderH if the mouse is over that, then it wouldn't be consumable by the sliderV of the scrollbox (which is usually placed after)
	// if HoveredNoBlock(hoverRect) || HoveredNoBlock(rect) {
	// 	_, scrollY := global.win.MouseScroll()
	// 	*val += (scrollY * step)
	// }

	*val = clamp(min, max, *val)
}

// // Lets you click to select and set the field as active. Returns true if the field is active
// func SelectField(label string, rect glitch.Rect) bool {
// 	id := getId(label)
// 	// text := removeDedup(label)

// 	resp := WidgetResp{}
// 	resp.hoverable(id, rect)
// 	if resp.selectableOnClick(id, rect) {
// 		global.cursorPos = len(*str)
// 	}
// 	isActive := global.activeId == id
// 	return isActive
// }

func TextInput(label string, str *string, rect glitch.Rect, style Style) {
	id := getId(label)
	// text := removeDedup(label)

	resp := WidgetResp{}
	resp.hoverable(id, rect)
	if resp.selectableOnClick(id, rect) {
		global.cursorPos = len(*str)
	}

	isActive := global.activeId == id
	if isActive {
		resp.recordTyped(str, &global.cursorPos)
	}

	mask := wmDrawText | wmDrawPanel
	drawStr := *str
	if drawStr == "" {
		drawStr = " "
	}
	textResp := doWidget(id, drawStr, mask, style, rect)

	if isActive {
		// .Color(glitch.RGBA{0.5, 0.5, 0.5, 0.5})) // TODO: CursorColor? Default to white
		cursoredTextRect := MeasureText((*str)[:global.cursorPos], rect, style.Text)
		// cursoredTextRect := drawText((*str)[:global.cursorPos], rect, textStyle)
		if global.cursorPos == 0 {
			cursoredTextRect = glm.R(0, 0, 0, textResp.textRect.H())
		}

		cursorWidth := 2.0                                             // TODO: Configurable?
		cursorRect := glm.R(0, 0, cursorWidth, textResp.textRect.H()). // cursoredTextRect.H()
										Moved(glitch.Vec2{textResp.textRect.Min.X + cursoredTextRect.W(), textResp.textRect.Min.Y})
		drawSprite(cursorRect, gStyle.textCursorStyle.Normal)
	}
}

// Returns the cursor rect and the anchor vector for the tooltip
func DefaultTooltipMount() (glm.Rect, glm.Vec2) {
	quadrant := Bounds().Center().Sub(global.mousePos).Norm()
	var movement glitch.Vec2
	if quadrant.X < 0 {
		movement.X = 1
	} else {
		movement.X = 0
	}
	cursorRect := glm.R(0, 0, 0, 0).WithCenter(global.mousePos)
	return cursorRect, movement
}

func Tooltip(label string, rect glitch.Rect) {
	TooltipExt(label, rect, gStyle.tooltipStyle)
}

func TooltipExt(label string, rect glitch.Rect, style Style) {
	id := getId(label)
	text := removeDedup(label)

	resp := WidgetResp{}
	resp.hoverOnly(id, rect)

	if global.hoverOnlyId != id {
		return // Exit early if not hovered
	}

	quadrant := Bounds().Center().Sub(global.mousePos).Norm()

	var movement glitch.Vec2
	if quadrant.X < 0 {
		movement.X = 1
	} else {
		movement.X = 0
	}

	// TODO: Maybe make this configurable? I removed the Y flip because most tooltips are just single lines, and the cursor ends up blocking the text if we anchor below
	// if quadrant.Y < 0 {
	// 	movement.Y = 1
	// } else {
	// 	movement.Y = 0
	// }

	cursorRect := glm.R(0, 0, 0, 0).WithCenter(global.mousePos)
	// style.Text = style.Text.Anchor(movement.Scaled(-1)).Pivot(glitch.Vec2{0.5, 0.5})
	tmpTextStyle := style.Text.Anchor(movement)

	// TODO: Panel drawing?
	resp.textRect = drawText(text, cursorRect, tmpTextStyle)

	// g.draw(tip, cursorRect, style.Normal, style.Text)
}

func MultiText(label string, rect glitch.Rect, textStyle TextStyle) glitch.Rect {
	// style := Style{
	// 	Text: NewTextStyle().Padding(glitch.R(5, 5, 5, 5)).WordWrap(true).Anchor(glitch.Vec2{0, 1.0}),
	// 	// Text: textStyle.WordWrap(true),
	// }
	// mask := wmDrawText
	// id := getId(label)
	text := removeDedup(label)
	// doWidget(id, text, mask, style, rect)
	return drawText(text, rect, textStyle)
}

func LineGraph(rect glitch.Rect, series []glitch.Vec2, textStyle TextStyle) {
	line := global.getGraph(rect)

	line.Line(series)
	line.Axes()
	line.DrawColorMask(global.sorter, glitch.Mat4Ident, textStyle.color)

	style := Style{
		Normal: NewSpriteStyle(line, textStyle.color),
	}

	PanelExt("##_graph", rect, style)

	// global.appendUnionBounds(rect)
	// global.debugRect(rect)

	// Draw text around axes
	axes := line.GetAxes()

	textStyle.anchor = glitch.Vec2{0, 0}
	textStyle.pivot = glitch.Vec2{1, 0.5}
	drawText(fmt.Sprintf("%.2f ms", axes.Min.Y), rect, textStyle)

	textStyle.anchor = glitch.Vec2{0, 1}
	textStyle.pivot = glitch.Vec2{1, 0.5}
	drawText(fmt.Sprintf("%.2f ms", axes.Max.Y), rect, textStyle)
}
