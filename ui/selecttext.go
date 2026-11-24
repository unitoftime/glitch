package ui

import (
	"math"
	"time"

	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
)

// --------------------------------------------------------------------------------
// - Selectable Text
// --------------------------------------------------------------------------------
// Text that can be read and selected but not edited, like a chat message. The owner draws the text as usual (ie with MultiText), places it with a PlacedText to hit test and highlight it, and keeps a TextSelection for it

// Where the runes of a text drawn by TextExt or MultiText land on the screen
// Note: Reused between frames, so that placing a text doesn't allocate
type PlacedText struct {
	carets []glitch.TextCaret
	origin glm.Vec2 // Where the origin of the text's mesh lands on the screen
	scale glm.Vec2 // From the units of the text's mesh to the screen
	rowHeight float64 // On the screen
}

// Places the text like drawText does in the rect with the style. Like MultiText, the text stops at a "##"
func (p *PlacedText) Place(str string, rect glitch.Rect, style TextStyle) {
	str = removeDedup(str)
	wrapWidth := rect.W() / style.scale // Note: Matches getText

	var bounds glitch.Rect
	p.carets, bounds = global.atlas.AppendCarets(p.carets[:0], str, global.fontScale, style.wordWrap, wrapWidth)
	drawn := placeText(bounds, rect, style)

	p.origin = drawn.Min
	p.scale = glm.Vec2{}
	if bounds.W() > 0 {
		p.scale.X = drawn.W() / bounds.W()
	}
	if bounds.H() > 0 {
		p.scale.Y = drawn.H() / bounds.H()
	}
	p.rowHeight = global.atlas.UngappedLineHeight() * global.fontScale * p.scale.Y
}

// The screen rect of a row, from the top of the text down
func (p *PlacedText) rowRect(row int, minX, maxX float64) glm.Rect {
	top := p.origin.Y + float64(1 - row) * p.rowHeight // Note: The first row sits above the mesh's origin, see glitch.AppendCarets
	return glm.R(p.origin.X + minX * p.scale.X, top - p.rowHeight, p.origin.X + maxX * p.scale.X, top)
}

// Returns the byte index of the caret nearest to the position. A position above or below the text picks from its first or last row
func (p *PlacedText) IndexAt(pos glm.Vec2) int {
	if len(p.carets) == 0 { return 0 }

	lastRow := p.carets[len(p.carets) - 1].Row
	row := lastRow
	if p.rowHeight > 0 {
		row = int(math.Floor((p.origin.Y + p.rowHeight - pos.Y) / p.rowHeight))
		row = glm.Clamp(0, lastRow, row)
	}

	best, bestDist := 0, math.Inf(1)
	for _, c := range p.carets {
		if c.Row < row { continue }
		if c.Row > row { break }
		dist := math.Abs(p.origin.X + c.X * p.scale.X - pos.X)
		if dist < bestDist {
			best, bestDist = c.Index, dist
		}
	}
	return best
}

// Appends a highlight rect for each row that the bytes from start to end cover
func (p *PlacedText) AppendRects(dst []glm.Rect, start, end int) []glm.Rect {
	if start >= end { return dst }

	row := -1
	minX, maxX := 0.0, 0.0
	for _, c := range p.carets {
		if c.Index < start { continue }
		if c.Index > end { break }
		if c.Row != row {
			if row >= 0 && maxX > minX {
				dst = append(dst, p.rowRect(row, minX, maxX))
			}
			row, minX = c.Row, c.X
		}
		maxX = c.X
	}
	if row >= 0 && maxX > minX {
		dst = append(dst, p.rowRect(row, minX, maxX))
	}
	return dst
}

// A selection in a text that can't be edited. The text between Anchor and Cursor is selected, in either order
type TextSelection struct {
	Anchor, Cursor int // Byte indexes into the text
	dragging bool // A left press on the text, which selects as it drags until the button is released
	clicks clickTracker
}

// Returns the selected bytes, start first
func (s *TextSelection) Range() (int, int) {
	return min(s.Anchor, s.Cursor), max(s.Anchor, s.Cursor)
}

func (s *TextSelection) HasSelection() bool {
	return s.Anchor != s.Cursor
}

// Returns true from a left press on the text until the button is released
func (s *TextSelection) Dragging() bool {
	return s.dragging
}

// Applies the mouse to the selection of the placed text. pressed is set on the frame that a left press lands on the text: it puts the selection there, or extends it with shift, or selects a word on a double click. Then dragging selects until the button is released
func (s *TextSelection) Update(str string, placed *PlacedText, pressed bool) {
	if pressed {
		idx := placed.IndexAt(global.mousePos)
		s.dragging = pressSelection(str, idx, &s.Anchor, &s.Cursor, &s.clicks)
	}
	if s.dragging {
		if !global.win.Pressed(glitch.MouseButtonLeft) {
			s.dragging = false
		} else {
			s.Cursor = placed.IndexAt(global.mousePos)
			global.mouseCaught = true // Note: The drag keeps the mouse even after it leaves the text, so it isn't also an attack
		}
	}
	s.Anchor = runeBoundary(str, s.Anchor)
	s.Cursor = runeBoundary(str, s.Cursor)
}

// Tells a double click from two clicks: the second press lands on the same spot soon after the first
type clickTracker struct {
	last time.Time
	lastIdx int
}

// Returns true if the press at the byte index is the second of a double click
func (c *clickTracker) press(idx int) bool {
	now := time.Now()
	double := now.Sub(c.last) < doubleClickTime && idx == c.lastIdx
	c.last, c.lastIdx = now, idx
	return double
}

// Applies a left press at the byte index to a selection: shift extends it, a double click selects the word, and a click puts both ends there. Returns true if a drag from the press should go on selecting
// Note: Shared by text inputs and selectable text
func pressSelection(str string, idx int, anchor, cursor *int, clicks *clickTracker) bool {
	double := clicks.press(idx)
	win := global.win
	switch {
	case win.Pressed(glitch.KeyLeftShift) || win.Pressed(glitch.KeyRightShift):
		*cursor = idx
	case double:
		*anchor, *cursor = wordAt(str, idx)
		return false // Note: So that the word stays selected
	default:
		*anchor, *cursor = idx, idx
	}
	return true
}
