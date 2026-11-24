package ui

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
)

// --------------------------------------------------------------------------------
// - Text Input
// --------------------------------------------------------------------------------
// A single line text box. It has the keyboard from a click on it until a click anywhere else (see focusOnClick), and while it does:
// - Typing, pasting, Backspace and Delete replace the selection, if there is one
// - Arrows, Home and End move the cursor. Ctrl (or Alt, ie Option on macOS) moves by word, and Super (Cmd on macOS) with an arrow goes to the start or end. Shift extends the selection
// - Ctrl (or Cmd) with A selects all, C copies, X cuts, V pastes, Z undoes, and Y or Shift+Z redoes
// - Clicking puts the cursor where it lands, dragging selects, shift-clicking extends the selection, and a double click selects a word
// Note: Cmd counts as Ctrl, and Option as the word key, everywhere, rather than only on macOS. So the browser build gets them on a Mac too, where it can't tell what it runs on. On the other systems those keys are mostly taken by the system anyway

// The options of a text input, see TextInputExt
type TextInputConfig struct {
	Placeholder string // Drawn dim while the text is empty, with the cursor at its start
	MaxLen int // The most bytes that the text can have, ie to match a limit that a server checks. 0 is no limit
}

func TextInput(label string, str *string, rect glitch.Rect, style Style) {
	TextInputExt(label, str, rect, style, TextInputConfig{})
}

func TextInputExt(label string, str *string, rect glitch.Rect, style Style, config TextInputConfig) {
	textInputId(getId(label), str, rect, style, config, gStyle.TextCursor.Normal, 2.0)
}

const (
	maxUndoSteps = 100
	doubleClickTime = 400 * time.Millisecond
	cursorBlinkPeriod = 1060 * time.Millisecond // One on and off. The cursor shows for the first half
	selectionAlpha = 0.35 // Of the cursor's color, for the highlight behind the selected text
	placeholderAlpha = 0.5 // Of the text's color
)

// The editing state of the text input that has the keyboard. Only one input has it at a time, so there is only one of these (see uiGlobals.edit)
type textEdit struct {
	id eid // The input that it belongs to. It starts over when another input gets the keyboard
	cursor int // Byte index into the text, always on a rune boundary
	anchor int // The other end of the selection, which is the text between it and the cursor. Equal to the cursor when nothing is selected
	scroll int // Byte index of the first rune that is shown, so that a text wider than the input keeps the cursor in view

	undo, redo []textSnapshot
	lastKind editKind // What the last edit was, so that a run of typing or deleting is undone as one step

	lastActive time.Time // The last edit, move or click. The cursor blink restarts from it, so the cursor shows while they're working
	clicks clickTracker
	dragging bool // A drag that started in the input, which selects as it goes
}

type textSnapshot struct {
	str string
	cursor, anchor int
}

type editKind uint8
const (
	editNone editKind = iota // Not an edit, ie a move. It ends a run of typing
	editTyping // Runs of these are one undo step
	editDeleting // Runs of these are one undo step
	editOther // Each of these is its own undo step, ie a paste or a cut
)

// Gives the keyboard to the input, with the cursor at the byte index
// Note: The index is clamped to the text when the input is drawn, so a large one puts the cursor at the end
func (g *uiGlobals) beginEdit(id eid, cursor int) {
	g.focusId = id
	g.focusSeen = true // Note: So that focusing before the element is first drawn doesn't drop the focus on the next Clear

	e := &g.edit
	undo, redo := e.undo[:0], e.redo[:0]
	*e = textEdit{
		id: id,
		cursor: cursor,
		anchor: cursor,
		undo: undo,
		redo: redo,
		lastActive: time.Now(),
	}
}

// Returns the start and end of the selection
func (e *textEdit) selection() (int, int) {
	return min(e.cursor, e.anchor), max(e.cursor, e.anchor)
}

func (e *textEdit) hasSelection() bool {
	return e.cursor != e.anchor
}

