package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

// On charge la photo de profil de l'utilisateur en fonction du perso choisi
func NewAvatar(path string, size float32) fyne.CanvasObject {
	img := canvas.NewImageFromFile(path)
	img.FillMode = canvas.ImageFillContain

	return container.New(layout.NewGridWrapLayout(fyne.NewSize(size, size)), img)
}