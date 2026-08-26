package commands

// LOOK, QUESTS,

import (
	"fmt"
	"the_answer_protocol/parsing"
)

// LOOK, cf. RFC 42TAP §5.1.2 : renvoie l'état de la salle courante.

func findPlayerByName(world parsing.WorldFile, name string) (parsing.Player, bool) {
	for _, p := range world.Players {
		if p.Name == name {
			return p, true
		}
	}
	return parsing.Player{}, false
}

// Look renvoie la parsing.Location où se trouve playerName, avec son champ
// Players rempli des joueurs actuellement présents dans la salle. world doit
// être une donnée à jour (lue sous WorldLock côté appelant), car plusieurs
// joueurs peuvent la modifier en même temps.
func Look(world parsing.WorldFile, playerName string) (parsing.Location, error) {
	player, ok := findPlayerByName(world, playerName)
	if !ok || player.Location.ID == "" {
		return parsing.Location{}, fmt.Errorf("player %q has no known location", playerName)
	}

	// On relit la salle depuis le monde partagé plutôt que d'utiliser
	// player.Location tel quel : ce dernier n'est qu'une copie prise au
	// moment de la connexion, qui ne reflète pas les objets pris/déposés
	// depuis par d'autres joueurs.
	loc, ok := world.World.Locations[player.Location.ID]
	if !ok {
		return parsing.Location{}, fmt.Errorf("player %q location %q is unknown", playerName, player.Location.ID)
	}

	for _, p := range world.Players {
		if p.Connected && p.Location.ID == loc.ID {
			loc.Players = append(loc.Players, p)
		}
	}

	return loc, nil
}
