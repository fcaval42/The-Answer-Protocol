package parsing

import ("gopkg.in/yaml.v3"
		"os"
)

type Spawn struct {
        NPCType string
        Count   int
}

type Location struct {
        Name        string
        Description string
        Exits       map[string]Location
        Items       []string
        Spawns      []Spawn
        NPCs        []string
}

type Item struct {
        Name        string
        Description string
        Obtainable  bool
}

type NPCStats struct {
        HP int
}

type NPC struct {
        Name        string
        Role        string
        Description string
        Dialogue    []string
        Stats       NPCStats
}

type World struct {
        Locations map[string]Location
        Items     map[string]Item
        NPCs      map[string]NPC
}

type worldFile struct {
        World World
        Items map[string]Item
        NPCs  map[string]NPC
}

func parsing_world(file_path string) worldFile{
	var world_file worldFile
	os.ReadFile(file_path)

	return world_file
}
