package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

var avatarFiles = map[string]string{
	"Gustave": "Gustave.png",
	"Lune":    "Lune.png",
	"Maëlle":  "Maelle.png",
	"Sciel":   "Sciel.png",
	"Verso":   "Verso.png",
}

func avatarPath(playerName string) string {
	if path, name := avatarFiles[playerName]; name {
		return "assets/players/" + path
	}
	return "assets/players/Gustave.png"
}

func NewAvatar(path string, size float32) fyne.CanvasObject {
	img := canvas.NewImageFromFile(path)
	img.FillMode = canvas.ImageFillContain

	return container.New(layout.NewGridWrapLayout(fyne.NewSize(size, size)), img)
}
