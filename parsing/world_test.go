package parsing_test

import (
	"os"
	"strings"
	"testing"

	"the_answer_protocol/parsing"
)

func loadFixture(t *testing.T, name string) []error {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	_, errs := parsing.ParseWorld(data)
	return errs
}

func assertErrors(t *testing.T, errs []error, expected []string) {
	t.Helper()

	matched := make([]bool, len(expected))
	for _, err := range errs {
		msg := err.Error()
		found := false
		for i, want := range expected {
			if strings.Contains(msg, want) {
				matched[i] = true
				found = true
			}
		}
		if !found {
			t.Errorf("unexpected error: %s", msg)
		}
	}
	for i, want := range expected {
		if !matched[i] {
			t.Errorf("expected an error containing %q, got none", want)
		}
	}
}

func TestValidWorld(t *testing.T) {
	errs := loadFixture(t, "valid.yaml")
	if len(errs) != 0 {
		t.Fatalf("valid.yaml should parse with no errors, got: %v", errs)
	}
}

func TestLocationErrors(t *testing.T) {
	errs := loadFixture(t, "location_errors.yaml")
	assertErrors(t, errs, []string{
		`a location has an empty id`,
		`location "": name is empty`,
		`location "": description is empty`,
		`location "": has no exits`,
		`location "self_loop": exit "east" points to itself`,
		`location "unknown_exit": exit "east" points to unknown location "does_not_exist"`,
		`location "blank_exit": an exit has an empty direction`,
		`location "blank_exit": exit "north" has an empty destination`,
		`location "dup_refs": item "potion" is listed more than once`,
		`location "dup_refs": npc "bob" is listed more than once`,
		`location "dup_refs": enemy "wolf" is listed more than once`,
		`location "dup_refs": "shared" is listed in both npcs and enemies`,
		`location "unknown_refs": unknown item "ghost_item"`,
		`location "unknown_refs": unknown npc "ghost_npc"`,
		`location "unknown_refs": unknown enemy "ghost_enemy"`,
		`location "bad_spawn": a spawn has an empty npc_type`,
		`location "bad_spawn": spawn references unknown npc_type "ghost_npc_type"`,
		`location "bad_spawn": spawn "bob" has a non-positive count 0`,
		`share the same name "Loop A"`,
		`npc "traveler" is placed in both`,
		`enemy "goblin" is placed in both`,
	})
}

func TestNoLoop(t *testing.T) {
	errs := loadFixture(t, "no_loop.yaml")
	assertErrors(t, errs, []string{
		`location "c": has no exits`,
		`the world has no loop`,
	})
}

func TestItemErrors(t *testing.T) {
	errs := loadFixture(t, "item_errors.yaml")
	assertErrors(t, errs, []string{
		`an item has an empty id`,
		`item "": name is empty`,
		`item "": description is empty`,
		`share the same name "Same Name"`,
		`item "bad_critical_low": critical -0.50 must be between 0 and 1`,
		`item "bad_critical_high": critical 1.50 must be between 0 and 1`,
	})
}

func TestNPCErrors(t *testing.T) {
	errs := loadFixture(t, "npc_errors.yaml")
	assertErrors(t, errs, []string{
		`an npc has an empty id`,
		`npc "": name is empty`,
		`npc "": role is empty`,
		`npc "": description is empty`,
		`npc "": dialogue is empty`,
		`share the same name "Same NPC"`,
		`npc "blank_line": dialogue line 0 is empty`,
	})
}

func TestEnemyErrors(t *testing.T) {
	errs := loadFixture(t, "enemy_errors.yaml")
	assertErrors(t, errs, []string{
		`an enemy has an empty id`,
		`enemy "": name is empty`,
		`enemy "": description is empty`,
		`enemy "": difficulty is empty`,
		`enemy "": dialogue is empty`,
		`enemy "": hp -1 is negative`,
		`enemy "": attack -2 is negative`,
		`enemy "": defense -3 is negative`,
		`enemy "": counter -4 is negative`,
		`share the same name "Same Enemy"`,
		`enemy "blank_line": dialogue line 0 is empty`,
	})
}

func TestPlayerErrors(t *testing.T) {
	errs := loadFixture(t, "player_errors.yaml")
	assertErrors(t, errs, []string{
		`a player has an empty id`,
		`player "": name is empty`,
		`player "": hp -1 is negative`,
		`player "": attack -2 is negative`,
		`player "": defense -3 is negative`,
		`player "": weapon name is empty`,
		`player "": weapon critical -0.10 must be between 0 and 1`,
		`player "": respawn hp -5 is negative`,
		`player "": respawn location is empty`,
		`share the same name "Same Player"`,
		`player "bad_respawn_location": respawn location "does_not_exist" is not a known location`,
		`player "bad_weapon_critical_high": weapon critical 1.50 must be between 0 and 1`,
	})
}

func TestQuestErrors(t *testing.T) {
	errs := loadFixture(t, "quest_errors.yaml")
	assertErrors(t, errs, []string{
		`a quest has an empty id`,
		`quest "": name is empty`,
		`quest "": description is empty`,
		`quest "": giver is empty`,
		`quest "": unknown objective type ""`,
		`quest "": unknown reward type ""`,
		`share the same name "Same Quest"`,
		`quest "unknown_giver": giver "ghost" is not a known npc`,
		`quest "bad_fetch": objective item "unknown_item" is not a known item`,
		`quest "bad_fetch": objective deliver_to "nowhere" is not a known location`,
		`quest "empty_fetch_item": fetch objective has an empty item`,
		`quest "bad_defeat": objective enemy "unknown_enemy" is not a known enemy`,
		`quest "empty_defeat_enemy": defeat objective has an empty enemy`,
		`quest "unknown_objective_type": unknown objective type "explore"`,
		`quest "bad_reward": reward item "unknown_item" is not a known item`,
		`quest "empty_reward_value": reward value is empty`,
		`quest "unknown_reward_type": unknown reward type "gold"`,
	})
}

func TestMalformedYAML(t *testing.T) {
	data, err := os.ReadFile("testdata/malformed.yaml")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	_, errs := parsing.ParseWorld(data)
	if len(errs) != 1 {
		t.Fatalf("expected exactly one error for malformed YAML, got %v", errs)
	}
	if !strings.Contains(errs[0].Error(), "parsing world:") {
		t.Errorf("expected a YAML syntax error, got: %v", errs[0])
	}
}

func TestRallyMissingFile(t *testing.T) {
	_, errs := parsing.Rally("testdata/does_not_exist.yaml")
	if len(errs) == 0 {
		t.Fatal("expected an error when the world file doesn't exist")
	}
}

func TestRealWorldYAML(t *testing.T) {
	_, errs := parsing.Rally("../yaml/world.yaml")
	if len(errs) != 0 {
		t.Fatalf("yaml/world.yaml should parse with no errors, got: %v", errs)
	}
}
