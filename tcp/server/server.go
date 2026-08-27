package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	. "the_answer_protocol/parsing"
	. "the_answer_protocol/tcp"
	. "the_answer_protocol/tcp/common"
)

func roomChat(conn net.Conn, room string) {

}

var (
	maps     map[string]Location
	mapsLock = sync.Mutex{}
)

var (
	world     WorldFile
	worldLock = sync.Mutex{}
)

func DeletePlayer(players []Player, player string) []Player {
	for i, p := range players {
		if p.Name == player {
			return append(players[:i], players[i+1:]...)
		}
	}
	return players
}

// Gestion de la connexion pour un client
func handleConnection(conn net.Conn) {
	var player string

	// Gestion de la fin de connexion
	defer func() {
		mapsLock.Lock()
		for _, loc := range maps {
			loc.Players = DeletePlayer(loc.Players, player)
			maps[loc.Name] = loc
		}
		mapsLock.Unlock()
		CloseSession(conn, player)
	}()
	player = Connexion(conn, player)
	if player == "" {
		return
	}
	// Charge le joueur et attribue la connexion
	worldLock.Lock()
	play := world.Players[player]
	play.Connect = conn
	world.Players[player] = play
	worldChar := ToString(world)
	worldLock.Unlock()

	// Mets le joueur dans la map a son point de respawn
	mapsLock.Lock()
	spawn := maps[play.Respawn.Location]
	spawn.Players = append(spawn.Players, play)
	maps[play.Respawn.Location] = spawn
	LogAndPrintf(conn, "%s", ToString(maps[play.Respawn.Location]))
	mapsLock.Unlock()

	// Charge le monde et le point de respawn du joueur pour lenvoyer au client
	fmt.Fprintf(conn, "%s\n", worldChar)
	loc := ToString(spawn)
	fmt.Fprintf(conn, "%s\n", loc)

	Dispatch(conn, player)
}

func main() {
	logfile, _ := os.Create("../protocol.log")
	log.SetOutput(logfile)
	maps = make(map[string]Location)
	// Ouverture de l'ecoute tcp sur le port 8080 du localhost
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}

	// Fermeture à la fin du programme
	defer ln.Close()
	var errs []error
	world, errs = Rally("../../yaml/world.yaml")
	if len(errs) != 0 {
		log.Fatalln(errs)
	}
	maps = world.World.Locations
	// Boucle infinie pour accepter la connexion des clients au serveur
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(conn.RemoteAddr().String() + "| " + err.Error())
			continue
		}

		// Ajout du client a la liste des clients
		ClientsLock.Lock()
		Clients[conn] = true
		ClientsLock.Unlock()

		LogAndPrintf(conn, "%s| [INFO] Nouveau client connecté\n", conn.RemoteAddr())

		// Gestion individuelle (Thread) du client
		go handleConnection(conn)
	}
}
