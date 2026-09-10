package main

import (
	"net"
	. "the_answer_protocol/tcp/common"
)

func CloseSession(conn net.Conn, player string) {
	ClientsLock.Lock()
	delete(Clients, conn)
	ClientsLock.Unlock()
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

	conn.Close()
	LogAndPrintf(player, conn, "%s| [INFO] Client déconnecté\n", conn.RemoteAddr())
}
