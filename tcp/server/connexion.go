package main

import (
	"net"
	. "the_answer_protocol/tcp/common"
)

func Connexion(conn net.Conn, player string) string {
	var previous string
	for {
		message := Recept(conn)
		if message == "" {
			return ""
		}
		if message == previous {
			LogServPrintf(conn, "%s| [WARN] %s", conn.RemoteAddr().String(), message)
			LogServFprintf(conn, "[SERVER][WARN] Attention à ne pas envoyer la même commande: %s\n", message)
		} else {
			LogServPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
		}
		previous = message
		if message == "CONNECT\n" {
			player = select_Player(conn)
			if player == "" {
				return ""
			}
			ConnectedLock.Lock()
			ConnectedClients[player] = conn
			ConnectedLock.Unlock()
			return player
		}
		if message == "QUIT\n" || message == "" {
			return ""
		}
	}
}
