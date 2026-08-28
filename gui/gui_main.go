package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"the_answer_protocol/parsing"
)

// ---- Données factices, en attendant le vrai serveur ----

type Room struct {
	name        string
	description string
	exits       []string
	items       []string
	npcs        []string
	background  string
	mapRow      int
	mapCol      int
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
	mapRow:      1,
	mapCol:      3,
}

var fakeRoom2 = Room{
	name:        "Cimetière Gestral",
	description: "De minuscules tombes, dont certaines sont vides.",
	exits:       []string{"Village Gestral"},
	background:  "assets/room/cimetiere_gestral.png",
	mapRow:      1,
	mapCol:      4,
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

// ----------------- Header (titre) ------------------

func buildHeader() fyne.CanvasObject {
	title := NewTitleText("The Answer Protocol", 60)

	return container.NewBorder(nil, Fix(0, 40), title, nil, nil)
}

// ----------------- Panneau de salle (gauche) ------------------

func buildRoomPanel(r Room) fyne.CanvasObject {
	// r Room = structure Room qui contient données de la salle
	// sortie fyne.CanvasObject = élément graphique (conteneur, widget...)

	// ------------------ TITRE ------------------
	title := NewTitleText(r.name, 40)
	desc := NewWrappingBodyText(r.description)

	// ------------------ SORTIE -----------------
	exitsRow := container.NewHBox()
	for _, e := range r.exits {
		exit := e
		exitsRow.Add(NewButton(exit, func() {
			fmt.Println("MOVE TO: ", exit)
		}))
	}

	// ------------------- ITEMS -----------------
	itemsRow := container.NewHBox()
	for _, i := range r.items {
		item := i // un bouton par objet
		itemsRow.Add(NewButton(item, func() {
			fmt.Println("TAKE: ", item)
		}))
	}

	// ------------------- PNJ --------------------
	npcsRow := container.NewHBox()
	for _, n := range r.npcs {
		npc := n
		npcsRow.Add(NewButton(npc+"TALK", func() {
			fmt.Println("TALK: ", npc)
		}))
		npcsRow.Add(NewButton(npc+"ATTACK", func() {
			fmt.Println("ATTACK:", npc)
		}))
	}

	// ----------------- ACTIONS --------------------
	actionsGrid := container.NewGridWithColumns(2,
		NewButton("Status", func() { fmt.Println("STATUS") }),
		NewButton("Quest", func() { fmt.Println("QUEST") }), 
		NewButton("Inventory", func() { fmt.Println("INVENTORY") }),
		NewButton("Quit", func() { fmt.Println("QUIT") }),
	)

	// conteneur vertical. Respecte ordre d'affichage
	return container.NewVBox(
		Fix(0, 8),
		title,
		Fix(0, 8),
		desc,
		Fix(0, 10),
		NewTitleText("Sortie", 25),
		Fix(0, 8),
		exitsRow,
		Fix(0, 10),
		NewTitleText("Objets", 25),
		Fix(0, 8),
		itemsRow,
		Fix(0, 15),
		NewTitleText("PNJ", 25),
		Fix(0, 8),
		npcsRow,
		Fix(0, 15),
		NewTitleText("Actions", 25),
		Fix(0, 8),
		actionsGrid,
	)
}

// --- Panneau personnage : PV + inventaire (droite, haute) -----

func statCell(label string, value string) fyne.CanvasObject {
	l := NewBodyText(label, 18)
	v := NewBodyText(value, 18)

	return container.NewVBox(l, v)
}

func buildCharacterPanel(p Player) fyne.CanvasObject {

	// ----------------- AVATAR + NOM PERSO --------------------
	name := NewTitleText(p.name, 70)
	name.TextSize = 48
	name.TextStyle = fyne.TextStyle{Bold: true}

	avatar := NewAvatar("assets/players/Gustave.png", 100)
	nameRow := container.NewHBox(avatar, name)

	// ----------------- BARRE HP --------------------

	hpBar, hpBarThemed := NewColoredProgressBar(HPColor)
	hpBar.Min = 0
	hpBar.Max = p.maxHP
	hpBar.Value = p.hp
	hpBar.TextFormatter = func() string { return "" }
	hpText := NewTitleText(fmt.Sprintf("%.0f / %.0f", p.hp, p.maxHP), 20)
	hpStack := container.NewStack(hpBarThemed, container.NewCenter(hpText))

	// ----------------- STATISTIQUES --------------------

	statsTitle := NewTitleText("Statistiques de combat", 25)
	statsGrid := container.NewGridWithColumns(2,
		statCell("Puissance d'attaque", fmt.Sprintf("%d", p.attack)),
		statCell("Défense", fmt.Sprintf("%d", p.defense)),
		statCell("Chances de critique", fmt.Sprintf("%d%%", p.critPct)),
	)

	// ----------------- INVENTAIRE ---------------------

	invLabel := NewTitleText("Inventaire", 25)
	invRow := container.NewHBox()
	for _, i := range p.inventory {
		item := i
		invRow.Add(NewButton(item, func() {
			fmt.Println("DROP: ", item)
		}))
	}

	return container.NewVBox(
		Fix(0, 8),
		nameRow,
		Fix(0, 15),
		hpStack,
		Fix(0, 15),
		statsTitle,
		Fix(0, 8),
		statsGrid,
		Fix(0, 15),
		invLabel,
		Fix(0, 8),
		invRow,
	)
}

// ---- Chat (droite, bas) : onglets Global/Salle/Groupe séparés des logs ----

func buildChatPanel() fyne.CanvasObject {

	makeScope := func(sample string) fyne.CanvasObject {
		messages := widget.NewLabel(sample)
		messages.Wrapping = fyne.TextWrapWord
		input := widget.NewEntry()
		input.SetPlaceHolder("Écrire un message...")
		send := NewButton("Envoyer", func() {
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

// ---- CHAT / MAP (droite, bas) : boutons pour ouvrir les panneaux dans des fenêtres séparées ----

const chatMapIconSize = 100

var (
	chatIconPath = "assets/icons/icon_chats.png"
	mapIconPath  = "assets/icons/icon_map.png"
)

func buildChatMapNav(w fyne.Window, getCurrentRoom func() Room) fyne.CanvasObject {
	titlesRow := container.NewGridWithColumns(2,
		container.NewCenter(NewTitleText("Chat", 25)),
		container.NewCenter(NewTitleText("Carte", 25)),
	)

	chatIcon := NewBigIconButton(chatIconPath, chatMapIconSize, func() {
		d := dialog.NewCustom("Chat", "Fermer", buildChatPanel(), w)
		d.Resize(fyne.NewSize(900, 650))
		d.Show()
	})
	mapIcon := NewBigIconButton(mapIconPath, chatMapIconSize, func() {
		room := getCurrentRoom()
		d := dialog.NewCustom("", "Fermer", BuildMapPanel(room.mapRow, room.mapCol), w)
		d.Resize(fyne.NewSize(1100, 650))
		d.Show()
	})

	iconsRow := container.NewGridWithColumns(2,
		container.NewCenter(chatIcon),
		container.NewCenter(mapIcon),
	)

	return container.NewVBox(titlesRow, Fix(0, 0), iconsRow)
}

// ------- Run ouvre la fenêtre du GUI quand perso vient d'être choisi côté server --------

func Run(playerName string, world parsing.WorldFile) {

	player := fakePlayer
	player.name = playerName

	fmt.Println(world)
	a := app.New()
	w := a.NewWindow("The Answer Protocol")
	w.Resize(fyne.NewSize(1400, 850))
	w.SetFixedSize(true) // empêche utilisateur de redimensionner fenêtre

	// DONNEES PROVISOIRES POUR TESTS MANUELS
	demoRooms := map[string]Room{
		"Village Gestral":   fakeRoom,
		"Cimetière gestral": fakeRoom2,
		"Cimetière Gestral": fakeRoom2,
	}

	background, bgImage := BuildBackground(fakeRoom.background)

	// garde une référence stable qu'on peut vider/remplir pour ne pas
	// reconstruire toute la fenêtre à chaque changement de salle.
	roomContainer := container.NewStack()

	currentRoom := fakeRoom // salle actuellement affichée, suivie pour la carte

	// pour changer d'affichage quand on change de room
	changeRoom := func(newRoom Room) {
		currentRoom = newRoom
		SetRoomBackground(bgImage, newRoom.background)
		roomContainer.Objects = []fyne.CanvasObject{buildRoomPanel(newRoom)} // VOIR
		roomContainer.Refresh()
	}
	changeRoom(fakeRoom) // affichage initial

	rightPanel := WithMinWidth(container.NewVBox(
		buildCharacterPanel(fakePlayer),
		Fix(0, 15),
		buildChatMapNav(w, func() Room { return currentRoom }),
	), 420)

	body := container.NewBorder(nil, nil, WithMinWidth(roomContainer, 420), rightPanel, nil)

	// NewBorder = conteneur avce header en haut, body au milieu, rightPanel à droite, roomContainer à gauche + bordures personnalisées
	content := container.NewBorder(buildHeader(), nil, nil, nil, body)

	full := container.NewStack(background, container.New(layout.NewCustomPaddedLayout(12, 4, 47, 47), content))

	w.SetContent(full)
	w.ShowAndRun()
}
