package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"sync"
	. "the_answer_protocol/parsing"
	. "the_answer_protocol/tcp"
	. "the_answer_protocol/tcp/common"
)

var (
	maps     map[string]Location
	mapsLock = sync.Mutex{}
)

var (
	world     WorldFile
	worldLock = sync.Mutex{}
)

func deletePlayerInMap(player string) map[string]Location {
	mapsLock.Lock()
	var playerInLoc []net.Conn
	var temp []Player
	for _, loc := range maps {
		for _, play := range loc.Players {
			if play.Name == player {
				temp = loc.Players
			}
		}
		loc.Players = DeletePlayer(loc.Players, player)
		maps[loc.ID] = loc
	}
	ConnectedLock.Lock()
	for _, play := range temp {
		for n, c := range ConnectedClients {
			if play.Name == n {
				if !slices.Contains(playerInLoc, c) {
					playerInLoc = append(playerInLoc, c)
				}
			}
		}
	}
	ConnectedLock.Unlock()
	BroadcastEVT(playerInLoc, fmt.Sprintf("[EVT] %s a quitté l'endroit.", player))
	mapsLock.Unlock()
	return maps
}

func disconnectPlayer(player string) {
	worldLock.Lock()
	if p, ok := world.Players[player]; ok {
		p.Connected = false
		world.Players[player] = p
	}
	worldLock.Unlock()
}

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
		var playerConn []net.Conn
		CloseSession(conn, player)
		ConnectedLock.Lock()
		for _, conn := range ConnectedClients {
			playerConn = append(playerConn, conn)
		}
		nb := len(playerConn)
		ConnectedLock.Unlock()
		deletePlayerInMap(player)
		BroadcastEVT(playerConn, fmt.Sprintf("[EVT] %s a quitté le jeu (%d joueurs connectés).", player, nb))
		disconnectPlayer(player)
	}()
	player = Connexion(conn, player)
	if player == "" {
		return
	}
	// Charge le joueur et attribue la connexion
	worldLock.Lock()
	play := world.Players[player]
	play.Connected = true
	worldLock.Unlock()

	// Mets le joueur dans la map a son point de respawn
	mapsLock.Lock()
	location := maps[play.Respawn.Location]
	location.Players = append(location.Players, play)
	maps[play.Respawn.Location] = location
	mapsLock.Unlock()
	var playerInLoc []net.Conn
	var receivers []net.Conn
	for _, play := range location.Players {
		for n, conn := range ConnectedClients {
			if play.Name == n {
				playerInLoc = append(playerInLoc, conn)
			}
		}
	}
	for _, c := range playerInLoc {
		if c != conn {
			receivers = append(receivers, c)
		}
	}
	BroadcastEVT(receivers, fmt.Sprintf("[EVT] %s a rejoint l'endroit.", play.Name))

	worldLock.Lock()
	world.Players[player] = play
	worldChar := ToString(world)
	worldLock.Unlock()

	// Charge le monde et le point de respawn du joueur pour lenvoyer au client
	fmt.Fprintf(conn, "%s\n", worldChar)
	fmt.Fprintf(conn, "%s\n", ToString(withEnemyHP(location)))

	Dispatch(conn, play, location)
}

func main() {
	logfile, _ := os.Create("protocol.log")
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
	world, errs = Rally("yaml/world.yaml")
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

		LogServPrintf(conn, "%s| [INFO] Nouveau client connecté\n", conn.RemoteAddr())

		// Gestion individuelle (Thread) du client
		go handleConnection(conn)
	}
}
