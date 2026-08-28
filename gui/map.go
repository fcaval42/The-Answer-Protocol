package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

// cache des images de la map pour ne les lire/décoder qu'une seule fois depuis le disque
var mapImageCache = map[string]fyne.Resource{}

func loadMapImage(path string) fyne.Resource {
	if res, ok := mapImageCache[path]; ok {
		return res
	}
	res, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		return nil
	}
	mapImageCache[path] = res
	return res
}

const mapCellSizeH = 220
const mapCellSizeV = 110
const rows, cols = 3, 3

var mapCellColor = color.NRGBA{R: 255, G: 255, B: 225, A: 20}        // couleur blanche de base
var mapCurrentCellColor = color.NRGBA{R: 232, G: 216, B: 176, A: 90} // beige comme le thème de base
var mapBackgroundPath = "assets/mapBackground.png"

type mapNode struct {
	row, col int
}

type mapPathfile struct {
	row, col int
}

var mapNames = map[mapNode]string{
	{0, 0}: "Lumière",
	{0, 1}: "",
	{0, 2}: "Vallons fleuries",
	{0, 3}: "Plage",
	{1, 0}: "",
	{1, 1}: "Nid d'Esquie",
	{1, 2}: "Océan Suspendu",
	{1, 3}: "Village Gestral",
	{1, 4}: "Cimetière Gestral",
	{2, 0}: "Vieille Lumière",
	{2, 1}: "",
	{2, 2}: "Terres oubliées",
	{2, 3}: "Falaise de Rochevague",
}

var mapPath = map[mapPathfile]string{
	{0, 0}: "assets/room/lumiere.png",
	{0, 1}: "",
	{0, 2}: "assets/room/vallons_fleuries.png",
	{0, 3}: "assets/room/plage.png",
	{1, 0}: "",
	{1, 1}: "assets/room/nid_esquie.png",
	{1, 2}: "assets/room/océan_suspendu.png",
	{1, 3}: "assets/room/village_gestral.png",
	{1, 4}: "assets/room/cimetiere_gestral.png",
	{2, 0}: "assets/room/vieille_lumiere.png",
	{2, 1}: "",
	{2, 2}: "assets/room/terres_oubliees.png",
	{2, 3}: "assets/room/falaise.png",
}


func mapCell(current bool, name string, img *canvas.Image) fyne.CanvasObject {
	if name == "" {
		// case invisible
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(100, 50))
		return spacer
	}

	fill := mapCellColor
	if current {
		fill = mapCurrentCellColor
	}

	background := canvas.NewRectangle(fill)
	background.SetMinSize(fyne.NewSize(mapCellSizeH, mapCellSizeV))

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = TitleColor // couleur du contour
	border.StrokeWidth = 1  // épaisseur de la ligne

	if current {
		border.StrokeColor = HPColor
		border.StrokeWidth = 4
	}

	label := NewBodyText(name, 14)

	return container.NewVBox(
		container.New(layout.NewGridWrapLayout(fyne.NewSize(mapCellSizeH, mapCellSizeV)), container.NewStack(background, img, border)),
		container.NewCenter(label),
	)
}

// on construit la carte du monde
func BuildMapPanel(currentRow, currentCol int) fyne.CanvasObject {
	rowsCols := []int{4, 5, 4}

	rows := container.NewVBox()
	for r, cols := range rowsCols {
		row := container.NewHBox()
		for c := 0; c < cols; c++ {
			if c > 0 { // ajouter espace que si pas première case
				row.Add(Fix(40, 0))
			}
			name := mapNames[mapNode{r, c}]
			var img *canvas.Image
			if name != "" {
				img = canvas.NewImageFromResource(loadMapImage(mapPath[mapPathfile{r, c}]))
				img.FillMode = canvas.ImageFillContain
			}
			row.Add(mapCell(r == currentRow && c == currentCol, name, img))
		}
		if r > 0 {
			rows.Add(Fix(0, 55))
		}
		rows.Add(row)
	}
	background := canvas.NewImageFromResource(loadMapImage(mapBackgroundPath))
	background.FillMode = canvas.ImageFillStretch

	return container.NewStack(background, container.New(layout.NewCustomPaddedLayout(0, 0, 20, 10), rows))
}
