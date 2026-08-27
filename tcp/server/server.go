package main

import (
	"io"
	"fmt"
	"errors"
	"log"
	"net"
	"os"

	. "the_answer_protocol/tcp/common"
	"the_answer_protocol/parsing"
)

func roomChat(conn net.Conn, room string) {

}

// Gestion de la connexion pour un client
func handleConnection(conn net.Conn) {
	var player string

	// Gestion de la fin de connexion
	defer func() {
		CloseSession(conn, player)
	}()
	Connexion(conn, player)
	Dispatch(conn, player)
}

func main() {
	logfile, err := os.Create("../protocol.log")
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logfile))

	// Parsing du monde une seule fois au démarrage : c'est cette instance,
	// partagée par toutes les connexions (voir tcp/common.World), qui garde
	// l'état des joueurs (Connected, Location, Inventory...) tant que le
	// serveur tourne.
	world, worldErrs := parsing.Rally("yaml/world.yaml")
	if len(worldErrs) > 0 {
		log.Fatalln(errors.Join(worldErrs...))
	}
	World = world

	// Ouverture de l'ecoute tcp sur le port 8080 du localhost
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}

	// Fermeture à la fin du programme
	defer ln.Close()

	fmt.Println("\033[0;32mServer Up and Running\033[0m")
	fmt.Println("")

	// Boucle infinie pour accepter la connexion des clients au serveur
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(conn.RemoteAddr().String() + "| " + err.Error())
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
