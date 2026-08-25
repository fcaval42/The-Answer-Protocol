package parsing

import "fmt"

type Weapon struct {
	Name     string  `yaml:"name"`
	Critical float64 `yaml:"critical"`
}

type Respawn struct {
	HP       int    `yaml:"hp"`
	Location string `yaml:"location"`
}

type Player struct {
	ID      string  `yaml:"-"`
	Name    string  `yaml:"name"`
	HP      int     `yaml:"hp"`
	Weapon  Weapon  `yaml:"weapon"`
	Attack  int     `yaml:"attack"`
	Defense int     `yaml:"defense"`
	Respawn Respawn `yaml:"respawn"`
	Location Location
	Inventory []Item
}

func validatePlayers(wf WorldFile) []error {
	var errs []error

	seenPlayerNames := make(map[string]string)
	for id, player := range wf.Players {
		if isBlank(id) {
			errs = append(errs, fmt.Errorf("a player has an empty id"))
		}
		if isBlank(player.Name) {
			errs = append(errs, fmt.Errorf("player %q: name is empty", id))
		} else if otherID, ok := seenPlayerNames[player.Name]; ok {
			errs = append(errs, fmt.Errorf("players %q and %q share the same name %q", otherID, id, player.Name))
		} else {
			seenPlayerNames[player.Name] = id
		}
		if player.HP < 0 {
			errs = append(errs, fmt.Errorf("player %q: hp %d is negative", id, player.HP))
		}
		if player.Attack < 0 {
			errs = append(errs, fmt.Errorf("player %q: attack %d is negative", id, player.Attack))
		}
		if player.Defense < 0 {
			errs = append(errs, fmt.Errorf("player %q: defense %d is negative", id, player.Defense))
		}
		if isBlank(player.Weapon.Name) {
			errs = append(errs, fmt.Errorf("player %q: weapon name is empty", id))
		}
		if player.Weapon.Critical < 0 || player.Weapon.Critical > 1 {
			errs = append(errs, fmt.Errorf("player %q: weapon critical %.2f must be between 0 and 1", id, player.Weapon.Critical))
		}
		if player.Respawn.HP < 0 {
			errs = append(errs, fmt.Errorf("player %q: respawn hp %d is negative", id, player.Respawn.HP))
		}
		if isBlank(player.Respawn.Location) {
			errs = append(errs, fmt.Errorf("player %q: respawn location is empty", id))
		} else if _, ok := wf.World.Locations[player.Respawn.Location]; !ok {
			errs = append(errs, fmt.Errorf("player %q: respawn location %q is not a known location", id, player.Respawn.Location))
		}
	}

	return errs
}
