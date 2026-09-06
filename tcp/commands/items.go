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
