package parsing

import "fmt"

// NPC is a non-combat character: dialogue and/or quest-giver. It never has
// combat stats — those belong to Enemy instead.
type NPC struct {
	ID          string   `yaml:"-"`
	Name        string   `yaml:"name"`
	Role        string   `yaml:"role"`
	Description string   `yaml:"description"`
	Dialogue    []string `yaml:"dialogue"`
}

func validateNPCs(wf WorldFile) []error {
	var errs []error

	seenNPCNames := make(map[string]string) // display name -> npc id
	for id, npc := range wf.NPCs {
		if isBlank(id) {
			errs = append(errs, fmt.Errorf("an npc has an empty id"))
		}
		if isBlank(npc.Name) {
			errs = append(errs, fmt.Errorf("npc %q: name is empty", id))
		} else if otherID, ok := seenNPCNames[npc.Name]; ok {
			errs = append(errs, fmt.Errorf("npcs %q and %q share the same name %q", otherID, id, npc.Name))
		} else {
			seenNPCNames[npc.Name] = id
		}
		if isBlank(npc.Description) {
			errs = append(errs, fmt.Errorf("npc %q: description is empty", id))
		}
		if isBlank(npc.Role) {
			errs = append(errs, fmt.Errorf("npc %q: role is empty", id))
		}
		if len(npc.Dialogue) == 0 {
			errs = append(errs, fmt.Errorf("npc %q: dialogue is empty", id))
		}
		for i, line := range npc.Dialogue {
			if isBlank(line) {
				errs = append(errs, fmt.Errorf("npc %q: dialogue line %d is empty", id, i))
			}
		}
	}

	return errs
}
