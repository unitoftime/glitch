package ui

import (
	"math"
	"strconv"
	"unsafe"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
	"golang.org/x/exp/constraints"
)

// --------------------------------------------------------------------------------
// - Debug widgets
// --------------------------------------------------------------------------------
// Dense, flat widgets for tools (eg the editor). They all draw with the DebugTheme, so they look the same whatever theme the game sets
// Most of them are laid out as fixed-height rows by a Rows, which skips the rows that are scrolled out of view rather than clipping them
// Note: A Rows doesn't call the widgets of rows it skips, so labels must be unique within their id scope (See: PushId). Otherwise the ids of duplicate labels shift as the list scrolls

type DebugTheme struct {
	RowHeight      float64
	Indent         float64 // How far each open Tree indents its children
	LabelWidth     float64 // The fraction of a Rows' width that Field labels take
	Pad            float64 // The gap between a label and its field, and between the parts of a field
	ScrollbarWidth float64
	CursorWidth    float64

	Text     TextStyle
	Panel    Style // Panel backgrounds
	Field    Style // Number fields, text inputs, dropdown buttons, toggles
	Header   Style // Tree headers, tabs
	Selected Style // The selected tab, the open dropdown, checked toggles
	Popup    Style // Dropdown list background and items
	Splitter Style
	Cursor   SpriteStyle

	Scrollbar SliderStyle

	PopupLayer int8 // The layer dropdown lists draw on, so they sit above every panel. Note: Must be inside the ui camera's depth range when the sorter depth tests, or the list is clipped
	PopupRows  int  // How many items a dropdown list shows before it scrolls
	SearchMin  int  // Dropdowns with at least this many items get a search box
}

var gDebug DebugTheme

func SetDebugTheme(theme DebugTheme) {
	gDebug = theme
}

func CurrentDebugTheme() DebugTheme {
	return gDebug
}

func defaultDebugTheme(white Drawer, grey []glitch.RGBA) DebugTheme {
	flat := func(normal, hovered, pressed glitch.RGBA) Style {
		return Style{
			Normal:  NewSpriteStyle(white, normal),
			Hovered: NewSpriteStyle(white, hovered),
			Pressed: NewSpriteStyle(white, pressed),
			Text:    NewTextStyle(),
		}
	}
	blue := glm.FromUint8(0x2f, 0x52, 0x80, 0xff)
	blueHover := glm.FromUint8(0x3a, 0x64, 0x9b, 0xff)

	return DebugTheme{
		RowHeight:      6,
		Indent:         3,
		LabelWidth:     0.4,
		Pad:            0.5,
		ScrollbarWidth: 3,
		CursorWidth:    0.5,

		Text: NewTextStyle().Scale(0.3).Anchor(glitch.Vec2{0, 0.5}).Padding(glm.R(1, 0, 1, 0)),

		Panel:    flat(grey[2], grey[2], grey[2]),
		Field:    flat(grey[4], grey[5], grey[6]),
		Header:   flat(grey[3], grey[4], grey[5]),
		Selected: flat(blue, blueHover, blueHover),
		Popup:    flat(grey[1], grey[4], grey[5]),
		Splitter: flat(grey[1], grey[7], grey[9]),
		Cursor:   NewSpriteStyle(white, grey[15]),

		Scrollbar: SliderStyle{
			Background: flat(grey[1], grey[1], grey[1]),
			Handle:     flat(grey[5], grey[6], grey[7]),
			Up:         flat(grey[3], grey[4], grey[5]),
			Down:       flat(grey[3], grey[4], grey[5]),
		},

		PopupLayer: 0, // Note: The sorter draws the lowest layer last. Not negative, because a depth tested sorter turns layers into depth, and the default camera depth range doesn't reach below 0
		PopupRows:  12,
		SearchMin:  10,
	}
}

// Draws a panel background that blocks the mouse from reaching what is behind it
func DebugPanel(rect glitch.Rect) {
	doWidget(getId("##debugpanel"), "", wmHoverable|wmDrawPanel, gDebug.Panel, rect)
}

