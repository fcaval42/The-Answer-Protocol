package main

import (
	// "fmt"
	"log"
	"net"
	"strings"
	. "the_answer_protocol/tcp/commands"
	. "the_answer_protocol/tcp/common"
)


// Gere la selection du joueur par le client
func select_Player(conn net.Conn) string{
	player := ConnectCom(conn)
	player = strings.Trim(player, "\n")
	switch player {
	case "":
		log.Println("[ERROR]: Player doesn't exist")
		return ""
	case "900":
		return ""
	}

	return player
}
