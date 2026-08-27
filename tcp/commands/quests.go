package commands

// QUEST, QUESTS

import (
	"encoding/json"
	"fmt"
	"the_answer_protocol/parsing"
)

// findNPCInRoom cherche, parmi les npcs presents dans loc, celui dont l'id
// ou le nom affiche vaut npcName  les ressources
// peuvent etre referencees par id ou par nom).
func findNPCInRoom(world parsing.WorldFile, loc parsing.Location, npcName string) (parsing.NPC, bool) {
	for _, npcID := range loc.NPCs {
		npc, ok := world.NPCs[npcID]
		if !ok {
			continue
		}
		if npc.ID == npcName || npc.Name == npcName {
			return npc, true
		}
	}
	return parsing.NPC{}, false
}

// findQuestForGiver cherche, dans le catalogue du monde, la quete confiee
// par le npc giverID. Les donnees actuelles n'associent qu'une seule quete
// par giver ; s'il y en avait plusieurs, la premiere trouvee serait retenue.
func findQuestForGiver(world parsing.WorldFile, giverID string) (parsing.Quest, bool) {
	for _, q := range world.Quests {
		if q.Giver == giverID {
			return q, true
		}
	}
	return parsing.Quest{}, false
}

// findPlayerQuest cherche l'entree de questID dans le journal de quetes du
// joueur.
func findPlayerQuest(player parsing.Player, questID string) (parsing.PlayerQuest, bool) {
	for _, pq := range player.Quests {
		if pq.QuestID == questID {
			return pq, true
		}
	}
	return parsing.PlayerQuest{}, false
}

func hasItem(player parsing.Player, itemID string) bool {
	for _, item := range player.Inventory {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

// questProgress calcule la progression courante de q pour player. Les
// quetes de ce monde ne referencent qu'un seul item/enemy cible (pas de
// quantite), la progression est donc binaire : "0/1" (pas encore rempli) ou
// "1/1" (objectif rempli). Les quetes de type "defeat" necessitent un
// systeme de combat qui n'est pas encore implemente ; elles restent donc a
// "0/1" jusqu'a ce qu'ATTACK puisse les faire progresser.
func questProgress(player parsing.Player, q parsing.Quest) string {
	switch q.Objective.Type {
	case "fetch":
		if hasItem(player, q.Objective.Item) && (q.Objective.DeliverTo == "" || player.Location.ID == q.Objective.DeliverTo) {
			return "1/1"
		}
	}
	return "0/1"
}

// Quest traite la commande QUEST <npc> : le joueur
// demande une quete au npc npcName, present dans sa salle courante.
//
// Design choice : il n'existe pas de commande
// ACCEPT_QUEST dans le RFC, donc demander une quete a son giver l'ajoute
// directement au journal du joueur (statut "active") si ce n'est pas deja
// fait ; redemander une quete deja en cours renvoie simplement ses
// informations, redemander une quete deja terminee renvoie
// NO_QUEST_AVAILABLE. Le champ "status" de la reponse reste "available",
// comme dans l'exemple du RFC : il decrit la quete telle qu'offerte par le
// giver, la nuance active/completed n'apparaissant que dans QUESTS.
func Quest(world parsing.WorldFile, playerName string, npcName string) (string, error) {
	id, player, ok := findPlayerByName(world, playerName)
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

	quest, ok := findQuestForGiver(world, npc.ID)
	if !ok {
		return "ERR 406 NO_QUEST_AVAILABLE", nil
	}

	if pq, known := findPlayerQuest(player, quest.ID); known {
		if pq.Status == "completed" {
			return "ERR 406 NO_QUEST_AVAILABLE", nil
		}
	} else {
		player.Quests = append(player.Quests, parsing.PlayerQuest{QuestID: quest.ID, Status: "active"})
		world.Players[id] = player
	}

	data, err := json.Marshal(struct {
		QuestID     string `json:"quest_id"`
		Description string `json:"description"`
		Reward      string `json:"reward"`
		Status      string `json:"status"`
	}{
		QuestID:     quest.ID,
		Description: quest.Description,
		Reward:      quest.Reward.Value,
		Status:      "available",
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}

// Quests traite la commande QUESTS : renvoie le
// journal de quetes du joueur. Le statut "completed" est recalcule a chaque
// appel via questProgress plutot que stocke : ca evite une etape de
// synchronisation separee tant qu'il n'existe pas de recompense a distribuer
// au moment precis de la completion
func Quests(world parsing.WorldFile, playerName string) (string, error) {
	_, player, ok := findPlayerByName(world, playerName)
	if !ok {
		return "", fmt.Errorf("player %q is unknown", playerName)
	}

	type questEntry struct {
		QuestID  string `json:"quest_id"`
		Status   string `json:"status"`
		Progress string `json:"progress,omitempty"`
	}

	entries := make([]questEntry, 0, len(player.Quests))
	for _, pq := range player.Quests {
		quest, ok := world.Quests[pq.QuestID]
		if !ok {
			continue
		}

		if pq.Status == "completed" {
			entries = append(entries, questEntry{QuestID: quest.ID, Status: "completed"})
			continue
		}

		progress := questProgress(player, quest)
		status := pq.Status
		if progress == "1/1" {
			status = "completed"
			entries = append(entries, questEntry{QuestID: quest.ID, Status: status})
			continue
		}
		entries = append(entries, questEntry{QuestID: quest.ID, Status: status, Progress: progress})
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}
