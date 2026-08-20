package parsing

import (
	"fmt"
	"os"
	"errors"
)

/**
Ce fichier lancer les deux parsing celui du monde et des entites
renvoie les deux monde plus les potentiels erreurs a utiliser pour charger le
monde
**/

func Rally(world_path string, entities_path string) (WorldFile, EntitiesFile, []error) {
	var er error
	var err []error
	var errs []error
	var world WorldFile
	var entities EntitiesFile

	// *************************************************************************
	//                           Entities Parsing
	// *************************************************************************

	entities_yaml, er := os.ReadFile(entities_path)
	if (er != nil){
		fmt.Println(err)
		errs = append(errs, er)
	} else {
		entities, err = ParseEntities(entities_yaml)
		if (err != nil){
			fmt.Println("\n---Errors of entities.yaml---")
			fmt.Println(errors.Join(err...))
			errs = append(errs, err...)
		}
	}

	// *************************************************************************
	//                            World Parsing
	// *************************************************************************

	world_yaml, er := os.ReadFile(world_path)
	if (er != nil){
		fmt.Println(err)
		errs = append(errs, er)
	} else {
		world, err = ParseWorld(world_yaml)
		if (err != nil){
			fmt.Println("\n---Errors of world.yaml---")
			fmt.Println(errors.Join(err...))
			errs = append(errs, err...)
		}
	}

	if (len(errs) != 0){
		fmt.Println("\nNumber of errors:", len(errs))
	}

	return world, entities, errs
}
