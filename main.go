package main

import (
	"fmt"

	//"image/color" // image = permet de toucher aux images en Go
	//// color = fournit couleurs/modèles prêts à l'emploi.

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

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
	level     int
	xp        float64
	maxXP     float64
	hp        float64
	maxHP     float64
	speed     int
	attack    int
	critPct   int
	defense   int
	inventory []string
}

var fakeRoom = Room{
	name:        "Village Gestral",
	description: "Une odeur de sueur et de métal flotte entre les huttes.",
	exits:       []string{"Océan suspendu", "Cimetière gestral"},
	items:       []string{"Chroma"},
	npcs:        []string{"Renoir"},
	background:  "assets/room/vallons_fleuries.png",
}

// deuxième salle factice, juste pour tester visuellement le changement de
// fond avant que le vrai MOVE réseau soit branché au serveur
var fakeRoom2 = Room{
	name:        "Cimetière Gestral",
	description: "De minuscules tombes, dont certaines sont vides.",
	exits:       []string{"Village Gestral"},
	background:  "assets/room/cimetiere_gestral.png",
}

var fakePlayer = Player{
	name:      "Gustave",
	hp:        478,
	maxHP:     478,
	attack:    316,
	critPct:   24,
	defense:   33,
	inventory: []string{"Lumina", "Pictos"},
}


// ----------------- Header (titre + infos salles) ------------------

func buildHeader() fyne.CanvasObject {
	title := gui.NewTitleText("The Answer Protocol", 60)
	//counters := gui.NewBodyText("2 dans la salle - 7 en ligne", 14)

	return container.NewBorder(nil, gui.Fix(0, 40), title, nil, nil)
}


// ----------------- Panneau de salle (gauche) ------------------

func buildRoomPanel(r Room) fyne.CanvasObject{
	// r Room = structure Room qui contient données de la salle
	// sortie fyne.CanvasObject = élément graphique (conteneur, widget...)

	// ------------------ TITRE ------------------
	title := gui.NewTitleText(r.name, 40)
	desc := gui.NewWrappingBodyText(r.description) // description, retour à la ligne auto

	// ------------------ SORTIE -----------------
	exitsRow := container.NewHBox() // conteneur horizontal qui contiendra les
	// boutons des sorties
	// NewHBox = crée nv conteneur + objets seront placés de gauche à droite et
	// tjrs affichés à leur taille min horizontale.
	for _, e := range r.exits {
		exit := e // on crée une nvlle variable locale à chaque tour.
		exitsRow.Add(gui.NewButton(exit, func() { // acxtion du bouton
			fmt.Println("MOVE TO: ", exit) // A REMPLACER par ex sendCommand
		}))
	}

	// ------------------- ITEMS -----------------
	itemsRow := container.NewHBox()
	for _, i := range r.items {
		item := i // un bouton par objet
		itemsRow.Add(gui.NewButton(item, func() {
			fmt.Println("TAKE: ", item) // A REMPLACER
		}))
	}

	// ------------------- PNJ --------------------
	npcsRow := container.NewHBox()
	for _, n := range r.npcs {
		npc := n
		npcsRow.Add(gui.NewButton(npc + "TALK", func() {
			fmt.Println("TALK: ", npc) // A REMPLACER
		}))
		npcsRow.Add(gui.NewButton(npc + "ATTACK", func() {
			fmt.Println("ATTACK:", npc) // A REMPLACER
		}))
	}

	// ----------------- ACTIONS --------------------
	actionsGrid := container.NewGridWithColumns(2,
		gui.NewButton("Status", func() { fmt.Println("STATUS") }), // A REMPLACER
		gui.NewButton("Quest", func() { fmt.Println("QUEST") }), // A REMPLACER
		gui.NewButton("Status", func() { fmt.Println("STATUS") }), // A REMPLACER
		gui.NewButton("Quit", func() { fmt.Println("QUIT") }), // A REMPLACER
		)

	// conteneur vertical. Respecte ordre d'affichage
	return container.NewVBox(
		gui.Fix(0, 8),
		title,
		gui.Fix(0, 8),
		desc,
		gui.Fix(0, 10),
		gui.NewTitleText("Sortie", 25),
		gui.Fix(0, 8),
		exitsRow,
		gui.Fix(0, 10),
		gui.NewTitleText("Objets", 25),
		gui.Fix(0, 8),
		itemsRow,
		gui.Fix(0, 15),
		gui.NewTitleText("PNJ", 25),
		gui.Fix(0, 8),
		npcsRow,
		gui.Fix(0, 15),
		gui.NewTitleText("Actions", 25),
		gui.Fix(0, 8),
		actionsGrid,
	)
}

// --- Panneau personnage : PV + inventaire (droite, haute) -----

func statCell(label string, value string) fyne.CanvasObject {
	l := gui.NewBodyText(label, 18)
	v := gui.NewBodyText(value, 18)

	return container.NewVBox(l, v)
}


