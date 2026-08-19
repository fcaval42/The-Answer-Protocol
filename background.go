package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func BuildBackground(imagePath string) (fyne.CanvasObject, *canvas.Image) {

	// on charge l'image
	bg := canvas.NewImageFromFile(imagePath)

	// FillMode = détermine comment une image doit être redimensionnée ou
	// ajustée pour remplir un espace donnée (ex: un conteneur)
	// ImageFillStretch = valeur possible pour FillMode. Signifie que l'image
	// est étirée pour remplir toute la zone dispo.
	// Donc ici on force l'image à s'étirer pour remplir tout l'espace alloué
	bg.FillMode = canvas.ImageFillStretch
	return bg, bg
}


func setRoomBackground(bg *canvas.Image, room Room) {
	bg.File = room.background
	bg.Refresh()
}