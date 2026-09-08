package gui

import (
	"encoding/json"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)


type questEntry struct {
	QuestID     string `json:"quest_id"`
	Name        string `json:"name"`
	Giver       string `json:"giver"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Progress    string `json:"progress"`
}


func questStatusLabel(q questEntry) string {
	switch q.Status {
	case "available":
		return "Disponible auprès de " + q.Giver + "."
	case "ready", "waiting":
		return "Objectif rempli — retournez voir " + q.Giver + " pour récupérer votre récompense."
	case "completed":
		return "Terminée."
	default:
		if q.Progress != "" {
			return "En cours — " + q.Progress
		}
		return "En cours."
	}
}

func openQuestsDialog(w fyne.Window, sess *Session, showMessage func(title, msg string)) {
	list := container.NewVBox()
	scroll := container.NewVScroll(list)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("", "Fermer", container.NewVBox(
		container.NewCenter(NewTitleText("Quêtes", 28)),
		Fix(0, 15),
		scroll,
	), w)
	d.Resize(fyne.NewSize(560, 520))

	render := func(quests []questEntry) {
		list.Objects = nil
		if len(quests) == 0 {
			list.Add(NewWrappingBodyText("Aucune quête pour le moment."))
		}
		for _, q := range quests {
			header := NewTitleText(fmt.Sprintf("%s (%s)", q.Name, q.Giver), 18)
			detail := NewWrappingBodyText(questStatusLabel(q))
			list.Add(container.NewVBox(header, detail, Fix(0, 10)))
		}
		list.Refresh()
	}

	load := func() {
		go func() {
			resp, err := sess.Send("QUESTS")
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
