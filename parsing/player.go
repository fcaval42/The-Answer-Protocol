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
	Name         string  `yaml:"name"`
	HP           int     `yaml:"hp"`
	MaxHP        int     `yaml:"-"`
	Weapon       Weapon  `yaml:"weapon"`
	Attack       int     `yaml:"attack"`
	Defense      int     `yaml:"defense"`
	Respawn      Respawn `yaml:"respawn"`
	Inventory    []Item
	Quests       []PlayerQuest
	Connected    bool
	BaseWeapon   Weapon `yaml:"-"`
	EquippedItem string `yaml:"-"`
}

func validatePlayers(wf WorldFile) []error {
	var errs []error

	for id, player := range wf.Players {
		if (id != player.Name) {
			errs = append(errs, fmt.Errorf("The ID and Name are different for %s - %s", id, player.Name))
		}
		if (player.HP < 1) {
			errs = append(errs, fmt.Errorf("HP of %s is negative %d", id, player.HP))
		}
		if (player.Weapon.Critical < 0) {
			errs = append(errs, fmt.Errorf("Critical of %s is negative %f", id, player.Weapon.Critical))
		}
		if (player.Attack < 1) {
			errs = append(errs, fmt.Errorf("Attack of %s is negative %d", id, player.Attack))
		}
		if (player.Defense < 1) {
			errs = append(errs, fmt.Errorf("Defense of %s is negative %d", id, player.Defense))
		}
		if (player.Respawn.HP < 1) {
			errs = append(errs, fmt.Errorf("Respawn HP of %s is negative %d", id, player.Respawn.HP))
		}
		if (player.Respawn.HP > player.HP) {
			errs = append(errs, fmt.Errorf("Respawn HP %d of %s is higher than %d max HP", player.Respawn.HP, id, player.HP))
		}

		loc_ok := false

		for l, _ := range wf.World.Locations {
			if (l == player.Respawn.Location){
				loc_ok = true
			}
		}

		if (loc_ok == false) {
			errs = append(errs, fmt.Errorf("Respawn rental for %s doesn't exist %s", id, player.Respawn.Location))
		}
	}

	return errs
}