// Returns true if the button was clicked
func DebugButton(label string, rect glitch.Rect) bool {
	return DebugButtonExt(label, rect, false)
}

// A button drawn with the selected style when selected is true. Returns true if it was clicked
func DebugButtonExt(label string, rect glitch.Rect, selected bool) bool {
	style := gDebug.Header
	if selected {
		style = gDebug.Selected
	}
	style.Text = gDebug.Text
	return ButtonExt(label, rect, style)
}

func DebugText(str string, rect glitch.Rect) {
	drawText(str, rect, gDebug.Text)
}

func DebugTextInput(label string, str *string, rect glitch.Rect) {
	textInputId(getId(label), str, rect, fieldStyle(), gDebug.Cursor, gDebug.CursorWidth)
}

// --------------------------------------------------------------------------------
// - State
// --------------------------------------------------------------------------------
// Note: Only one value can be dragged, one field edited and one dropdown open at a time, so each keeps a single slot rather than per-widget state

type debugState struct {
	treeOpen map[eid]bool // Only holds trees that were toggled. The rest use their default

	// The number field or splitter being dragged. Values are set relative to where the drag started, so clamping doesn't make the value drift from the mouse
	dragId    eid
	dragStart float64
	dragMouse glitch.Vec2
	dragMoved bool // False until the mouse moves past dragThreshold, so a click without movement can start editing instead

	// The number field being typed into. It also holds the focus while editing
	editId   eid
	editText string
	editOrig string // The text when editing started, so committing unchanged text doesn't round the value

	// The open dropdown
	popupId     eid
	popupSeen   bool // Set when the open dropdown is drawn. It closes if it isn't drawn for a frame (See: Clear)
	searchId    eid  // The open dropdown's search box
	popupSearch string
	popupScroll int
	filter      []int  // The indices of the items that match popupSearch
	filterKey   string // The search that filter was built for
	filterLen   int    // The item count that filter was built for. -1 forces a rebuild
}

var dbg = debugState{
	treeOpen: make(map[eid]bool),
}

const dragThreshold = 1.0 // How far the mouse moves before a press becomes a drag

func clearDebug() {
	if dbg.popupId != invalidId && !dbg.popupSeen {
		closePopup()
	}
	dbg.popupSeen = false
}

// Starts a value drag if the widget was just pressed, and returns the dragged value while it is held
func (resp *WidgetResp) valueDrag(id eid, value float64, axis glitch.Vec2, speed float64) (float64, bool) {
	if resp.Pressed {
		dbg.dragId = id
		dbg.dragStart = value
		dbg.dragMouse = global.mousePos
		dbg.dragMoved = false
	}
	if !resp.Held || dbg.dragId != id {
		return value, false
	}

	global.mouseCaught = true // Note: The drag keeps the mouse even after it leaves the widget
	delta := global.mousePos.Sub(dbg.dragMouse).Dot(axis)
	if !dbg.dragMoved && math.Abs(delta) <= dragThreshold {
		return value, false
	}
	dbg.dragMoved = true
	return dbg.dragStart + delta*speed, true
}

// --------------------------------------------------------------------------------
// - Rows
// --------------------------------------------------------------------------------

// Lays out fixed-height rows top to bottom and scrolls them. Keep one per list across frames, it holds the scroll position
//
//	rows.Begin(rect)
//	if r, ok := rows.Field("Speed"); ok { ui.DragFloat("Speed", &speed, 0.1, r) }
//	if rows.Tree("Children", false) { ...; rows.TreeEnd() }
//	rows.End()
type Rows struct {
	Scroll    int
	lastCount int // The row count of the last frame, which decides if a scrollbar is needed before the rows are known

	area   glitch.Rect // The whole region, which the wheel scrolls
	rest   glitch.Rect // The part of the region that rows haven't taken yet
	bar    glitch.Rect
	labelX float64 // Where Field labels end
	shown  int
	count  int
	indent float64
}

