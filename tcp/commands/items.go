package commands

import (
	"encoding/json"
	"fmt"
	. "the_answer_protocol/parsing"
	// "errors"
)

func Inventory(player Player) (string, error) {

	items := make([]string, 0, len(player.Inventory))
	for _, item := range player.Inventory {
		items = append(items, item.ID)
	}

	data, err := json.Marshal(items)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}

// Equip remplace l'arme du joueur par un objet de son inventaire. Seuls les
// objets portant une valeur de critique (armes) sont équipables ; les autres
// objets (Chroma, Pictos, Lumina...) sont juste consultables.
func Equip(player Player, itemID string) (string, Player, error) {
	var item Item
	found := false
	for _, it := range player.Inventory {
		if it.ID == itemID {
			item = it
			found = true
			break
		}
	}
	if !found {
		return "ERR 404 ITEM_NOT_FOUND", player, nil
	}
	if item.Critical == nil {
		return "ERR 406 NOT_EQUIPABLE", player, nil
	}

	player.Weapon = Weapon{Name: item.Name, Critical: *item.Critical}

	data, err := json.Marshal(struct {
		Weapon   string  `json:"weapon"`
		Critical float64 `json:"critical"`
	}{
		Weapon:   player.Weapon.Name,
		Critical: player.Weapon.Critical,
	})
	if err != nil {
		return "", player, err
	}

	return fmt.Sprintf("OK %s", data), player, nil
}
