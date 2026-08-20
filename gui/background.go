package gui

import (
	"image/color"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// je met un rectangle en fond pour les deux colonnes gauche et droite
func WithMinWidth(obj fyne.CanvasObject, w float32) fyne.CanvasObject {
	sizer := canvas.NewRectangle(color.Black)
	sizer.SetMinSize(fyne.NewSize(w, 0))

	//panelBg := canvas.NewRectangle(color.NRGBA{R: 0, G:0, B: 0, A: 160})

	return container.NewStack(sizer, container.NewPadded(obj))
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
		A: 125, // 50% de transparent pour l'image background
	})
	bgStack := container.NewStack(bg, overlay)
	return bgStack, bg
}


func SetRoomBackground(bg *canvas.Image, backgroundPath string) {
	bg.File = backgroundPath
	bg.Refresh()
}