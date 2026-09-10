package main

import (
	"fmt"
	"net"
	"strings"
	. "the_answer_protocol/parsing"
	. "the_answer_protocol/tcp"
	. "the_answer_protocol/tcp/commands"
	. "the_answer_protocol/tcp/common"
)

func dropItem(conn net.Conn, player Player, loc Location, message string) Location {
	itemName := strings.TrimSpace(strings.TrimPrefix(message, "DROP"))
	mapsLock.Lock()
	defer mapsLock.Unlock()
	for i, p := range loc.Players {
		if p.Name == player.Name {
			for j, item := range p.Inventory {
				if item.Name == itemName {
					loc.Players[i].Inventory = append(loc.Players[i].Inventory[:j], loc.Players[i].Inventory[j+1:]...)
					if loc.Players[i].EquippedItem == item.ID {
						loc.Players[i].Weapon = loc.Players[i].BaseWeapon
						loc.Players[i].EquippedItem = ""
					}
					loc.Items = append(loc.Items, itemName)
					maps[loc.ID] = loc
					LogAndFprintf(player.Name, conn, "[SERVER] Vous avez déposé l'objet: %s\n", itemName)
					LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(loc)))
					return loc
				}
			}
		}
	}
	LogAndFprintf(player.Name, conn, "ERR 404 ITEM_NOT_OWNED: Vous ne possédez pas l'objet %s.\n", itemName)
	return loc
}

func Move(conn net.Conn, player Player, loc string, message string) (newloc Location, moved bool) {
	mapsLock.Lock()
	defer mapsLock.Unlock()
	direction := strings.TrimSpace(strings.TrimPrefix(message, "MOVE"))
	locate := maps[loc]

	if locate.Exits[direction] != "" {
		current := player
		for _, p := range locate.Players {
			if p.Name == player.Name {
				current = p
				break
			}
		}
		mapsLock.Unlock()
		deletePlayerInMap(player.Name)
		mapsLock.Lock()
		direction = locate.Exits[direction]
		locate = maps[direction]
		locate.Players = append(locate.Players, current)
		var playerInLoc []net.Conn
		for _, play := range locate.Players {
			for n, conn := range ConnectedClients {
				if play.Name == n {
					playerInLoc = append(playerInLoc, conn)
				}
			}
		}
		BroadcastEVT(playerInLoc, fmt.Sprintf("[EVT] %s a rejoint l'endroit.", current.Name))
		maps[direction] = locate
		LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(locate)))
		LogAndFprintf(player.Name, conn, "%s | Vous êtes maintenant dans %s\n", conn.RemoteAddr().String(), locate.Name)
		return locate, true
	}
	LogAndFprintf(player.Name, conn, "ERR 301 NO_EXIT: Vous ne pouvez pas vous déplacer dans cette direction.\n")
	return locate, false
}

func takeItem(conn net.Conn, player Player, loc Location, message string) Location {
	mapsLock.Lock()
	defer mapsLock.Unlock()
	itemName := strings.TrimSpace(strings.TrimPrefix(message, "TAKE"))
	loc, _, ok := FindPlayerByName(player.Name, maps)
	if !ok {
		LogAndFprintf(player.Name, conn, "ERR 404 ITEM_NOT_FOUND: L'objet %s n'est pas disponible dans cet endroit.\n", itemName)
		return loc
	}
	for i, p := range loc.Players {
		if p.Name == player.Name {
			for j, item := range loc.Items {
				worldLock.Lock()
				if item == itemName {
					for _, it := range world.Items {
						if it.Name == itemName && !it.Obtainable {
							LogAndFprintf(player.Name, conn, "ERR 406 ITEM_NOT_OBTAINABLE: L'objet %s n'est pas obtenable.\n", itemName)
							worldLock.Unlock()
							return loc
						}
						if it.Name == itemName {
							loc.Players[i].Inventory = append(loc.Players[i].Inventory, it)
							loc.Items = append(loc.Items[:j], loc.Items[j+1:]...)
							maps[loc.ID] = loc
							worldLock.Unlock()
							LogAndFprintf(player.Name, conn, "[SERVER] Vous avez pris l'objet: %s\n", itemName)
							LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(loc)))
							return loc
						}
					}
				}
				worldLock.Unlock()
			}
		}
	}
	LogAndFprintf(player.Name, conn, "ERR 404 ITEM_NOT_FOUND: L'objet %s n'est pas disponible dans cet endroit.\n", itemName)
	return loc
}

func withEnemyHP(room Location) Location {
	if len(room.Enemies) == 0 {
		return room
	}
	worldLock.Lock()
	defer worldLock.Unlock()
	hp := make(map[string]int, len(room.Enemies))
	for _, id := range room.Enemies {
		if e, ok := world.Enemies[id]; ok {
			hp[id] = e.HP
		}
	}
	room.EnemyHP = hp
	return room
}