// Keeps the cursor, the selection and the scroll inside the text and on rune boundaries. The text can change from outside the input between frames
func (e *textEdit) clamp(str string) {
	e.cursor = runeBoundary(str, e.cursor)
	e.anchor = runeBoundary(str, e.anchor)
	e.scroll = runeBoundary(str, e.scroll)
}

func (e *textEdit) touch() {
	e.lastActive = time.Now()
}

// Moves the cursor. The selection goes with it, unless it is being extended
func (e *textEdit) moveTo(idx int, extend bool) {
	e.cursor = idx
	if !extend {
		e.anchor = idx
	}
	e.lastKind = editNone
	e.touch()
}

// Saves the text for undo before an edit of the kind. A run of typing or deleting only saves before its first edit
func (e *textEdit) record(str string, kind editKind) {
	if kind != editOther && kind == e.lastKind { return }

	e.undo = append(e.undo, textSnapshot{str, e.cursor, e.anchor})
	if len(e.undo) > maxUndoSteps {
		e.undo = append(e.undo[:0], e.undo[1:]...)
	}
	e.redo = e.redo[:0]
	e.lastKind = kind
}

// Replaces the selection with the text, and puts the cursor after it. Returns the new text
// Note: With a MaxLen, only as much of the insert goes in as fits
func (e *textEdit) replaceSelection(str, insert string, maxLen int, kind editKind) string {
	start, end := e.selection()
	if maxLen > 0 {
		insert = fitBytes(insert, maxLen - (len(str) - (end - start)))
	}
	if insert == "" && start == end { return str } // Nothing changes

	e.record(str, kind)
	str = str[:start] + insert + str[end:]
	e.cursor = start + len(insert)
	e.anchor = e.cursor
	e.touch()
	return str
}

// Deletes the text between the two byte indexes, or the selection if there is one. Returns the new text
func (e *textEdit) deleteRange(str string, a, b int) string {
	if !e.hasSelection() {
		e.anchor, e.cursor = a, b
	}
	return e.replaceSelection(str, "", 0, editDeleting)
}

// Steps back through the undo history, or forward through the redo history, keeping the text it leaves for the other one. Returns the new text
func (e *textEdit) step(str string, from, to *[]textSnapshot) string {
	if len(*from) == 0 { return str }

	last := (*from)[len(*from)-1]
	*from = (*from)[:len(*from)-1]
	*to = append(*to, textSnapshot{str, e.cursor, e.anchor})

	e.cursor, e.anchor = last.cursor, last.anchor
	e.lastKind = editNone
	e.touch()
	return last.str
}

func (e *textEdit) selectedText(str string) string {
	start, end := e.selection()
	return str[start:end]
}