func (r *Rows) Begin(rect glitch.Rect) {
	r.area = rect
	r.shown = max(0, int(rect.H()/gDebug.RowHeight))
	r.bar = glitch.Rect{}
	if r.lastCount > r.shown {
		r.bar = rect.CutRight(gDebug.ScrollbarWidth)
	}
	r.rest = rect
	r.labelX = rect.Min.X + rect.W()*gDebug.LabelWidth
	r.Scroll = clamp(0, max(0, r.lastCount-r.shown), r.Scroll)
	r.count = 0
	r.indent = 0
}

// Draws the scrollbar. Call after the last row
func (r *Rows) End() {
	r.lastCount = r.count
	if r.count > r.shown && r.bar.W() > 0 {
		gDebug.Scrollbar.ScrollRows(&r.Scroll, r.count, r.shown, r.bar, r.area)
	}
}

// Takes the next row. Returns false if the row is scrolled out of view, in which case nothing should be drawn for it
func (r *Rows) Next() (glitch.Rect, bool) {
	i := r.count
	r.count++
	if i < r.Scroll || i >= r.Scroll+r.shown {
		return glitch.Rect{}, false
	}
	row := r.rest.CutTop(gDebug.RowHeight)
	row.CutLeft(r.indent)
	return row, true
}

// Takes the next row and draws the label on its left. Returns the rect for the field
func (r *Rows) Field(label string) (glitch.Rect, bool) {
	row, ok := r.Next()
	if !ok {
		return row, false
	}
	labelRect := row.CutLeft(max(0, r.labelX-row.Min.X))
	row.CutLeft(gDebug.Pad)
	drawText(removeDedup(label), labelRect, gDebug.Text)
	return row, true
}

// Takes the next row for a line of text
func (r *Rows) Text(str string) {
	row, ok := r.Next()
	if !ok {
		return
	}
	drawText(str, row, gDebug.Text)
}

// Takes the next row for a collapsible header and returns true if it is open. If it is open, its children are indented until TreeEnd
func (r *Rows) Tree(label string, defaultOpen bool) bool {
	return r.TreeText(label, removeDedup(label), defaultOpen)
}

// A Tree that shows text instead of its label, so the text can change (eg a count) without losing the open state, which is kept by label
func (r *Rows) TreeText(label, text string, defaultOpen bool) bool {
	id := getId(label) // Note: Always taken, even when the row is scrolled out, because the open state is looked up by it
	open, toggled := dbg.treeOpen[id]
	if !toggled {
		open = defaultOpen
	}

	if row, ok := r.Next(); ok {
		resp := doWidget(id, "", wmHoverable|wmClickable|wmDrawPanel, gDebug.Header, row)
		arrow := row.CutLeft(row.H())
		sign := "+"
		if open {
			sign = "-"
		}
		drawText(sign, arrow, gDebug.Text.Anchor(glitch.Vec2{0.5, 0.5}))
		drawText(text, row, gDebug.Text)

		if resp.Released {
			open = !open
			dbg.treeOpen[id] = open
		}
	}

	if open {
		r.indent += gDebug.Indent
	}
	return open
}

// Ends the children of an open Tree
func (r *Rows) TreeEnd() {
	r.indent = max(0, r.indent-gDebug.Indent)
}

// --------------------------------------------------------------------------------
// - Number fields
// --------------------------------------------------------------------------------
// Dragging left and right changes the value by speed per ui unit. Clicking without dragging types a value in: Enter or clicking away commits it, Escape cancels

// Returns true if the value changed
func DragFloat[T constraints.Float](label string, v *T, speed float64, rect glitch.Rect) bool {
	bits := 8 * int(unsafe.Sizeof(*v))
	display := strconv.FormatFloat(float64(*v), 'f', 3, bits)

	edit := numberField(getId(label), display, float64(*v), speed, rect)
	switch {
	case edit.dragged:
		return setChanged(v, T(edit.value))
	case edit.committed:
		val, err := strconv.ParseFloat(edit.text, bits)
		if err != nil {
			return false
		}
		return setChanged(v, T(val))
	}
	return false
}

