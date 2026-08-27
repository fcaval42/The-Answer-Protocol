package tcp

import (
	"fmt"
	"the_answer_protocol/parsing"
	"encoding/json"
)

func worldInString(world parsing.WorldFile) string {
	data, err := json.Marshal(world)
	if err != nil {
		fmt.Println("Error during conversion", err)
		return ""
	}

	return string(data)
}

func stringInWorld(word string) parsing.WorldFile {

	var w parsing.WorldFile

	err := json.Unmarshal([]byte(word), &w)

	if err != nil {
		fmt.Println("Error during conversion", err)
		return w
	}

	return w
}



func mapInString(world parsing.Location) string {
	data, err := json.Marshal(world)
	if err != nil {
		fmt.Println("Error during conversion", err)
		return ""
	}

	return string(data)
}

func stringInMap(word string) parsing.Location {

	var w parsing.Location

	err := json.Unmarshal([]byte(word), &w)

	if err != nil {
		fmt.Println("Error during conversion", err)
		return w
	}

	return w
}

