package tcp

import (
	"fmt"
	"encoding/json"
)

func ToString[T any](data T) string {
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error during conversion", err)
		return ""
	}
	return string(b)
}

func FromString[T any](s string) T {
	var v T
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		fmt.Println("Error during conversion", err)
	}
	return v
}


//Utilisation example:
// worldStr := toString(world)
// w := fromString[parsing.WorldFile](worldStr)

// locStr := toString(location)
// loc := fromString[parsing.Location](locStr)