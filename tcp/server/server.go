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

var (
    pendingInvites     = make(map[string]chan string)
    pendingInvitesLock sync.Mutex
)

type Group struct {
	ID      string
	Members map[string]net.Conn
	messages []string
	Mu      sync.RWMutex
}

var (
	groups     = make(map[string]*Group)
	groupsLock sync.Mutex
)

var (
	ConnectedClients     = make(map[string]net.Conn)
	ConnectedLock sync.Mutex
)

var (
	globaleMessage     []string
	globaleMessageLock sync.Mutex
)
var (inGroupChat	 []net.Conn
	inGroupChatLock sync.Mutex
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
	groupID := fmt.Sprintf("%d", len(groups)+1)

	// Creer une nouvelle instance de Group
	newGroup := &Group{
		ID:      groupID,
		Members: make(map[string]net.Conn),
	}

	// Ajouter le createur du groupe à la liste des membres
	newGroup.Members[player] = conn

	// Ajouter le groupe a la liste de groupes globale
	groupsLock.Lock()
	groups[groupID] = newGroup
	groupsLock.Unlock()

	LogAndFprintf(conn, "[SERVER] Groupe créé avec succès ! ID du groupe : %s\n", groupID)
}

func searchPlayerGroup(player string) *Group {
	var currentGroup *Group
	groupsLock.Lock()
	for _, g := range groups {
		if _, exists := g.Members[player]; exists {
			currentGroup = g
			break
		}
	}
	groupsLock.Unlock()
	return currentGroup
}

func GroupInvite(conn net.Conn, player string) {
	// Trouver le groupe du joueur actuel
	var currentGroup *Group
	currentGroup = searchPlayerGroup(player)
	if currentGroup == nil {
		LogAndFprintf(conn, "[SERVER] Vous n'êtes pas dans un groupe. Créez-en un d'abord.\n")
		return
	}

	// Recupere la liste des joueurs disponibles pour l'invitation
	availablePlayers := getAvailablePlayers(player)

	if len(availablePlayers) == 0 {
		LogAndFprintf(conn, "[SERVER] Aucun joueur disponible pour l'invitation.\n")
		return
	}	
	LogAndFprintf(conn, "[SERVER] Joueurs disponibles : %s\n", strings.Join(availablePlayers, ", "))
	var invitee string
	var inviteeConn net.Conn
	var exists bool
	for{
		// Reponse du joueur pour choisir qui inviter
		invitee = recept(conn)
		invitee = strings.TrimSpace(invitee)
	
		if !isPlayerAvailable(invitee, availablePlayers) {
			LogAndFprintf(conn, "[SERVER] Joueur invalide ou déjà dans un groupe.\\Joueurs disponibles : %s\n", strings.Join(availablePlayers, ", "))
			continue
		}
		inviteeConn, exists = ConnectedClients[invitee]
		if !exists {
			LogAndFprintf(conn, "[SERVER] Joueur non connecté.\n")
			continue
		}
		LogAndFprintf(inviteeConn, "[INFO] %s vous a invité à rejoindre un groupe. ACCEPT/DECLINE\n", player)
		break
	}
	replyChan := make(chan string)

    pendingInvitesLock.Lock()
    pendingInvites[invitee] = replyChan
    pendingInvitesLock.Unlock()

    // Nettoyage après l'invitation
    defer func() {
        pendingInvitesLock.Lock()
        delete(pendingInvites, invitee)
        pendingInvitesLock.Unlock()
    }()

    fmt.Fprintf(inviteeConn, "[INFO] %s vous a invité à rejoindre un groupe. ACCEPT/DECLINE\n", player)

    // 2. Attendre la réponse transmise par le channel
    response := <-replyChan

    if response == "DECLINE\n" {
        LogAndFprintf(conn, "[SERVER] %s a refusé l'invitation.\n", invitee)
        return
    }
    if response == "ACCEPT\n" {
        LogAndFprintf(conn, "[SERVER] %s a accepté l'invitation.\n", invitee)
    } else {
        LogAndFprintf(inviteeConn, "[INFO] Réponse invalide.\n")
        return
    }

    groupsLock.Lock()
    defer groupsLock.Unlock()
    currentGroup.Members[invitee] = inviteeConn
    LogAndFprintf(conn, "[SERVER] %s a rejoint le groupe %s !\n", invitee, currentGroup.ID)
}

