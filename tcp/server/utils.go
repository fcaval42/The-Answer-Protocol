package main


import (
	"fmt"
	"net"
	"strings"
	. "the_answer_protocol/tcp/common"
	"log"
	"io"
	"bufio"
)

func DisplayGroup(conn net.Conn) bool{
	groupsLock.Lock()
	defer groupsLock.Unlock()
	var result string
	if len(groups) == 0 {
		LogAndFprintf(conn, "[SERVER] Aucun groupe n'existe actuellement.\n")
		return false
	}
	for _, group := range groups {
		group.Mu.Lock()
		result = fmt.Sprintf("Groupe %s :\\", group.ID)
		for player, connect := range group.Members {
			if conn == connect {
				LogAndFprintf(conn, "[SERVER] Vous êtes déjà dans le groupe %s.\n", group.ID)
				group.Mu.Unlock()
				return false
			}
			result += fmt.Sprintf("  - %s\\", player)
		}
		group.Mu.Unlock()
	}
	result += "Entrez l'ID du groupe que vous souhaitez rejoindre :\\"
	fmt.Fprintf(conn, "[SERVER] %s\n", result)
	return true
}

func GetAvailablePlayers(self string) []string {
	ClientsLock.Lock()
	defer ClientsLock.Unlock()
	var availablePlayers []string
	for player := range ConnectedClients {
		if player != self {
			if SearchPlayerGroup(player) == nil {
				availablePlayers = append(availablePlayers, player)
			}
		}
	}
	return availablePlayers
}

func IsPlayerAvailable(player string, availablePlayers []string) bool {
	for _, availablePlayer := range availablePlayers {
		if availablePlayer == player {
			if SearchPlayerGroup(player) != nil {
				return false
			}
			return true
		}
	}
	return false
}

func Recept(conn net.Conn) string {
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

func SearchPlayerGroup(player string) *Group {
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


func PrintChat(conn net.Conn, messages []string) {
	chat := ""
	for _, m := range messages {
		chat += m
	}
	chat = strings.ReplaceAll(chat, "\n", "\\")
	fmt.Fprintf(conn, "%s\n", chat)
}

func Broadcast(receiver []net.Conn, messages []string) {
	for _, client := range receiver {
		PrintChat(client, messages)
	}
}