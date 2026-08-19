package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	. "the_answer_protocol/tcp/common"
)


// Permet de transmettre le message envoyé par un client aux autres
func broadcast(message string, sender net.Conn, player string) {
	// Gestion du mutex
	clientsLock.Lock()
	defer clientsLock.Unlock()

	// Préparer le message à envoyer (avec le nom du personnage du client expéditeur)
	fullMessage := fmt.Sprintf("[%s] : %s\n", player, message)

	// Envoyer à tous les clients sauf l'expéditeur
	for client := range clients {
		if client == sender {
			continue
		}
		// Envoie du message a un client
		_, err := client.Write([]byte(fullMessage))
		if err != nil {
			log.Printf("[ERROR]: lors de l'envoi à %s : %v\n", client.RemoteAddr(), err)
			// Supprimer le client en erreur
			delete(clients, client)
			client.Close()
		}
	}
}

// Gestion de la connexion pour un client
func handleConnection(conn net.Conn) {
	var player string
	// Gestion de la fin de connexion
	defer func() {
		clientsLock.Lock()
		// Supprimer le client de la liste à la fermeture
		delete(clients, conn)
		clientsLock.Unlock()
		// Si un personnage avait été utilisé ont le remer
		// dans la liste des personnages jouables
		if player != "" && player != "900" {
			PersoLock.Lock()
			PersoChoice = append(PersoChoice, player)
			PersoLock.Unlock()
			LogAndPrintf(conn, "[INFO] [%s]: \"Pour ceux qui viendront après\"\n", player)
		}
		// Fermeture de la connexion
		conn.Close()
		LogAndPrintf(conn, "%s| [INFO] Client déconnecté\n", conn.RemoteAddr())
	}()

	// Buffer en écoute sur le client
	reader := bufio.NewReader(conn)
	// Séléction du personnage jouable
	player = selectPlayer(conn)
	player = strings.Trim(player, "\n")
	switch player {
	case "":
		log.Println("[ERROR]: Player doesn't exist")
		return
	case "900":
		return
	}

	// Boucle infinie d'écoute des clients
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println(conn.RemoteAddr().String() + "| " + err.Error())
			return
		}
		// Affichage du message
		LogAndPrintf(conn, "%s| [%s]: %s", conn.RemoteAddr().String(), player, message)
		// Redistribution aux autres clients
		broadcast(message, conn, player)
	}
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
		clientsLock.Lock()
		clients[conn] = true
		clientsLock.Unlock()

		LogAndPrintf(conn, "%s| [INFO] Nouveau client connecté\n", conn.RemoteAddr())

		// Gestion individuelle (Thread) du client
		go handleConnection(conn)
	}
}

// Command (CONNECT, LOOK, MOVE, CHAT, TAKE, DROP, INVENTORY,
// TALK, ATTACK, STATUS, QUEST, QUESTS, WHO, GROUP, QUIT)
