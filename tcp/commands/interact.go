package commands

// TALK, STATUS, ATTACK

import (
	"encoding/json"
	"fmt"
	"the_answer_protocol/parsing"
)

func findEnemyInRoom(world parsing.WorldFile, loc parsing.Location, enemyName string) (parsing.Enemy, bool) {
	for _, enemyID := range loc.Enemies {
		enemy, ok := world.Enemies[enemyID]
		if !ok {
			continue
		}
		if enemy.ID == enemyName || enemy.Name == enemyName {
			return enemy, true
		}
	}
	return parsing.Enemy{}, false
}

func Talk(world parsing.WorldFile, playerName string, npcName string) (string, error) {
	_, player, ok := findPlayerByName(world, playerName)
	if !ok || player.Location.ID == "" {
		return "", fmt.Errorf("player %q has no known location", playerName)
	}

	loc, ok := world.World.Locations[player.Location.ID]
	if !ok {
		return "", fmt.Errorf("player %q location %q is unknown", playerName, player.Location.ID)
	}

	npc, ok := findNPCInRoom(world, loc, npcName)
	if !ok {
		return "ERR 404 NPC_NOT_FOUND", nil
	}

	return fmt.Sprintf("OK %s", npc.Dialogue[0]), nil
}

func statusLabel(hp int, maxHP int) string {
	if hp <= 0 {
		return "dead"
	}
	if maxHP <= 0 {
		return "healthy"
	}
	ratio := float64(hp) / float64(maxHP)
	switch {
	case ratio < 0.3:
		return "critical"
	case ratio < 0.75:
		return "wounded"
	default:
		return "healthy"
	}
}

func Status(world parsing.WorldFile, playerName string) (string, error) {
	_, player, ok := findPlayerByName(world, playerName)
	if !ok {
		return "", fmt.Errorf("player %q is unknown", playerName)
	}

	maxHP := player.MaxHP

	data, err := json.Marshal(struct {
		HP     int    `json:"hp"`
		MaxHP  int    `json:"max_hp"`
		Status string `json:"status"`
	}{
		HP:     player.HP,
		MaxHP:  maxHP,
		Status: statusLabel(player.HP, maxHP),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}

func Attack(world parsing.WorldFile, playerName string, enemyName string) (string, error) {
	playerID, player, ok := findPlayerByName(world, playerName)
	if !ok || player.Location.ID == "" {
		return "", fmt.Errorf("player %q has no known location", playerName)
	}

	loc, ok := world.World.Locations[player.Location.ID]
	if !ok {
		return "", fmt.Errorf("player %q location %q is unknown", playerName, player.Location.ID)
	}

	enemy, ok := findEnemyInRoom(world, loc, enemyName)
	if !ok {
		if _, isNPC := findNPCInRoom(world, loc, enemyName); isNPC {
			return "ERR 405 NPC_NOT_HOSTILE", nil
		}
		return "ERR 404 NPC_NOT_FOUND", nil
	}

	damage := player.Attack - enemy.Defense
	if damage < 1 {
		damage = 1
	}
	enemy.HP -= damage
	if enemy.HP < 0 {
		enemy.HP = 0
	}

	status := "combat"
	if enemy.HP == 0 {
		status = "victory"
	} else {
		player.HP -= enemy.Counter
		if player.HP < 0 {
			player.HP = 0
		}
		if player.HP == 0 {
			status = "defeat"
		}
	}

	world.Enemies[enemy.ID] = enemy
	world.Players[playerID] = player

	data, err := json.Marshal(struct {
		AttackerHP int    `json:"attacker_hp"`
		TargetHP   int    `json:"target_hp"`
		Damage     int    `json:"damage"`
		Status     string `json:"status"`
	}{
		AttackerHP: player.HP,
		TargetHP:   enemy.HP,
		Damage:     damage,
		Status:     status,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}
