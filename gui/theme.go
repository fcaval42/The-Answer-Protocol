package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func Fix(largeur float32, hauteur float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	return container.New(layout.NewGridWrapLayout(fyne.NewSize(largeur, hauteur)), rect)
}

var TitleColor = color.NRGBA{R: 0xe8, G: 0xd8, B: 0xb0, A: 0xff}
var BodyColor = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
var HPColor = color.NRGBA{R: 125, G: 84, B: 7, A: 255}
var ButtonTextSize float32 = 16
var DescSize float32 = 19

func loadFont(path string) fyne.Resource {
	res, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		return nil
	}
	return res
}

var (
	titleFont = loadFont("assets/fonts/Cinzel-VariableFont_wght.ttf")
	bodyFont  = loadFont("assets/fonts/IMFellDoublePica-Regular.ttf")
)

func newText(text string, size float32, col color.Color, font fyne.Resource, style fyne.TextStyle) *canvas.Text {
	t := canvas.NewText(text, col)
	t.TextSize = size
	t.TextStyle = style
	t.FontSource = font
	return t
}

func NewTitleText(text string, size float32) *canvas.Text {
	return newText(text, size, TitleColor, titleFont, fyne.TextStyle{Bold: true})
}

func NewBodyText(text string, size float32) *canvas.Text {
	return newText(text, size, BodyColor, bodyFont, fyne.TextStyle{})
}

type fontTheme struct {
	fyne.Theme
	font    fyne.Resource
	color   color.Color
	size    float32
	bgAlpha uint8
}

func newFontTheme(font fyne.Resource, col color.Color, size float32) fyne.Theme {
	return &fontTheme{Theme: theme.DefaultTheme(), font: font, color: col, size: size}
}

func (t *fontTheme) Font(style fyne.TextStyle) fyne.Resource {
	if t.font != nil {
		return t.font
	}
	return t.Theme.Font(style)
}

func (t *fontTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameForeground {
		return t.color
	}
	if t.bgAlpha > 0 && (name == theme.ColorNameButton || name == theme.ColorNameDisabledButton) {
		base := t.Theme.Color(name, variant)
		r, g, b, _ := base.RGBA()
		return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: t.bgAlpha}
	}
	return t.Theme.Color(name, variant)
}

func (t *fontTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText && t.size > 0 {
		return t.size
	}
	return t.Theme.Size(name)
}

type primaryColorTheme struct {
	fyne.Theme
	color color.Color
}

func (t *primaryColorTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNamePrimary {
		return t.color
	}
	return t.Theme.Color(name, variant)
}

func NewColoredProgressBar(col color.Color) (*widget.ProgressBar, fyne.CanvasObject) {
	bar := widget.NewProgressBar()
	themed := container.NewThemeOverride(bar, &primaryColorTheme{Theme: theme.DefaultTheme(), color: col})
	return bar, themed
}

type iconSizeTheme struct {
	fyne.Theme
	size float32
}

func (t *iconSizeTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameInlineIcon {
		return t.size
	}
	return t.Theme.Size(name)
}

func NewBigIconButton(path string, size float32, tapped func()) fyne.CanvasObject {
	res, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		res = theme.BrokenImageIcon()
	}
	btn := widget.NewButtonWithIcon("", res, tapped)
	btn.Importance = widget.LowImportance
	return container.NewThemeOverride(btn, &iconSizeTheme{Theme: theme.DefaultTheme(), size: size})
}

func NewButton(label string, tapped func()) fyne.CanvasObject {
	btn := widget.NewButton(label, tapped)
	return container.NewThemeOverride(btn, newFontTheme(bodyFont, BodyColor, ButtonTextSize))
}

func NewSemiTransparentButton(label string, alpha uint8, tapped func()) fyne.CanvasObject {
	btn := widget.NewButton(label, tapped)
	th := &fontTheme{Theme: theme.DefaultTheme(), font: bodyFont, color: BodyColor, size: ButtonTextSize, bgAlpha: alpha}
	return container.NewThemeOverride(btn, th)
}

func NewWrappingBodyText(text string) fyne.CanvasObject {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapWord
	l.Alignment = fyne.TextAlignCenter
	return container.NewThemeOverride(l, newFontTheme(bodyFont, BodyColor, DescSize))
}

type CustomTheme struct {
	fyne.Theme
}

func NewCustomTheme() *CustomTheme {
	return &CustomTheme{Theme: theme.DarkTheme()}
}

func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x2b, G: 0x1e, B: 0x0c, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xf5, G: 0xec, B: 0xd8, A: 0xff}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0xe6, G: 0xc8, B: 0x6e, A: 0xff}
	case theme.ColorNameButton:
		return HPColor
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0x5c, G: 0x3d, B: 0x05, A: 0xff}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x1a, G: 0x12, B: 0x08, A: 0xcc}
	default:
		return t.Theme.Color(name, variant)
	}
}
