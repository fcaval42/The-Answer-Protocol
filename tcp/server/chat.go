package main

import (
	"fmt"
	"net"
	. "the_answer_protocol/tcp/common"
	"slices"
)

func GroupChat(conn net.Conn, player string) {
	inGroupChatLock.Lock()
	inGroupChat = append(inGroupChat, conn)
	inGroupChatLock.Unlock()
	defer func() {
		inGroupChatLock.Lock()
		for i, con := range inGroupChat {
			if conn == con {
				inGroupChat = slices.Delete(inGroupChat, i, i+1)
				break
			}
		}
		inGroupChatLock.Unlock()
	}()
	var currentGroup *Group
	currentGroup = SearchPlayerGroup(player)
	if currentGroup == nil {
		LogAndFprintf(conn, "[SERVER] Vous n'êtes pas dans un groupe. Créez-en un d'abord.\n")
		return
	}
	receivers := make([]net.Conn, 0, len(currentGroup.Members))
	for{
		for _, memberConn := range currentGroup.Members {
			if slices.Contains(inGroupChat, memberConn) {
				receivers = append(receivers, memberConn)
			}
		}
		Broadcast(receivers, currentGroup.messages)
		send := Recept(conn)
		if send == "EXIT\n" || send == "" || send == "QUIT\n" {
			return
		} else {
			send = fmt.Sprintf("[%s]: %s", player, send)
		}
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), send)
		currentGroup.Mu.Lock()
		currentGroup.messages = append(currentGroup.messages, send)
		if len(currentGroup.messages) > 10 {
			currentGroup.messages = currentGroup.messages[len(currentGroup.messages)-10:]
		}
		currentGroup.Mu.Unlock()
		}
}

func GlobalChat(conn net.Conn, player string) {
	inGlobalLock.Lock()
	inGlobal = append(inGlobal, conn)
	inGlobalLock.Unlock()
	defer func() {
		inGlobalLock.Lock()
		for i, con := range inGlobal {
			if conn == con {
				inGlobal = slices.Delete(inGlobal, i, i+1)
				break
			}
		}
		inGlobalLock.Unlock()
	}()
	for {
		Broadcast(inGlobal, globaleMessage)
		send := Recept(conn)
		if send == "EXIT\n" || send == "" || send == "QUIT\n" {
			return
		} else {
			send = fmt.Sprintf("[%s]: %s", player, send)
		}
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), send)
		globaleMessageLock.Lock()
		globaleMessage = append(globaleMessage, send)
		if len(globaleMessage) > 10 {
			globaleMessage = globaleMessage[len(globaleMessage)-10:]
		}
		globaleMessageLock.Unlock()
	}
}