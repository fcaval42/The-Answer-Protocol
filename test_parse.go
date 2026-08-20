package main

import ("fmt"
		"os"
		"the_answer_protocol/parsing")

func main() {
	world_path := "yaml/world.yaml"
	entities_path := "yaml/entities.yaml"
	world, entities, err := parsing.Rally(world_path, entities_path)

	if (err != nil){
		os.Exit(1)
	} else {

		fmt.Sprintf("\n\n", world)
		fmt.Sprintf("\n\n", entities)
	}
}
