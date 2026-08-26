package parsing

import "fmt"

type Spawn struct {
	NPCType string `yaml:"npc_type"`
	Count   int    `yaml:"count"`
}

type Location struct {
	ID          string            `yaml:"-"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Exits       map[string]string `yaml:"exits"`
	Items       []string          `yaml:"items"`
	NPCs        []string          `yaml:"npcs"`
	Enemies     []string          `yaml:"enemies"`
	Spawns      []Spawn           `yaml:"spawns"`
	Players     []Player
}

func validateLocations(wf WorldFile) []error {
	var errs []error

	seenLocationNames := make(map[string]string)
	placedNPCs := make(map[string]string)
	placedEnemies := make(map[string]string)
	for id, loc := range wf.World.Locations {
		if isBlank(id) {
			errs = append(errs, fmt.Errorf("a location has an empty id"))
		}
		if isBlank(loc.Name) {
			errs = append(errs, fmt.Errorf("location %q: name is empty", id))
		} else {
			if slug := slugifyLocationName(loc.Name); slug != id {
				errs = append(errs, fmt.Errorf("location %q: name %q does not match its id (expected id %q)", id, loc.Name, slug))
			}
			if otherID, ok := seenLocationNames[loc.Name]; ok {
				errs = append(errs, fmt.Errorf("locations %q and %q share the same name %q", otherID, id, loc.Name))
			} else {
				seenLocationNames[loc.Name] = id
			}
		}
		if isBlank(loc.Description) {
			errs = append(errs, fmt.Errorf("location %q: description is empty", id))
		}

		if len(loc.Exits) == 0 {
			errs = append(errs, fmt.Errorf("location %q: has no exits", id))
		}
		for dir, dest := range loc.Exits {
			if isBlank(dir) {
				errs = append(errs, fmt.Errorf("location %q: an exit has an empty direction", id))
			}
			if isBlank(dest) {
				errs = append(errs, fmt.Errorf("location %q: exit %q has an empty destination", id, dir))
				continue
			}
			if dest == id {
				errs = append(errs, fmt.Errorf("location %q: exit %q points to itself", id, dir))
			}
			if _, ok := wf.World.Locations[dest]; !ok {
				errs = append(errs, fmt.Errorf("location %q: exit %q points to unknown location %q", id, dir, dest))
			}
		}

		for _, itemID := range loc.Items {
			if isBlank(itemID) {
				errs = append(errs, fmt.Errorf("location %q: an item reference is empty", id))
				continue
			}
			if _, ok := wf.Items[itemID]; !ok {
				errs = append(errs, fmt.Errorf("location %q: unknown item %q", id, itemID))
			}
		}
		for _, dup := range duplicatesIn(loc.Items) {
			errs = append(errs, fmt.Errorf("location %q: item %q is listed more than once", id, dup))
		}

		for _, npcID := range loc.NPCs {
			if isBlank(npcID) {
				errs = append(errs, fmt.Errorf("location %q: an npc reference is empty", id))
				continue
			}
			if _, ok := wf.NPCs[npcID]; !ok {
				errs = append(errs, fmt.Errorf("location %q: unknown npc %q", id, npcID))
			} else if otherID, ok := placedNPCs[npcID]; ok && otherID != id {
				errs = append(errs, fmt.Errorf("npc %q is placed in both %q and %q", npcID, otherID, id))
			} else {
				placedNPCs[npcID] = id
			}
		}
		for _, dup := range duplicatesIn(loc.NPCs) {
			errs = append(errs, fmt.Errorf("location %q: npc %q is listed more than once", id, dup))
		}

		for _, enemyID := range loc.Enemies {
			if isBlank(enemyID) {
				errs = append(errs, fmt.Errorf("location %q: an enemy reference is empty", id))
				continue
			}
			if _, ok := wf.Enemies[enemyID]; !ok {
				errs = append(errs, fmt.Errorf("location %q: unknown enemy %q", id, enemyID))
			} else if otherID, ok := placedEnemies[enemyID]; ok && otherID != id {
				errs = append(errs, fmt.Errorf("enemy %q is placed in both %q and %q", enemyID, otherID, id))
			} else {
				placedEnemies[enemyID] = id
			}
		}
		for _, dup := range duplicatesIn(loc.Enemies) {
			errs = append(errs, fmt.Errorf("location %q: enemy %q is listed more than once", id, dup))
		}

		inNPCs := make(map[string]bool, len(loc.NPCs))
		for _, npcID := range loc.NPCs {
			inNPCs[npcID] = true
		}
		for _, enemyID := range loc.Enemies {
			if inNPCs[enemyID] {
				errs = append(errs, fmt.Errorf("location %q: %q is listed in both npcs and enemies", id, enemyID))
			}
		}

		for _, spawn := range loc.Spawns {
			if isBlank(spawn.NPCType) {
				errs = append(errs, fmt.Errorf("location %q: a spawn has an empty npc_type", id))
				continue
			}
			if _, ok := wf.NPCs[spawn.NPCType]; !ok {
				errs = append(errs, fmt.Errorf("location %q: spawn references unknown npc_type %q", id, spawn.NPCType))
			}
			if spawn.Count <= 0 {
				errs = append(errs, fmt.Errorf("location %q: spawn %q has a non-positive count %d", id, spawn.NPCType, spawn.Count))
			}
		}
	}

	if len(wf.World.Locations) > 0 && !worldHasLoop(wf.World.Locations) {
		errs = append(errs, fmt.Errorf("the world has no loop: following exits forward never leads back to an already-visited location, dead ends are fine but the map as a whole needs at least one circuit"))
	}

	return errs
}

func worldHasLoop(locations map[string]Location) bool {
	const (
		unvisited = iota
		visiting
		done
	)
	state := make(map[string]int, len(locations))
	found := false

	var visit func(id string)
	visit = func(id string) {
		if found {
			return
		}
		state[id] = visiting
		for _, dest := range locations[id].Exits {
			if dest == id {
				continue
			}
			if _, ok := locations[dest]; !ok {
				continue
			}
			switch state[dest] {
			case unvisited:
				visit(dest)
			case visiting:
				found = true
				return
			}
		}
		state[id] = done
	}

	for id := range locations {
		if state[id] == unvisited {
			visit(id)
		}
		if found {
			break
		}
	}

	return found
}
