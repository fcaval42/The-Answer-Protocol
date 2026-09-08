package commands

// TALK, STATUS, ATTACK

import (
	"encoding/json"
	"fmt"
	. "the_answer_protocol/parsing"
	"math/rand/v2"
)

func Talk(room Location, world WorldFile, npcName string) (string) {
	npcid, ok := findNPCInRoom(room, npcName)
	if !ok {
		return "ERR 404 NPC_NOT_FOUND"
	}
	npc := room.NPCs[npcid]
	npcChar := world.NPCs[npc]
	var response string
	response += fmt.Sprintf("[%s]: ", npcName)
	for _, speech := range npcChar.Dialogue{
		response += fmt.Sprintf("%s ", speech)
	}
	// prefixe "OK " : c'est la seule ligne "OK .../ERR .../{..." de cette
	// reponse, elle doit donc pouvoir etre reconnue par le GUI (cf.
	// Session.Deliver, qui n'accepte que ces prefixes).
	return fmt.Sprintf("OK %s", response)
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
		Critical float64 `json:"critical"`
		Status string `json:"status"`
	}{
		HP:     player.HP,
		MaxHP:  maxHP,
		Critical: player.Weapon.Critical,
		Status: statusLabel(player.HP, maxHP),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}

func Attack(player Player, enemy Enemy) (string, string, string, Player, Enemy) {
	var status string
	var critMsg string
	damage := player.Attack - enemy.Defense

	if rand.Float64() < (player.Weapon.Critical) {
		dam_base := damage
		damage += int(float64(damage) * player.Weapon.Critical)
		critMsg = fmt.Sprintf("COUP CRITIQUE: Vous infligez %d au lieu de %d", damage, dam_base)
	}

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
		return "", critMsg, status, player, enemy
	}

	return fmt.Sprintf("OK %s", data), critMsg, status, player, enemy
}
