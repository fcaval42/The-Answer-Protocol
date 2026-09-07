package main

import (
	"fmt"
	"net"
	"strings"
	"time"
	. "the_answer_protocol/tcp/common"
)


func GroupCreate(conn net.Conn, player string) {
	if SearchPlayerGroup(player) != nil {
		LogAndFprintf(conn, "[SERVER] ERR 402 ALREADY_IN_GROUP: Vous êtes déjà dans un groupe. Quittez-le avant d'en créer un nouveau.\n")
		return
	}
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

func GroupJoin(conn net.Conn, player string, groupID string) {
	if SearchPlayerGroup(player) != nil {
		LogAndFprintf(conn, "[SERVER] ERR 402 ALREADY_IN_GROUP: Vous êtes déjà dans un groupe. Quittez-le avant d'en rejoindre un autre.\n")
		return
	}

	groupsLock.Lock()
	defer groupsLock.Unlock()
	group, exists := groups[groupID]
	if !exists {
		LogAndFprintf(conn, "[SERVER] Groupe introuvable.\n")
		return
	}

	group.Mu.Lock()
	if _, memberExists := group.Members[player]; memberExists {
		LogAndFprintf(conn, "[SERVER] Vous êtes déjà dans ce groupe.\n")
		return
	}
	group.Members[player] = conn
	group.messages = append(group.messages, fmt.Sprintf("%s a rejoint le groupe.\n", player))
	group.Mu.Unlock()
	LogAndFprintf(conn, "[SERVER] Vous avez rejoint le groupe %s !\n", groupID)
}


func GroupInvite(conn net.Conn, player string) {
	// Trouver le groupe du joueur actuel
	var currentGroup *Group
	currentGroup = SearchPlayerGroup(player)
	if currentGroup == nil {
		LogAndFprintf(conn, "[SERVER] ERR 401 NOT_IN_GROUP: Vous n'êtes pas dans un groupe. Créez-en un d'abord.\n")
		return
	}

	// Recupere la liste des joueurs disponibles pour l'invitation
	availablePlayers := GetAvailablePlayers(player)

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
		invitee = Recept(conn)
		invitee = strings.TrimSpace(invitee)
	
		if !IsPlayerAvailable(invitee, availablePlayers) {
			LogAndFprintf(conn, "[SERVER] Joueur invalide ou déjà dans un groupe.\\Joueurs disponibles : %s\n", strings.Join(availablePlayers, ", "))
			continue
		}
		inviteeConn, exists = ConnectedClients[invitee]
		if !exists {
			LogAndFprintf(conn, "[SERVER] Joueur non connecté.\n")
			continue
		}
		break
	}
	replyChan := make(chan string, 1)

    pendingInvitesLock.Lock()
    pendingInvites[invitee] = replyChan
    pendingInvitesLock.Unlock()

    // Nettoyage après l'invitation
    defer func() {
        pendingInvitesLock.Lock()
        delete(pendingInvites, invitee)
        pendingInvitesLock.Unlock()
    }()

    fmt.Fprintf(inviteeConn, "[INFO] %s vous a invité à rejoindre un groupe. ACCEPT/DECLINE (10s)\n", player)

	var response string
	select {
	case response = <-replyChan:
	case <-time.After(10 * time.Second):
		response = "DECLINE\n"
		fmt.Fprintf(inviteeConn, "[INFO] Invitation expirée.\n")
	}

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
	currentGroup.messages = append(currentGroup.messages, fmt.Sprintf("%s a rejoint le groupe.\n", invitee))
    LogAndFprintf(conn, "[SERVER] %s a rejoint le groupe %s !\n", invitee, currentGroup.ID)
}

func GroupLeave(player string) bool{
	if SearchPlayerGroup(player) != nil {
		currentGroup := SearchPlayerGroup(player)
		currentGroup.Mu.Lock()
		currentGroup.messages = append(currentGroup.messages, fmt.Sprintf("%s a quitté le groupe.\n", player))
		delete(currentGroup.Members, player)
		currentGroup.Mu.Unlock()
		return true
	}
	return false
}
