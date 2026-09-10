package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"image/color"
)

const panelWidth = 420
const panelHeight = 650

const bgOverflow = 38

func WithMinWidth(obj fyne.CanvasObject, w float32) fyne.CanvasObject {
	sizer := canvas.NewRectangle(color.Transparent)
	sizer.SetMinSize(fyne.NewSize(w, panelHeight))

	bg := canvas.NewImageFromFile("assets/gauche.png")
	bg.FillMode = canvas.ImageFillStretch
	bg.Translucency = 0.04
	bg.Resize(fyne.NewSize(w+bgOverflow*2, panelHeight+bgOverflow*2))
	bg.Move(fyne.NewPos(-bgOverflow, -bgOverflow))
	bgHolder := container.NewWithoutLayout(bg)

	return container.NewStack(sizer, bgHolder, container.NewPadded(obj))
}

func BuildBackground(imagePath string) (fyne.CanvasObject, *canvas.Image) {

	bg := canvas.NewImageFromFile(imagePath)

	bg.FillMode = canvas.ImageFillStretch
	overlay := canvas.NewRectangle(color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 60,
	})
	bgStack := container.NewStack(bg, overlay)
	return bgStack, bg
}

func SetRoomBackground(bg *canvas.Image, backgroundPath string) {
	bg.File = backgroundPath
	bg.Refresh()
}
