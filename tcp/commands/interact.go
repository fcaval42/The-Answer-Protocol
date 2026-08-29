package commands

// TALK, STATUS, ATTACK

import (
	"encoding/json"
	"fmt"
	"the_answer_protocol/parsing"
)

// findEnemyInRoom cherche, parmi les enemies presents dans loc, celui dont
// l'id ou le nom affiche vaut enemyName (memes regles de reference que
// findNPCInRoom pour les npcs, cf. quests.go).
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

// Talk traite la commande TALK <npc> : renvoie la premiere ligne de
// dialogue du npc npcName, present dans la salle courante du joueur.
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

// statusLabel derive un statut lisible a partir du ratio hp/max_hp. Choix
// simple (pas d'effets de statut type poison/buff, cf. RFC 6.1.1) qui garde
// "healthy" pour une blessure legere, comme dans l'exemple du RFC
// (80/100 -> "healthy").
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

// Status traite la commande STATUS : renvoie les points de vie courants du
// joueur. parsing.Player.MaxHP est capture une fois au parsing (cf.
// parsing.ParseWorld) a partir du hp de depart du yaml : Respawn.HP n'aurait
// pas convenu, c'est le hp (reduit) auquel le joueur revient apres une
// defaite, pas son plafond.
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

// Attack traite la commande ATTACK <npc> : engage le combat contre l'enemy
// enemyName present dans la salle courante du joueur.
//
// Design choice (cf. RFC 6.1.1, laisse a l'implementation) : pas de systeme
// de tour explicite - chaque ATTACK est un echange complet : le joueur
// frappe en premier, puis l'enemy contre-attaque s'il est encore en vie.
// Les degats sont attaque - defense de la cible, plancher a 1 pour qu'un
// combat progresse toujours meme contre une grosse defense. L'enemy est
// partage au niveau du monde (pas une instance par joueur) : ses HP restent
// donc reduits pour tous les joueurs tant qu'il n'est pas vaincu.
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
