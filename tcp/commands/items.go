package commands

import (
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