// Applies this frame's keyboard to the text. Returns the new text
func (e *textEdit) handleKeys(str string, config TextInputConfig) string {
	win := global.win
	ctrl := win.Pressed(glitch.KeyLeftControl) || win.Pressed(glitch.KeyRightControl)
	alt := win.Pressed(glitch.KeyLeftAlt) || win.Pressed(glitch.KeyRightAlt)
	super := win.Pressed(glitch.KeyLeftSuper) || win.Pressed(glitch.KeyRightSuper)
	shift := win.Pressed(glitch.KeyLeftShift) || win.Pressed(glitch.KeyRightShift)

	shortcut := ctrl || super // Note: Cmd on macOS
	byWord := ctrl || alt // Note: Option on macOS
	toEnd := super // Note: Cmd with an arrow on macOS
	pressed := func(key glitch.Key) bool {
		return win.JustPressed(key) || win.Repeated(key)
	}

	// Note: Not while a shortcut is held, so that Ctrl+A doesn't also type an 'a'. Except with Alt too, because AltGr sends Ctrl and Alt together, and it does type
	if typed := win.Typed(); len(typed) > 0 && !(shortcut && !alt) {
		str = e.replaceSelection(str, singleLine(string(typed)), config.MaxLen, editTyping)
	}

	if shortcut {
		switch {
		case win.JustPressed(glitch.KeyA):
			e.anchor = 0
			e.moveTo(len(str), true)
		case win.JustPressed(glitch.KeyC):
			if e.hasSelection() {
				win.SetClipboard(e.selectedText(str))
			}
		case win.JustPressed(glitch.KeyX):
			if e.hasSelection() {
				win.SetClipboard(e.selectedText(str))
				str = e.replaceSelection(str, "", 0, editOther)
			}
		case win.JustPressed(glitch.KeyV):
			// Note: Empty in the browser, where pasted text comes in through Window.Pasted instead
			str = e.replaceSelection(str, singleLine(win.Clipboard()), config.MaxLen, editOther)
		case pressed(glitch.KeyY), shift && pressed(glitch.KeyZ):
			str = e.step(str, &e.redo, &e.undo)
		case pressed(glitch.KeyZ):
			str = e.step(str, &e.undo, &e.redo)
		}
	}
	if pasted := win.Pasted(); pasted != "" {
		str = e.replaceSelection(str, singleLine(pasted), config.MaxLen, editOther)
	}

	if pressed(glitch.KeyBackspace) {
		start := prevRune(str, e.cursor)
		if toEnd {
			start = 0
		} else if byWord {
			start = findWordIndexLeft(str, e.cursor)
		}
		str = e.deleteRange(str, start, e.cursor)
	}
	if pressed(glitch.KeyDelete) {
		end := nextRune(str, e.cursor)
		if toEnd {
			end = len(str)
		} else if byWord {
			end = findWordIndexRight(str, e.cursor)
		}
		str = e.deleteRange(str, e.cursor, end)
	}

	if pressed(glitch.KeyLeft) {
		start, _ := e.selection()
		switch {
		case toEnd:
			e.moveTo(0, shift)
		case byWord:
			e.moveTo(findWordIndexLeft(str, e.cursor), shift)
		case e.hasSelection() && !shift:
			e.moveTo(start, false) // Note: Collapses to the left end of the selection, rather than moving past it
		default:
			e.moveTo(prevRune(str, e.cursor), shift)
		}
	}
	if pressed(glitch.KeyRight) {
		_, end := e.selection()
		switch {
		case toEnd:
			e.moveTo(len(str), shift)
		case byWord:
			e.moveTo(findWordIndexRight(str, e.cursor), shift)
		case e.hasSelection() && !shift:
			e.moveTo(end, false)
		default:
			e.moveTo(nextRune(str, e.cursor), shift)
		}
	}
	if win.JustPressed(glitch.KeyHome) {
		e.moveTo(0, shift)
	}
	if win.JustPressed(glitch.KeyEnd) {
		e.moveTo(len(str), shift)
	}

	return str
}

// Applies this frame's mouse to the cursor: a click puts it where it lands, a drag selects, shift extends the selection, and a double click selects a word
func (e *textEdit) handleMouse(str string, hot bool, layout *textLayout) {
	win := global.win
	if hot && win.JustPressed(glitch.MouseButtonLeft) {
		idx := layout.indexAt(str, global.mousePos.X)
		e.dragging = pressSelection(str, idx, &e.anchor, &e.cursor, &e.clicks)
		e.lastKind = editNone
		e.touch()
	}

	if e.dragging {
		if !win.Pressed(glitch.MouseButtonLeft) {
			e.dragging = false
		} else {
			e.moveTo(layout.indexAt(str, global.mousePos.X), true)
			global.mouseCaught = true // Note: The drag keeps the mouse even after it leaves the input, so it isn't also an attack
		}
	}
}

// Where the text of an input sits on the screen, and the part of it that is shown
type textLayout struct {
	area glitch.Rect // Inside of the text padding
	line glitch.Rect // Where a line of the text is drawn, which the cursor and the selection fill the height of
	scale float64 // Of the text, on top of the font scale
	start, end int // The bytes of the text that are shown

	// Note: Reused between frames, see uiGlobals.textLayout
	bounds []int // Every rune boundary from start to end
	offsets []float64 // The x of each of the bounds, from the left of the area
}

