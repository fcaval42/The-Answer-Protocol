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

// waitFirst attend jusqu'à maxWait la première ligne reçue sur ch, puis
// regroupe les éventuelles lignes suivantes rapprochées comme collectQuiet.
// Contrairement à collectQuiet - qui abandonne dès que `quiet` s'écoule sans
// qu'aucune ligne ne soit jamais arrivée, ce qui convient à un aller-retour
// synchrone - il faut ici bloquer jusqu'à la première ligne : le serveur
// reste volontairement silencieux sur cette connexion tant que l'invité n'a
// pas répondu, ou que son délai n'a pas expiré (cf. GroupInvite côté
// serveur, tcp/server/group.go, qui attend jusqu'à 10s).
func waitFirst(ch <-chan string, maxWait time.Duration) []string {
	select {
	case l, ok := <-ch:
		if !ok {
			return nil
		}
		var lines []string
		for _, part := range strings.Split(strings.TrimRight(l, "\n"), "\n") {
			if part != "" {
				lines = append(lines, part)
			}
		}
		return append(lines, collectQuiet(ch, 300*time.Millisecond, 300*time.Millisecond)...)
	case <-time.After(maxWait):
		return nil
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

// availablePlayersFrom extrait la liste des joueurs invitables du message
// serveur "[SERVER] Joueurs disponibles : a, b, c\n" (cf. GroupInvite côté
// serveur, tcp/server/group.go). Renvoie (nil, false) si le texte ne contient
// pas ce marqueur (autre message : erreur, aucun joueur dispo...).
func availablePlayersFrom(text string) ([]string, bool) {
	const marker = "Joueurs disponibles : "
	idx := strings.Index(text, marker)
	if idx == -1 {
		return nil, false
	}
	rest := text[idx+len(marker):]
	if nl := strings.IndexByte(rest, '\n'); nl != -1 {
		rest = rest[:nl]
	}
	var names []string
	for _, n := range strings.Split(rest, ",") {
		n = strings.TrimSpace(n)
		if n != "" {
			names = append(names, n)
		}
	}
	return names, len(names) > 0
}

// openInvitePicker affiche la liste des joueurs invitables et envoie le choix
// de l'utilisateur sur ch/sess. Une fois "GROUP INVITE" envoyé, le serveur
// reste bloqué (Recept en boucle côté GroupInvite) tant qu'un nom valide ne
// lui est pas fourni sur CETTE connexion : contrairement à GROUP JOIN, il n'y
// a donc pas moyen d'annuler proprement ici, d'où l'absence de bouton
// Fermer/Annuler (NewCustomWithoutButtons) - annuler laisserait la connexion
// bloquée pour le reste de la session.
func openInvitePicker(w fyne.Window, sess *Session, ch <-chan string, names []string, showMessage func(title, msg string)) {
	body := container.NewVBox(NewWrappingBodyText("Qui voulez-vous inviter ?"), Fix(0, 8))
	var d dialog.Dialog

	for _, name := range names {
		n := name
		body.Add(NewButton(n, func() {
			d.Hide()
			go func() {
				if err := sess.SendRaw(n + "\n"); err != nil {
					sess.ExitRaw()
					fyne.Do(func() { showMessage("Erreur", err.Error()) })
					return
				}
				// Nom invalide (joueur parti entre-temps...) : le serveur
				// reprint aussitôt la liste et attend un nouveau choix sur
				// cette même connexion, sans jamais revenir à Dispatch.
				first := collectQuiet(ch, 300*time.Millisecond, 2*time.Second)
				if retryNames, ok := availablePlayersFrom(strings.Join(first, "\n")); ok {
					fyne.Do(func() { openInvitePicker(w, sess, ch, retryNames, showMessage) })
					return
				}
				// Nom valide : le serveur reste maintenant silencieux sur
				// cette connexion jusqu'à ce que l'invité réponde (jusqu'à
				// 10s), pendant lesquelles toute autre action GUI (Attack,
				// Move...) resterait sans réponse - d'où cette popup
				// d'attente bloquante, cohérente avec la modalité déjà
				// acceptée pour le chat (cf. commentaire de package).
				var waiting dialog.Dialog
				fyne.Do(func() {
					waiting = dialog.NewCustomWithoutButtons("Invitation envoyée",
						NewWrappingBodyText("En attente de la réponse de "+n+" (jusqu'à 10s)..."), w)
					waiting.Resize(fyne.NewSize(340, 140))
					waiting.Show()
				})
				rest := waitFirst(ch, 11*time.Second)
				sess.ExitRaw()
				result := append(first, rest...)
				if len(result) == 0 {
					result = []string{"Pas de réponse du serveur."}
				}
				fyne.Do(func() {
					waiting.Hide()
					showMessage("Groupe", strings.Join(result, "\n"))
				})
			}()
		}))
	}

	d = dialog.NewCustomWithoutButtons("Inviter", body, w)
	d.Resize(fyne.NewSize(360, 260))
	d.Show()
}

// showInviteDialog affiche par-dessus tout le reste (parent = fenêtre
// principale) l'invitation de groupe reçue d'un autre joueur, cf.
// Session.OnInvite. ACCEPT/DECLINE part directement sur la connexion via
// SendRaw plutôt que Send() : la confirmation renvoyée par le serveur est un
// texte libre "[SERVER] ..." (cf. dispatch.go), pas une ligne OK/ERR/{ que
// Send() saurait reconnaître.
func showInviteDialog(w fyne.Window, sess *Session, from string) {
	var d dialog.Dialog
	answered := false
	respond := func(answer string) {
		if answered {
			return
		}
		answered = true
		d.Hide()
		sess.SendRaw(answer + "\n")
	}

	body := container.NewVBox(
		NewWrappingBodyText(from+" vous invite à rejoindre son groupe."),
		Fix(0, 16),
		container.NewGridWithColumns(2,
			NewButton("Accepter", func() { respond("ACCEPT") }),
			NewButton("Refuser", func() { respond("DECLINE") }),
		),
	)
	d = dialog.NewCustomWithoutButtons("Invitation de groupe", body, w)
	d.Resize(fyne.NewSize(380, 180))
	d.Show()

	// Le serveur abandonne l'invitation après 10s sans réponse (cf.
	// GroupInvite) : la popup se referme d'elle-même pour ne pas laisser un
	// choix devenu sans effet affiché à l'écran.
	go func() {
		time.Sleep(10 * time.Second)
		fyne.Do(func() {
			if !answered {
				answered = true
				d.Hide()
			}
		})
	}()
}

// openGroupDialog propose Créer / Rejoindre / Inviter / Quitter un groupe.
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
	body.Add(NewButton("Inviter un joueur", func() {
		go func() {
			ch := sess.EnterRaw()
			if err := sess.SendRaw("GROUP INVITE\n"); err != nil {
				sess.ExitRaw()
				fyne.Do(func() { showMessage("Erreur", err.Error()) })
				return
			}
			lines := collectQuiet(ch, 200*time.Millisecond, 2*time.Second)
			text := strings.Join(lines, "\n")
			names, ok := availablePlayersFrom(text)
			if !ok {
				sess.ExitRaw()
				fyne.Do(func() { showMessage("Groupe", text) })
				return
			}
			fyne.Do(func() { openInvitePicker(w, sess, ch, names, showMessage) })
		}()
	}))
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
	d.Resize(fyne.NewSize(360, 320))
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
