package ui

import "github.com/unitoftime/glitch"

// --------------------------------------------------------------------------------
// - Layout
// --------------------------------------------------------------------------------
type LayoutType uint8

const (
	CutLeft LayoutType = iota
	CutTop
	CutRight
	CutBottom
	Centered
)

type SizeType uint8

const (
	// SizeNone SizeType = iota
	SizePixels SizeType = iota
	SizeText
	SizeParent // Percent of parent
	SizeChildren
)

func (s SizeType) Calc(val float64, parentSize float64, textSize float64) float64 {
	switch s {
	case SizePixels:
		return val
	case SizeText:
		return val * textSize
	case SizeParent:
		return val * parentSize
	}
	return 100 // TODO
}

type Size struct {
	TypeX, TypeY SizeType
	Value        glitch.Vec2

	// Type string // Null, Pixels, TextContent, PercentOfParent, ChildrenSum
	// value float32
	// strictness float32
}

type Layout struct {
	Type    LayoutType
	Bounds  glitch.Rect
	Padding glitch.Rect
	Size    Size
}

func (l *Layout) Next(textBounds glitch.Rect) glitch.Rect {
	size := l.Size

	// TODO: If percent of parent
	// cutX := size.value.X * l.Bounds.W()
	// cutY := size.value.Y * l.Bounds.H()
	cutX := size.TypeX.Calc(size.Value.X, l.Bounds.W(), textBounds.W())
	cutY := size.TypeY.Calc(size.Value.Y, l.Bounds.H(), textBounds.H())

	ret := l.Bounds
	switch l.Type {
	case CutLeft:
		ret = l.Bounds.CutLeft(cutX)
	case CutRight:
		ret = l.Bounds.CutRight(cutX)
	case CutTop:
		ret = l.Bounds.CutTop(cutY)
	case CutBottom:
		ret = l.Bounds.CutBottom(cutY)
	case Centered:
		ret = ret.SliceHorizontal(cutY).SliceVertical(cutX)
	}

	return ret.Unpad(l.Padding)
}

// --------------------------------------------------------------------------------
type vList struct {
	rect       glitch.Rect
	last       glitch.Rect
	size       float64
	padNext    float64
	fromBottom bool
}

func (l *vList) Next() glitch.Rect {
	if l.fromBottom {
		if l.padNext != 0 {
			l.rect.CutBottom(l.padNext)
		}
		l.last = l.rect.CutBottom(l.size)
	} else {
		if l.padNext != 0 {
			l.rect.CutTop(l.padNext)
		}
		l.last = l.rect.CutTop(l.size)
	}
	return l.last
}

func (l vList) Last() glitch.Rect {
	return l.last
}

func VList(rect glitch.Rect, num int) vList {
	elementHeight := rect.H() / float64(num)
	return vList{
		rect: rect,
		last: rect,
		size: elementHeight,
	}
}

func VList2(rect glitch.Rect, size float64) vList {
	return vList{
		rect: rect,
		last: rect,
		size: size,
	}
}

func (l vList) Bottom(val bool) vList {
	l.fromBottom = val
	return l
}

func (l vList) Pad(val float64) vList {
	l.padNext = val
	return l
}

type hList struct {
	rect glitch.Rect
	last glitch.Rect
	size float64
}

func HList(rect glitch.Rect, num int) hList {
	size := rect.W() / float64(num)
	return hList{
		rect: rect,
		last: rect,
		size: size,
	}
}
func HList2(rect glitch.Rect, size float64) hList {
	return hList{
		rect: rect,
		last: rect,
		size: size,
	}
}
func (l *hList) Next() glitch.Rect {
	l.last = l.rect.CutLeft(l.size)
	return l.last
}

func (l hList) Last() glitch.Rect {
	return l.last
}

type gridList struct {
	rect       glitch.Rect
	last       glitch.Rect
	sizeX      float64
	sizeY      float64
	numX, numY int
	// Type // TODO: RowOrder or ColumnOrder

	currentRect glitch.Rect
	currentNum  int
}

func GridList(rect glitch.Rect, numX, numY int) gridList {
	sizeX := rect.W() / float64(numX)
	sizeY := rect.H() / float64(numY)
	return gridList{
		rect:  rect,
		last:  rect,
		sizeX: sizeX,
		sizeY: sizeY,
		numX:  numX,
		numY:  numY,
	}
}

func (l *gridList) Next() glitch.Rect {
	// Row Order (ie Horizontal first)
	if l.currentNum%l.numX == 0 {
		l.currentRect = l.rect.CutTop(l.sizeY)
	}
	l.currentNum++

	l.last = l.currentRect.CutLeft(l.sizeX)

	return l.last
}

func (l gridList) Last() glitch.Rect {
	return l.last
}
