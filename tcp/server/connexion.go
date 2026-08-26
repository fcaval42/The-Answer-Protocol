package main

import (
	"net"
	. "the_answer_protocol/tcp/common"
)

func Connexion(conn net.Conn, player string) {	
	for {
		message := Recept(conn)
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
	}}