// Returns true if the value changed
func DragInt[T constraints.Integer](label string, v *T, speed float64, rect glitch.Rect) bool {
	var zero T
	unsigned := zero-1 > 0
	bits := 8 * int(unsafe.Sizeof(*v))

	// Note: Formatted and parsed exactly, because large ids don't survive a round trip through float64. Only dragging goes through a float
	var display string
	if unsigned {
		display = strconv.FormatUint(uint64(*v), 10)
	} else {
		display = strconv.FormatInt(int64(*v), 10)
	}

	edit := numberField(getId(label), display, float64(*v), speed, rect)
	switch {
	case edit.dragged:
		val := math.Round(edit.value)
		if unsigned {
			val = max(0, val)
		}
		return setChanged(v, T(val))
	case edit.committed:
		if unsigned {
			val, err := strconv.ParseUint(edit.text, 10, bits)
			if err != nil {
				return false
			}
			return setChanged(v, T(val))
		}
		val, err := strconv.ParseInt(edit.text, 10, bits)
		if err != nil {
			return false
		}
		return setChanged(v, T(val))
	}
	return false
}

func setChanged[T comparable](v *T, val T) bool {
	if *v == val {
		return false
	}
	*v = val
	return true
}

type numberEdit struct {
	dragged bool
	value   float64 // The dragged value

	committed bool
	text      string // The typed value
}

func numberField(id eid, display string, value, speed float64, rect glitch.Rect) numberEdit {
	ret := numberEdit{}

	if dbg.editId == id && global.focusId == id {
		textInputId(id, &dbg.editText, rect, fieldStyle(), gDebug.Cursor, gDebug.CursorWidth)

		if global.win.JustPressed(glitch.KeyEscape) {
			dbg.editId = invalidId
			ClearFocus()
			return ret
		}

		clickedAway := global.focusId != id
		enter := global.win.JustPressed(glitch.KeyEnter) || global.win.JustPressed(glitch.KeyKPEnter)
		if clickedAway || enter {
			if dbg.editText != dbg.editOrig {
				ret.committed = true
				ret.text = dbg.editText
			}
			dbg.editId = invalidId
			ClearFocus()
		}
		return ret
	}

	resp := WidgetResp{}
	resp.clickable(id)
	resp.hoverable(id, rect)
	ret.value, ret.dragged = resp.valueDrag(id, value, glitch.Vec2{1, 0}, speed)

	if resp.Released && dbg.dragId == id && !dbg.dragMoved {
		dbg.editId = id
		dbg.editText = display
		dbg.editOrig = display
		global.focusId = id
		global.focusSeen = true
		global.cursorPos = len(display)
	}

	doWidget(id, display, wmDrawPanel|wmDrawText, fieldStyle(), rect)
	return ret
}

// The field style with the debug text style
func fieldStyle() Style {
	style := gDebug.Field
	style.Text = gDebug.Text
	return style
}

// Returns true if either part changed
func Vec2Field(label string, v *glm.Vec2, speed float64, rect glitch.Rect) bool {
	PushId(getId(label))
	defer PopId()

	list := HList(rect, 2)
	pad := glm.R(gDebug.Pad/2, 0, gDebug.Pad/2, 0)
	changedX := DragFloat("##x", &v.X, speed, list.Next().Unpad(pad))
	changedY := DragFloat("##y", &v.Y, speed, list.Next().Unpad(pad))
	return changedX || changedY
}

var colorLabels = [4]string{"##r", "##g", "##b", "##a"}

