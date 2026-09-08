package main

import (
	"fmt"
	"net"
	"slices"
	"sync"
	. "the_answer_protocol/tcp/common"
)

var (
	inRoomChat   map[string][]net.Conn
	inRoomChatMu sync.Mutex
)

var (
	roomMessages   = make(map[string][]string)
	roomMessagesMu sync.Mutex
)

func RoomChat(conn net.Conn, loc string, player string) {
	inRoomChatMu.Lock()
	if inRoomChat == nil {
		inRoomChat = make(map[string][]net.Conn)
	}
	inRoomChat[loc] = append(inRoomChat[loc], conn)
	inRoomChatMu.Unlock()
	defer func() {
		inRoomChatMu.Lock()
		for i, con := range inRoomChat[loc] {
			if conn == con {
				inRoomChat[loc] = slices.Delete(inRoomChat[loc], i, i+1)
				break
			}
		}
		inRoomChatMu.Unlock()
	}()
	for {
		inRoomChatMu.Lock()
		Broadcast(inRoomChat[loc], roomMessages[loc])
		inRoomChatMu.Unlock()
		send := Recept(conn)
		if send == "EXIT\n" || send == "" || send == "QUIT\n" {
			LogAndFprintf(conn, "[SERVER] Vous avez quitté le chat de la salle.\n")
			return
		} else {
			send = fmt.Sprintf("[%s]: %s", player, send)
		}
		var otherMembers []net.Conn
		mapsLock.Lock()
		locate := maps[loc]
		mapsLock.Unlock()
		ConnectedLock.Lock()
		inRoomChatMu.Lock()
		for _, play := range locate.Players{
			c := ConnectedClients[play.Name]
			if !slices.Contains(inRoomChat[loc], c) && c != conn{
				otherMembers = append(otherMembers, c)
			}
		}
		ConnectedLock.Unlock()
		inRoomChatMu.Unlock()
		BroadcastEVT(otherMembers, fmt.Sprintf("[EVT] %s", send))
		roomMessagesMu.Lock()
		roomMessages[loc] = append(roomMessages[loc], send)
		if len(roomMessages[loc]) > 10 {
			roomMessages[loc] = roomMessages[loc][len(roomMessages[loc])-10:]
		}
		roomMessagesMu.Unlock()
	}
}

func GroupChat(conn net.Conn, player string) {
	inGroupChatLock.Lock()
	if inGroupChat == nil {
		inGroupChat = make(map[string][]net.Conn)
	}
	var currentGroup *Group
	currentGroup = SearchPlayerGroup(player)
	if currentGroup == nil {
		LogAndFprintf(conn, "[SERVER] Vous n'êtes pas dans un groupe. Créez-en un d'abord.\n")
		inGroupChatLock.Unlock()
		return
	}
	inGroupChat[currentGroup.ID] = append(inGroupChat[currentGroup.ID], conn)
	inGroupChatLock.Unlock()
	defer func() {
		inGroupChatLock.Lock()
		for i, con := range inGroupChat[currentGroup.ID] {
			if conn == con {
				group := inGroupChat[currentGroup.ID]
				group = slices.Delete(inGroupChat[currentGroup.ID], i, i+1)
				inGroupChat[currentGroup.ID] = group
				break
			}
		}
		inGroupChatLock.Unlock()
	}()
	for {
		receivers := make([]net.Conn, 0, len(currentGroup.Members))
		inGroupChatLock.Lock()
		for _, memberConn := range inGroupChat[currentGroup.ID] {
			receivers = append(receivers, memberConn)
		}
		inGroupChatLock.Unlock()
		var otherMembers []net.Conn
		for _, c := range currentGroup.Members{
			if !slices.Contains(receivers, c) && c != conn{
				otherMembers = append(otherMembers, c)
			}
		}
		Broadcast(receivers, currentGroup.messages)
		send := Recept(conn)
		if send == "EXIT\n" || send == "" || send == "QUIT\n" {
			LogAndFprintf(conn, "[SERVER] Vous avez quitté le chat de groupe.\n")
			return
			} else {
				send = fmt.Sprintf("[%s]: %s", player, send)
			}
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), send)
		currentGroup.Mu.Lock()
		BroadcastEVT(otherMembers, fmt.Sprintf("[EVT] %s", send))
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
			LogAndFprintf(conn, "[SERVER] Vous avez quitté le chat global.\n")
			return
		} else {
			send = fmt.Sprintf("[%s]: %s", player, send)
		}
		ConnectedLock.Lock()
		var otherMembers []net.Conn
		for _, c := range ConnectedClients{
			if !slices.Contains(inGlobal, c) && c != conn{
				otherMembers = append(otherMembers, c)
			}
		}
		ConnectedLock.Unlock()
		BroadcastEVT(otherMembers, fmt.Sprintf("[EVT] %s", send))
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), send)
		globaleMessageLock.Lock()
		globaleMessage = append(globaleMessage, send)
		if len(globaleMessage) > 10 {
			globaleMessage = globaleMessage[len(globaleMessage)-10:]
		}
		globaleMessageLock.Unlock()
	}
}

