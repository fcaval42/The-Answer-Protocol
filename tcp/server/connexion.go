package main

import (
	"net"
	. "the_answer_protocol/tcp/common"
)

// Gere la connexion d'un client,
// en attendant que le client envoie la commande "CONNECT"
// et en traitant la selection du joueur.
// Renvoie le nom du joueur choisi ("" si la connexion s'est terminee avant
// que la selection aboutisse), pour que l'appelant (handleConnection) puisse
// le transmettre a Dispatch : sans ce retour, Dispatch ne pourrait pas savoir
// quel joueur executed la commande (LOOK, CHAT, GROUP...).
func Connexion(conn net.Conn, player string) string {
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
	}
}