func broadcastRoomExcept(room Location, exclude string, format string, args ...interface{}) {
	ConnectedLock.Lock()
	receivers := make([]net.Conn, 0, len(room.Players))
	for _, p := range room.Players {
		if p.Name == exclude {
			continue
		}
		if c, ok := ConnectedClients[p.Name]; ok {
			receivers = append(receivers, c)
		}
	}
	ConnectedLock.Unlock()
	for _, c := range receivers {
		LogServFprintf(c, format, args...)
	}
}

func Dispatch(conn net.Conn, player Player, loc Location) {
	var previous string
	for {
		message := Recept(conn)
		LogAndPrintf(player.Name, conn, "%s", message)
		pendingInvitesLock.Lock()
		ch, hasPendingInvite := pendingInvites[player.Name]
		pendingInvitesLock.Unlock()

		if message == previous && strings.HasPrefix(message, "ATTACK") == false {
			LogAndPrintf(player.Name, conn, "%s| [WARN] %s", conn.RemoteAddr(), message)
			LogAndFprintf(player.Name, conn, "[SERVER][WARN] Attention à ne pas envoyer la même commande: %s", message)
		} else {
			LogAndPrintf(player.Name, conn, "%s| %s", conn.RemoteAddr().String(), message)
		}
		previous = message

		switch {
		case message == "":
			return

		case message == "QUIT\n":
			return

		case message == "WHO\n":
			room, _, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			ConnectedLock.Lock()
			serverCount := len(ConnectedClients)
			ConnectedLock.Unlock()
			resp, err := Who(room, serverCount)
			if err != nil {
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			LogAndFprintf(player.Name, conn, "%s\n", resp)

		case message == "INVENTORY\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, err := Inventory(room.Players[index])
			if err != nil {
				LogAndFprintf(player.Name, conn, "ERR COMMAND FAILED\n")
				continue
			} else {
				LogAndFprintf(player.Name, conn, "%s\n", resp)
				LogAndFprintf(player.Name, conn, "[SERVER] Arme équipée : %s\n", room.Players[index].Weapon.Name)
				LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(room)))
			}

		case message == "QUESTS\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			worldLock.Lock()
			for i := range room.Players[index].Quests {
				if IsQuestGloballyCompleted(room.Players[index].Quests[i].QuestID) && room.Players[index].Quests[i].Status != "completed" {
					room.Players[index].Quests[i].Status = "completed"
				}
			}
			world.Players[player.Name] = room.Players[index]
			mapsLock.Lock()
			maps[room.ID] = room
			mapsLock.Unlock()
			worldLock.Unlock()

			resp, err := Quests(room, room.Players[index], world)
			if err != nil {
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			} else {
				LogAndFprintf(player.Name, conn, "%s\n", resp)
			}

		case message == "STATUS\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				return
			}
			resp, err := Status(room, room.Players[index])
			if err != nil {
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			} else {
				LogAndFprintf(player.Name, conn, "%s\n", resp)
			}

		case message == "ACCEPT\n", message == "DECLINE\n":
			if hasPendingInvite {
				ch <- message
				LogAndFprintf(player.Name, conn, "[SERVER] Vous avez répondu à l'invitation avec: %s\n", strings.TrimSpace(message))
				continue
			}

		case message == "LOOK\n":
			room, _, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, err := Look(room)
			if err != nil {
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			LogAndFprintf(player.Name, conn, "%s\n", resp)

		case message == "CHAT\n":
		chat:
			for {
				LogAndFprintf(player.Name, conn, "[SERVER] À qui souhaitez vous envoyer un message (GLOBAL, ROOM, GROUP)?\n")
				message := Recept(conn)
				if message == "" {
					continue
				}
				LogAndPrintf(player.Name, conn, "%s| %s", conn.RemoteAddr().String(), message)
				switch message {
				case "GLOBAL\n":
					GlobalChat(conn, player.Name)
					break chat
				case "ROOM\n":
					RoomChat(conn, loc.Name, player.Name)
					break chat
				case "GROUP\n":
					GroupChat(conn, player.Name)
					break chat
				case "EXIT\n":
					break chat
				}
			}
		case strings.HasPrefix(message, "CHAT "):
			chatMessage := strings.TrimPrefix(message, "CHAT ")
			switch {
			case strings.HasPrefix(chatMessage, "GLOBAL "):
				messages := strings.TrimPrefix(chatMessage, "GLOBAL ")
				messages = fmt.Sprintf("[%s]: %s", player.Name, messages)
				GlobalChatBroadcast(conn, messages)
			case strings.HasPrefix(chatMessage, "ROOM "):
				messages := strings.TrimPrefix(chatMessage, "ROOM ")
				RoomChatBroadcast(conn, loc.Name, player.Name, messages)
			case strings.HasPrefix(chatMessage, "GROUP "):
				messages := strings.TrimPrefix(chatMessage, "GROUP ")
				GroupChatBroadcast(conn, messages, player.Name)
			default:
				LogAndFprintf(player.Name, conn, "[SERVER] ERR 400 INVALID_CHAT_COMMAND: Commande de chat invalide.\n")
			}
		case message == "GROUP CREATE\n":
			GroupCreate(conn, player.Name)

		case message == "GROUP INVITE\n":
			GroupInvite(conn, player.Name)

		case message == "GROUP JOIN\n":
			if DisplayGroup(conn) {
				groupID := Recept(conn)
				groupID = strings.TrimSpace(groupID)
				GroupJoin(conn, player.Name, groupID)
			}

		case message == "GROUP LEAVE\n":
			if GroupLeave(player.Name) {
				LogAndFprintf(player.Name, conn, "[SERVER] Vous avez quitté le groupe.\n")
			} else {
				LogAndFprintf(player.Name, conn, "[SERVER] Vous n'êtes pas dans un groupe.\n")
			}

		case strings.HasPrefix(message, "ATTACK"):
			enemyName := strings.TrimSpace(strings.TrimPrefix(message, "ATTACK"))
			mapsLock.Lock()
			var roomID string
			var playerI int
			var ok bool
			for id, loca := range maps {
				for i, p := range loca.Players {
					if p.Name == player.Name {
						roomID = id
						playerI = i
						ok = true
						break
					}
				}
				if ok {
					break
				}
			}
			if !ok {
				mapsLock.Unlock()
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			var hasNpc bool
			if len(maps[roomID].NPCs) > 0 {
				hasNpc = true
			}
			room := maps[roomID]
			mapsLock.Unlock()
			worldLock.Lock()
			enemy, enemyI, ok := FindEnemyInRoom(enemyName, room, world)
			worldLock.Unlock()
			if !ok && !hasNpc {
				LogAndFprintf(player.Name, conn, "ERR 404 NPC_NOT_FOUND\n")
				continue
			}
			if !ok && hasNpc {
				LogAndFprintf(player.Name, conn, "ERR 405 NPC_NOT_HOSTILE\n")
				continue
			}
			mapsLock.Lock()
			room = maps[roomID]
			playerI = -1
			for i, p := range room.Players {
				if p.Name == player.Name {
					playerI = i
					break
				}
			}
			if playerI == -1 {
				mapsLock.Unlock()
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, status, majplayer, majEnemy := Attack(room.Players[playerI], enemy)
			if resp == "" {
				mapsLock.Unlock()
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			switch status {
			case "defeat":
				room.Players = append(room.Players[:playerI], room.Players[playerI+1:]...)
				maps[roomID] = room
				respawnLoc := maps[majplayer.Respawn.Location]
				majplayer.HP = majplayer.Respawn.HP
				respawnLoc.Players = append(respawnLoc.Players, majplayer)
				maps[majplayer.Respawn.Location] = respawnLoc
				worldLock.Lock()
				world.Enemies[enemyName] = majEnemy
				worldLock.Unlock()
				mapsLock.Unlock()

				LogAndFprintf(player.Name, conn, "%s\n", resp)
				LogAndFprintf(player.Name, conn, "[SERVER] Vous êtes mort. Vous allez être téléporté à votre point de respawn.\n")
				LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(respawnLoc)))
				LogAndFprintf(player.Name, conn, "[SERVER] Vous êtes mort. Vous allez être téléporté à votre point de respawn.\\%s\n", resp)
				broadcastRoomExcept(room, player.Name, "%s\n", ToString(withEnemyHP(room)))
				broadcastRoomExcept(room, player.Name, "[SERVER] %s a été vaincu par %s et respawn ailleurs.\n", player.Name, enemyName)
			case "victory":
				room.Enemies = append(room.Enemies[:enemyI], room.Enemies[enemyI+1:]...)
				majplayer, questMsg := CompleteDefeatQuest(world, majplayer, enemyName)
				room.Players[playerI] = majplayer
				maps[roomID] = room
				worldLock.Lock()
				world.Enemies[enemyName] = majEnemy
				worldLock.Unlock()
				mapsLock.Unlock()

				LogAndFprintf(player.Name, conn, "%s\n", resp)
				if questMsg != "" {
					LogAndFprintf(player.Name, conn, "%s\n[SERVER] Vous avez vaincu l'ennemi %s.\n", questMsg, enemyName)
				} else {
					LogAndFprintf(player.Name, conn, "[SERVER] Vous avez vaincu l'ennemi %s.\n", enemyName)
				}
				broadcastRoomExcept(room, player.Name, "%s\n", ToString(withEnemyHP(room)))
				broadcastRoomExcept(room, player.Name, "[SERVER] %s a vaincu l'ennemi %s.\n", player.Name, enemyName)
				LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(room)))
			case "combat":
				room.Players[playerI] = majplayer
				room.Enemies[enemyI] = majEnemy.ID
				maps[roomID] = room
				worldLock.Lock()
				world.Enemies[enemyName] = majEnemy
				worldLock.Unlock()
				mapsLock.Unlock()

				LogAndFprintf(player.Name, conn, "%s\n", resp)
				LogAndFprintf(player.Name, conn, "[SERVER] Vous êtes en combat avec %s.\n", enemyName)
				LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(room)))
				LogAndFprintf(player.Name, conn, "[SERVER] Vous êtes en combat avec %s.\\%s\n", enemyName, resp)
				broadcastRoomExcept(room, player.Name, "%s\n", ToString(withEnemyHP(room)))
				broadcastRoomExcept(room, player.Name, "[SERVER] %s combat %s.\n", player.Name, enemyName)
			}

		case strings.HasPrefix(message, "TALK"):
			room, _, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp := Talk(room, world, strings.TrimSpace(strings.TrimPrefix(message, "TALK")))
			LogAndFprintf(player.Name, conn, "%s\n", resp)
			LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(room)))

		case strings.HasPrefix(message, "QUEST"):
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, updatedPlayer, completedQuestID, err := ReturnQuest(world, room, room.Players[index], strings.TrimSpace(strings.TrimPrefix(message, "QUEST")))
			if err != nil {
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			} else {
				room.Players[index] = updatedPlayer
				mapsLock.Lock()
				maps[room.ID] = room
				mapsLock.Unlock()

				if completedQuestID != "" {
					worldLock.Lock()
					world.Players[player.Name] = updatedPlayer

					questName := ""
					for _, q := range world.Quests {
						if q.ID == completedQuestID {
							questName = q.Name
							break
						}
					}

					ConnectedLock.Lock()
					for otherPlayerName, otherConn := range ConnectedClients {
						if otherPlayerName == player.Name {
							continue
						}
						for i := range world.Players[otherPlayerName].Quests {
							if world.Players[otherPlayerName].Quests[i].QuestID == completedQuestID {
								world.Players[otherPlayerName].Quests[i].Status = "completed"
								if questName != "" {
									msg := fmt.Sprintf("[EVT] %s a terminé la quête %s.\n", player.Name, questName)
									LogServFprintf(otherConn, "%s", msg)
								}
								break
							}
						}
					}
					ConnectedLock.Unlock()
					worldLock.Unlock()
				}

				LogAndFprintf(player.Name, conn, "%s\n", resp)
				LogAndFprintf(player.Name, conn, "%s\n", ToString(withEnemyHP(room)))
			}

		case strings.HasPrefix(message, "EQUIP"):
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, updatedPlayer, err := Equip(room.Players[index], strings.TrimSpace(strings.TrimPrefix(message, "EQUIP")))
			if err != nil {
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			room.Players[index] = updatedPlayer
			mapsLock.Lock()
			maps[room.ID] = room
			mapsLock.Unlock()
			LogAndFprintf(player.Name, conn, "%s\n", resp)

		case message == "UNEQUIP\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, updatedPlayer, err := Unequip(room.Players[index])
			if err != nil {
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			room.Players[index] = updatedPlayer
			mapsLock.Lock()
			maps[room.ID] = room
			mapsLock.Unlock()
			LogAndFprintf(player.Name, conn, "%s\n", resp)

		case strings.HasPrefix(message, "MOVE"):
			var tempLoc Location
			tempLoc, moved := Move(conn, player, loc.ID, message)
			if moved {
				loc = tempLoc
			}

		case strings.HasPrefix(message, "TAKE"):
			loc = takeItem(conn, player, loc, message)

		case strings.HasPrefix(message, "DROP"):
			loc = dropItem(conn, player, loc, message)

		default:
			aide := strings.Join([]string{
				"ERR 400 UNKNOWN_COMMAND: Commande inconnue.",
				"[SERVER] --- Commandes disponibles ---",
				"[SERVER] Général       : LOOK, WHO, STATUS, QUIT",
				"[SERVER] Communication : CHAT <GLOBAL|ROOM|GROUP> <message>",
				"[SERVER] Groupe        : GROUP CREATE, GROUP INVITE, GROUP JOIN, GROUP LEAVE",
				"[SERVER] Objets        : INVENTORY, TAKE <item>, DROP <item>, EQUIP <item>, UNEQUIP",
				"[SERVER] Interaction   : TALK <npc>, ATTACK <npc>, QUEST <npc>, QUESTS",
				"[SERVER] Déplacement   : MOVE <direction>",
			}, "\\")
			LogAndFprintf(player.Name, conn, "%s\n", aide)
		}
	}
}