func getAvailablePlayers(self string) []string {
	ClientsLock.Lock()
	defer ClientsLock.Unlock()
	var availablePlayers []string
	for player := range ConnectedClients {
		if player != self {
			if searchPlayerGroup(player) == nil {
				availablePlayers = append(availablePlayers, player)
			}
		}
	}
	return availablePlayers
}

func isPlayerAvailable(player string, availablePlayers []string) bool {
	for _, availablePlayer := range availablePlayers {
		if availablePlayer == player {
			if searchPlayerGroup(player) != nil {
				return false
			}
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
	if strings.Contains(message, "\\") {
		message = strings.ReplaceAll(message, "\\", "\n")
	}
	return message
}

func roomChat(conn net.Conn, room string) {

}

func groupChat(conn net.Conn, player string) {
	inGroupChatLock.Lock()
	inGroupChat = append(inGroupChat, conn)
	inGroupChatLock.Unlock()
	defer func() {
		inGroupChatLock.Lock()
		for i, con := range inGroupChat {
			if conn == con {
				inGroupChat = slices.Delete(inGroupChat, i, i+1)
				break
			}
		}
		inGroupChatLock.Unlock()
	}()
	var currentGroup *Group
	currentGroup = searchPlayerGroup(player)
	if currentGroup == nil {
		LogAndFprintf(conn, "[SERVER] Vous n'êtes pas dans un groupe. Créez-en un d'abord.\n")
		return
	}
	receivers := make([]net.Conn, 0, len(currentGroup.Members))
	for{
		for _, memberConn := range currentGroup.Members {
			if slices.Contains(inGroupChat, memberConn) {
				receivers = append(receivers, memberConn)
			}
		}
		broadcast(receivers, currentGroup.messages)
		send := recept(conn)
		if send == "EXIT\n" || send == "" || send == "QUIT\n" {
			return
		} else {
			send = fmt.Sprintf("[%s]: %s", player, send)
		}
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), send)
		currentGroup.Mu.Lock()
		currentGroup.messages = append(currentGroup.messages, send)
		if len(currentGroup.messages) > 10 {
			currentGroup.messages = currentGroup.messages[len(currentGroup.messages)-10:]
		}
		currentGroup.Mu.Unlock()
		}
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
		broadcast(inGlobal, globaleMessage)
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

func printChat(conn net.Conn, messages []string) {
	chat := ""
	for _, m := range messages {
		chat += m
	}
	chat = strings.ReplaceAll(chat, "\n", "\\")
	fmt.Fprintf(conn, "%s\n", chat)
}

// Permet de transmettre le message envoye par un client aux autres
func broadcast(receiver []net.Conn, messages []string) {
	for _, client := range receiver {
		printChat(client, messages)
	}
}

// Gestion de la connexion pour un client
func handleConnection(conn net.Conn) {
	var player string
	var message string

	// Gestion de la fin de connexion
	defer func() {
		ClientsLock.Lock()
		// Supprime le client de la liste a la fermeture
		delete(Clients, conn)
		ClientsLock.Unlock()
		// Si un personnage avait ete utilise ont le remet
		// dans la liste des personnages jouables
		if searchPlayerGroup(player) != nil {
			currentGroup := searchPlayerGroup(player)
			currentGroup.Mu.Lock()
			delete(currentGroup.Members, player)
			currentGroup.Mu.Unlock()
		}
		if player != "" && player != "900" {
			PersoLock.Lock()
			PersoChoice = append(PersoChoice, player)
			PersoLock.Unlock()
			ConnectedLock.Lock()
			delete(ConnectedClients, player)
			ConnectedLock.Unlock()
		}
			LogAndPrintf(conn, "%s| [INFO] [%s]: \"Pour ceux qui viendront après\"\n", conn.RemoteAddr().String(), player)
		
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
			ConnectedLock.Lock()
			ConnectedClients[player] = conn
			ConnectedLock.Unlock()
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

		pendingInvitesLock.Lock()
		ch, hasPendingInvite := pendingInvites[player]
		pendingInvitesLock.Unlock()

		if hasPendingInvite && (message == "ACCEPT\n" || message == "DECLINE\n") {
			ch <- message // On envoie le choix dans le channel vers GroupInvite
			continue      // On passe au message suivant sans exécuter la suite de handleConnection
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
					case "GROUP\n":
						groupChat(conn, player)
						break chatLoop
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
