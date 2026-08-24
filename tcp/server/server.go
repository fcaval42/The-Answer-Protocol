package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"slices"
	"strings"
	"sync"
	. "the_answer_protocol/tcp/common"
)

type Group struct {
	ID      string
	Members map[net.Conn]string // map[connexion]nom_du_joueur
	Mu      sync.RWMutex
}

var (
	groups     = make(map[string]*Group) // Map d'ID de groupe vers la struct Group
	groupsLock sync.RWMutex
)

var (
	globaleMessage     []string
	globaleMessageLock sync.Mutex
)
var (
	inGlobal     []net.Conn
	inGlobalLock sync.Mutex
)
var (
	group     [][]net.Conn
	groupLock sync.Mutex
)

func GroupCreate(conn net.Conn, player string) {
	LogAndFprintf(conn, "[SERVER] Création d'un nouveau groupe...\n")
	// Générer un ID unique pour le groupe (vous pouvez utiliser un UUID ou un compteur)
	groupID := fmt.Sprintf("group-%d", len(groups)+1)

	// Créer une nouvelle instance de Group
	newGroup := &Group{
		ID:      groupID,
		Members: make(map[net.Conn]string),
	}

	// Ajouter le créateur du groupe à la liste des membres
	newGroup.Members[conn] = player

	// Ajouter le groupe à la map globale
	groupsLock.Lock()
	groups[groupID] = newGroup
	groupsLock.Unlock()

	LogAndFprintf(conn, "[SERVER] Groupe créé avec succès ! ID du groupe : %s\n", groupID)
}

func GroupInvite(conn net.Conn, player string) {
	LogAndFprintf(conn, "[SERVER] Quel membre d'expédition souhaitez-vous inviter dans votre groupe ?\n")
	// Récupérer la liste des joueurs disponibles pour l'invitation
	availablePlayers := getAvailablePlayers()

	if len(availablePlayers) == 0 {
		LogAndFprintf(conn, "[SERVER] Aucun joueur disponible pour l'invitation.\n")
		return
	}

	// Afficher la liste des joueurs disponibles
	LogAndFprintf(conn, "[SERVER] Joueurs disponibles : %s\n", strings.Join(availablePlayers, ", "))

	// Attendre la réponse du joueur pour choisir qui inviter
	invitee := recept(conn)
	invitee = strings.TrimSpace(invitee)

	if !isPlayerAvailable(invitee, availablePlayers) {
		LogAndFprintf(conn, "[SERVER] Joueur invalide ou indisponible pour l'invitation.\n")
		return
	}

	// Ajouter le joueur invité au groupe (vous pouvez choisir le groupe approprié ici)
	groupsLock.Lock()
	defer groupsLock.Unlock()

	// Trouver le groupe du joueur actuel
	var currentGroup *Group
	for _, g := range groups {
		if _, exists := g.Members[conn]; exists {
			currentGroup = g
			break
		}
	}

	if currentGroup == nil {
		LogAndFprintf(conn, "[SERVER] Vous n'êtes pas dans un groupe. Créez-en un d'abord.\n")
		return
	}

	// Ajouter le joueur invité au groupe
	currentGroup.Members[conn] = invitee
	LogAndFprintf(conn, "[SERVER] Joueur %s invité avec succès dans le groupe %s !\n", invitee, currentGroup.ID)
}

func getAvailablePlayers() []string {
	PersoLock.Lock()
	defer PersoLock.Unlock()

	var availablePlayers []string
	for _, player := range PersoChoice {
		availablePlayers = append(availablePlayers, player)
	}
	return availablePlayers
}

func isPlayerAvailable(player string, availablePlayers []string) bool {
	for _, availablePlayer := range availablePlayers {
		if availablePlayer == player {
			return true
		}
	}
	return false
}

func recept(conn net.Conn) string {
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

func roomChat(conn net.Conn, room string) {

}

func groupChat(conn net.Conn) {

}

func globalChat(conn net.Conn, player string) {
	inGlobalLock.Lock()
	inGlobal = append(inGlobal, conn)
	inGlobalLock.Unlock()
	defer func() {
		inGlobalLock.Lock()
		for i, con := range inGlobal {
			if conn == con {
				inGlobal = slices.Delete(inGlobal, i, i+1)
				break
			}
		}
		inGlobalLock.Unlock()
	}()
	for {
		printChat(conn)
		broadcast(conn)
		send := recept(conn)
		if send == "EXIT\n" || send == "" || send == "QUIT\n" {
			return
		} else {
			send = fmt.Sprintf("[%s]: %s", player, send)
		}
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), send)
		globaleMessageLock.Lock()
		globaleMessage = append(globaleMessage, send)
		if len(globaleMessage) > 10 {
			globaleMessage = globaleMessage[len(globaleMessage)-10:]
		}
		globaleMessageLock.Unlock()
	}
}

func printChat(conn net.Conn) {
	chat := ""
	for _, m := range globaleMessage {
		chat += m
	}
	chat = strings.ReplaceAll(chat, "\n", "\\")
	fmt.Fprintf(conn, "%s\n", chat)
}

// Permet de transmettre le message envoyé par un client aux autres
func broadcast(sender net.Conn) {
	// Gestion du mutex
	inGlobalLock.Lock()
	defer inGlobalLock.Unlock()

	// Envoyer à tous les clients sauf l'expéditeur
	for _, client := range inGlobal {
		if client == sender {
			continue
		}
		printChat(client)
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
	for {
		message = recept(conn)
		if message != "" {
			LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
		}
		if message == "CONNECT\n" {
			player = select_Player(conn)
			if player == "" {
				return
			}
			break
		}
		if message == "QUIT\n" || message == "" {
			return
		}
	}
	for {
		message = recept(conn)
		if message == "" {
			break
		}
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
		if message == "WHO\n" {
			nb_player := 5 - len(PersoChoice)
			LogAndFprintf(conn, "[SERVER] Il y a actuellement %d membre.s d'expédition\n", nb_player)
		}
		if message == "QUIT\n" || message == "" {
			break
		}
		if message == "CHAT\n" {
		chatLoop:
			for {
				LogAndFprintf(conn, "[SERVER] À qui souhaitez vous envoyer un message (GLOBAL, ROOM, GROUP)?\n")
				message = recept(conn)
				if message == "" {
					break
				}
				LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
				switch message {
				case "GLOBAL\n":
					globalChat(conn, player)
					break chatLoop
					// case "ROOM":
					// 	roomChat()
					// 	break chatLoop
					// case "GROUP":
					// 	groupeChat()
					// 	break chatLoop
				}
			}
		}
		if message == "GROUP CREATE\n" {
			GroupCreate(conn, player)
		}
		if message == "GROUP INVITE\n" {
			GroupInvite(conn, player)

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