// Lays out the text in the rect. The scroll keeps the cursor in view when the edit is given, ie for the input that has the keyboard. Others show the start of their text
func (l *textLayout) update(str string, rect glitch.Rect, t TextStyle, e *textEdit) {
	l.area = rect.Unpad(t.padding)
	l.scale = inputTextScale(l.area, t)

	lineHeight := global.atlas.UngappedLineHeight() * global.fontScale * l.scale
	l.line = l.area.FullAnchorOLD(glm.R(0, 0, 0, lineHeight), glitch.Vec2{0, t.anchor.Y}, glitch.Vec2{0, t.pivot.Y})

	l.start = 0
	if e != nil {
		e.scroll = min(e.scroll, e.cursor)
		for e.scroll < e.cursor && l.width(str[e.scroll:e.cursor]) > l.area.W() {
			e.scroll = nextRune(str, e.scroll) // Note: Moves the start up until the cursor fits
		}
		for e.scroll > 0 && l.width(str[prevRune(str, e.scroll):]) <= l.area.W() {
			e.scroll = prevRune(str, e.scroll) // Note: Pulls the start back while the rest still fits, ie after deleting from the end
		}
		l.start = e.scroll
	}

	l.bounds = append(l.bounds[:0], l.start)
	l.offsets = append(l.offsets[:0], 0)
	l.end = l.start
	x := 0.0
	for i, r := range str[l.start:] {
		x += global.atlas.RuneAdvance(r, global.fontScale) * l.scale
		if x > l.area.W() { break }

		l.end = l.start + i + utf8.RuneLen(r)
		l.bounds = append(l.bounds, l.end)
		l.offsets = append(l.offsets, x)
	}
}

// The width that the text takes up at the layout's scale
func (l *textLayout) width(str string) float64 {
	w := 0.0
	for _, r := range str {
		w += global.atlas.RuneAdvance(r, global.fontScale) * l.scale
	}
	return w
}

// Returns the x on the screen of the byte index. An index outside of what is shown is put at the nearest edge of it
func (l *textLayout) xOf(idx int) float64 {
	for i, b := range l.bounds {
		if b >= idx {
			return l.area.Min.X + l.offsets[i]
		}
	}
	return l.area.Min.X + l.offsets[len(l.offsets)-1]
}

// Returns the rune boundary nearest to the x on the screen. Past an edge it is the rune just beyond what is shown, so that a drag past the edge scrolls
func (l *textLayout) indexAt(str string, x float64) int {
	if x < l.area.Min.X && l.start > 0 {
		return prevRune(str, l.start)
	}
	if x > l.area.Max.X && l.end < len(str) {
		return nextRune(str, l.end)
	}

	best := l.bounds[0]
	bestDist := -1.0
	for i, b := range l.bounds {
		dist := abs(l.area.Min.X + l.offsets[i] - x)
		if bestDist < 0 || dist < bestDist {
			best, bestDist = b, dist
		}
	}
	return best
}

// Returns the rect across the line from one x to another
func (l *textLayout) span(x1, x2 float64) glitch.Rect {
	return glm.R(x1, l.line.Min.Y, x2, l.line.Max.Y)
}

// The scale that the text of an input is drawn at, on top of the font scale
// Note: An auto fit style fits a line of text to the height of the input, rather than fitting the whole text into it, so the text doesn't shrink as it's typed
func inputTextScale(area glitch.Rect, t TextStyle) float64 {
	if !t.autoFit { return t.scale }

	scale := area.H() / (global.atlas.UngappedLineHeight() * global.fontScale)
	if t.fitInteger {
		scale = max(1, float64(int(scale)))
	}
	return scale
}

// Draws one line of text at the layout's scale, starting from the x
func (l *textLayout) drawLine(str string, x float64, t TextStyle) {
	if str == "" { return }

	text := global.getText(str, t, l.area)
	w := text.Bounds().W() * l.scale
	text.RectDrawColorMask(global.sorter, glm.R(x, l.line.Min.Y, x + w, l.line.Max.Y), t.color)
}

