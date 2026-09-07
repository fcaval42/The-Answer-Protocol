package gui

import (
	"encoding/json"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"

	"the_answer_protocol/parsing"
)

// questEntry reflète une entrée JSON renvoyée par la commande QUESTS :
// quêtes disponibles auprès des PNJ de la salle courante (non encore
// acceptées), quêtes en cours, prêtes à être rendues, ou déjà terminées.
type questEntry struct {
	QuestID     string `json:"quest_id"`
	Name        string `json:"name"`
	Giver       string `json:"giver"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Progress    string `json:"progress"`
}

// removeQuest retourne quests privé de l'entrée questID (utilisé quand le
// joueur décline une quête disponible : rien à envoyer au serveur puisque
// tant qu'elle n'est pas acceptée, aucun état n'a été créé côté joueur — on
// se contente de la retirer de la liste affichée).
func removeQuest(quests []questEntry, questID string) []questEntry {
	out := make([]questEntry, 0, len(quests))
	for _, q := range quests {
		if q.QuestID != questID {
			out = append(out, q)
		}
	}
	return out
}

// openQuestsDialog affiche la popup du bouton Quest : liste les quêtes
// disponibles (proposées par les PNJ de la salle), en cours, prêtes à être
// rendues, ou terminées, avec les boutons d'action correspondants
// directement sur chaque quête.
func openQuestsDialog(w fyne.Window, sess *Session, world parsing.WorldFile, player parsing.Player, refreshCharacter func(parsing.Player), showMessage func(title, msg string)) {
	// snapshot local de l'inventaire du joueur, mis à jour à chaque
	// récompense récupérée pendant que cette popup reste ouverte (pour ne
	// pas perdre les objets déjà ajoutés si plusieurs quêtes sont rendues
	// l'une après l'autre sans refermer la popup).
	currentPlayer := player

	list := container.NewVBox()
	scroll := container.NewVScroll(list)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("", "Fermer", container.NewVBox(
		container.NewCenter(NewTitleText("Quêtes", 28)),
		Fix(0, 15),
		scroll,
	), w)
	d.Resize(fyne.NewSize(560, 520))

	// quêtes pour lesquelles une demande (acceptation ou récupération de
	// récompense) est déjà partie vers le serveur : le bouton correspondant
	// est masqué tant qu'on n'a pas la réponse, pour qu'un double-clic (ou
	// un clic pendant que le serveur répond) ne puisse jamais déclencher une
	// deuxième requête pour la même quête depuis cette popup.
	pending := map[string]bool{}

	var render func(quests []questEntry)
	var load func()

	// sendQuestCommand envoie "QUEST <giver>" en verrouillant la quête (un
	// seul essai a la fois depuis cette popup) : le bouton disparait pendant
	// l'attente, puis onSuccess traite la réponse en cas de réussite. En cas
	// d'échec ou d'erreur, le verrou est relâché pour permettre un nouvel
	// essai.
	sendQuestCommand := func(q questEntry, quests []questEntry, onSuccess func(resp string)) {
		if pending[q.QuestID] {
			return
		}
		pending[q.QuestID] = true
		render(quests)
		go func() {
			resp, err := sess.Send("QUEST " + q.Giver)
			if resp == "" && err == nil {
				fyne.Do(func() {
					delete(pending, q.QuestID)
					render(quests)
				})
				return
			}
			if err != nil {
				fyne.Do(func() {
					delete(pending, q.QuestID)
					showMessage("Erreur", err.Error())
					render(quests)
				})
				return
			}
			if strings.HasPrefix(resp, "ERR") {
				fyne.Do(func() {
					delete(pending, q.QuestID)
					showMessage("Erreur", resp)
					render(quests)
				})
				return
			}
			fyne.Do(func() { onSuccess(resp) })
		}()
	}

	render = func(quests []questEntry) {
		list.Objects = nil
		if len(quests) == 0 {
			list.Add(NewWrappingBodyText("Aucune quête pour le moment."))
		}
		for _, q := range quests {
			q := q
			header := NewTitleText(fmt.Sprintf("%s (%s)", q.Name, q.Giver), 18)

			var detail fyne.CanvasObject
			switch {
			case q.Status == "available" && pending[q.QuestID]:
				detail = container.NewVBox(
					NewWrappingBodyText(q.Description),
					NewWrappingBodyText("Acceptation en cours..."),
				)
			case q.Status == "available":
				detail = container.NewVBox(
					NewWrappingBodyText(q.Description),
					container.NewHBox(
						NewButton("Accepter", func() {
							sendQuestCommand(q, quests, func(resp string) {
								showMessage("Quête acceptée", q.Name+" est maintenant en cours.")
								load() // recharge depuis le serveur : la quête passe en "active" et ne peut plus etre re-acceptee
							})
						}),
						NewButton("Décliner", func() {
							render(removeQuest(quests, q.QuestID))
						}),
					),
				)
			case q.Status == "ready" && pending[q.QuestID]:
				detail = container.NewVBox(
					NewWrappingBodyText(q.Description),
					NewWrappingBodyText("Récupération en cours..."),
				)
			case q.Status == "waiting":
				// objectif rempli, mais le joueur n'est pas dans la salle du
				// PNJ : pas de bouton, juste l'indication de ce qu'il reste à
				// faire pour toucher la récompense.
				detail = container.NewVBox(
					NewWrappingBodyText(q.Description),
					NewWrappingBodyText("Objectif rempli — retournez voir "+q.Giver+" pour récupérer votre récompense."),
				)
			case q.Status == "ready":
				detail = container.NewVBox(
					NewWrappingBodyText(q.Description),
					NewWrappingBodyText(q.Giver+" est là : vous pouvez récupérer votre récompense !"),
					NewButton("Récupérer", func() {
						sendQuestCommand(q, quests, func(resp string) {
							var result struct {
								Reward  string `json:"reward"`
								Message string `json:"message"`
							}
							json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &result)

							updated := currentPlayer
							if it, ok := world.Items[result.Reward]; ok {
								updated.Inventory = append(updated.Inventory, it)
							}
							currentPlayer = updated
							refreshCharacter(updated) // affiche la récompense dans le panneau perso

							msg := result.Message
							if msg == "" {
								msg = q.Name + " est terminée."
							}
							showMessage("Quête terminée", msg)
							load() // recharge depuis le serveur : la quête passe en "completed" et ne peut plus etre re-recuperee
						})
					}),
				)
			case q.Status == "completed":
				detail = NewWrappingBodyText("Terminée")
			default:
				detail = NewWrappingBodyText(fmt.Sprintf("En cours — %s\n%s", q.Progress, q.Description))
			}

			list.Add(container.NewVBox(header, detail, Fix(0, 10)))
		}
		list.Refresh()
	}

	load = func() {
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
			fyne.Do(func() { render(quests) })
		}()
	}

	d.Show()
	load()
}
