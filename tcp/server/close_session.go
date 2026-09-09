package main

import (
	"net"
	. "the_answer_protocol/tcp/common"
)

// Gere la fermeture de la session d'un client,
// en supprimant le client de la liste des clients
// connectés et en libérant les ressources associées.
func CloseSession(conn net.Conn, player string) {		
	ClientsLock.Lock()
		// Supprime le client de la liste a la fermeture
		delete(Clients, conn)
		ClientsLock.Unlock()
		// Si un personnage avait ete utilise ont le remet
		// dans la liste des personnages jouables
		GroupLeave(player)
		if player != "" && player != "900" {
			PersoLock.Lock()
			PersoChoice = append(PersoChoice, player)
			PersoLock.Unlock()
			ConnectedLock.Lock()
			delete(ConnectedClients, player)
			ConnectedLock.Unlock()
		}
			LogAndPrintf(player, conn, "%s| [INFO] [%s]: \"Pour ceux qui viendront après\"\n", conn.RemoteAddr().String(), player)
		
		// Fermeture de la connexion
		conn.Close()
		LogAndPrintf(player, conn, "%s| [INFO] Client déconnecté\n", conn.RemoteAddr())
	}