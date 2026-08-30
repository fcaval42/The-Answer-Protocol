package commands

// LOOK

import (
	"encoding/json"
	"fmt"
	"the_answer_protocol/parsing"
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

func findPlayerByName(world parsing.WorldFile, name string) (id string, player parsing.Player, ok bool) {
	for id, p := range world.Players {
		if p.Name == name {
			return id, p, true
		}
	}
	return "", parsing.Player{}, false
}

func Look(world parsing.WorldFile, playerName string) (string, error) {
	_, player, ok := findPlayerByName(world, playerName)
	if !ok || player.Location.ID == "" {
		return "", fmt.Errorf("player %q has no known location", playerName)
	}

	loc, ok := world.World.Locations[player.Location.ID]
	if !ok {
		return "", fmt.Errorf("player %q location %q is unknown", playerName, player.Location.ID)
	}

	players := make([]string, 0)
	for _, p := range world.Players {
		if p.Connected && p.Location.ID == loc.ID {
			players = append(players, p.Name)
		}
	}

	npcs := make([]string, 0, len(loc.NPCs)+len(loc.Enemies))
	npcs = append(npcs, loc.NPCs...)
	npcs = append(npcs, loc.Enemies...)

	items := loc.Items
	if items == nil {
		items = make([]string, 0)
	}

	data, err := json.Marshal(lookResponse{
		Room: lookRoom{
			ID:          loc.ID,
			Name:        loc.Name,
			Description: loc.Description,
			Exits:       loc.Exits,
		},
		Players: players,
		Items:   items,
		NPCs:    npcs,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}
