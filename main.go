package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ---- Données factices, en attendant le vrai serveur ----

type Room struct {
	name        string
	description string
	exits       []string
	items       []string
	npcs        []string
}

type Player struct {
	name      string
	hp        float64 // 0-100
	inventory []string
}

var fakeRoom = Room{
	name:        "Village Gestral",
	description: "Une odeur de sueur et de métal flotte entre les huttes.",
	exits:       []string{"Océan suspendu", "Cimetière gestral"},
	items:       []string{"Chroma"},
	npcs:        []string{"Renoir"},
}

var fakePlayer = Player{
	name:      "Gustave",
	hp:        100,
	inventory: []string{"Lumina", "Pictos"},
}


// ----------------- Panneau de salle (gauche) ------------------

// r Room = structure Room qui contient données de la salle
func buildRoomPanel(r Room) fyne.CanvasObject{
	// sortie fyne.CanvasObject = élément graphique (conteneur, widget...)

	// ------------------ TITRE ------------------
	title := widget.NewLabelWithStyle(
		r.name,
		fyne.TextAlignLeading,  // alignement à gauche
		fyne.TextStyle{Bold: true},  // texte en gras
	)
	desc := widget.NewLabel(r.description) // label pour la description
	desc.Wrapping = fyne.TextWrapWord // active retour à la ligne auto (texte
	// s'adapte à la largeur du conteneur)

	// ------------------ SORTIE -----------------
	exitsRow := container.NewHBox() // conteneur horizontal qui contiendra les
	// boutons des sorties
	for _, e := range r.exits {
		exit := e // on crée une nvlle variable locale à chaque tour.
		exitsRow.Add(widget.NewButton(exit, func() { // acxtion du bouton
			fmt.Println("MOVE TO: ", exit) // A REMPLACER par ex sendCommand
		}))
	}

	// ------------------- ITEMS -----------------
	itemsRow := container.NewHBox()
	for _, i := range r.items {
		item := i // un bouton par objet
		itemsRow.Add(widget.NewButton(item, func() {
			fmt.Println("TAKE: ", item) // A REMPLACER
		}))
	}

	// ------------------- PNJ --------------------
	ncpsRow := container.NewHBox()
	for _, n := range r.npcs {
		ncp := n
		ncpsRow.Add(widget.NewButton((npc + "TALK", func() {
			fmt.Println("TALK: ")
		})))
	}


}



func main() {
	a := app.New()
	w := a.NewWindow("The Answer Protocol")

	roomPanel := buildRoomPanel(fakeRoom)

	w.ShowAndRun()
}