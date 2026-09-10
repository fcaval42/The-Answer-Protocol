package commands

import (
	"encoding/json"
	"fmt"
	. "the_answer_protocol/parsing"
)

func findNPCInRoom(loc Location, npcName string) (id int, ok bool) {
	for npcID, npc := range loc.NPCs {
		if npcName == npc {
			return npcID, true
		}
	}
	return 0, false
}

func findQuestForGiver(world WorldFile, giverID string) (Quest, bool) {
	for _, q := range world.Quests {
		if q.Giver == giverID {
			return q, true
		}
	}
	return Quest{}, false
}

func findPlayerQuest(player Player, questID string) (PlayerQuest, bool) {
	for _, pq := range player.Quests {
		if pq.QuestID == questID {
			return pq, true
		}
	}
	return PlayerQuest{}, false
}

func hasItem(player Player, itemID string) bool {
	for _, item := range player.Inventory {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

func questProgress(player Player, pq PlayerQuest, q Quest) string {
	switch q.Objective.Type {
	case "fetch":
		if hasItem(player, q.Objective.Item) {
			return "1/1"
		}
	case "defeat":
		if pq.Status == "ready" || pq.Status == "completed" {
			return "1/1"
		}
	}
	return "0/1"
}

func ReturnQuest(world WorldFile, room Location, player Player, npcName string) (string, Player, string, error) {
	npcId, ok := findNPCInRoom(room, npcName)
	if !ok {
		return "ERR 404 NPC_NOT_FOUND", player, "", nil
	}

	quest, ok := findQuestForGiver(world, room.NPCs[npcId])
	if !ok {
		return "ERR 406 NO_QUEST_AVAILABLE", player, "", nil
	}

	pq, known := findPlayerQuest(player, quest.ID)
	if known && pq.Status == "completed" {
		return "ERR 406 NO_QUEST_AVAILABLE", player, "", nil
	}

	if !known && IsQuestGloballyCompleted(quest.ID) {
		return "ERR 406 NO_QUEST_AVAILABLE", player, "", nil
	}

	if known && questProgress(player, pq, quest) == "1/1" {
		for i := range player.Quests {
			if player.Quests[i].QuestID == quest.ID {
				player.Quests[i].Status = "completed"
			}
		}
		MarkQuestCompleted(quest.ID)
		rewardName := quest.Reward.Value
		if quest.Reward.Type == "item" {
			if item, ok := world.Items[quest.Reward.Value]; ok {
				player.Inventory = append(player.Inventory, item)
				rewardName = item.Name
			}
		}
		data, err := json.Marshal(struct {
			QuestID string `json:"quest_id"`
			Status  string `json:"status"`
			Reward  string `json:"reward"`
			Message string `json:"message"`
		}{
			QuestID: quest.ID,
			Status:  "completed",
			Reward:  quest.Reward.Value,
			Message: fmt.Sprintf("Quête terminée : %s ! Vous recevez : %s.", quest.Name, rewardName),
		})
		if err != nil {
			return "", player, "", err
		}
		return fmt.Sprintf("OK %s", data), player, quest.ID, nil
	}

	if !known {
		player.Quests = append(player.Quests, PlayerQuest{QuestID: quest.ID, Status: "active"})
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
		return "", player, "", err
	}

	return fmt.Sprintf("OK %s", data), player, "", nil
}

func CompleteDefeatQuest(world WorldFile, player Player, enemyName string) (Player, string) {
	for i, pq := range player.Quests {
		if pq.Status != "active" {
			continue
		}
		quest, ok := world.Quests[pq.QuestID]
		if !ok || quest.Objective.Type != "defeat" || quest.Objective.Enemy != enemyName {
			continue
		}
		player.Quests[i].Status = "ready"
		return player, fmt.Sprintf("[SERVER] Objectif de quête rempli : %s ! Retournez voir %s pour récupérer votre récompense.", quest.Name, quest.Giver)
	}
	return player, ""
}

func Quests(room Location, player Player, world WorldFile) (string, error) {
	type questEntry struct {
		QuestID     string `json:"quest_id"`
		Name        string `json:"name"`
		Giver       string `json:"giver"`
		Status      string `json:"status"`
		Description string `json:"description,omitempty"`
		Progress    string `json:"progress,omitempty"`
	}

	entries := make([]questEntry, 0, len(player.Quests))
	known := make(map[string]bool)

	for _, pq := range player.Quests {
		quest, ok := world.Quests[pq.QuestID]
		if !ok {
			continue
		}
		known[quest.ID] = true

		if IsQuestGloballyCompleted(pq.QuestID) && pq.Status != "completed" {
			pq.Status = "completed"
		}

		if pq.Status == "completed" {
			entries = append(entries, questEntry{QuestID: quest.ID, Name: quest.Name, Giver: quest.Giver, Status: "completed"})
			continue
		}

		progress := questProgress(player, pq, quest)
		if progress == "1/1" {
			status := "waiting"
			if _, present := findNPCInRoom(room, quest.Giver); present {
				status = "ready"
			}
			entries = append(entries, questEntry{QuestID: quest.ID, Name: quest.Name, Giver: quest.Giver, Status: status, Description: quest.Description})
			continue
		}
		entries = append(entries, questEntry{QuestID: quest.ID, Name: quest.Name, Giver: quest.Giver, Status: pq.Status, Description: quest.Description, Progress: progress})
	}

	for _, npcID := range room.NPCs {
		quest, ok := findQuestForGiver(world, npcID)
		if !ok || known[quest.ID] || IsQuestGloballyCompleted(quest.ID) {
			continue
		}
		entries = append(entries, questEntry{QuestID: quest.ID, Name: quest.Name, Giver: quest.Giver, Status: "available", Description: quest.Description})
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}
