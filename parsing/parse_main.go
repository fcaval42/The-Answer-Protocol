package parsing

// ca viens
import ("fmt"
		"os")

func rally(world_path string, entities_path string){

	entities_yaml, err := os.ReadFile(entities_path)
	if (err != nil){
		fmt.Println(err)
		return
	}
	entities, err := ParseEntities(entities_yaml)
	if (err != nil){
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", entities)

	world_yaml, err := os.ReadFile(world_path)
	if (err != nil){
		fmt.Println(err)
		return
	}
	world, err := ParseWorld(world_yaml)
	if (err != nil){
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", world)
}


func main(){
	rally("world.yaml", "entities.yaml")
}
