package game

import (
	"graphics.gd/classdb/Button"
	"graphics.gd/classdb/ColorRect"
	"graphics.gd/classdb/Control"
	"graphics.gd/classdb/GridContainer"
	"graphics.gd/classdb/HBoxContainer"
	"graphics.gd/classdb/Label"
	"graphics.gd/classdb/MarginContainer"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/PanelContainer"
	"graphics.gd/classdb/Polygon2D"
	"graphics.gd/classdb/ProgressBar"
	"graphics.gd/classdb/ScrollContainer"
	"graphics.gd/classdb/StyleBoxFlat"
	"graphics.gd/classdb/TextServer"
	"graphics.gd/classdb/Theme"
	"graphics.gd/classdb/VBoxContainer"
	"graphics.gd/variant/Color"
	"graphics.gd/variant/Float"
	"graphics.gd/variant/Vector2"
)

// The palette. Placeholder art is shapes, text and these colours.
var (
	colBack    = rgb(0x12, 0x10, 0x16)
	colPanel   = rgb(0x1e, 0x1b, 0x26)
	colPanelHi = rgb(0x2a, 0x26, 0x35)
	colLine    = rgb(0x3a, 0x35, 0x48)
	colText    = rgb(0xd8, 0xd2, 0xc4)
	colDim     = rgb(0x8c, 0x86, 0x7a)
	colBlood   = rgb(0x9c, 0x24, 0x2e)
	colGold    = rgb(0xc9, 0xa2, 0x27)
	colGood    = rgb(0x5a, 0x96, 0x66)
	colWarn    = rgb(0xc7, 0x7b, 0x2b)
	colDream   = rgb(0x7a, 0x5c, 0xb8)
	colHunger  = rgb(0xcf, 0xbd, 0x9f)
	colRot     = rgb(0x6f, 0x8f, 0x4a)
)

func rgb(r, g, b uint8) Color.RGBA {
	return Color.RGBA{R: Float.X(r) / 255, G: Float.X(g) / 255, B: Float.X(b) / 255, A: 1}
}

func flat(bg, border Color.RGBA, width, margin int) StyleBoxFlat.Instance {
	sb := StyleBoxFlat.New()
	sb.SetBgColor(bg).SetBorderColor(border)
	sb.SetBorderWidthAll(width)
	sb.SetCornerRadiusAll(3)
	sb.AsStyleBox().SetContentMarginAll(Float.X(margin))
	return sb
}

// newTheme builds the one Theme the whole game uses.
func newTheme() Theme.Instance {
	t := Theme.New()
	t.SetDefaultFontSize(15)
	t.SetColor("font_color", "Label", colText)
	t.SetStylebox("panel", "PanelContainer", flat(colPanel, colLine, 1, 10).AsStyleBox())
	t.SetStylebox("normal", "Button", flat(colPanelHi, colLine, 1, 6).AsStyleBox())
	t.SetStylebox("hover", "Button", flat(colLine, colGold, 1, 6).AsStyleBox())
	t.SetStylebox("pressed", "Button", flat(colBlood, colGold, 1, 6).AsStyleBox())
	t.SetStylebox("disabled", "Button", flat(colPanel, colPanelHi, 1, 6).AsStyleBox())
	t.SetStylebox("focus", "Button", flat(Color.RGBA{}, Color.RGBA{}, 0, 6).AsStyleBox())
	t.SetColor("font_color", "Button", colText)
	t.SetColor("font_hover_color", "Button", colGold)
	t.SetColor("font_pressed_color", "Button", colText)
	t.SetColor("font_disabled_color", "Button", colDim)
	t.SetStylebox("background", "ProgressBar", flat(colBack, colLine, 1, 0).AsStyleBox())
	t.SetStylebox("fill", "ProgressBar", flat(colGood, colGood, 0, 0).AsStyleBox())
	t.SetStylebox("panel", "TooltipPanel", flat(colBack, colGold, 1, 8).AsStyleBox())
	t.SetColor("font_color", "TooltipLabel", colText)
	return t
}

// node is anything that can go in the tree.
type node = Node.Any

func add(parent node, kids ...node) {
	for _, k := range kids {
		parent.AsNode().AddChild(k.AsNode())
	}
}

func text(s string, size int, col Color.RGBA) Label.Instance {
	l := Label.New()
	l.SetText(s)
	l.AsControl().AddThemeFontSizeOverride("font_size", size)
	l.AsControl().AddThemeColorOverride("font_color", col)
	return l
}

