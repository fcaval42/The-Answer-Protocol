package commands

// QUEST, QUESTS

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

func questProgress(room Location, player Player, q Quest) string {
	switch q.Objective.Type {
	case "fetch":
		if hasItem(player, q.Objective.Item) && (q.Objective.DeliverTo == "" || room.ID == q.Objective.DeliverTo) {
			return "1/1"
		}
	}
	return "0/1"
}





func ReturnQuest(world WorldFile, room Location, player Player, npcName string) (string, Player, error) {
	npcId, ok := findNPCInRoom(room, npcName)
	if !ok {
		return "ERR 404 NPC_NOT_FOUND", player, nil
	}

	quest, ok := findQuestForGiver(world, room.NPCs[npcId])
	if !ok {
		return "ERR 406 NO_QUEST_AVAILABLE", player, nil
	}

	if pq, known := findPlayerQuest(player, quest.ID); known {
		if pq.Status == "completed" {
			return "ERR 406 NO_QUEST_AVAILABLE", player, nil
		}
	} else {
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
		return "", player, err
	}

	return fmt.Sprintf("OK %s", data), player, nil
}

func Quests(room Location, player Player, world WorldFile) (string, error) {
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

		progress := questProgress(room, player, quest)
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
