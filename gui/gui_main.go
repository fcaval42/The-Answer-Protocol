package gui

import (
	"fmt"

	"encoding/json"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"the_answer_protocol/parsing"
)


// ----------------- Header (titre) ------------------

func buildHeader() fyne.CanvasObject {
	title := NewTitleText("The Answer Protocol", 60)

	return container.NewBorder(nil, Fix(0, 40), title, nil, nil)
}

// ----------------- Panneau de salle (gauche) ------------------

func buildRoomPanel(loc parsing.Location, world parsing.WorldFile, sess *Session,
	changeRoom func(parsing.Location), refreshCharacter func(hp int),
	showMessage func(title, msg string)) fyne.CanvasObject {

	// ------------------ TITRE ------------------
	title := NewTitleText(loc.Name, 40)
	desc := NewWrappingBodyText(loc.Description)

	// ------------------ SORTIE -----------------
	exitsRow := container.New(layout.NewGridWrapLayout(fyne.NewSize(180, 40)))
	seenExits := make(map[string]bool)
	for direction, destID := range loc.Exits {
		if seenExits[destID] {
			continue // plusieurs directions peuvent mener à la même salle
		}
		seenExits[destID] = true

		destLoc, ok := world.World.Locations[destID]
		exitLabel := destID
		if ok {
			exitLabel = destLoc.Name
		}
		dir := direction

		exitsRow.Add(NewButton(exitLabel, func() {
			go func() {
				resp, err := sess.Send("MOVE " + dir)
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				var newLoc parsing.Location
				if err := json.Unmarshal([]byte(resp), &newLoc); err != nil {
					fyne.Do(func() { showMessage("Erreur", "réponse du serveur illisible : "+err.Error()) })
					return
				}
				fyne.Do(func() { changeRoom(newLoc) })
			}()
		}))
	}

	// ------------------- ITEMS -----------------
	itemsRow := container.NewHBox()
	for _, itemID := range loc.Items {
		itemName := itemID
		if it, ok := world.Items[itemID]; ok {
			itemName = it.Name
		}
		item := itemName // un bouton par objet
		itemsRow.Add(NewButton(item, func() {
			fmt.Println("TAKE: ", item)
		}))
	}

	// ------------------- PNJ --------------------
	npcsRow := container.NewHBox()
	for _, npcID := range loc.NPCs {
		npcName := npcID
		if n, ok := world.NPCs[npcID]; ok {
			npcName = n.Name
		}
		id := npcID
		npc := npcName

		npcsRow.Add(NewButton(npc+" TALK", func() {
			go func() {
				resp, err := sess.Send("TALK " + id)
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				fyne.Do(func() { showMessage(npc, strings.TrimPrefix(resp, "OK ")) })
			}()
		}))
	}

	for _, enemyID := range loc.Enemies {
		enemyName := enemyID
		if e, ok := world.Enemies[enemyID]; ok {
			enemyName = e.Name
		}
		id := enemyID
		enemy := enemyName
		// PV de depart de l'ennemi, tels que charges au demarrage du client :
		// world cote GUI n'est jamais resynchronise pendant le combat, donc
		// ca reste un plafond correct pour l'affichage "restant / max".
		enemyMaxHP := 0
		if e, ok := world.Enemies[id]; ok {
			enemyMaxHP = e.HP
		}

		npcsRow.Add(NewButton(enemy+" ATTACK", func() {
			go func() {
				resp, err := sess.Send("ATTACK " + id)
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				var result struct {
					AttackerHP int    `json:"attacker_hp"`
					TargetHP   int    `json:"target_hp"`
					Damage     int    `json:"damage"`
					Status     string `json:"status"`
				}
				json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &result)
				fyne.Do(func() {
					refreshCharacter(result.AttackerHP)
					showMessage("Combat", fmt.Sprintf(
						"Dégâts infligés : %d\nPV de %s : %d / %d (%s)",
						result.Damage, enemy, result.TargetHP, enemyMaxHP, result.Status,
					))
				})
			}()
		}))
	}

	// ----------------- ACTIONS --------------------
	actionsGrid := container.NewGridWithColumns(2,
		NewButton("Status", func() {
			go func() {
				resp, err := sess.Send("STATUS")
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				var s struct {
					HP     int    `json:"hp"`
					MaxHP  int    `json:"max_hp"`
					Status string `json:"status"`
				}
				json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &s)
				fyne.Do(func() {
					refreshCharacter(s.HP)
					showMessage("Statut", fmt.Sprintf("%d / %d PV — %s", s.HP, s.MaxHP, s.Status))
				})
			}()
		}),
		NewButton("Quest", func() {
			go func() {
				resp, err := sess.Send("QUESTS")
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				fyne.Do(func() { showMessage("Quêtes", strings.TrimPrefix(resp, "OK ")) })
			}()
		}),
		NewButton("Inventory", func() {
			go func() {
				resp, err := sess.Send("INVENTORY")
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				fyne.Do(func() { showMessage("Inventaire", strings.TrimPrefix(resp, "OK ")) })
			}()
		}),
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

