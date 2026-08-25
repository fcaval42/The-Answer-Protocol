package tcp

import (
	"fmt"
	"the_answer_protocol/parsing"
	"encoding/json"
)

func Struct_in_string(world parsing.WorldFile) string {
	data, err := json.Marshal(world)
	if err != nil {
		fmt.Println("Error during conversion", err)
		return ""
	}

	return string(data)
}

func String_in_struct(word string) parsing.WorldFile {

	var w parsing.WorldFile

	err := json.Unmarshal([]byte(word), &w)

	if err != nil {
		fmt.Println("Error during conversion", err)
		return w
	}

	return w
}
