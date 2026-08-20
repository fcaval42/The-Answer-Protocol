package parsing

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

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
	Spawns      []Spawn           `yaml:"spawns"`
}

type Item struct {
	ID          string `yaml:"-"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Obtainable  bool   `yaml:"obtainable"`
}

type NPCStats struct {
	HP int `yaml:"hp"`
}

type NPC struct {
	ID          string   `yaml:"-"`
	Name        string   `yaml:"name"`
	Role        string   `yaml:"role"`
	Description string   `yaml:"description"`
	Dialogue    []string `yaml:"dialogue"`
	Stats       NPCStats `yaml:"stats"`
}

type QuestObjective struct {
	Type      string `yaml:"type"`
	Item      string `yaml:"item,omitempty"`
	NPC       string `yaml:"npc,omitempty"`
	DeliverTo string `yaml:"deliver_to,omitempty"`
}

type QuestReward struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

type Quest struct {
	ID          string         `yaml:"-"`
	Name        string         `yaml:"name"`
	Giver       string         `yaml:"giver"`
	Description string         `yaml:"description"`
	Objective   QuestObjective `yaml:"objective"`
	Reward      QuestReward    `yaml:"reward"`
}

type World struct {
	Locations map[string]Location `yaml:"locations"`
}

type WorldFile struct {
	World  World            `yaml:"world"`
	Items  map[string]Item  `yaml:"items"`
	NPCs   map[string]NPC   `yaml:"npcs"`
	Quests map[string]Quest `yaml:"quests"`
}

func ParseWorld(worldYAML []byte) (WorldFile, error) {
	var wf WorldFile
	if err := yaml.Unmarshal(worldYAML, &wf); err != nil {
		return WorldFile{}, fmt.Errorf("parsing world: %w", err)
	}

	for id, loc := range wf.World.Locations {
		loc.ID = id
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
	for id, quest := range wf.Quests {
		quest.ID = id
		wf.Quests[id] = quest
	}

	return wf, nil
}

func (wf WorldFile) Validate() error {
	var errs []error

	for id, loc := range wf.World.Locations {
		for dir, dest := range loc.Exits {
			if _, ok := wf.World.Locations[dest]; !ok {
				errs = append(errs, fmt.Errorf("location %q: exit %q points to unknown location %q", id, dir, dest))
			}
		}
		for _, itemID := range loc.Items {
			if _, ok := wf.Items[itemID]; !ok {
				errs = append(errs, fmt.Errorf("location %q: unknown item %q", id, itemID))
			}
		}
		for _, npcID := range loc.NPCs {
			if _, ok := wf.NPCs[npcID]; !ok {
				errs = append(errs, fmt.Errorf("location %q: unknown npc %q", id, npcID))
			}
		}
		for _, spawn := range loc.Spawns {
			if _, ok := wf.NPCs[spawn.NPCType]; !ok {
				errs = append(errs, fmt.Errorf("location %q: spawn references unknown npc_type %q", id, spawn.NPCType))
			}
		}
	}

	for id, q := range wf.Quests {
		if _, ok := wf.NPCs[q.Giver]; !ok {
			errs = append(errs, fmt.Errorf("quest %q: giver %q is not a known npc", id, q.Giver))
		}

		switch q.Objective.Type {
		case "fetch":
			if _, ok := wf.Items[q.Objective.Item]; !ok {
				errs = append(errs, fmt.Errorf("quest %q: objective item %q is not a known item", id, q.Objective.Item))
			}
			if q.Objective.DeliverTo != "" {
				if _, ok := wf.World.Locations[q.Objective.DeliverTo]; !ok {
					errs = append(errs, fmt.Errorf("quest %q: objective deliver_to %q is not a known location", id, q.Objective.DeliverTo))
				}
			}
		case "defeat":
			if _, ok := wf.NPCs[q.Objective.NPC]; !ok {
				errs = append(errs, fmt.Errorf("quest %q: objective npc %q is not a known npc", id, q.Objective.NPC))
			}
		default:
			errs = append(errs, fmt.Errorf("quest %q: unknown objective type %q", id, q.Objective.Type))
		}

		if q.Reward.Type == "item" {
			if _, ok := wf.Items[q.Reward.Value]; !ok {
				errs = append(errs, fmt.Errorf("quest %q: reward item %q is not a known item", id, q.Reward.Value))
			}
		}
	}

	return errors.Join(errs...)
}
