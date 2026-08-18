package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"the_answer_protocol/tcp/common"
)

// Création du'une map (dictionnaire) pour gérer tous les clients
// Protégé par un mutex
var (
	clients     = make(map[net.Conn]bool) // Map des connexions actives
	clientsLock sync.Mutex                // Mutex pour synchroniser l'accès à la map
)

// Fonction isdigit pour eviter de faire un atoi sur des caractères
func IsDigit(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Permet de vérifier que le choix de l'utilisateur est dans
// la liste de personnage
func isInList(num int) bool {
	persoLen := len(common.PersoChoice)
	return num >= 1 && num <= persoLen
}

// Permet de retirer l'indice i lorsque le joueur à choisi son personnage,
// le rendant indisponible pour les autres
func removeUnordered(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// Selection du personnage joueur
func selectPlayer(conn net.Conn) string {
	// Lock du mutex pour le rendre indisponible pour les autres
	common.PersoLock.Lock()
	// Le mutex sera unlock à la fin de la fonction selectPlayer
	defer common.PersoLock.Unlock()
	// Initialisation d'un buffer en écoute sur le client
	reader := bufio.NewReader(conn)

	// Permet l'affichage formaté "1. nom, 2. nom, etc..."
	charactere := make([]string, len(common.PersoChoice))
	for i, perso := range common.PersoChoice {
		charactere[i] = fmt.Sprintf("%d. %s", i+1, perso)
	}
	msg := fmt.Sprintf("[WARN] Choisissez votre joueur parmi en entrant le numéro correspondant: %s\n", strings.Join(charactere, ", "))
	common.LogAndFprintf(conn, msg)

	// Boucle infinie tant que le joueur n'a pas effectué un bon choix
	for {
		// Recois le choix du joueur par le client
		player, err := reader.ReadString('\n')
		common.LogAndPrintf(conn, conn.RemoteAddr().String()+"| " + player)
		if player != "\n" {
			if err != nil {
				log.Println(conn.RemoteAddr().String() + "| [ERROR] " + err.Error())
				return ""
			}

			// Verifie qu'il s'agit bien de nombres
			if IsDigit(strings.Trim(player, "\n")) {
				// Conversion en int
				indextest, err := strconv.Atoi(strings.Trim(player, "\n"))
				if err != nil {
					return ""
				}
				// On verifie que le choix correspond à la liste de personnage
				if isInList(indextest) {
					player = common.PersoChoice[indextest-1]
					// On supprime le choix fait par le joueur des choix pour les prochains joueurs
					common.PersoChoice = removeUnordered(common.PersoChoice, indextest-1)
					// On retourne le personnage choisi
					return player
				}
			}
		}
		common.LogAndFprintf(conn, msg)
	}
}

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
		if player != "" {
			common.PersoLock.Lock()
			common.PersoChoice = append(common.PersoChoice, player)
			common.PersoLock.Unlock()
		}
		// Fermeture de la connexion
		conn.Close()
		common.LogAndPrintf(conn, "[INFO] [%s]: \"Pour ceux qui viendront après\"\n", player)
		common.LogAndPrintf(conn, "[INFO] Client déconnecté : %s\n", conn.RemoteAddr())
	}()

	// Buffer en écoute sur le client
	reader := bufio.NewReader(conn)
	// Séléction du personnage jouable
	player = selectPlayer(conn)
	player = strings.Trim(player, "\n")
	if player == "" {
		log.Println("[ERROR]: Player doesn't exist")
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
		common.LogAndPrintf(conn, "%s| [%s]: %s", conn.RemoteAddr().String(), player, message)
		// Redistribution aux autres clients
		broadcast(message, conn, player)
	}
}

func main() {
	logfile, _ := os.Create("protocol.log")
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

		common.LogAndPrintf(conn, "[INFO] Nouveau client connecté : %s\n", conn.RemoteAddr())

		// Gestion individuelle (Thread) du client
		go handleConnection(conn)
	}
}

// Command (CONNECT, LOOK, MOVE, CHAT, TAKE, DROP, INVENTORY,
// TALK, ATTACK, STATUS, QUEST, QUESTS, WHO, GROUP, QUIT)
