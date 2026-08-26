package main

import (
	"log"
	"net"
	"os"
	. "the_answer_protocol/tcp/common"
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
	logfile, _ := os.Create("../protocol.log")
	log.SetOutput(logfile)
	// Ouverture de l'ecoute tcp sur le port 8080 du localhost
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}

	// Fermeture à la fin du programme
	defer ln.Close()

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