// A swatch and a 0 to 1 field per channel. Returns true if any channel changed
func ColorField(label string, c *glitch.RGBA, rect glitch.Rect) bool {
	PushId(getId(label))
	defer PopId()

	swatch := rect.CutLeft(rect.H())
	rect.CutLeft(gDebug.Pad)
	drawSprite(swatch, gDebug.Field.Normal.Color(*c))

	list := HList(rect, 4)
	pad := glm.R(gDebug.Pad/2, 0, gDebug.Pad/2, 0)
	channels := [4]*float64{&c.R, &c.G, &c.B, &c.A}
	changed := false
	for i, ch := range channels {
		if DragFloat(colorLabels[i], ch, 0.005, list.Next().Unpad(pad)) {
			*ch = clamp(0, 1, *ch)
			changed = true
		}
	}
	return changed
}

// --------------------------------------------------------------------------------
// - Toggle, Tabs, Splitter
// --------------------------------------------------------------------------------

// A checkbox at the left of the rect. Returns true if it was toggled
func Toggle(label string, v *bool, rect glitch.Rect) bool {
	box := rect.CutLeft(rect.H())
	resp := doWidget(getId(label), "", wmHoverable|wmClickable|wmDrawPanel, gDebug.Field, box)
	if *v {
		drawSprite(box.Unpad(glm.R(1, 1, 1, 1)), gDebug.Selected.Normal)
	}
	if resp.Released {
		*v = !*v
		return true
	}
	return false
}

// A row of equal-width tabs. Returns true if the selected tab changed
func Tabs(label string, idx *int, names []string, rect glitch.Rect) bool {
	PushId(getId(label))
	defer PopId()

	list := HList(rect, len(names))
	changed := false
	for i := range names {
		style := gDebug.Header
		if i == *idx {
			style = gDebug.Selected
		}
		style.Text = gDebug.Text.Anchor(glitch.Vec2{0.5, 0.5})

		PushId(i)
		resp := doWidget(getId("##tab"), names[i], wmHoverable|wmClickable|wmDrawPanel|wmDrawText, style, list.Next())
		PopId()

		if resp.Released && *idx != i {
			*idx = i
			changed = true
		}
	}
	return changed
}

// A handle that drags the x coordinate of a vertical split line (eg the edge of a side panel). Returns true if it moved
func SplitterX(label string, x *float64, min, max float64, rect glitch.Rect) bool {
	return splitter(getId(label), x, min, max, glitch.Vec2{1, 0}, rect)
}

// A handle that drags the y coordinate of a horizontal split line. Returns true if it moved
func SplitterY(label string, y *float64, min, max float64, rect glitch.Rect) bool {
	return splitter(getId(label), y, min, max, glitch.Vec2{0, 1}, rect)
}

func splitter(id eid, v *float64, min, max float64, axis glitch.Vec2, rect glitch.Rect) bool {
	resp := doWidget(id, "", wmHoverable|wmClickable|wmDrawPanel, gDebug.Splitter, rect)
	val, dragged := resp.valueDrag(id, *v, axis, 1)
	if !dragged {
		return false
	}
	return setChanged(v, clamp(min, max, val))
}

// --------------------------------------------------------------------------------
// - Dropdown
// --------------------------------------------------------------------------------

// A button that opens a list to pick from. Lists with at least SearchMin items get a search box, where Enter picks the first match
// Returns true if the selection changed
func Dropdown(label string, idx *int, items []string, rect glitch.Rect) bool {
	id := getId(label)
	open := dbg.popupId == id

	current := ""
	if *idx >= 0 && *idx < len(items) {
		current = items[*idx]
	}
	style := fieldStyle()
	if open {
		style.Normal = gDebug.Selected.Normal
	}
	resp := doWidget(id, current, wmHoverable|wmClickable|wmDrawPanel|wmDrawText, style, rect)

	if resp.Released {
		if open {
			closePopup()
			return false
		}
		openPopup(id, *idx)
		open = true
	}
	if !open {
		return false
	}
	dbg.popupSeen = true

	return drawPopup(id, idx, items, rect)
}

