package parsing

import "fmt"

type Enemy struct {
	ID          string   `yaml:"-"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Dialogue    []string `yaml:"dialogue"`
	HP          int      `yaml:"hp"`
	Attack      int      `yaml:"attack"`
	Defense     int      `yaml:"defense"`
	Counter     int      `yaml:"counter"`
	Difficulty  string   `yaml:"difficulty"`
}

func validateEnemies(wf WorldFile) []error {
	var errs []error

	seenEnemyNames := make(map[string]string)
	for id, enemy := range wf.Enemies {
		if isBlank(id) {
			errs = append(errs, fmt.Errorf("an enemy has an empty id"))
		}
		if isBlank(enemy.Name) {
			errs = append(errs, fmt.Errorf("enemy %q: name is empty", id))
		} else if otherID, ok := seenEnemyNames[enemy.Name]; ok {
			errs = append(errs, fmt.Errorf("enemies %q and %q share the same name %q", otherID, id, enemy.Name))
		} else {
			seenEnemyNames[enemy.Name] = id
		}
		if isBlank(enemy.Description) {
			errs = append(errs, fmt.Errorf("enemy %q: description is empty", id))
		}
		if isBlank(enemy.Difficulty) {
			errs = append(errs, fmt.Errorf("enemy %q: difficulty is empty", id))
		}
		if len(enemy.Dialogue) == 0 {
			errs = append(errs, fmt.Errorf("enemy %q: dialogue is empty", id))
		}
		for i, line := range enemy.Dialogue {
			if isBlank(line) {
				errs = append(errs, fmt.Errorf("enemy %q: dialogue line %d is empty", id, i))
			}
		}
		if enemy.HP < 0 {
			errs = append(errs, fmt.Errorf("enemy %q: hp %d is negative", id, enemy.HP))
		}
		if enemy.Attack < 0 {
			errs = append(errs, fmt.Errorf("enemy %q: attack %d is negative", id, enemy.Attack))
		}
		if enemy.Defense < 0 {
			errs = append(errs, fmt.Errorf("enemy %q: defense %d is negative", id, enemy.Defense))
		}
		if enemy.Counter < 0 {
			errs = append(errs, fmt.Errorf("enemy %q: counter %d is negative", id, enemy.Counter))
		}
	}

	return errs
}
