package gui

import (
	"encoding/json"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"

	"the_answer_protocol/parsing"
)

// openItemDialog affiche la description d'un objet de l'inventaire, avec la
// possibilité de le déposer (Drop) et, s'il est équipable, de l'équiper /
// déséquiper + Un seul objet peut être équipé à la fois
func openItemDialog(w fyne.Window, sess *Session, item parsing.Item, player parsing.Player, refreshCharacter func(parsing.Player), changeRoom func(parsing.Location), showMessage func(title, msg string)) {
	body := container.NewVBox()
	var d dialog.Dialog
	var render func()
	sending := false

	dropItem := func() {
		if sending {
			return
		}
		sending = true
		render()
		go func() {
			resp, err := sess.Send("DROP " + item.Name)
			if resp == "" && err == nil {
				fyne.Do(func() {
					sending = false
					render()
				})
				return
			}
			if err != nil {
				fyne.Do(func() {
					sending = false
					showMessage("Erreur", err.Error())
					render()
				})
				return
			}
			if strings.HasPrefix(resp, "ERR") {
				fyne.Do(func() {
					sending = false
					showMessage("Erreur", resp)
					render()
				})
				return
			}
			var newRoom parsing.Location
			if err := json.Unmarshal([]byte(resp), &newRoom); err != nil {
				fyne.Do(func() {
					sending = false
					showMessage("Erreur", "réponse du serveur illisible : "+err.Error())
					render()
				})
				return
			}
			updated := player
			newInv := make([]parsing.Item, 0, len(updated.Inventory))
			removed := false
			for _, invItem := range updated.Inventory {
				if !removed && invItem.Name == item.Name {
					removed = true
					continue
				}
				newInv = append(newInv, invItem)
			}
			updated.Inventory = newInv
			if updated.EquippedItem == item.ID {
				updated.EquippedItem = ""
				updated.Weapon = updated.BaseWeapon
			}
			fyne.Do(func() {
				refreshCharacter(updated)
				changeRoom(newRoom)
				d.Hide()
				showMessage("Objet déposé", item.Name+" a été déposé.")
			})
		}()
	}

	sendCommand := func(cmd string, onSuccess func(weapon string, critical float64)) {
		if sending {
			return
		}
		sending = true
		render()
		go func() {
			resp, err := sess.Send(cmd)
			if resp == "" && err == nil {
				fyne.Do(func() {
					sending = false
					render()
				})
				return
			}
			if err != nil {
				fyne.Do(func() {
					sending = false
					showMessage("Erreur", err.Error())
					render()
				})
				return
			}
			if strings.HasPrefix(resp, "ERR") {
				fyne.Do(func() {
					sending = false
					showMessage("Erreur", resp)
					render()
				})
				return
			}
			var result struct {
				Weapon   string  `json:"weapon"`
				Critical float64 `json:"critical"`
			}
			json.Unmarshal([]byte(strings.TrimPrefix(resp, "OK ")), &result)
			fyne.Do(func() { onSuccess(result.Weapon, result.Critical) })
		}()
	}

	render = func() {
		body.Objects = nil
		body.Add(container.NewCenter(NewTitleText(item.Name, 24)))
		body.Add(Fix(0, 10))
		body.Add(NewWrappingBodyText(item.Description))
		body.Add(Fix(0, 15))

		if sending {
			body.Add(NewWrappingBodyText("Action en cours..."))
			body.Refresh()
			return
		}

		switch {
		case item.Critical != nil && player.EquippedItem == item.ID:
			body.Add(NewWrappingBodyText("Déjà équipé."))
			body.Add(NewButton("Déséquiper", func() {
				sendCommand("UNEQUIP", func(weapon string, critical float64) {
					updated := player
					updated.Weapon = parsing.Weapon{Name: weapon, Critical: critical}
					updated.EquippedItem = ""
					player = updated
					refreshCharacter(updated) // met a jour les stats affichees (chances de critique)
					d.Hide()
					showMessage("Équipement", item.Name+" a été retiré.")
				})
			}))
		case item.Critical != nil:
			body.Add(NewButton("Équiper", func() {
				sendCommand("EQUIP "+item.ID, func(weapon string, critical float64) {
					updated := player
					updated.Weapon = parsing.Weapon{Name: weapon, Critical: critical}
					updated.EquippedItem = item.ID
					player = updated
					refreshCharacter(updated) // met a jour les stats affichees (chances de critique)
					d.Hide()
					showMessage("Équipement", item.Name+" est maintenant équipé.")
				})
			}))
		}
		body.Add(NewButton("Drop", dropItem))
		body.Refresh()
	}

	render()

	d = dialog.NewCustom("", "Fermer", body, w)
	d.Resize(fyne.NewSize(420, 260))
	d.Show()
}
