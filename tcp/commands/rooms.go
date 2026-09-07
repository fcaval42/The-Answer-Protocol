package commands

// LOOK

import (
	"encoding/json"
	"fmt"
	. "the_answer_protocol/parsing"
)

type lookRoom struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}

type lookResponse struct {
	Room    lookRoom `json:"room"`
	Players []string `json:"players"`
	Items   []string `json:"items,omitempty"`
	NPCs    []string `json:"npcs,omitempty"`
	Enemies []string `json:"enemies,omitempty"`
}

func Look(room Location) (string, error) {
	players := make([]string, 0, len(room.Players))
	for _, p := range room.Players {
		players = append(players, p.Name)
	}

	data, err := json.Marshal(lookResponse{
		Room: lookRoom{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			Exits:       room.Exits,
		},
		Players: players,
		Items:   room.Items,
		NPCs:    room.NPCs,
		Enemies: room.Enemies,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}

func FindPlayerByName(name string, maps map[string]Location) (room Location, index int, ok bool) {
	for _, loc := range maps {
		for i, p := range loc.Players {
			if p.Name == name {
				return loc, i, true
			}
		}
	}
	return Location{}, 0, false
}


func FindEnemyInRoom(enemyName string, loc Location, world WorldFile) (Enemy, int, bool) {
	for _, npc := range loc.NPCs {
		if npc == enemyName {
			return Enemy{}, 1, false
		}
	}
	for i, enemy := range loc.Enemies {
		if enemy == enemyName {
			enemyStats := world.Enemies[enemy]
			return enemyStats, i, true
		}
	}
	return Enemy{}, 0, false
}