// para is text that wraps within a width.
func para(s string, size int, col Color.RGBA, width int) Label.Instance {
	l := text(s, size, col)
	l.SetAutowrapMode(TextServer.AutowrapWordSmart)
	l.AsControl().SetCustomMinimumSize(Vector2.New(width, 0))
	return l
}

// tip gives a control hover text. Tooltips for generated things come straight
// from their Cause.
func tip[T interface{ AsControl() Control.Instance }](c T, s string) T {
	c.AsControl().SetTooltipText(s)
	c.AsControl().SetMouseFilter(Control.MouseFilterPass)
	return c
}

func vbox(sep int, kids ...node) VBoxContainer.Instance {
	b := VBoxContainer.New()
	b.AsControl().AddThemeConstantOverride("separation", sep)
	add(b, kids...)
	return b
}

func hbox(sep int, kids ...node) HBoxContainer.Instance {
	b := HBoxContainer.New()
	b.AsControl().AddThemeConstantOverride("separation", sep)
	add(b, kids...)
	return b
}

func grid(cols, sep int, kids ...node) GridContainer.Instance {
	g := GridContainer.New()
	g.SetColumns(cols)
	g.AsControl().AddThemeConstantOverride("h_separation", sep)
	g.AsControl().AddThemeConstantOverride("v_separation", sep)
	add(g, kids...)
	return g
}

func panel(kids ...node) PanelContainer.Instance {
	p := PanelContainer.New()
	add(p, kids...)
	return p
}

// framed is a panel with a coloured border, for the thing that matters now.
func framed(border Color.RGBA, kids ...node) PanelContainer.Instance {
	p := panel(kids...)
	p.AsControl().AddThemeStyleboxOverride("panel", flat(colPanel, border, 2, 10).AsStyleBox())
	return p
}

func margin(px int, kid node) MarginContainer.Instance {
	m := MarginContainer.New()
	for _, side := range []string{"margin_left", "margin_right", "margin_top", "margin_bottom"} {
		m.AsControl().AddThemeConstantOverride(side, px)
	}
	add(m, kid)
	return m
}

func scroll(w, h int, kid node) ScrollContainer.Instance {
	s := ScrollContainer.New()
	s.SetHorizontalScrollMode(ScrollContainer.ScrollModeDisabled)
	s.AsControl().SetCustomMinimumSize(Vector2.New(w, h))
	add(s, kid)
	return s
}

func sized[T interface{ AsControl() Control.Instance }](c T, w, h int) T {
	c.AsControl().SetCustomMinimumSize(Vector2.New(w, h))
	return c
}

func grow[T interface{ AsControl() Control.Instance }](c T) T {
	c.AsControl().SetSizeFlagsHorizontal(Control.SizeExpandFill)
	return c
}

func gap(w, h int) Control.Instance {
	c := Control.New()
	c.SetCustomMinimumSize(Vector2.New(w, h))
	return c
}

func spacer() Control.Instance { return grow(gap(0, 0)) }

// meter is a bar from 0 to max, in a colour.
func meter(value, top int, col Color.RGBA, w, h int) ProgressBar.Instance {
	p := ProgressBar.New()
	p.SetShowPercentage(false)
	p.AsRange().SetMaxValue(Float.X(max(top, 1)))
	p.AsRange().SetValue(Float.X(min(max(value, 0), top)))
	p.AsControl().SetCustomMinimumSize(Vector2.New(w, h))
	p.AsControl().AddThemeStyleboxOverride("fill", flat(col, col, 0, 0).AsStyleBox())
	p.AsControl().SetMouseFilter(Control.MouseFilterPass)
	return p
}

// shape is a Polygon2D inside a control of its own size, so containers can
// lay it out. Points are in a w by h box.
func shape(w, h int, col Color.RGBA, pts ...Vector2.XY) Control.Instance {
	c := gap(w, h)
	c.SetMouseFilter(Control.MouseFilterIgnore)
	p := Polygon2D.New()
	p.SetPolygon(pts)
	p.SetColor(col)
	add(c, p)
	return c
}

func rect(w, h int, col Color.RGBA) ColorRect.Instance {
	r := ColorRect.New()
	r.SetColor(col)
	r.AsControl().SetCustomMinimumSize(Vector2.New(w, h))
	r.AsControl().SetMouseFilter(Control.MouseFilterIgnore)
	return r
}

func pt(x, y float64) Vector2.XY { return Vector2.XY{X: Float.X(x), Y: Float.X(y)} }

// rawButton is a button with no behaviour yet; App.button wires it.
func rawButton(label string, enabled bool) Button.Instance {
	b := Button.New()
	b.SetText(label)
	b.AsBaseButton().SetDisabled(!enabled)
	return b
}
