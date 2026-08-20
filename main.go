package main

import (
	"fmt"
	"os"

	"the_answer_protocol/parsing"
)

func main() {
	worldYAML, err := os.ReadFile("world.yaml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	world, err := parsing.ParseWorld(worldYAML)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n\n", world)

	if err := world.Validate(); err != nil {
		fmt.Println("world validation errors:")
		fmt.Println(err)
		fmt.Println()
	} else {
		fmt.Println("world validation: OK")
		fmt.Println()
	}

	entitiesYAML, err := os.ReadFile("entities.yaml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	entities, err := parsing.ParseEntities(entitiesYAML)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", entities)
}
