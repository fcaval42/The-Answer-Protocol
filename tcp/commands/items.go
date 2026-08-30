package commands

import (
	"encoding/json"
	"fmt"
	"the_answer_protocol/parsing"
	// "errors"
)

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
