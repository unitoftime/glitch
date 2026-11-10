package ui

import (
	"github.com/unitoftime/flow/glm"
	"github.com/unitoftime/glitch"
)

// The styles that the widgets which don't take a style argument draw with
type Theme struct {
	Button   Style
	Panel    Style
	DragSlot Style
	DragItem Style

	ScrollbarTop    Style
	ScrollbarBottom Style
	ScrollbarHandle Style
	ScrollbarBg     Style

	CheckboxTrue  Style
	CheckboxFalse Style

	TextInputPanel Style
	TextCursor     Style
	Tooltip        Style
}

var gStyle Theme

// Swaps the theme that widgets draw with. Takes effect on the next widget drawn
func SetTheme(theme Theme) {
	gStyle = theme
}

func CurrentTheme() Theme {
	return gStyle
}

//--------------------------------------------------------------------------------

type SpriteStyle struct {
	sprite Drawer
	color  glitch.RGBA
	rot    float64
}

func NewSpriteStyle(sprite Drawer, color glitch.RGBA) SpriteStyle {
	return SpriteStyle{
		sprite, color, 0.0,
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
func (s SpriteStyle) Rotation(rot float64) SpriteStyle {
	s.rot = rot
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