func GlobalChatBroadcast(conn net.Conn, message string) {
	globaleMessageLock.Lock()
	globaleMessage = append(globaleMessage, message)
	if len(globaleMessage) > 10 {
		globaleMessage = globaleMessage[len(globaleMessage)-10:]
	}
	globaleMessageLock.Unlock()
	inGlobalLock.Lock()
	ConnectedLock.Lock()
	var receivers []net.Conn
	for _, c := range ConnectedClients {
		if !slices.Contains(inGlobal, c) && c != conn{
			receivers = append(receivers, c)
		}
	}
	ConnectedLock.Unlock()
	Broadcast(inGlobal, globaleMessage)
	inGlobalLock.Unlock()
	message = fmt.Sprintf("[EVT] %s", message)
	BroadcastEVT(receivers, message)
}

func RoomChatBroadcast(conn net.Conn, loc string, player string, message string) {
	roomMessagesMu.Lock()
	message = fmt.Sprintf("[%s]: %s", player, message)
	roomMessages[loc] = append(roomMessages[loc], message)
	if len(roomMessages[loc]) > 10 {
		roomMessages[loc] = roomMessages[loc][len(roomMessages[loc])-10:]
	}
	inRoomChatMu.Lock()
	Broadcast(inRoomChat[loc], roomMessages[loc])
	var receivers []net.Conn
	mapsLock.Lock()
	ConnectedLock.Lock()
	for _, c := range maps[loc].Players {
		if c.Name != player && !slices.Contains(inRoomChat[loc], ConnectedClients[c.Name]) {
			receivers = append(receivers, ConnectedClients[c.Name])
		}
	}
	inRoomChatMu.Unlock()
	ConnectedLock.Unlock()
	mapsLock.Unlock()
	roomMessagesMu.Unlock()
	BroadcastEVT(receivers, fmt.Sprintf("[EVT] %s", message))
}

func GroupChatBroadcast(conn net.Conn, message string, player string) {
	currentGroup := SearchPlayerGroup(player)
	if currentGroup == nil {
		LogAndFprintf(conn, "[SERVER] ERR 401 NOT_IN_GROUP: Vous n'êtes pas dans un groupe. Créez-en un d'abord.\n")
		return
	}
	currentGroup.Mu.Lock()
	message = fmt.Sprintf("[%s]: %s", player, message)
	currentGroup.messages = append(currentGroup.messages, message)
	if len(currentGroup.messages) > 10 {
		currentGroup.messages = currentGroup.messages[len(currentGroup.messages)-10:]
	}
	inGroupChatLock.Lock()
	var receivers []net.Conn
	for _, c := range currentGroup.Members {
		if !slices.Contains(inGroupChat[currentGroup.ID], c) && c != conn {
			receivers = append(receivers, c)
		}
	}
	receiver := make([]net.Conn, 0, len(currentGroup.Members))
	for _, memberConn := range inGroupChat[currentGroup.ID] {
		receiver = append(receiver, memberConn)
	}
	Broadcast(receiver, currentGroup.messages)
	currentGroup.Mu.Unlock()
	inGroupChatLock.Unlock()
	message = fmt.Sprintf("[EVT] %s", message)
	BroadcastEVT(receivers, message)
}
