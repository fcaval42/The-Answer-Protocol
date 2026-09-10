package main

import (
	"log"
	"net"
	"strings"
	. "the_answer_protocol/tcp/commands"
)

func select_Player(conn net.Conn) string {
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
