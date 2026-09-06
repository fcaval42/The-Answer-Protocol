package commands

// TALK, STATUS, ATTACK

import (
	"encoding/json"
	"fmt"
	. "the_answer_protocol/parsing"
)

func Talk(room Location, world WorldFile, npcName string) (string) {
	npcid, ok := findNPCInRoom(room, npcName)
	npc := room.NPCs[npcid]
	if !ok {
		return "ERR 404 NPC_NOT_FOUND"
	}
	npcChar := world.NPCs[npc]
	var response string
	response += fmt.Sprintf("[%s]: ", npcName)
	for _, speech := range npcChar.Dialogue{
		response += fmt.Sprintf("%s ", speech)
	}
	return fmt.Sprintf("%s\n", response)
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

func Status(room Location, player Player) (string, error) {
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

func Attack(player Player, enemy Enemy) (string, string, Player, Enemy) {
	var status string

	damage := player.Attack - enemy.Defense
	if damage < 1 {
		damage = 1
	}
	enemy.HP -= damage
	if enemy.HP < 0 {
		enemy.HP = 0
	}

	status = "combat"
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
		return "", status, player, enemy
	}

	return fmt.Sprintf("OK %s", data), status, player, enemy
}
