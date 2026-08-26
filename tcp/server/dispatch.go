package main

import (
	. "the_answer_protocol/tcp/common"
	"net"
	"strings"
)

// Dispatch des commandes du client après la connexion,
// en fonction de la commande reçue.
func Dispatch(conn net.Conn, player string) {	
	for {
		message := Recept(conn)
		pendingInvitesLock.Lock()
		ch, hasPendingInvite := pendingInvites[player]
		pendingInvitesLock.Unlock()
		switch message{
		case "":
			return
		case "WHO\n":
			nb_player := 5 - len(PersoChoice)
			LogAndFprintf(conn, "[SERVER] Il y a actuellement %d membre.s d'expédition\n", nb_player)
		case "ACCEPT\n", "DECLINE\n":
			if hasPendingInvite{
				ch <- message
				continue
			}
		case "CHAT\n":
			for {
				LogAndFprintf(conn, "[SERVER] À qui souhaitez vous envoyer un message (GLOBAL, ROOM, GROUP)?\n")
				message := Recept(conn)
				if message == "" {
					return
				}
				LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
				switch message {
				case "GLOBAL\n":
					GlobalChat(conn, player)
					return
					// case "ROOM":
					// 	roomChat()
					// 	break chatLoop
					case "GROUP\n":
						GroupChat(conn, player)
						return
				}
			}
		case "GROUP CREATE\n":
			GroupCreate(conn, player)
		case "GROUP INVITE\n":
			GroupInvite(conn, player)
		case "GROUP JOIN\n":
			if DisplayGroup(conn){
				groupID := Recept(conn)
				groupID = strings.TrimSpace(groupID)
				GroupJoin(conn, player, groupID)
			}
		case "GROUP LEAVE\n":
			if GroupLeave(player) {
				LogAndFprintf(conn, "[SERVER] Vous avez quitté le groupe.\n")
			} else {
				LogAndFprintf(conn, "[SERVER] Vous n'êtes pas dans un groupe.\n")
			}
		case "QUIT\n":
			return
		}
		}
	}