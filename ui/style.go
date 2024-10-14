package ui

import (
	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
)

type FullStyle struct {
	buttonStyle   Style
	panelStyle    Style
	dragSlotStyle Style
	dragItemStyle Style

	scrollbarTopStyle    Style
	scrollbarBotStyle    Style
	scrollbarHandleStyle Style
	scrollbarBgStyle     Style

	checkboxStyleTrue  Style
	checkboxStyleFalse Style

	textInputPanelStyle Style
	textCursorStyle     Style
	tooltipStyle        Style

	// textStyle TextStyle // TODO: This should also include the atlas
}

var gStyle FullStyle

func SetButtonStyle(style Style) {
	gStyle.buttonStyle = style
}
func SetPanelStyle(style Style) {
	gStyle.panelStyle = style
}
func SetDragSlotStyle(style Style) {
	gStyle.dragSlotStyle = style
}
func DragSlotStyle() Style {
	return gStyle.dragSlotStyle
}

//	func SetDragItemStyle(style Style) {
//		gStyle.dragItemStyle = style
//	}
func DragItemStyle() Style {
	return gStyle.dragItemStyle
}
func SetCheckboxStyleTrue(style Style) {
	gStyle.checkboxStyleTrue = style
}
func SetCheckboxStyleFalse(style Style) {
	gStyle.checkboxStyleFalse = style
}

// func SetTextStyle(style TextStyle) {
// 	gStyle.textStyle = style
// }

func SetTooltipStyle(style Style) {
	gStyle.tooltipStyle = style
}

func SetScrollbarTopStyle(style Style) {
	gStyle.scrollbarTopStyle = style
}
func SetScrollbarBottomStyle(style Style) {
	gStyle.scrollbarBotStyle = style
}
func SetScrollbarBgStyle(style Style) {
	gStyle.scrollbarBgStyle = style
}
func SetScrollbarHandleStyle(style Style) {
	gStyle.scrollbarHandleStyle = style
}

//--------------------------------------------------------------------------------

type SpriteStyle struct {
	sprite Drawer
	color  glitch.RGBA
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
	Text                     TextStyle
}

func NewStyle(normal Drawer, color glitch.RGBA) Style {
	return Style{
		Normal:  NewSpriteStyle(normal, color),
		Hovered: NewSpriteStyle(normal, color),
		Pressed: NewSpriteStyle(normal, color),
		Text:    NewTextStyle(),
	}
}

func ButtonStyle(normal, hovered, pressed Drawer) Style {
	return Style{
		Normal:  NewSpriteStyle(normal, glitch.White),
		Hovered: NewSpriteStyle(hovered, glitch.White),
		Pressed: NewSpriteStyle(pressed, glitch.White),
		Text:    NewTextStyle(),
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
	padding       glitch.Rect
	color         glitch.RGBA
	scale         float64
	autoFit       bool // Auto scale the text to fit the rectangle
	fitInteger    bool // If autoscaling, then only scale by integers (for pixel fonts)
	wordWrap      bool
	shadow        glitch.Vec2
}

// TODO: I kind of feel like the string needs to be in here, I'm not sure though
func NewTextStyle() TextStyle {
	return TextStyle{
		anchor:  glitch.Vec2{0.5, 0.5},
		pivot:   glitch.Vec2{0.5, 0.5},
		padding: glm.R(0, 0, 0, 0),
		color:   glitch.White,
		scale:   1.0,
		autoFit: false,
		shadow:  glitch.Vec2{0.0, 0.0},
	}
}

func (s TextStyle) Anchor(v glitch.Vec2) TextStyle {
	s.anchor = v
	s.pivot = v
	return s
}

func (s TextStyle) Pivot(v glitch.Vec2) TextStyle {
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
func (s TextStyle) Autofit(v bool) TextStyle {
	s.autoFit = v
	if s.autoFit {
		s.wordWrap = false
	}
	return s
}
func (s TextStyle) FitInteger(v bool) TextStyle {
	s.fitInteger = v
	return s
}

func (s TextStyle) Shadow(v glitch.Vec2) TextStyle {
	s.shadow = v
	return s
}
func (s TextStyle) WordWrap(v bool) TextStyle {
	s.wordWrap = v
	if s.wordWrap {
		s.autoFit = false
	}
	return s
}
