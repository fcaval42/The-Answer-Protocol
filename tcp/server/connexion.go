package main

import (
	"net"
	. "the_answer_protocol/tcp/common"
)

// Gere la connexion d'un client,
// en attendant que le client envoie la commande "CONNECT"
// et en traitant la selection du joueur.
func Connexion(conn net.Conn, player string) string{	
	for {
		message := Recept(conn)
		if message != "" {
			LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
		}
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
	}}