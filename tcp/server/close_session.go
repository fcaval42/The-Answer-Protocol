package main

import (
	"net"
	. "the_answer_protocol/tcp/common"
)

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
			LogAndPrintf(conn, "%s| [INFO] [%s]: \"Pour ceux qui viendront après\"\n", conn.RemoteAddr().String(), player)
		
		// Fermeture de la connexion
		conn.Close()
		LogAndPrintf(conn, "%s| [INFO] Client déconnecté\n", conn.RemoteAddr())
	}