package main

import (
	"net"
	. "the_answer_protocol/tcp/common"
)

// Gere la connexion d'un client,
// en attendant que le client envoie la commande "CONNECT"
// et en traitant la selection du joueur.
func Connexion(conn net.Conn, player string) string{
	var previous string
	for {
		message := Recept(conn)
		if message == "" {
			return ""
		}
		if message == previous {
			LogAndPrintf(conn, "%s| [WARN] %s", conn.RemoteAddr().String(), message)
			LogAndFprintf(conn, "[SERVER][WARN] Attention à ne pas envoyer la même commande: %s\n", message)
			} else {
			LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
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
