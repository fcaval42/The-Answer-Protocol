package commands

import (
	"fmt"
	"the_answer_protocol/parsing"
)

// TAKE, DROP, INVENTORY


// pour take je pense quil faudra cest deux param
func take(world parsing.WorldFile, item string) string {

	print(world)

	return (fmt.Fprint("OK taken %s", item))
}
