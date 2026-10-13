package parsing

import (
	"errors"
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

type Player_t struct {
	Name         string
	Hp           float64
	Localisation string
	Inventory    []string
}

type WorldMap struct {
	Locations map[string]Location `yaml:"locations"`
}

type WorldFile struct {
	World   WorldMap          `yaml:"world"`
	Items   map[string]Item   `yaml:"items"`
	NPCs    map[string]NPC    `yaml:"npcs"`
	Enemies map[string]Enemy  `yaml:"enemies"`
	Players map[string]Player `yaml:"players"`
	Quests  map[string]Quest  `yaml:"quests"`
}

func ParseWorld(worldYAML []byte) (WorldFile, []error) {
	var wf WorldFile

	if err := yaml.Unmarshal(worldYAML, &wf); err != nil {
		return WorldFile{}, []error{fmt.Errorf("parsing world: %w", err)}
	}

	for id, loc := range wf.World.Locations {
		loc.ID = id
		loc.Background = resolveBackground(loc.Background)
		wf.World.Locations[id] = loc
	}

	for id, item := range wf.Items {
		item.ID = id
		wf.Items[id] = item
	}

	for id, npc := range wf.NPCs {
		npc.ID = id
		wf.NPCs[id] = npc
	}

	for id, enemy := range wf.Enemies {
		enemy.ID = id
		wf.Enemies[id] = enemy
	}

	for id, quest := range wf.Quests {
		quest.ID = id
		wf.Quests[id] = quest
	}

	for id, player := range wf.Players {
		player.MaxHP = player.HP
		player.BaseWeapon = player.Weapon
		wf.Players[id] = player
	}

	err := wf.Validate()
	if err != nil {
		return wf, err
	}

	return wf, nil
}

func (wf WorldFile) Validate() []error {
	var errs []error

	errs = append(errs, validateLocations(wf)...)
	errs = append(errs, validateItems(wf)...)
	errs = append(errs, validateNPCs(wf)...)
	errs = append(errs, validatePlayers(wf)...)
	errs = append(errs, validateEnemies(wf)...)
	errs = append(errs, validateQuests(wf)...)

	return errs
}

func Rally(world_path string) (WorldFile, []error) {
	var world WorldFile
	var errs []error

	world_yaml, er := os.ReadFile(world_path)
	if len(world_yaml) == 0{
		fmt.Println("\n---Errors of world.yaml---")
		fmt.Println("YAML VIDE")
		errs = append(errs, fmt.Errorf("YAML VIDE"))
	}

	if er != nil {
		errs = append(errs, er)
	} else {
		var err []error
		world, err = ParseWorld(world_yaml)
		if err != nil {
			fmt.Println("\n---Errors of world.yaml---")
			fmt.Println(errors.Join(err...))
			errs = append(errs, err...)
		}
	}

	if len(errs) != 0 {
		fmt.Println("\nNumber of errors:", len(errs))
	}

	return world, errs
}
