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

// -------------------- ESPACE FIXE -----------------

// Crée un bloc invisible pour bloquer un espace précis
func Fix(largeur float32, hauteur float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	// applique taille fixe
	return container.New(layout.NewGridWrapLayout(fyne.NewSize(largeur, hauteur)), rect)
}

// -------------------- POLICES ET COULEURS -----------------

// couleur beige, utilisée pour les titres
var TitleColor = color.NRGBA{R: 0xe8, G: 0xd8, B: 0xb0, A: 0xff}

// blanc, utilisé pour le texte courant (corps de texte + boutons)
var BodyColor = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

// rouge, utilisé pour la barre de vie
var HPColor = color.NRGBA{R: 125, G: 84, B: 7, A: 255}

// taille du texte des boutons
var ButtonTextSize float32 = 16
var DescSize float32 = 19

// loadFont charge une police depuis assets/, ou nil si le fichier est
// introuvable (le texte retombe alors sur la police par défaut du thème).
func loadFont(path string) fyne.Resource {
	res, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		return nil
	}
	return res
}

var (
	titleFont = loadFont("assets/Cinzel-VariableFont_wght.ttf") // MERVEILLEUSE police des titres
	bodyFont  = loadFont("assets/IMFellDoublePica-Regular.ttf") // police du texte courant
)

// -------------------- TEXTES STATIQUES (canvas.Text) -----------------

func newText(text string, size float32, col color.Color, font fyne.Resource, style fyne.TextStyle) *canvas.Text {
	t := canvas.NewText(text, col)
	t.TextSize = size
	t.TextStyle = style
	t.FontSource = font
	return t
}

// NewTitleText crée un texte de titre : police custom + couleur beige.
func NewTitleText(text string, size float32) *canvas.Text {
	return newText(text, size, TitleColor, titleFont, fyne.TextStyle{Bold: true})
}

// NewBodyText crée un texte basique : police + couleur du texte blanc.
func NewBodyText(text string, size float32) *canvas.Text {
	return newText(text, size, BodyColor, bodyFont, fyne.TextStyle{})
}

// -------------------- WIDGETS THÉMÉS (boutons, texte avec retour à la ligne) --

type fontTheme struct {
	fyne.Theme
	font  fyne.Resource
	color color.Color
	size  float32
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
	return t.Theme.Color(name, variant)
}

// -------------------------- BARRE DE HP -----------------------------

// primaryColorTheme change uniquement la couleur "primary" du thème, celle
// que widget.ProgressBar utilise pour dessiner sa barre remplie.
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

// NewColoredProgressBar crée une widget.ProgressBar dont la barre remplie
// utilise la couleur donnée (au lieu du bleu par défaut du thème). Renvoie
// la barre elle-même (pour régler Min/Max/Value) et l'objet thémé à
// afficher (à placer dans le conteneur final).
func NewColoredProgressBar(col color.Color) (*widget.ProgressBar, fyne.CanvasObject) {
	bar := widget.NewProgressBar()
	themed := container.NewThemeOverride(bar, &primaryColorTheme{Theme: theme.DefaultTheme(), color: col})
	return bar, themed
}

// ------------------- BOUTON CLIQUABLE ENORME(MAP + CHAT) ---------------------

// iconSizeTheme force la taille des icônes de bouton (normalement fixée
// par le thème) à une valeur donnée, pour faire des icônes plus grandes.
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

// NewBigIconButton crée un bouton composé uniquement d'une icône (chargée
// depuis un fichier), affichée à la taille donnée, cliquable via tapped.
func NewBigIconButton(path string, size float32, tapped func()) fyne.CanvasObject {
	res, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		res = theme.BrokenImageIcon()
	}
	btn := widget.NewButtonWithIcon("", res, tapped)
	btn.Importance = widget.LowImportance // pas de fond, juste l'icône
	return container.NewThemeOverride(btn, &iconSizeTheme{Theme: theme.DefaultTheme(), size: size})
}

// ------------------------ BOUTON AVEC THÈME BASE-----------------------------

func (t *fontTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText && t.size > 0 {
		return t.size
	}
	return t.Theme.Size(name)
}

// NewButton crée un bouton avec la police régulière et un texte blanc.
func NewButton(label string, tapped func()) fyne.CanvasObject {
	btn := widget.NewButton(label, tapped)
	return container.NewThemeOverride(btn, newFontTheme(bodyFont, BodyColor, ButtonTextSize))
}

// ------------------------- THEME DU TEXTE DE BASE ----------------------------

// NewWrappingBodyText crée un texte qui retourne à la ligne automatiquement
// (contrairement à NewBodyText, basé sur un widget.Label), avec la police
// régulière et la couleur du texte courant.
func NewWrappingBodyText(text string) fyne.CanvasObject {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapWord
	l.Alignment = fyne.TextAlignCenter
	return container.NewThemeOverride(l, newFontTheme(bodyFont, BodyColor, DescSize))
}
