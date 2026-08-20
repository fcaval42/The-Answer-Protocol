package parsing

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Respawn struct {
	HP       int    `yaml:"hp"`
	Location string `yaml:"location"`
}

type PlayerEntity struct {
	ID      string  `yaml:"-"`
	Name    string  `yaml:"name"`
	HP      int     `yaml:"hp"`
	Attack  int     `yaml:"attack"`
	Defense int     `yaml:"defense"`
	Speed   int     `yaml:"speed"`
	Respawn Respawn `yaml:"respawn"`
}

type EnemyEntity struct {
	ID         string `yaml:"-"`
	Name       string `yaml:"name"`
	HP         int    `yaml:"hp"`
	Attack     int    `yaml:"attack"`
	Defense    int    `yaml:"defense"`
	Speed      int    `yaml:"speed"`
	Difficulty string `yaml:"difficulty"`
}

type Entities struct {
	Players map[string]PlayerEntity `yaml:"players"`
	Enemies map[string]EnemyEntity  `yaml:"enemies"`
}

type EntitiesFile struct {
	Entities Entities `yaml:"entities"`
}

func ParseEntities(entitiesYAML []byte) (EntitiesFile, error) {
	var ef EntitiesFile
	if err := yaml.Unmarshal(entitiesYAML, &ef); err != nil {
		return EntitiesFile{}, fmt.Errorf("parsing entities: %w", err)
	}

	for id, p := range ef.Entities.Players {
		p.ID = id
		ef.Entities.Players[id] = p
	}
	for id, e := range ef.Entities.Enemies {
		e.ID = id
		ef.Entities.Enemies[id] = e
	}

	return ef, nil
}
