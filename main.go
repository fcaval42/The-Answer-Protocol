package main

import (
	"fmt"
	"os"
	"image/color" // image = permet de toucher aux images en Go
	// color = fournit couleurs/modèles prêts à l'emploi.

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"the_answer_protocol/parsing"
	"the_answer_protocol/gui"
)

// ---- Données factices, en attendant le vrai serveur ----

type Room struct {
	name        string
	description string
	exits       []string
	items       []string
	npcs        []string
	background  string // chemin vers le gif de fond propre à cette salle
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
	background:  "assets/village_gestral.png",
}

// deuxième salle factice, juste pour tester visuellement le changement de
// fond avant que le vrai MOVE réseau soit branché au serveur
var fakeRoom2 = Room{
	name:        "Cimetière Gestral",
	description: "De minuscules tombes, dont certaines sont vides.",
	exits:       []string{"Village Gestral"},
	background:  "assets/cimetiere_gestral.png",
}

var fakePlayer = Player{
	name:      "Gustave",
	hp:        100,
	inventory: []string{"Lumina", "Pictos"},
}


// je met un rectangle en fond pour les deux colonnes gauche et droite
func withMinWidth(obj fyne.CanvasObject, w float32) fyne.CanvasObject {
	sizer := canvas.NewRectangle(color.Transparent)
	sizer.SetMinSize(fyne.NewSize(w, 0))
	return container.NewStack(sizer, obj)
}

// ----------------- Header (titre + infos salles) ------------------

func buildHeader() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"The Answer Protocol",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	counters := widget.NewLabel("2 dans la salle - 7 en ligne")

	return container.NewBorder(nil, nil, title, counters, nil)
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
	hpLabel := widget.NewLabel(fmt.Sprintf("PV %g/100", p.hp))

	invLabel := widget.NewLabelWithStyle("Inventaire", fyne.TextAlignLeading,
		fyne.TextStyle{Bold : true})
	invRow := container.NewHBox()
	for _, i := range p.inventory {
		item := i
		invRow.Add(widget.NewButton(item, func() {
			fmt.Println("DROP: ", item)
		}))
	}

	return container.NewVBox(name, hpLabel, hpBar, invLabel, invRow)
}


// ---------------- Bouton carte (centre) ------------

func buildMapButton() fyne.CanvasObject {
	btn := widget.NewButton("Voir Map", func() {
		fmt.Println("MAP")  // FAIRE une vue carte du monde
	})

	return container.NewCenter(btn)
}


// ---- Chat (droite, bas) : onglets Global/Salle/Groupe séparés des logs ----

func buildChatPanel() fyne.CanvasObject {
	// fonction dans une fonction -> Makescope = fonction interne. Définie à
	// l'intérieur de chat pour créer un conteneur réutilisable avec un sample
	// différent pour chaque objet
	makeScope := func(sample string) fyne.CanvasObject {
		messages := widget.NewLabel(sample)
		messages.Wrapping = fyne.TextWrapWord
		input := widget.NewEntry()
		input.SetPlaceHolder("Écrire un message...")
		send := widget.NewButton("Envoyer", func() {
			println("CHAT:", input.Text)
			input.SetText("")
		})
		return container.NewBorder(nil, container.NewBorder(nil, nil, nil, send, input), nil, nil, messages)
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Global", makeScope("Alice: ...")),
		container.NewTabItem("Salle", makeScope("Venus: Parle mieux on comprend rien")),
		container.NewTabItem("Groupe", makeScope("")),
	)

	return container.NewBorder(widget.NewLabelWithStyle("Chat",
	fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, tabs)
}


func main() {

	world_path := "yaml/world.yaml"
	world, err := parsing.Rally(world_path)

	if err != nil {
		os.Exit(1)
	}
	fmt.Println(world.Enemies)


	a := app.New()
	w := a.NewWindow("The Answer Protocol")
	w.Resize(fyne.NewSize(1300, 800))
	w.SetFixedSize(true) // empêche utilisateur de redimensionner fenêtre

	// petit annuaire de démo : nom de sortie -> salle cible.
	// Plus tard, ça n'existera plus : le serveur renverra directement
	// la nouvelle Room en réponse à MOVE, via le protocole réseau.
	demoRooms := map[string]Room{
		"Village Gestral":   fakeRoom,
		"Cimetière gestral": fakeRoom2,
		"Cimetière Gestral": fakeRoom2,
	}

	background, bgImage := gui.BuildBackground(fakeRoom.background)

	// garde une référence stable qu'on peut vider/remplir pour ne pas
	// reconstruire toute la fenêtre à chaque changement de salle.
	roomContainer := container.NewStack()

	changeRoom := func(newRoom Room) {
		gui.SetRoomBackground(bgImage, newRoom.background)
		roomContainer.Objects = []fyne.CanvasObject{buildRoomPanel(newRoom)} // VOIR
		roomContainer.Refresh()
	}
	changeRoom(fakeRoom) // affichage initial
	_ = demoRooms // gardé pour un futur test manuel : changeRoom(demoRooms["cimetière gestral"])


	//roomPanel := withMinWidth(buildRoomPanel(fakeRoom), 420)
	mapButton := buildMapButton()
	rightPanel := withMinWidth(container.NewVBox(
		buildCharacterPanel(fakePlayer),
		widget.NewSeparator(),
		buildChatPanel(),
	), 420)

	body := container.NewBorder(nil, nil, withMinWidth(roomContainer, 420), rightPanel, mapButton)

	// NewBorder = créer un conteneur avec des bordures personnalisées pour
	// organiser les éléments d'une interface (top, bottom, left, right, content)
	content := container.NewBorder(buildHeader(), nil, nil, nil, body)

	full := container.NewStack(background, container.NewPadded(content))

	w.SetContent(full)
	w.ShowAndRun()
}
