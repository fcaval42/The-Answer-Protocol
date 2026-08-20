package parsing

// ca viens
import ("fmt"
		"os")

func Rally(world_path string, entities_path string) (WorldFile, EntitiesFile, []error) {
	var errs []error
	var world WorldFile
	var entities EntitiesFile

	entities_yaml, err := os.ReadFile(entities_path)
	if (err != nil){
		fmt.Println(err)
		errs = append(errs, err)
	} else {
		entities, err = ParseEntities(entities_yaml)
		if (err != nil){
			fmt.Println(err)
			errs = append(errs, err)
		}
	}

	world_yaml, err := os.ReadFile(world_path)
	if (err != nil){
		fmt.Println(err)
		errs = append(errs, err)
	} else {
		world, err = ParseWorld(world_yaml)
		if (err != nil){
			fmt.Println(err)
			errs = append(errs, err)
		}
	}

	return world, entities, errs
}
