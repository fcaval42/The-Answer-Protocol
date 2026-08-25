package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/container"
)


const mapCellSizeH = 150
const mapCellSizeV = 120
const rows, cols = 3, 3

var mapCellColor = color.NRGBA{R: 255, G: 255, B: 225, A: 20} // couleur blanche de base
var mapCurrentCellColor = color.NRGBA{R: 232, G: 216, B: 176, A: 90} // beige comme le thème de base


func mapCell(current bool) fyne.CanvasObject {
	fill := mapCellColor
	if current {
		fill = mapCurrentCellColor
	}

	frame := canvas.NewRectangle(fill)
	frame.StrokeColor = TitleColor // couleur du contour (bordure du rectangle)
	frame.StrokeWidth = 1 // épaisseur de la ligne

	if current {
		frame.StrokeWidth = 2
	}

	frame.SetMinSize(fyne.NewSize(mapCellSizeH, mapCellSizeV))

	label := NewBodyText("", 14)

	return container.NewVBox(
		container.New(layout.NewGridWrapLayout(fyne.NewSize(mapCellSizeH, mapCellSizeV)), frame),
		container.NewCenter(label),
	)
}


// on construit la carte du monde 
func BuildMapPanel(currentRow, currentCol int) fyne.CanvasObject {
	rowsCols := []int{3, 4, 2, 2}

	rows := container.NewVBox()
	for r, cols := range rowsCols{
		row := container.NewHBox()
		for c := 0; c < cols; c++ {
			if c > 0 { // ajouter espace que si pas première case
				row.Add(Fix(70, 0))
			}
			row.Add(mapCell(r == currentRow && c == currentCol))
		}
		if r > 0 {
			rows.Add(Fix(0, 15))
		}
		rows.Add(row)
	}

	return container.New(layout.NewCustomPaddedLayout(0, 0, 70, 0), rows)
}
