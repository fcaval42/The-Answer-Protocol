package commands

import (
	"encoding/json"
	"fmt"
	"the_answer_protocol/parsing"
	// "errors"
)

// TAKE, DROP, INVENTORY

// pour take je pense quil faudra cest trois param
func Take(world parsing.WorldFile, item string, Player parsing.Player_t) (string, error) {
	_, exist := world.Items[item]

	if exist {
		return fmt.Sprintf("OK taken=%s", item), nil
	}
	return "", fmt.Errorf("ERR 404 ITEM_NOT_FOUND: %q", item)

}

// Inventory renvoie la reponse "OK [...]" listant les ids des items
// actuellement dans l'inventaire de playerName
func Inventory(world parsing.WorldFile, playerName string) (string, error) {
	_, player, ok := findPlayerByName(world, playerName)
	if !ok {
		return "", fmt.Errorf("player %q is unknown", playerName)
	}

	ids := make([]string, 0, len(player.Inventory))
	for _, item := range player.Inventory {
		ids = append(ids, item.ID)
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}