func textInputId(id eid, str *string, rect glitch.Rect, style Style, config TextInputConfig, cursor SpriteStyle, cursorWidth float64) {
	resp := WidgetResp{}
	resp.hoverable(id, rect)
	justFocused := resp.focusOnClick(id)

	// Note: Always left aligned, so that the cursor and the selection line up with the text, and an empty input has its cursor at the start
	t := style.Text
	t.anchor.X, t.pivot.X = 0, 0

	layout := &global.textLayout
	e := &global.edit
	focused := global.focusId == id
	if focused {
		global.focusSeen = true
		// Note: A new edit every time the input gets the keyboard, so that the undo history of an earlier one doesn't carry over
		if justFocused || e.id != id {
			global.beginEdit(id, len(*str))
		}
		e.clamp(*str)

		layout.update(*str, rect, t, e)
		e.handleMouse(*str, global.hotId == id, layout)
		*str = e.handleKeys(*str, config)
		e.clamp(*str)
		layout.update(*str, rect, t, e)
	} else {
		layout.update(*str, rect, t, nil)
	}

	doWidget(id, "", wmDrawPanel, style, rect)

	if *str == "" {
		placeholder := t
		placeholder.color = t.color.Mult(glm.Alpha(placeholderAlpha))
		hint := config.Placeholder
		for hint != "" && layout.width(hint) > layout.area.W() {
			hint = hint[:prevRune(hint, len(hint))] // Note: Cut to the input, like a long text is
		}
		layout.drawLine(hint, layout.area.Min.X, placeholder)
	} else {
		if focused && e.hasSelection() {
			start, end := e.selection()
			selected := SpriteStyle{cursor.sprite, cursor.color.Mult(glm.Alpha(selectionAlpha)), 0}
			drawSprite(layout.span(layout.xOf(start), layout.xOf(end)), selected)
		}
		layout.drawLine((*str)[layout.start:layout.end], layout.area.Min.X, t)
	}

	if focused && time.Since(e.lastActive) % cursorBlinkPeriod < cursorBlinkPeriod / 2 {
		x := layout.xOf(e.cursor)
		drawSprite(layout.span(x, x + cursorWidth), cursor)
	}
}

// --------------------------------------------------------------------------------
// - Text helpers
// --------------------------------------------------------------------------------
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

// Returns the start and end of the word at the byte index, or of the run of spaces if it's on one
func wordAt(str string, idx int) (int, int) {
	if str == "" { return 0, 0 }

	at := idx
	if at >= len(str) {
		at = prevRune(str, len(str))
	}
	r, _ := utf8.DecodeRuneInString(str[at:])
	onSpace := r == ' '

	start := at
	for start > 0 {
		r, size := utf8.DecodeLastRuneInString(str[:start])
		if (r == ' ') != onSpace { break }
		start -= size
	}
	end := at
	for end < len(str) {
		r, size := utf8.DecodeRuneInString(str[end:])
		if (r == ' ') != onSpace { break }
		end += size
	}
	return start, end
}

// Returns the byte index of the rune before the index, or 0 at the start
func prevRune(str string, idx int) int {
	_, size := utf8.DecodeLastRuneInString(str[:idx])
	return idx - size
}

// Returns the byte index of the rune after the index, or the length at the end
func nextRune(str string, idx int) int {
	_, size := utf8.DecodeRuneInString(str[idx:])
	return idx + size
}

// Returns the index clamped into the string, and moved back onto a rune boundary
func runeBoundary(str string, idx int) int {
	idx = clamp(0, len(str), idx)
	for idx > 0 && idx < len(str) && !utf8.RuneStart(str[idx]) {
		idx--
	}
	return idx
}

// Returns as much of the start of the string as fits in n bytes, without cutting a rune in half
func fitBytes(str string, n int) string {
	if n <= 0 { return "" }
	if len(str) <= n { return str }
	return str[:runeBoundary(str, n)]
}

// Drops newlines and other control characters, so pasted text fits on one line
func singleLine(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, str)
}

func abs(v float64) float64 {
	if v < 0 { return -v }
	return v
}
