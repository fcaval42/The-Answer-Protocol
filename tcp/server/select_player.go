package main

import (
	// "fmt"
	"log"
	"net"
	"strings"
	. "the_answer_protocol/tcp/commands"
	. "the_answer_protocol/tcp/common"
)

func select_Player(conn net.Conn) string{
	player := ConnectCom(conn)
	player = strings.Trim(player, "\n")
	switch player {
	case "":
		log.Println("[ERROR]: Player doesn't exist")
		return ""
	case "900":
		return ""
	}

	// Marque le joueur comme connecté dans le monde partagé par le serveur.
	// Un Location.ID vide signifie qu'il n'a encore jamais été joué depuis le
	// lancement du serveur : on lui donne alors un point de départ.
	// Si Location est déjà renseignée, on la garde telle quelle (idem pour
	// Inventory), pour ne rien perdre d'une session précédente.
	WorldLock.Lock()
	for id, p := range World.Players {
		if p.Name != player {
			continue
		}
		p.Connected = true
		if p.Location.ID == "" {
			p.Location = World.World.Locations["Lumière"]
		}
		World.Players[id] = p
		break
	}
	WorldLock.Unlock()

	return player
}
