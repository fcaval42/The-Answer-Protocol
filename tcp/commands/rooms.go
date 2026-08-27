package commands

// LOOK

import (
	"encoding/json"
	"fmt"
	"the_answer_protocol/parsing"
)

// Structures dediees a la reponse LOOK (cf. RFC 42TAP §5.1.2), separees du
// modele parsing.Location : ce dernier porte des champs internes (Spawns,
// Players en tant que parsing.Player complets...) qui ne correspondent pas
// au format attendu par le protocole.
type lookRoom struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}

type lookResponse struct {
	Room    lookRoom `json:"room"`
	Players []string `json:"players"`
	Items   []string `json:"items"`
	NPCs    []string `json:"npcs"`
}

// findPlayerByName cherche un parsing.Player par son nom affiche plutot que
// par la cle de la map World.Players : c'est ce nom (celui choisi via
// ConnectCom/PersoChoice, ex: "Maëlle") qui circule comme identifiant de
// session dans tcp/server, alors que la cle yaml peut differer legerement
// (accents, tirets...). La cle est tout de meme renvoyee : les commandes qui
// modifient l'etat du joueur (cf. Quest dans quests.go) en ont besoin pour
// ecrire leur mise a jour dans world.Players.
func findPlayerByName(world parsing.WorldFile, name string) (id string, player parsing.Player, ok bool) {
	for id, p := range world.Players {
		if p.Name == name {
			return id, p, true
		}
	}
	return "", parsing.Player{}, false
}

// Look renvoie la reponse "OK {json}" decrivant la salle ou se trouve
// playerName : la salle elle-meme, les joueurs connectes qui s'y trouvent,
// ses items, et ses npcs (enemies inclus : ce sont des npcs attaquables du
// point de vue du protocole, meme si le monde les separe dans leur propre
// table). world doit etre une donnee a jour (lue sous WorldLock cote
// appelant), car plusieurs joueurs peuvent la modifier en meme temps.
func Look(world parsing.WorldFile, playerName string) (string, error) {
	_, player, ok := findPlayerByName(world, playerName)
	if !ok || player.Location.ID == "" {
		return "", fmt.Errorf("player %q has no known location", playerName)
	}

	// On relit la salle depuis le monde partage plutot que d'utiliser
	// player.Location tel quel : ce dernier n'est qu'une copie prise au
	// moment de la connexion, qui ne reflete pas les objets pris/deposes
	// depuis par d'autres joueurs.
	loc, ok := world.World.Locations[player.Location.ID]
	if !ok {
		return "", fmt.Errorf("player %q location %q is unknown", playerName, player.Location.ID)
	}

	players := make([]string, 0)
	for _, p := range world.Players {
		if p.Connected && p.Location.ID == loc.ID {
			players = append(players, p.Name)
		}
	}

	npcs := make([]string, 0, len(loc.NPCs)+len(loc.Enemies))
	npcs = append(npcs, loc.NPCs...)
	npcs = append(npcs, loc.Enemies...)

	items := loc.Items
	if items == nil {
		items = make([]string, 0)
	}

	data, err := json.Marshal(lookResponse{
		Room: lookRoom{
			ID:          loc.ID,
			Name:        loc.Name,
			Description: loc.Description,
			Exits:       loc.Exits,
		},
		Players: players,
		Items:   items,
		NPCs:    npcs,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("OK %s", data), nil
}
