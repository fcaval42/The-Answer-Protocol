package gui

import (
	"image/color"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// withMinWidth force un conteneur à respecter une largeur minimale, et lui
// ajoute un fond semi-transparent (pour rester lisible par-dessus l'image
// de fond de la salle) avec un peu de marge interne autour du contenu.
const panelWidth = 420
const panelHeight = 650

// bgOverflow = de combien de pixels l'image de fond déborde de chaque côté
// par rapport à la zone réelle du panneau (qui, elle, ne change pas).
const bgOverflow = 38

// je met un rectangle en fond pour les deux colonnes gauche et droite
func WithMinWidth(obj fyne.CanvasObject, w float32) fyne.CanvasObject {
	sizer := canvas.NewRectangle(color.Transparent)
	sizer.SetMinSize(fyne.NewSize(w, panelHeight))

	bg := canvas.NewImageFromFile("assets/gauche.png")
	bg.FillMode = canvas.ImageFillStretch
	bg.Translucency = 0.04 // 0 = opaque, 1 = invisible
	// taille et position fixées à la main, plus grandes que le panneau : le
	// fond déborde visuellement sans agrandir la zone occupée par le contenu.
	bg.Resize(fyne.NewSize(w+bgOverflow*2, panelHeight+bgOverflow*2))
	bg.Move(fyne.NewPos(-bgOverflow, -bgOverflow))
	// conteneur sans layout : il ignore les Resize que lui envoie le Stack
	// parent, donc la taille/position manuelles de bg ci-dessus sont conservées.
	bgHolder := container.NewWithoutLayout(bg)

	return container.NewStack(sizer, bgHolder, container.NewPadded(obj))
}


func BuildBackground(imagePath string) (fyne.CanvasObject, *canvas.Image) {

	// on charge l'image
	bg := canvas.NewImageFromFile(imagePath)

	// FillMode = détermine comment une image doit être redimensionnée ou
	// ajustée pour remplir un espace donnée (ex: un conteneur)
	// ImageFillStretch = valeur possible pour FillMode. Signifie que l'image
	// est étirée pour remplir toute la zone dispo.
	// Donc ici on force l'image à s'étirer pour remplir tout l'espace alloué
	bg.FillMode = canvas.ImageFillStretch
	overlay := canvas.NewRectangle(color.RGBA {
		R: 0,
		G: 0,
		B: 0,
		A: 60, // % de transparent pour l'image background
	})
	bgStack := container.NewStack(bg, overlay)
	return bgStack, bg
}


func SetRoomBackground(bg *canvas.Image, backgroundPath string) {
	bg.File = backgroundPath
	bg.Refresh()
}