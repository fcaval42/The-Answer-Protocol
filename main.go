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

func buildRoomPanel(r Room) fyne.CanvasObject{
	// r Room = structure Room qui contient données de la salle
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
	// NewHBox = crée nv conteneur + objets seront placés de gauche à droite et
	// tjrs affichés à leur taille min horizontale.
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
	npcsRow := container.NewHBox()
	for _, n := range r.npcs {
		npc := n
		npcsRow.Add(widget.NewButton(npc + "TALK", func() {
			fmt.Println("TALK: ", npc) // A REMPLACER
		}))
		npcsRow.Add(widget.NewButton(npc + "ATTACK", func() {
			fmt.Println("ATTACK:", npc) // A REMPLACER
		}))
	}

	// ----------------- ACTIONS --------------------
	actionsGrid := container.NewGridWithColumns(2, 
		widget.NewButton("Status", func() { fmt.Println("STATUS") }), // A REMPLACER
		widget.NewButton("Quest", func() { fmt.Println("QUEST") }), // A REMPLACER
		widget.NewButton("Status", func() { fmt.Println("STATUS") }), // A REMPLACER
		widget.NewButton("Quit", func() { fmt.Println("QUIT") }), // A REMPLACER
		)

	// conteneur vertical. Respecte ordre d'affichage
	return container.NewVBox(
		title,
		desc,
		widget.NewSeparator(),      // ligne de séparation
		widget.NewLabel("Sortie"),
		exitsRow,
		widget.NewLabel("Objets"),
		itemsRow,
		widget.NewLabel("PNJ"),
		npcsRow,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading,
	fyne.TextStyle{Bold: true}),
		actionsGrid,
	)
}

// --- Panneau personnage : PV + inventaire (droite, haute) -----

func buildCharacterPanel(p Player) fyne.CanvasObject {
	name := widget.NewLabelWithStyle(p.name, fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true})

	// widget progress Bar = widget déjà dispo de base youpi
	hpBar := widget.NewProgressBar()
	hpBar.Min = 0
	hpBar.Max = 100
	hpBar.Value = p.hp
	hpLabel := widget.NewLabel(fmt.Sprintf("PV %0f/100", p.hp))

	invLabel := widget.NewLabelWithStyle("Inventaire", fyne.TextAlignLeading,
		fyne.TextStyle{Bold : true})
	invRow := container.NewHBox()
	for _, i := range p.inventory {
		item := i
		invRow.Add(widget.NewButton(item, func() {
			fmt.Println("LÂCHER: ", item)
		}))
	}

	return container.NewVBox(name, hpLabel, hpBar, invLabel, invRow)
}


// ---------------- Bouton carte (centre) ------------

func buildMapButton() fyne.CanvasObject {
	btn := widget.NewButton("Voir Map", func() {
		fmt.Println("Ouvrir carte")  // FAIRE une vue carte du monde
	})

	return container.NewCenter(btn)
}



func main() {
	a := app.New()
	w := a.NewWindow("The Answer Protocol")

	roomPanel := buildRoomPanel(fakeRoom)
	mapButton := buildMapButton()
	rightPanel := container.NewVBox(
		buildCharacterPanel(fakePlayer),
	)

	w.SetContent(container.NewGridWithColumns(3, roomPanel, mapButton, rightPanel))
	w.Resize(fyne.NewSize(700, 500))
	w.ShowAndRun()
}
