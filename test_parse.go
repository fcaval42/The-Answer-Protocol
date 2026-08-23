package main

import (
	"fmt"
	"os"
	"the_answer_protocol/parsing"
)

func main() {
	world_path := "yaml/world.yaml"
	world, err := parsing.Rally(world_path)

	if err != nil {
		os.Exit(1)
	} else {
		fmt.Sprintf("\n\n", world)
	}
}