func buildCharacterPanel(p Player) fyne.CanvasObject {
	name := gui.NewTitleText(p.name, 70)
	name.TextSize = 48
	name.TextStyle = fyne.TextStyle{Bold: true}

	avatar := gui.NewAvatar("assets/players/Gustave.png", 100)
	nameRow := container.NewHBox(avatar, name)

	// widget progress Bar = widget déjà dispo de base youpi
	hpBar, hpBarThemed := gui.NewColoredProgressBar(gui.HPColor)
	hpBar.Min = 0
	hpBar.Max = p.maxHP
	hpBar.Value = p.hp
	hpBar.TextFormatter = func() string { return "" }
	hpText := gui.NewTitleText(fmt.Sprintf("%.0f / %.0f", p.hp, p.maxHP), 20)
	hpStack := container.NewStack(hpBarThemed, container.NewCenter(hpText))

	statsTitle := gui.NewTitleText("Statistiques de combat", 25)
	statsGrid := container.NewGridWithColumns(2,
		statCell("Puissance d'attaque", fmt.Sprintf("%d", p.attack)),
		statCell("Défense", fmt.Sprintf("%d", p.defense)),
		statCell("Chances de critique", fmt.Sprintf("%d%%", p.critPct)),
	)

	invLabel := gui.NewTitleText("Inventaire", 25)
	invRow := container.NewHBox()
	for _, i := range p.inventory {
		item := i
		invRow.Add(gui.NewButton(item, func() {
			fmt.Println("DROP: ", item)
		}))
	}

	return container.NewVBox(
		gui.Fix(0, 8),
		nameRow,
		gui.Fix(0, 15),
		hpStack,
		gui.Fix(0, 15),
		statsTitle,
		gui.Fix(0, 8),
		statsGrid,
		gui.Fix(0, 15),
		invLabel,
		gui.Fix(0, 8),
		invRow,
	)
}

// ---- Chat (droite, bas) : onglets Global/Salle/Groupe séparés des logs ----
// Affiché dans une popup ouverte depuis l'icône "Chat" (voir buildChatMapNav).

func buildChatPanel() fyne.CanvasObject {

	makeScope := func(sample string) fyne.CanvasObject {
		messages := widget.NewLabel(sample)
		messages.Wrapping = fyne.TextWrapWord
		input := widget.NewEntry()
		input.SetPlaceHolder("Écrire un message...")
		send := gui.NewButton("Envoyer", func() {
			println("CHAT:", input.Text)
			input.SetText("")
		})
		return container.NewBorder(nil, container.NewBorder(nil, nil, nil, send, input), nil, nil, messages)
	}

	return container.NewAppTabs(
		container.NewTabItem("Global", makeScope("Alice: ...")),
		container.NewTabItem("Salle", makeScope("Venus: Parle mieux on comprend rien")),
		container.NewTabItem("Groupe", makeScope("")),
	)
}

// ---- Nav Chat / Carte (droite, bas) : titres + grandes icônes cliquables --
// qui ouvrent chacune une popup (chat rempli, carte vide pour l'instant).

const chatMapIconSize = 100

var (
	chatIconPath = "assets/icons/icon_chats.png"
	mapIconPath  = "assets/icons/icon_map.png"
)

func buildChatMapNav(w fyne.Window) fyne.CanvasObject {
	titlesRow := container.NewGridWithColumns(2,
		container.NewCenter(gui.NewTitleText("Chat", 25)),
		container.NewCenter(gui.NewTitleText("Carte", 25)),
	)

	chatIcon := gui.NewBigIconButton(chatIconPath, chatMapIconSize, func() {
		d := dialog.NewCustom("Chat", "Fermer", buildChatPanel(), w)
		d.Resize(fyne.NewSize(900, 650))
		d.Show()
	})
	mapIcon := gui.NewBigIconButton(mapIconPath, chatMapIconSize, func() {
		d := dialog.NewCustom("", "Fermer", gui.BuildMapPanel(1, 1), w)
		d.Resize(fyne.NewSize(1100, 700))
		d.Show()
	})


	iconsRow := container.NewGridWithColumns(2,
		container.NewCenter(chatIcon),
		container.NewCenter(mapIcon),
	)

	return container.NewVBox(titlesRow, gui.Fix(0, 0), iconsRow)
}


func main() {
	a := app.New()
	w := a.NewWindow("The Answer Protocol")
	w.Resize(fyne.NewSize(1400, 850))
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


	rightPanel := gui.WithMinWidth(container.NewVBox(
		buildCharacterPanel(fakePlayer),
		gui.Fix(0, 15),
		buildChatMapNav(w),
	), 420)

	body := container.NewBorder(nil, nil, gui.WithMinWidth(roomContainer, 420), rightPanel, nil)

	// NewBorder = créer un conteneur avec des bordures personnalisées pour
	// organiser les éléments d'une interface (top, bottom, left, right, content)
	content := container.NewBorder(buildHeader(), nil, nil, nil, body)

	full := container.NewStack(background, container.New(layout.NewCustomPaddedLayout(12, 4, 47, 47), content))

	w.SetContent(full)
	w.ShowAndRun()
}
