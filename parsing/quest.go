package parsing

import (
	"fmt"
	"sync"
)

type QuestObjective struct {
	Type      string `yaml:"type"`
	Item      string `yaml:"item,omitempty"`
	Enemy     string `yaml:"enemy,omitempty"`
	DeliverTo string `yaml:"deliver_to,omitempty"`
}

type QuestReward struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

type Quest struct {
	ID          string         `yaml:"-"`
	Name        string         `yaml:"name"`
	Giver       string         `yaml:"giver"`
	Description string         `yaml:"description"`
	Objective   QuestObjective `yaml:"objective"`
	Reward      QuestReward    `yaml:"reward"`
}

type PlayerQuest struct {
	QuestID string
	Status  string
}

var (
	globallyCompletedQuests   = make(map[string]bool)
	globallyCompletedQuestsMu sync.Mutex
)

func MarkQuestCompleted(questID string) {
	globallyCompletedQuestsMu.Lock()
	defer globallyCompletedQuestsMu.Unlock()
	globallyCompletedQuests[questID] = true
}

func IsQuestGloballyCompleted(questID string) bool {
	globallyCompletedQuestsMu.Lock()
	defer globallyCompletedQuestsMu.Unlock()
	return globallyCompletedQuests[questID]
}

func validateQuests(wf WorldFile) []error {
	var errs []error

	seenQuestNames := make(map[string]string)
	for id, q := range wf.Quests {
		if isBlank(id) {
			errs = append(errs, fmt.Errorf("a quest has an empty id"))
		}

		if isBlank(q.Name) {
			errs = append(errs, fmt.Errorf("quest %q: name is empty", id))
		} else if otherID, ok := seenQuestNames[q.Name]; ok {
			errs = append(errs, fmt.Errorf("quests %q and %q share the same name %q", otherID, id, q.Name))
		} else {
			seenQuestNames[q.Name] = id
		}

		if isBlank(q.Description) {
			errs = append(errs, fmt.Errorf("quest %q: description is empty", id))
		}

		if isBlank(q.Giver) {
			errs = append(errs, fmt.Errorf("quest %q: giver is empty", id))
		} else if _, ok := wf.NPCs[q.Giver]; !ok {
			errs = append(errs, fmt.Errorf("quest %q: giver %q is not a known npc", id, q.Giver))
		}

		switch q.Objective.Type {
		case "fetch":
			if isBlank(q.Objective.Item) {
				errs = append(errs, fmt.Errorf("quest %q: fetch objective has an empty item", id))
			} else if _, ok := wf.Items[q.Objective.Item]; !ok {
				errs = append(errs, fmt.Errorf("quest %q: objective item %q is not a known item", id, q.Objective.Item))
			}
			if q.Objective.DeliverTo != "" {
				if _, ok := wf.World.Locations[q.Objective.DeliverTo]; !ok {
					errs = append(errs, fmt.Errorf("quest %q: objective deliver_to %q is not a known location", id, q.Objective.DeliverTo))
				}
			}
		case "defeat":
			if isBlank(q.Objective.Enemy) {
				errs = append(errs, fmt.Errorf("quest %q: defeat objective has an empty enemy", id))
			} else if _, ok := wf.Enemies[q.Objective.Enemy]; !ok {
				errs = append(errs, fmt.Errorf("quest %q: objective enemy %q is not a known enemy", id, q.Objective.Enemy))
			}
		default:
			errs = append(errs, fmt.Errorf("quest %q: unknown objective type %q", id, q.Objective.Type))
		}

		switch q.Reward.Type {
		case "item":
			if isBlank(q.Reward.Value) {
				errs = append(errs, fmt.Errorf("quest %q: reward value is empty", id))
			} else if _, ok := wf.Items[q.Reward.Value]; !ok {
				errs = append(errs, fmt.Errorf("quest %q: reward item %q is not a known item", id, q.Reward.Value))
			}
		default:
			errs = append(errs, fmt.Errorf("quest %q: unknown reward type %q", id, q.Reward.Type))
		}
	}

	return errs
}
