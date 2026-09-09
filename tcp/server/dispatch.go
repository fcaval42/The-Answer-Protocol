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
	// Récupère le nom de l'objet à déposer à partir du message
	itemName := strings.TrimSpace(strings.TrimPrefix(message, "DROP"))
	mapsLock.Lock()
	defer mapsLock.Unlock()
	for i, p := range loc.Players {
		if p.Name == player.Name {
			for j, item := range p.Inventory {
				if item.Name == itemName {
					// Supprime l'objet de l'inventaire du joueur
					loc.Players[i].Inventory = append(loc.Players[i].Inventory[:j], loc.Players[i].Inventory[j+1:]...)
					// Si l'objet depose est celui actuellement equipe, on le desequipe
					if loc.Players[i].EquippedItem == item.ID {
						loc.Players[i].Weapon = loc.Players[i].BaseWeapon
						loc.Players[i].EquippedItem = ""
					}
					// Ajoute l'objet à la liste des objets de la location
					loc.Items = append(loc.Items, itemName)
					maps[loc.ID] = loc
					LogAndFprintf(player.Name, conn, "[SERVER] Vous avez déposé l'objet: %s\n", itemName)
					LogAndFprintf(player.Name, conn, "%s\n", ToString(loc))
					return loc
				}
			}
		}
	}
	LogAndFprintf(player.Name, conn, "ERR 404 ITEM_NOT_OWNED: Vous ne possédez pas l'objet %s.\n", itemName)
	return loc
}

func Move(conn net.Conn, player Player, loc string, message string) (newloc Location, moved bool) {
	// Récupère la direction du mouvement à partir du message
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
		LogAndFprintf(player.Name, conn, "%s\n", ToString(locate))
		LogAndFprintf(player.Name, conn, "%s | Vous êtes maintenant dans %s\n", conn.RemoteAddr().String(), locate.Name)
		return locate, true
	}
	LogAndFprintf(player.Name, conn, "ERR 301 NO_EXIT: Vous ne pouvez pas vous déplacer dans cette direction.\n")
	return locate, false
}

