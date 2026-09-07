package gui

import (
	"encoding/json"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"

	"the_answer_protocol/parsing"
)

// questForGiver retourne la quête (si elle existe) dont ce PNJ est le
// donneur — sert uniquement à savoir s'il faut afficher le bouton Quest à
// côté de lui dans la salle.
func questForGiver(world parsing.WorldFile, giverID string) (parsing.Quest, bool) {
	for _, q := range world.Quests {
		if q.Giver == giverID {
			return q, true
		}
	}
	return parsing.Quest{}, false
}

// openNpcQuestDialog affiche la quête proposée par ce PNJ précis. Comme ce
// bouton n'existe que lorsque le PNJ est dans la salle courante, c'est ici
// (et seulement ici, contrairement au popup Quêtes global en lecture seule)
// qu'on peut accepter une quête disponible, ou récupérer sa
// récompense une fois l'objectif rempli.
func openNpcQuestDialog(w fyne.Window, sess *Session, world parsing.WorldFile, player parsing.Player, npcName, giverID string, refreshCharacter func(parsing.Player), showMessage func(title, msg string)) {
	body := container.NewVBox()
	var d dialog.Dialog
	var render func(q *questEntry)
	sending := false

	// sendQuestCmd envoie "QUEST <giverID>" (utilisé pour Accepter comme
	// pour Récupérer, ReturnQuest gère les deux selon l'état de la quête).
	// Le bouton disparait pendant l'attente pour empêcher de spammer la
	// même action.
	sendQuestCmd := func(q questEntry, onSuccess func(resp string)) {
		if sending {
			return
		}
		sending = true
		render(&q)
		go func() {
			resp, err := sess.Send("QUEST " + giverID)
			if resp == "" && err == nil {
				fyne.Do(func() {
					sending = false
					render(&q)
				})
				return
			}
			if err != nil {
				fyne.Do(func() {
					sending = false
					showMessage("Erreur", err.Error())
					render(&q)
				})
				return
			}
			if strings.HasPrefix(resp, "ERR") {
				fyne.Do(func() {
					sending = false
					showMessage("Erreur", resp)
					render(&q)
				})
				return
			}
			fyne.Do(func() { onSuccess(resp) })
		}()
	}

	render = func(q *questEntry) {
		body.Objects = nil
		body.Add(container.NewCenter(NewTitleText(npcName, 24)))
		body.Add(Fix(0, 15))

		if q == nil {
			body.Add(NewWrappingBodyText("Rien à faire pour l'instant."))
			body.Refresh()
			return
		}

		body.Add(NewTitleText(q.Name, 18))
		body.Add(Fix(0, 8))
		body.Add(NewWrappingBodyText(q.Description))
		body.Add(Fix(0, 15))

		switch {
		case sending:
			body.Add(NewWrappingBodyText("Action en cours..."))
		case q.Status == "available":
			body.Add(NewButton("Accepter", func() {
				sendQuestCmd(*q, func(resp string) {
					showMessage("Quête acceptée", q.Name+" est maintenant en cours.")
					d.Hide()
				})
			}))
		case q.Status == "ready" || q.Status == "waiting":
			body.Add(NewButton("Récupérer", func() {
				sendQuestCmd(*q, func(resp string) {
					var result struct {
						Reward  string `json:"reward"`
						Message string `json:"message"`
					}
					json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &result)

					updated := player
					if it, ok := world.Items[result.Reward]; ok {
						updated.Inventory = append(updated.Inventory, it)
					}
					refreshCharacter(updated) // affiche la récompense dans le panneau perso

					msg := result.Message
					if msg == "" {
						msg = q.Name + " est terminée."
					}
					showMessage("Quête terminée", msg)
					d.Hide()
				})
			}))
		case q.Status == "completed":
			body.Add(NewWrappingBodyText("Terminée."))
		default:
			progress := NewWrappingBodyText("En cours.")
			if q.Progress != "" {
				progress = NewWrappingBodyText("En cours — " + q.Progress)
			}
			body.Add(progress)
		}
		body.Refresh()
	}

	load := func() {
		go func() {
			resp, err := sess.Send("QUESTS")
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
			var quests []questEntry
			if err := json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &quests); err != nil {
				fyne.Do(func() { showMessage("Erreur", "réponse du serveur illisible : "+err.Error()) })
				return
			}
			var found *questEntry
			for i := range quests {
				if quests[i].Giver == giverID {
					found = &quests[i]
					break
				}
			}
			fyne.Do(func() { render(found) })
		}()
	}

	body.Add(container.NewCenter(NewTitleText(npcName, 24)))
	d = dialog.NewCustom("", "Fermer", body, w)
	d.Resize(fyne.NewSize(420, 320))
	d.Show()
	load()
}
