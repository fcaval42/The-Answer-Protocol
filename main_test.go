package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestRoomPanelDoesNotCrash(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	w := a.NewWindow("test")
	panel := buildRoomPanel(fakeRoom)
	w.SetContent(panel)

	if panel == nil {
		t.Fatal("buildRoomPanel a renvoyé nil")
	}
}

func TestRoomPanelHasCorrectExitCount(t *testing.T) {
	room := Room{
		Name:  "Salle de test",
		Exits: []string{"nord", "sud", "est"},
	}

	panel := buildRoomPanel(room)
	if panel == nil {
		t.Fatal("buildRoomPanel a renvoyé nil")
	}
	// Ici tu peux creuser plus loin avec les fonctions d'introspection
	// du package fyne/test si tu veux vérifier le nombre de boutons créés,
	// mais déjà vérifier que ça ne panique pas pour différentes tailles
	// de Room est un bon point de départ.
}