func takeItem(conn net.Conn, player Player, loc Location, message string) Location {
	// Récupère le nom de l'objet à prendre à partir du message
	mapsLock.Lock()
	defer mapsLock.Unlock()
	itemName := strings.TrimSpace(strings.TrimPrefix(message, "TAKE"))
	loc, _, ok := FindPlayerByName(player.Name, maps)
	if !ok{
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
							// Ajoute l'objet à l'inventaire du joueur
							loc.Players[i].Inventory = append(loc.Players[i].Inventory, it)
							// Supprime l'objet de la liste des objets de la location
							loc.Items = append(loc.Items[:j], loc.Items[j+1:]...)
							maps[loc.ID] = loc
							worldLock.Unlock()
							LogAndFprintf(player.Name, conn, "[SERVER] Vous avez pris l'objet: %s\n", itemName)
							LogAndFprintf(player.Name, conn, "%s\n", ToString(loc))
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

// broadcastRoomExcept envoie un message a tous les joueurs de la salle demander dans le sujet. // modif
func broadcastRoomExcept(room Location, exclude string, format string, args ...interface{}) { // modif
	ConnectedLock.Lock()                                // modif
	receivers := make([]net.Conn, 0, len(room.Players)) // modif
	for _, p := range room.Players {                    // modif
		if p.Name == exclude {                          // modif
			continue                                    // modif
		}                                                // modif
		if c, ok := ConnectedClients[p.Name]; ok {       // modif
			receivers = append(receivers, c)             // modif
		}                                                // modif
	}                                                    // modif
	ConnectedLock.Unlock()                               // modif
	for _, c := range receivers {                        // modif
		LogServFprintf(c, format, args...)                // modif
	}                                                    // modif
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
				// resp ("OK [...]") doit arriver avant l'etat de la salle : le
				// GUI ne recupere que la toute premiere ligne OK/ERR/{ d'une
				// reponse (cf. Session.Deliver), sinon il recupere l'etat de
				// la salle a la place de l'inventaire.
				LogAndFprintf(player.Name, conn, "%s\n", resp)
				LogAndFprintf(player.Name, conn, "%s\n", ToString(room))
			}

		case message == "QUESTS\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			// Vérifier si des quêtes ont été terminées globalement et marquer celles du joueur comme completed
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
			resp, critMsg, status, majplayer, majEnemy := Attack(room.Players[playerI], enemy)
			if resp == "" {
				mapsLock.Unlock()
				LogAndFprintf(player.Name, conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			if critMsg != "" {
				resp = critMsg + "\\" + resp
			}
			switch status {
			case "defeat":
				// Supprimer le joueur de la room actuelle
				room.Players = append(room.Players[:playerI], room.Players[playerI+1:]...)
				maps[roomID] = room
				// Ajouter le joueur au point de respawn
				respawnLoc := maps[majplayer.Respawn.Location]
				majplayer.HP = majplayer.Respawn.HP
				respawnLoc.Players = append(respawnLoc.Players, majplayer)
				maps[majplayer.Respawn.Location] = respawnLoc
				worldLock.Lock()
				world.Enemies[enemyName] = majEnemy
				worldLock.Unlock()
				mapsLock.Unlock()

				// resp doit etre envoye comme sa propre ligne commencant par
				// "OK " : le GUI ne recupere que la toute premiere ligne
				// OK/ERR/{ d'une reponse (cf. Session.Deliver). Avant, resp
				// etait fusionne dans le message [SERVER] via un "\" litteral,
				// donc la ligne entiere commencait par "[SERVER]" et n'etait
				// jamais reconnue : c'est l'etat de la salle (qui commence par
				// "{") qui etait capte a la place du resultat du combat.
				LogAndFprintf(player.Name, conn, "%s\n", resp)
				LogAndFprintf(player.Name, conn, "[SERVER] Vous êtes mort. Vous allez être téléporté à votre point de respawn.\n")
				LogAndFprintf(player.Name, conn, "%s\n", ToString(respawnLoc))
				LogAndFprintf(player.Name, conn, "[SERVER] Vous êtes mort. Vous allez être téléporté à votre point de respawn.\\%s\n", resp)
				broadcastRoomExcept(room, player.Name, "%s\n", ToString(room))                                                          // modif
				broadcastRoomExcept(room, player.Name, "[SERVER] %s a été vaincu par %s et respawn ailleurs.\n", player.Name, enemyName) // modif
			case "victory":
				room.Enemies = append(room.Enemies[:enemyI], room.Enemies[enemyI+1:]...)
				majplayer, questMsg := CompleteDefeatQuest(world, majplayer, enemyName) // ajout
				room.Players[playerI] = majplayer
				maps[roomID] = room
				worldLock.Lock()
				world.Enemies[enemyName] = majEnemy
				worldLock.Unlock()
				mapsLock.Unlock()

				// resp doit etre sa propre ligne "OK ...", meme raison que
				// pour le cas "defeat" ci-dessus.
				LogAndFprintf(player.Name, conn, "%s\n", resp)
				if questMsg != "" { // ajout
					LogAndFprintf(player.Name, conn, "%s\n[SERVER] Vous avez vaincu l'ennemi %s.\n", questMsg, enemyName) // ajout
				} else { // ajout
					LogAndFprintf(player.Name, conn, "[SERVER] Vous avez vaincu l'ennemi %s.\n", enemyName) // ajout
				} // ajout
				broadcastRoomExcept(room, player.Name, "%s\n", ToString(room))                                        // modif
				broadcastRoomExcept(room, player.Name, "[SERVER] %s a vaincu l'ennemi %s.\n", player.Name, enemyName) // modif
				LogAndFprintf(player.Name, conn, "%s\n", ToString(room))
			case "combat":
				room.Players[playerI] = majplayer
				room.Enemies[enemyI] = majEnemy.ID
				maps[roomID] = room
				worldLock.Lock()
				world.Enemies[enemyName] = majEnemy
				worldLock.Unlock()
				mapsLock.Unlock()

				// resp doit etre sa propre ligne "OK ...", meme raison que
				// pour le cas "defeat" ci-dessus.
				LogAndFprintf(player.Name, conn, "%s\n", resp)
				LogAndFprintf(player.Name, conn, "[SERVER] Vous êtes en combat avec %s.\n", enemyName)
				LogAndFprintf(player.Name, conn, "%s\n", ToString(room))
				LogAndFprintf(player.Name, conn, "[SERVER] Vous êtes en combat avec %s.\\%s\n", enemyName, resp)
				broadcastRoomExcept(room, player.Name, "%s\n", ToString(room))                             // modif
				broadcastRoomExcept(room, player.Name, "[SERVER] %s combat %s.\n", player.Name, enemyName) // modif
			}

		case strings.HasPrefix(message, "TALK"):
			room, _, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(player.Name, conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp := Talk(room, world, strings.TrimSpace(strings.TrimPrefix(message, "TALK")))
			// resp ("OK [npc]: ...") doit arriver avant l'etat de la salle,
			// meme raison que pour INVENTORY/QUEST : sinon le GUI recupere
			// l'etat de la salle au lieu du dialogue.
			LogAndFprintf(player.Name, conn, "%s\n", resp)
			LogAndFprintf(player.Name, conn, "%s\n", ToString(room))

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

				// Si une quête a été terminée, marquer la quête comme terminée pour tous les joueurs qui l'ont en cours
				if completedQuestID != "" {
					worldLock.Lock()
					world.Players[player.Name] = updatedPlayer

					// Trouver le nom de la quête
					questName := ""
					for _, q := range world.Quests {
						if q.ID == completedQuestID {
							questName = q.Name
							break
						}
					}

					// Notifier tous les joueurs connectés
					ConnectedLock.Lock()
					for otherPlayerName, otherConn := range ConnectedClients {
						if otherPlayerName == player.Name {
							continue
						}
						// Marquer la quête comme terminée pour ce joueur si il l'a en cours
						for i := range world.Players[otherPlayerName].Quests {
							if world.Players[otherPlayerName].Quests[i].QuestID == completedQuestID {
								world.Players[otherPlayerName].Quests[i].Status = "completed"
								// Envoyer une notification
								if questName != "" {
									msg := fmt.Sprintf("[EVT] %s a terminé la quête %s.\n", player.Name, questName)
									LogServFprintf(otherConn, "%s" ,msg)
								}
								break
							}
						}
					}
					ConnectedLock.Unlock()
					worldLock.Unlock()
				}

				// resp est envoye avant l'etat de la salle : c'est la seule
				// ligne "OK .../ERR ..." de cette reponse, elle doit donc
				// arriver en premier pour que le GUI (qui ne recupere que la
				// toute premiere ligne OK/ERR/{ d'une reponse, cf. Session.Deliver)
				// recoive bien le resultat de la commande QUEST et pas l'etat de
				// la salle.
				LogAndFprintf(player.Name, conn, "%s\n", resp)
				LogAndFprintf(player.Name, conn, "%s\n", ToString(room))
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