func buildCharacterPanel(player parsing.Player) fyne.CanvasObject {

	// ----------------- AVATAR + NOM PERSO --------------------
	name := NewTitleText(player.Name, 70)
	name.TextSize = 48
	name.TextStyle = fyne.TextStyle{Bold: true}

	avatar := NewAvatar(avatarPath(player.Name), 100)
	nameRow := container.NewHBox(avatar, name)

	// ----------------- BARRE HP --------------------

	hpBar, hpBarThemed := NewColoredProgressBar(HPColor)
	hpBar.Min = 0
	hpBar.Max = float64(player.MaxHP)
	hpBar.Value = float64(player.HP)
	hpBar.TextFormatter = func() string { return "" }
	hpText := NewTitleText(fmt.Sprintf("%d / %d", player.HP, player.MaxHP), 20)
	hpStack := container.NewStack(hpBarThemed, container.NewCenter(hpText))

	// ----------------- STATISTIQUES --------------------

	statsTitle := NewTitleText("Statistiques de combat", 25)
	statsGrid := container.NewGridWithColumns(2,
		statCell("Puissance d'attaque", fmt.Sprintf("%d", player.Attack)),
		statCell("Défense", fmt.Sprintf("%d", player.Defense)),
		statCell("Chances de critique", fmt.Sprintf("%.0f%%", player.Weapon.Critical*100)),
	)

	// ----------------- INVENTAIRE ---------------------

	invLabel := NewTitleText("Inventaire", 25)
	invRow := container.NewHBox()
	for _, i := range player.Inventory {
		item := i.Name
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

func buildChatMapNav(w fyne.Window, getCurrentRoom func() parsing.Location) fyne.CanvasObject {
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
		row, col, ok := roomGridPosition(room.Name)
		if !ok {
			row, col = -1, -1
		}
		d := dialog.NewCustom("", "Fermer", BuildMapPanel(row, col), w)
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

func findPlayer(world parsing.WorldFile, name string) (parsing.Player, bool) {
	for _, player := range world.Players {
		if player.Name == name {
			return player, true
		}
	}
	return parsing.Player{}, false
}


func Run(playerName string, world parsing.WorldFile, sess *Session) {

	player, ok := findPlayer(world, playerName)
	if !ok {
		fmt.Println("[ATTENTION] joueur inconnu dans world.Players: ", player.Name)
	}

	startLoc := world.World.Locations[player.Respawn.Location]

	a := app.New()
	w := a.NewWindow("The Answer Protocol")
	w.Resize(fyne.NewSize(1400, 850))
	w.SetFixedSize(true) // empêche utilisateur de redimensionner fenêtre

	background, bgImage := BuildBackground(startLoc.Background)

	// garde une référence stable qu'on peut vider/remplir pour ne pas
	// reconstruire toute la fenêtre à chaque changement de salle.
	roomContainer := container.NewStack()
	currentRoom := startLoc // salle actuellement affichée, suivie pour la carte


	// permet de redessiner le panneau perso à jour (HP)
	// characterContainer = boîte pour vider/remplir --- refreshCharacter = reconstruit avce nv HP
	// -- showMessage = raccourci pour ouvrir une boîte de dialogue
	characterContainer := container.NewStack(buildCharacterPanel(player))
	refreshCharacter := func(hp int) {
		player.HP = hp
		characterContainer.Objects = []fyne.CanvasObject{buildCharacterPanel(player)}
		characterContainer.Refresh()
	}
	showMessage := func(title, msg string) {
		// Bug connu de Fyne/harfbuzz (go-text/typesetting) : le rendu de
		// certains textes dans une popup peut paniquer en interne. Sans ce
		// recover, ça tue tout le processus GUI. On perd juste ce message
		// (loggue au terminal) plutot que de crasher toute la partie.
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("[GUI] échec d'affichage de la popup", title, ":", msg, "(", r, ")")
			}
		}()
		dialog.NewInformation(title, msg, w).Show()
	}

	// pour changer d'affichage quand on change de room
	var changeRoom func(newRoom parsing.Location)
	changeRoom = func(newRoom parsing.Location) {
		currentRoom = newRoom
		SetRoomBackground(bgImage, newRoom.Background)
		roomContainer.Objects = []fyne.CanvasObject{buildRoomPanel(newRoom, world, sess, changeRoom, refreshCharacter, showMessage)}
		roomContainer.Refresh()
	}
	changeRoom(startLoc) // affichage initial

	rightPanel := WithMinWidth(container.NewVBox(
		characterContainer,
		Fix(0, 15),
		buildChatMapNav(w, func() parsing.Location { return currentRoom }),
	), 420)

	body := container.NewBorder(nil, nil, WithMinWidth(roomContainer, 420), rightPanel, nil)

	// NewBorder = conteneur avce header en haut, body au milieu, rightPanel à droite, roomContainer à gauche + bordures personnalisées
	content := container.NewBorder(buildHeader(), nil, nil, nil, body)

	full := container.NewStack(background, container.New(layout.NewCustomPaddedLayout(12, 4, 47, 47), content))

	w.SetContent(full)
	w.ShowAndRun()
}
