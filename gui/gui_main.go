package gui

import (
	"fmt"
	"os"
	"time"

	"encoding/json"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"

	"the_answer_protocol/parsing"
)

// ----------------- Header (titre) ------------------

func buildHeader() fyne.CanvasObject {
	title := NewTitleText("The Answer Protocol", 60)
	return container.NewBorder(nil, nil, title, nil, nil)
}

// ----------------- Panneau de salle (gauche) ------------------

func buildRoomPanel(loc parsing.Location, world parsing.WorldFile, player parsing.Player, sess *Session, w fyne.Window,
	changeRoom func(parsing.Location), refreshCharacter func(parsing.Player),
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
				if resp == "" && err == nil {
					return // ignore silencieusement (cooldown / action deja en cours)
				}
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
		id := itemID
		label := itemName
		itemsRow.Add(NewButton(label, func() {
			go func() {
				resp, err := sess.Send("TAKE " + id)
				if resp == "" && err == nil {
					return // ignore silencieusement (cooldown / action deja en cours)
				}
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				var newRoom parsing.Location
				if err := json.Unmarshal([]byte(resp), &newRoom); err != nil {
					fyne.Do(func() { showMessage("Erreur", "réponse du serveur illisible : "+err.Error()) })
					return
				}
				updated := player
				if it, ok := world.Items[id]; ok {
					updated.Inventory = append(updated.Inventory, it)
				}
				fyne.Do(func() {
					refreshCharacter(updated)
					changeRoom(newRoom)
				})
			}()
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

		npcButtons := container.NewHBox(
			NewButton(npc+" TALK", func() {
				go func() {
					resp, err := sess.Send("TALK " + id)
					if resp == "" && err == nil {
						return // ignore silencieusement (cooldown / action deja en cours)
					}
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
			}),
		)
		// le bouton Quest n'apparait que pour les PNJ donneurs de quete
		// (Sophie, Esquie...)
		if _, ok := questForGiver(world, id); ok {
			npcButtons.Add(NewButton(npc+" Quest", func() {
				openNpcQuestDialog(w, sess, world, player, npc, id, refreshCharacter, showMessage)
			}))
		}
		npcsRow.Add(npcButtons)
	}

	for _, enemyID := range loc.Enemies {
		enemyName := enemyID
		if e, ok := world.Enemies[enemyID]; ok {
			enemyName = e.Name
		}
		id := enemyID
		enemy := enemyName
		enemyMaxHP := 0
		if e, ok := world.Enemies[id]; ok {
			enemyMaxHP = e.HP
		}

		enemyBar, enemyBarThemed := NewColoredProgressBar(HPColor)
		enemyBar.Min = 0
		enemyBar.Max = float64(enemyMaxHP)
		enemyBar.Value = float64(enemyMaxHP)
		enemyBar.TextFormatter = func() string { return "" }
		enemyHPText := NewTitleText(fmt.Sprintf("%d / %d", enemyMaxHP, enemyMaxHP), 14)
		enemyHPStack := container.NewStack(enemyBarThemed, container.NewCenter(enemyHPText))
		enemyHPFixed := container.New(layout.NewGridWrapLayout(fyne.NewSize(160, 22)), enemyHPStack)

		npcsRow.Add(container.NewVBox(
			enemyHPFixed,
			NewButton(enemy+" ATTACK", func() {
				go func() {
					resp, err := sess.Send("ATTACK " + id)
					if resp == "" && err == nil {
						return // ignore silencieusement (cooldown / action deja en cours)
					}
					if err != nil {
						fyne.Do(func() { showMessage("Erreur", err.Error()) })
						return
					}
					if strings.HasPrefix(resp, "ERR") {
						if strings.Contains(resp, "NPC_NOT_FOUND") {
							fyne.Do(func() { showMessage("Combat", "Victory") })
							return
						}
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
					updated := player
					updated.HP = result.AttackerHP
					fyne.Do(func() {
						enemyBar.Value = float64(result.TargetHP)
						enemyBar.Refresh()
						enemyHPText.Text = fmt.Sprintf("%d / %d", result.TargetHP, enemyMaxHP)
						enemyHPText.Refresh()
						if result.Status == "victory" {
							refreshCharacter(updated)
							showMessage("Combat", "Victoire")
							return
						}
						if result.Status == "defeat" {
							updated.HP = player.Respawn.HP
							refreshCharacter(updated)
							if respawnLoc, ok := world.World.Locations[player.Respawn.Location]; ok {
								changeRoom(respawnLoc)
							}
							showMessage("Combat", "Vous êtes mort... Retour à "+player.Respawn.Location)
							return
						}
						refreshCharacter(updated)
						showMessage("Combat", fmt.Sprintf(
							"Dégâts infligés : %d\nPV de %s : %d / %d (%s)",
							result.Damage, enemy, result.TargetHP, enemyMaxHP, result.Status,
						))
					})
				}()
			}),
		))
	}

	// ----------------- ACTIONS --------------------
	actionsGrid := container.NewGridWithColumns(2,
		NewButton("Status", func() {
			go func() {
				resp, err := sess.Send("STATUS")
				if resp == "" && err == nil {
					return // ignore silencieusement (cooldown / action deja en cours)
				}
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
				updated := player
				updated.HP = s.HP
				fyne.Do(func() {
					refreshCharacter(updated)
					showMessage("Statut", fmt.Sprintf("%d / %d PV — %s", s.HP, s.MaxHP, s.Status))
				})
			}()
		}),
		NewButton("Quests", func() {
			openQuestsDialog(w, sess, showMessage)
		}),
		NewButton("Inventory", func() {
			go func() {
				resp, err := sess.Send("INVENTORY")
				if resp == "" && err == nil {
					return // ignore silencieusement (cooldown / action deja en cours)
				}
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				var itemIDs []string
				if err := json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &itemIDs); err != nil {
					fyne.Do(func() { showMessage("Erreur", "réponse du serveur illisible : "+err.Error()) })
					return
				}
				names := make([]string, 0, len(itemIDs))
				for _, id := range itemIDs {
					if it, ok := world.Items[id]; ok {
						names = append(names, it.Name)
					} else {
						names = append(names, id)
					}
				}
				msg := "Inventaire vide."
				if len(names) > 0 {
					msg = strings.Join(names, ", ")
				}
				fyne.Do(func() { showMessage("Inventaire", msg) })
			}()
		}),
		NewButton("Look", func() {
			go func() {
				resp, err := sess.Send("LOOK")
				if resp == "" && err == nil {
					return // ignore silencieusement (cooldown / action deja en cours)
				}
				if err != nil {
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				if strings.HasPrefix(resp, "ERR") {
					fyne.Do(func() { showMessage("Erreur", resp) })
					return
				}
				var look struct {
					Players []string `json:"players"`
					Items   []string `json:"items"`
					NPCs    []string `json:"npcs"`
				}
				json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &look)
				fyne.Do(func() {
					showMessage("Salle", fmt.Sprintf(
						"Joueurs présents : %s\nObjets : %s\nPNJ : %s",
						strings.Join(look.Players, ", "), strings.Join(look.Items, ", "), strings.Join(look.NPCs, ", "),
					))
				})
			}()
		}),
	)

	// conteneur vertical. Respecte ordre d'affichage
	return container.NewVBox(
		Fix(0, 5),
		title,
		Fix(0, 5),
		desc,
		Fix(0, 8),
		NewTitleText("Sortie", 25),
		Fix(0, 5),
		exitsRow,
		Fix(0, 8),
		NewTitleText("Objets", 25),
		Fix(0, 5),
		itemsRow,
		Fix(0, 8),
		NewTitleText("PNJ", 25),
		Fix(0, 5),
		npcsRow,
		Fix(0, 10),
		NewTitleText("Actions", 25),
		Fix(0, 5),
		actionsGrid,
	)
}

// --- Panneau personnage : PV + inventaire (droite, haute) -----

func statCell(label string, value string) fyne.CanvasObject {
	l := NewBodyText(label, 18)
	v := NewBodyText(value, 18)

	return container.NewVBox(l, v)
}

func buildCharacterPanel(player parsing.Player, sess *Session, w fyne.Window, refreshCharacter func(parsing.Player), changeRoom func(parsing.Location), showMessage func(title, msg string)) fyne.CanvasObject {

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
		i := i
		itemLabel := i.Name
		invRow.Add(NewButton(itemLabel, func() {
			openItemDialog(w, sess, i, player, refreshCharacter, changeRoom, showMessage)
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

// ---- CHAT / MAP (droite, bas) : boutons pour ouvrir les panneaux dans des fenêtres séparées ----

const chatMapIconSize = 100

var (
	chatIconPath = "assets/icons/icon_chats.png"
	mapIconPath  = "assets/icons/icon_map.png"
)

// openChatEntryDialog propose Chat et Groupe : deux fonctionnalités
// distinctes côté serveur (cf. gui/chat_gui.go), regroupées ici derrière
// l'icône Chat plutôt que d'avoir chacune leur bouton dans les actions.
func openChatEntryDialog(w fyne.Window, sess *Session, player parsing.Player, showMessage func(title, msg string)) {
	var d dialog.Dialog
	body := container.NewVBox(
		container.NewCenter(NewTitleText("Chat", 24)),
		Fix(0, 10),
		NewButton("Chat", func() {
			d.Hide()
			openChatDialog(w, sess, player, showMessage)
		}),
		NewButton("Groupe", func() {
			d.Hide()
			openGroupDialog(w, sess, showMessage)
		}),
	)
	d = dialog.NewCustom("Chat", " Fermer", body, w)
	d.Resize(fyne.NewSize(320, 220))
	d.Show()
}

func buildChatMapNav(w fyne.Window, sess *Session, player parsing.Player, showMessage func(title, msg string), getCurrentRoom func() parsing.Location) fyne.CanvasObject {
	titlesRow := container.NewGridWithColumns(2,
		container.NewCenter(NewTitleText("Chat", 25)),
		container.NewCenter(NewTitleText("Carte", 25)),
	)

	chatIcon := NewBigIconButton(chatIconPath, chatMapIconSize, func() {
		openChatEntryDialog(w, sess, player, showMessage)
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
	a.Settings().SetTheme(NewCustomTheme())
	w := a.NewWindow("The Answer Protocol")
	w.Resize(fyne.NewSize(1400, 850))
	w.SetFixedSize(true) // empêche utilisateur de redimensionner fenêtre

	background, bgImage := BuildBackground(startLoc.Background)

	// garde une référence stable qu'on peut vider/remplir pour ne pas
	// reconstruire toute la fenêtre à chaque changement de salle.
	roomContainer := container.NewStack()
	currentRoom := startLoc // salle actuellement affichée, suivie pour la carte

	showMessage := func(title, msg string) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("[GUI] échec d'affichage de la popup", title, ":", msg, "(", r, ")")
			}
		}()
		body := container.NewVBox()
		lines := strings.Split(msg, "\n")
		for i, line := range lines {
			body.Add(NewWrappingBodyText(line))
			if i < len(lines)-1 {
				body.Add(Fix(10, 6))
			}
		}
		content := container.NewVBox(
			container.NewCenter(NewTitleText(title, 28)),
			Fix(0, 25),
			body,
		)
		d := dialog.NewCustom("", "Fermer", content, w)
		d.Resize(fyne.NewSize(480, 320))
		d.Show()
	}

	// Invitation de groupe reçue d'un autre joueur : cf. Session.OnInvite,
	// qui l'intercepte avant qu'elle ne finisse imprimée dans le terminal.
	sess.OnInvite(func(from string) {
		fyne.Do(func() { showInviteDialog(w, sess, from) })
	})

	// Notification de quête terminée par un autre joueur
	sess.OnQuestComplete(func(questName, completer string) {
		fyne.Do(func() {
			showMessage("Quête terminée", fmt.Sprintf("%s a terminé la quête : %s\n\nCette quête est maintenant marquée comme terminée pour vous.", completer, questName))
		})
	})

	// Détection de déconnexion du serveur
	sess.OnDisconnect(func() {
		fyne.Do(func() {
			showMessage("Serveur déconnecté", "Server déconnecté")
		})
		go func() {
			time.Sleep(3 * time.Second)
			fyne.Do(func() {
				w.Close()
				os.Exit(0)
			})
		}()
	})

	var changeRoom func(newRoom parsing.Location)
	var refreshCharacter func(p parsing.Player)

	characterContainer := container.NewStack()
	refreshCharacter = func(p parsing.Player) {
		player = p
		characterContainer.Objects = []fyne.CanvasObject{buildCharacterPanel(p, sess, w, refreshCharacter, func(newRoom parsing.Location) { changeRoom(newRoom) }, showMessage)}
		characterContainer.Refresh()
	}
	refreshCharacter(player) // affichage initial du panneau perso

	changeRoom = func(newRoom parsing.Location) {
		currentRoom = newRoom
		SetRoomBackground(bgImage, newRoom.Background)
		roomContainer.Objects = []fyne.CanvasObject{buildRoomPanel(newRoom, world, player, sess, w, changeRoom, refreshCharacter, showMessage)}
		roomContainer.Refresh()
	}
	changeRoom(startLoc) // affichage initial

	quitBtn := NewSemiTransparentButton("Quit", 140, func() {
		go func() {
			sess.Quit()
			os.Exit(0)
		}()
	})

	// Boutons Who et Quit à droite de la fenêtre (même taille que les autres : 180x40)
	whoBtn := NewSemiTransparentButton("Who", 140, func() {
		go func() {
			resp, err := sess.Send("WHO")
			if resp == "" && err == nil {
				return
			}
			if err != nil {
				fyne.Do(func() { showMessage("Erreur", err.Error()) })
				return
			}
			if strings.HasPrefix(resp, "ERR") {
				fyne.Do(func() { showMessage("Erreur", resp) })
				return
			}
			var who struct {
				Room   []string `json:"room"`
				Server int      `json:"server"`
			}
			json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &who)
			fyne.Do(func() {
				showMessage("Qui est là", fmt.Sprintf(
					"Dans cette salle : %s\nConnectés au serveur : %d",
					strings.Join(who.Room, ", "), who.Server,
				))
			})
		}()
	})

	// Boutons haut à droite (who et quit)
	fixedSize := func(obj fyne.CanvasObject) fyne.CanvasObject {
		return container.New(layout.NewGridWrapLayout(fyne.NewSize(180, 40)), obj)
	}
	buttonsContainer := container.NewHBox(fixedSize(whoBtn), fixedSize(quitBtn))

	rightPanel := WithMinWidth(container.NewVBox(
		characterContainer,
		Fix(0, 15),
		buildChatMapNav(w, sess, player, showMessage, func() parsing.Location { return currentRoom }),
	), 420)

	body := container.NewBorder(nil, nil, WithMinWidth(roomContainer, 420), rightPanel, nil)

	// Header avec titre au centre et boutons à droite
	headerContent := container.NewBorder(nil, Fix(0, 23), buildHeader(), container.NewCenter(buttonsContainer), nil)

	// NewBorder = conteneur avec header en haut, body au milieu
	content := container.NewBorder(headerContent, nil, nil, nil, body)

	full := container.NewStack(background, container.New(layout.NewCustomPaddedLayout(20, 4, 47, 47), content))

	w.SetContent(full)
	w.ShowAndRun()
}
