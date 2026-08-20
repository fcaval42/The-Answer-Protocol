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


var globaleMessage []string


func recept(conn net.Conn) string{
	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return ""
		}
		log.Println(conn.RemoteAddr().String() + "| " + err.Error())
		return ""
	}
	return message
}


// Permet de transmettre le message envoyé par un client aux autres
func broadcast(message string, sender net.Conn, player string) {
	// Gestion du mutex
	ClientsLock.Lock()
	defer ClientsLock.Unlock()

	// Préparer le message à envoyer (avec le nom du personnage du client expéditeur)
	fullMessage := fmt.Sprintf("[%s] : %s\n", player, message)

	// Envoyer à tous les clients sauf l'expéditeur
	for client := range Clients {
		if client == sender {
			continue
		}
		// Envoie du message a un client
		_, err := client.Write([]byte(fullMessage))
		if err != nil {
			log.Printf("[ERROR]: lors de l'envoi à %s : %v\n", client.RemoteAddr(), err)
			// Supprimer le client en erreur
			delete(Clients, client)
			client.Close()
		}
	}
}

// Gestion de la connexion pour un client
func handleConnection(conn net.Conn) {
	var player string
	var message string
	// Gestion de la fin de connexion
	defer func() {
		ClientsLock.Lock()
		// Supprimer le client de la liste à la fermeture
		delete(Clients, conn)
		ClientsLock.Unlock()
		// Si un personnage avait été utilisé ont le remer
		// dans la liste des personnages jouables
		if player != "" && player != "900" {
			PersoLock.Lock()
			PersoChoice = append(PersoChoice, player)
			PersoLock.Unlock()
			LogAndPrintf(conn, "%s| [INFO] [%s]: \"Pour ceux qui viendront après\"\n", conn.RemoteAddr().String(), player)
		}
		// Fermeture de la connexion
		conn.Close()
		LogAndPrintf(conn, "%s| [INFO] Client déconnecté\n", conn.RemoteAddr())
	}()

	// Buffer en écoute sur le client
	// reader := bufio.NewReader(conn)
	// Séléction du personnage jouable
	// 	message, err := reader.ReadString('\n')
	// 	if err != nil {
		// 		if err == io.EOF {
			// 			return
			// 		}
			// 		log.Println(conn.RemoteAddr().String() + "| " + err.Error())
			// 		return
			// 	}
	for{
		message = recept(conn)
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
		if message == "CONNECT\n"{
			player = select_Player(conn)
			if player == ""{
				return
			}
		break
		}
		if message == "QUIT\n" || message == ""{
			return
		}
	}
	for {
		message = recept(conn)
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
		if message == "WHO\n"{
			nb_player := 5 - len(PersoChoice)
			LogAndFprintf(conn, "[SERVER] Il y a actuellement %d membre.s d'expédition\n", nb_player)
		}
		if message == "QUIT\n" || message == ""{
			break
		}
		if message == "CHAT\n"{
			chatLoop:
			for{
				LogAndFprintf(conn, "[SERVER] À qui souhaitez vous envoyer un message (GLOBAL, ROOM, GROUP)?\n")
				message = recept(conn)
				LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
				switch message{
				case "GLOBAL\n":
					for{
						chat := ""
						for _, m := range globaleMessage{
							chat += m
						}
						chat = strings.ReplaceAll(chat, "\n", "\\")
						fmt.Fprintf(conn, "%s\n", chat)
						send := recept(conn)
						if send == "EXIT\n" || send == "" || send == "QUIT\n"{
							break chatLoop
							} else{send = fmt.Sprintf("[%s]: %s", player, send)}
						LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), send)
						// Ajouter le nouveau message
						globaleMessage = append(globaleMessage, send)

						// Garder uniquement les 10 derniers éléments
						if len(globaleMessage) > 10 {
							globaleMessage = globaleMessage[len(globaleMessage)-10:]
						}
							// -> iL FAUT PUSH VERS LE HAUT
					}
				}

			}
		}

		// Affichage du message
		// LogAndPrintf(conn, "%s| [%s]: %s", conn.RemoteAddr().String(), player, message)
		// Redistribution aux autres clients
		// broadcast(message, conn, player)
	}
	// Boucle infinie d'écoute des clients
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

// Command (CONNECT, LOOK, MOVE, CHAT, TAKE, DROP, INVENTORY,
// TALK, ATTACK, STATUS, QUEST, QUESTS, WHO, GROUP, QUIT)