// Closes the open dropdown and gives back the keyboard its search box held
func closePopup() {
	dbg.popupId = invalidId
	if global.focusId == dbg.searchId {
		ClearFocus()
	}
}

func openPopup(id eid, selected int) {
	closePopup() // Note: Another dropdown might be open
	dbg.popupId = id
	dbg.popupSearch = ""
	dbg.filterLen = -1
	dbg.popupScroll = max(0, selected-gDebug.PopupRows/2) // Note: ScrollRows clamps it once the list is drawn
}

func drawPopup(id eid, idx *int, items []string, button glitch.Rect) bool {
	lastLayer := Layer()
	SetLayer(gDebug.PopupLayer)
	defer SetLayer(lastLayer)

	PushId(id)
	defer PopId()

	searching := len(items) >= gDebug.SearchMin
	if dbg.filterLen != len(items) || dbg.filterKey != dbg.popupSearch {
		rebuildFilter(items)
	}

	// Below the button, or above it if the list would leave the screen
	rows := min(len(dbg.filter), gDebug.PopupRows)
	if searching {
		rows++
	}
	height := float64(max(rows, 1)) * gDebug.RowHeight
	popup := glm.R(button.Min.X, button.Min.Y-height, button.Max.X, button.Min.Y)
	if popup.Min.Y < Bounds().Min.Y {
		popup = glm.R(button.Min.X, button.Max.Y, button.Max.X, button.Max.Y+height)
	}

	clickedAway := global.win.JustPressed(glitch.MouseButtonLeft) &&
		!popup.Contains(global.mousePos) && !button.Contains(global.mousePos)
	if clickedAway || global.win.JustPressed(glitch.KeyEscape) {
		closePopup()
		return false
	}

	doWidget(getId("##popupbg"), "", wmHoverable|wmDrawPanel, gDebug.Popup, popup)

	list := popup
	if searching {
		dbg.searchId = getId("##search")
		global.focusId = dbg.searchId // Note: The search box keeps the keyboard while the list is open
		global.focusSeen = true
		textInputId(dbg.searchId, &dbg.popupSearch, list.CutTop(gDebug.RowHeight), fieldStyle(), gDebug.Cursor, gDebug.CursorWidth)

		enter := global.win.JustPressed(glitch.KeyEnter) || global.win.JustPressed(glitch.KeyKPEnter)
		if enter && len(dbg.filter) > 0 {
			closePopup()
			return setChanged(idx, dbg.filter[0])
		}
	}

	start, end := 0, len(dbg.filter)
	if len(dbg.filter) > gDebug.PopupRows {
		bar := list.CutRight(gDebug.ScrollbarWidth)
		start, end = gDebug.Scrollbar.ScrollRows(&dbg.popupScroll, len(dbg.filter), gDebug.PopupRows, bar, list)
	}

	itemStyle := gDebug.Popup
	itemStyle.Text = gDebug.Text
	selectedStyle := gDebug.Selected
	selectedStyle.Text = gDebug.Text
	for _, item := range dbg.filter[start:end] {
		style := itemStyle
		if item == *idx {
			style = selectedStyle
		}

		PushId(item)
		resp := doWidget(getId("##item"), items[item], wmHoverable|wmClickable|wmDrawPanel|wmDrawText, style, list.CutTop(gDebug.RowHeight))
		PopId()

		if resp.Released {
			closePopup()
			return setChanged(idx, item)
		}
	}
	return false
}

func rebuildFilter(items []string) {
	dbg.filterKey = dbg.popupSearch
	dbg.filterLen = len(items)
	dbg.filter = dbg.filter[:0]
	for i := range items {
		if containsFold(items[i], dbg.popupSearch) {
			dbg.filter = append(dbg.filter, i)
		}
	}
}

// Returns true if sub is in s, ignoring ASCII case
func containsFold(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if lowerASCII(s[i+j]) != lowerASCII(sub[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func lowerASCII(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
