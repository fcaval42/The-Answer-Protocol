package main

import ("fmt"
		"the_answer_protocol/parsing")

func main(){
	world, entities, err := parsing.Rally("world.yaml", "entities.yaml")
	if (err != nil){
		fmt.Println("\n\n", err)
	} else{
		fmt.Println("\n\n", world)
		fmt.Println("\n\n", entities)
	}
}
