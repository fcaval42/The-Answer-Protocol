package commands

// LOOK

import (
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
	Items   []string `json:"items"`
	NPCs    []string `json:"npcs"`
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
	for i, enemy := range loc.Enemies {
		if enemy == enemyName {
			enemyStats := world.Enemies[enemy]
			return enemyStats, i, true
		}
	}
	return Enemy{}, 0, false
}

