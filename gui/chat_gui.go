package gui

import (
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"the_answer_protocol/parsing"
)

// Le chat et les groupes (CHAT, GROUP CREATE/JOIN/LEAVE...) sont gérés côté
// serveur par des boucles bloquantes qui monopolisent la connexion (Recept
// en boucle, cf. tcp/server/chat.go et group.go) et répondent en texte libre
// "[SERVER] ..." — jamais par une ligne OK/ERR/{ comme le reste du jeu. Ça
// ne passe donc pas par Session.Send() : ce fichier pilote la connexion en
// mode brut (Session.EnterRaw/SendRaw) le temps de l'échange, exactement
// comme le ferait le client texte. Conséquence assumée : pendant qu'une
// session de chat est active dans son popup, les autres actions du GUI
// (Attack, Move...) sont bloquées par la modalité du popup — il faut
// explicitement Quitter le chat pour rendre la main.

// collectQuiet lit ch jusqu'à un silence de `quiet`, ou au plus `maxWait`,
// et renvoie les lignes reçues entre-temps (une entrée par ligne logique,
// les messages multi-lignes du serveur étant déjà rejoints par de vrais
// "\n" à ce stade, cf. client_main.go).
func collectQuiet(ch <-chan string, quiet, maxWait time.Duration) []string {
	var lines []string
	deadline := time.After(maxWait)
	timer := time.NewTimer(quiet)
	defer timer.Stop()
	for {
		select {
		case l, ok := <-ch:
			if !ok {
				return lines
			}
			for _, part := range strings.Split(strings.TrimRight(l, "\n"), "\n") {
				if part != "" {
					lines = append(lines, part)
				}
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(quiet)
		case <-timer.C:
			return lines
		case <-deadline:
			return lines
		}
	}
}

// sendRawCollect ouvre une session brute le temps d'un aller-retour ponctuel
// (GROUP CREATE / GROUP LEAVE...), sans conversation qui suit.
func sendRawCollect(sess *Session, cmd string) ([]string, error) {
	ch := sess.EnterRaw()
	defer sess.ExitRaw()
	if err := sess.SendRaw(cmd); err != nil {
		return nil, err
	}
	return collectQuiet(ch, 200*time.Millisecond, 2*time.Second), nil
}

// openGroupDialog propose Créer / Rejoindre / Quitter un groupe. GROUP
// INVITE n'est volontairement pas câblé ici : c'est un échange multi-étapes
// avec un AUTRE joueur (choix du destinataire puis attente de son
// ACCEPT/DECLINE sur sa propre connexion), d'une autre nature que les trois
// actions ponctuelles ci-dessous.
func openGroupDialog(w fyne.Window, sess *Session, showMessage func(title, msg string)) {
	body := container.NewVBox(container.NewCenter(NewTitleText("Groupe", 24)), Fix(0, 10))

	act := func(cmd string) {
		go func() {
			lines, err := sendRawCollect(sess, cmd)
			if err != nil {
				fyne.Do(func() { showMessage("Erreur", err.Error()) })
				return
			}
			fyne.Do(func() { showMessage("Groupe", strings.Join(lines, "\n")) })
		}()
	}

	body.Add(NewButton("Créer un groupe", func() { act("GROUP CREATE\n") }))
	body.Add(NewButton("Quitter le groupe", func() { act("GROUP LEAVE\n") }))
	body.Add(NewButton("Rejoindre un groupe", func() {
		go func() {
			ch := sess.EnterRaw()
			if err := sess.SendRaw("GROUP JOIN\n"); err != nil {
				sess.ExitRaw()
				fyne.Do(func() { showMessage("Erreur", err.Error()) })
				return
			}
			lines := collectQuiet(ch, 200*time.Millisecond, 2*time.Second)
			text := strings.Join(lines, "\n")
			if !strings.Contains(text, "Entrez l'ID du groupe") {
				sess.ExitRaw()
				fyne.Do(func() { showMessage("Groupe", text) })
				return
			}
			fyne.Do(func() {
				idEntry := widget.NewEntry()
				idEntry.SetPlaceHolder("ID du groupe")
				dialog.ShowForm("Rejoindre un groupe", "Rejoindre", "Annuler",
					[]*widget.FormItem{widget.NewFormItem("", idEntry)},
					func(ok bool) {
						id := strings.TrimSpace(idEntry.Text)
						go func() {
							defer sess.ExitRaw()
							if !ok || id == "" {
								// Le serveur attend toujours un ID sur cette
								// connexion (Recept bloquant) : sans réponse,
								// le prochain Send() du GUI recevrait ce
								// texte au lieu de sa vraie réponse. Une
								// ligne vide le débloque proprement (ID
								// invalide côté serveur, sans effet de bord).
								sess.SendRaw("\n")
								return
							}
							sess.SendRaw(id + "\n")
							more := collectQuiet(ch, 200*time.Millisecond, 2*time.Second)
							fyne.Do(func() { showMessage("Groupe", strings.Join(more, "\n")) })
						}()
					}, w)
			})
		}()
	}))

	d := dialog.NewCustom("", "Fermer", body, w)
	d.Resize(fyne.NewSize(360, 260))
	d.Show()
}

// openChatDialog reproduit le déroulé du chat texte (CHAT -> GLOBAL/ROOM/
// GROUP -> conversation -> EXIT) dans une fenêtre GUI. cf. le commentaire de
// package en tête de fichier pour le pourquoi du mode brut.
func openChatDialog(w fyne.Window, sess *Session, player parsing.Player, showMessage func(title, msg string)) {
	body := container.NewVBox()
	var d dialog.Dialog

	log := widget.NewLabel("")
	log.Wrapping = fyne.TextWrapWord
	logScroll := container.NewVScroll(log)
	logScroll.SetMinSize(fyne.NewSize(420, 260))

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Votre message...")

	done := make(chan struct{})
	// inChatLoop : vrai des l'envoi d'un canal (GLOBAL/ROOM/GROUP) tant qu'on
	// n'a pas la certitude que le serveur en est deja ressorti tout seul
	// (cas "GROUP" hors groupe, cf. plus bas) - c'est ce qui determine si
	// quitter la fenetre doit envoyer EXIT pour debloquer la connexion.
	inChatLoop := false
	active := false
	closed := false
	scopeLabel := ""

	linesCh := sess.EnterRaw()

	var renderPicker func()
	var renderActive func()
	var chooseScope func(cmd, label string)

	enterChatPrompt := func() { sess.SendRaw("CHAT\n") }

	leave := func() {
		if closed {
			return
		}
		closed = true
		close(done)
		if inChatLoop {
			sess.SendRaw("EXIT\n")
		}
		sess.ExitRaw()
	}

	renderPicker = func() {
		body.Objects = nil
		body.Add(container.NewCenter(NewTitleText("Chat", 24)))
		body.Add(Fix(0, 10))
		body.Add(NewWrappingBodyText("Choisissez un canal :"))
		body.Add(Fix(0, 8))
		body.Add(container.NewGridWithColumns(3,
			NewButton("Global", func() { chooseScope("GLOBAL\n", "Chat global") }),
			NewButton("Salle", func() { chooseScope("ROOM\n", "Chat de la salle") }),
			NewButton("Groupe", func() { chooseScope("GROUP\n", "Chat de groupe") }),
		))
		body.Refresh()
	}

	chooseScope = func(cmd, label string) {
		inChatLoop = true
		scopeLabel = label
		sess.SendRaw(cmd)
		body.Objects = nil
		body.Add(container.NewCenter(NewTitleText(label, 24)))
		body.Add(NewWrappingBodyText("Connexion..."))
		body.Refresh()
	}

	renderActive = func() {
		active = true
		body.Objects = nil
		body.Add(container.NewCenter(NewTitleText(scopeLabel, 24)))
		body.Add(Fix(0, 8))
		body.Add(logScroll)
		body.Add(Fix(0, 8))
		body.Add(container.NewBorder(nil, nil, nil, NewButton("Envoyer", func() {
			text := strings.TrimSpace(entry.Text)
			if text == "" {
				return
			}
			sess.SendRaw(text + "\n")
			entry.SetText("")
		}), entry))
		body.Add(Fix(0, 8))
		body.Add(NewButton("Retour", func() {
			d.Hide() // déclenche leave() via SetOnClosed : EXIT + ExitRaw
			openChatEntryDialog(w, sess, player, showMessage)
		}))
		body.Refresh()
	}
	entry.OnSubmitted = func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		sess.SendRaw(text + "\n")
		entry.SetText("")
	}

	go func() {
		for {
			select {
			case <-done:
				return
			case line, ok := <-linesCh:
				if !ok {
					return
				}
				// La reception d'une ligne (meme vide : un salon/groupe sans
				// historique se traduit par un broadcast vide, cf. Broadcast
				// / PrintChat cote serveur) est notre seul signal que le
				// canal a bien ete rejoint - il ne faut donc pas filtrer les
				// lignes vides ici avant de decider de passer en mode actif,
				// seulement pour l'affichage.
				text := strings.TrimRight(line, "\n")
				fyne.Do(func() {
					switch {
					case !inChatLoop:
						// prompt initial ("GLOBAL, ROOM, GROUP ?") : on
						// affiche deja nos propres boutons, on l'ignore.
					case !active && strings.Contains(text, "Vous n'êtes pas dans un groupe"):
						// GroupChat n'est jamais entree en boucle : le
						// serveur est deja revenu au Dispatch principal,
						// il faut renvoyer CHAT pour recommencer.
						inChatLoop = false
						showMessage("Chat", text)
						renderPicker()
						enterChatPrompt()
					case !active:
						renderActive()
						log.SetText(text)
					default:
						log.SetText(text)
					}
				})
			}
		}
	}()

	renderPicker()
	d = dialog.NewCustom("", "Fermer", body, w)
	d.Resize(fyne.NewSize(460, 420))
	d.SetOnClosed(leave)
	d.Show()
	enterChatPrompt()
}
