package gui

import (
	"encoding/json"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"

	"the_answer_protocol/parsing"
)

// openItemDialog affiche la description d'un objet équipable de
// l'inventaire (Critical non nil : une arme) avec un bouton Équiper qui
// envoie EQUIP au serveur et met à jour les statistiques de combat du
// joueur (chances de critique) dans le panneau perso.
func openItemDialog(w fyne.Window, sess *Session, item parsing.Item, player parsing.Player, refreshCharacter func(parsing.Player), showMessage func(title, msg string)) {
	var d dialog.Dialog

	equipBtn := NewButton("Équiper", func() {
		go func() {
			resp, err := sess.Send("EQUIP " + item.ID)
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
			var result struct {
				Weapon   string  `json:"weapon"`
				Critical float64 `json:"critical"`
			}
			json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &result)

			updated := player
			updated.Weapon = parsing.Weapon{Name: result.Weapon, Critical: result.Critical}

			fyne.Do(func() {
				refreshCharacter(updated) // met a jour les stats affichees (chances de critique)
				d.Hide()
				showMessage("Équipement", item.Name+" est maintenant équipé.")
			})
		}()
	})

	content := container.NewVBox(
		container.NewCenter(NewTitleText(item.Name, 24)),
		Fix(0, 10),
		NewWrappingBodyText(item.Description),
		Fix(0, 15),
		equipBtn,
	)

	d = dialog.NewCustom("", "Fermer", content, w)
	d.Resize(fyne.NewSize(420, 260))
	d.Show()
}
