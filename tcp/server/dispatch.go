package main

import (
	. "the_answer_protocol/tcp/commands"
	. "the_answer_protocol/tcp/common"
	. "the_answer_protocol/tcp"
	. "the_answer_protocol/parsing"
	"net"
	"strings"
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
					// Ajoute l'objet à la liste des objets de la location
					loc.Items = append(loc.Items, itemName)
					maps[loc.ID] = loc
					LogAndFprintf(conn, "[SERVER] Vous avez déposé l'objet: %s\n", itemName)
					LogAndFprintf(conn, "%s\n", ToString(loc))
					return loc
				}
			}
		}
	}
	LogAndFprintf(conn, "[SERVER] ERRNO 404: Vous ne possédez pas l'objet %s.\n", itemName)
	return loc
}

func Move(conn net.Conn, player Player, loc string, message string) (newloc Location, moved bool) {
	// Récupère la direction du mouvement à partir du message
	mapsLock.Lock()
	defer mapsLock.Unlock()
	direction := strings.TrimSpace(strings.TrimPrefix(message, "MOVE"))
	locate := maps[loc]

	if locate.Exits[direction] != ""{  // ajout
		current := player  // ajout
		for _, p := range locate.Players {  // ajout
			if p.Name == player.Name {  // ajout
				current = p  // ajout
				break  // ajout
			}  // ajout
		}
		mapsLock.Unlock()
		deletePlayerInMap(player.Name)
		mapsLock.Lock()
		direction = locate.Exits[direction]
		locate = maps[direction]
		locate.Players = append(locate.Players, current) // ajout
		maps[direction] = locate
		LogAndFprintf(conn, "%s\n", ToString(locate))
		LogAndFprintf(conn, "%s | Vous êtes maintenant dans %s\n", conn.RemoteAddr().String(), locate.Name)
		return locate, true
	}
	LogAndFprintf(conn, "[SERVER] ERRNO 301 NO_EXIT: Vous ne pouvez pas vous déplacer dans cette direction.\n")
	return locate, false
}

func takeItem(conn net.Conn, player Player, loc Location, message string) Location {
	// Récupère le nom de l'objet à prendre à partir du message
	itemName := strings.TrimSpace(strings.TrimPrefix(message, "TAKE"))
	mapsLock.Lock()
	defer mapsLock.Unlock()
	for i, p := range loc.Players {
		if p.Name == player.Name {
			for j, item := range loc.Items {
				worldLock.Lock()
				if item == itemName {
					for _, it := range world.Items {
						if it.Name == itemName && !it.Obtainable {
							LogAndFprintf(conn, "[SERVER] L'objet %s n'est pas obtenable.\n", itemName)
							worldLock.Unlock()
							return loc
						}
						if it.Name == itemName{
						// Ajoute l'objet à l'inventaire du joueur. La quête de type
						// "fetch" associee (s'il y en a une) ne se termine pas ici :
						// le joueur doit retourner voir le donneur de quete pour
						// recevoir sa recompense (cf. ReturnQuest / commande QUEST).
						loc.Players[i].Inventory = append(loc.Players[i].Inventory, it)
						// Supprime l'objet de la liste des objets de la location
						loc.Items = append(loc.Items[:j], loc.Items[j+1:]...)
						maps[loc.ID] = loc
						worldLock.Unlock()
						LogAndFprintf(conn, "[SERVER] Vous avez pris l'objet: %s\n", itemName)
						LogAndFprintf(conn, "%s\n", ToString(loc))
						return loc
						}
					}
				}
				worldLock.Unlock()
			}
		}
	}
	LogAndFprintf(conn, "[SERVER] ERRNO 404: L'objet %s n'est pas disponible dans cet endroit.\n", itemName)
	return loc
}

// Dispatch des commandes du client après la connexion,
// en fonction de la commande reçue.
func Dispatch(conn net.Conn, player Player, loc Location) {
	var previous string
	for {
		message := Recept(conn)
		LogAndPrintf(conn, "%s", message)
		pendingInvitesLock.Lock()
		ch, hasPendingInvite := pendingInvites[player.Name]
		pendingInvitesLock.Unlock()

		if message == previous {
			LogAndPrintf(conn, "%s| [WARN] %s", conn.RemoteAddr(), message)
			LogAndFprintf(conn, "[SERVER][WARN] Attention à ne pas envoyer la même commande: %s\n", message)
		} else {
			LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
		}
		previous = message

		switch {
		case message == "":
			return

		case message == "QUIT\n":
			return

		case message == "WHO\n":
			nb_player := 5 - len(PersoChoice)
			LogAndFprintf(conn, "[SERVER] Il y a actuellement %d membre.s d'expédition\n", nb_player)

		case message == "INVENTORY\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, err := Inventory(room.Players[index])
			if err != nil {
				LogAndFprintf(conn, "ERR COMMAND FAILED\n")
				continue
			} else {
				// resp ("OK [...]") doit arriver avant l'etat de la salle : le
				// GUI ne recupere que la toute premiere ligne OK/ERR/{ d'une
				// reponse (cf. Session.Deliver), sinon il recupere l'etat de
				// la salle a la place de l'inventaire.
				LogAndFprintf(conn, "%s\n", resp)
				LogAndFprintf(conn, "%s\n", ToString(room))
			}

		case message == "QUESTS\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, err := Quests(room, room.Players[index], world)
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			} else {
				LogAndFprintf(conn, "%s\n", resp)
			}

		case message == "STATUS\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				return
			}
			resp, err := Status(room, room.Players[index])
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			} else {
				LogAndFprintf(conn, "%s\n", resp)
			}

		case message == "ACCEPT\n", message == "DECLINE\n":
			if hasPendingInvite {
				ch <- message
				LogAndFprintf(conn, "[SERVER] Vous avez répondu à l'invitation avec: %s\n", strings.TrimSpace(message))
				continue
			}

		case message == "LOOK\n":
			room, _, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, err := Look(room)
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			LogAndFprintf(conn, "%s\n", resp)

		case message == "CHAT\n":
		chat:
			for {
				LogAndFprintf(conn, "[SERVER] À qui souhaitez vous envoyer un message (GLOBAL, ROOM, GROUP)?\n")
				message := Recept(conn)
				if message == "" {
					continue
				}
				LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
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
				}
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
				LogAndFprintf(conn, "[SERVER] Vous avez quitté le groupe.\n")
			} else {
				LogAndFprintf(conn, "[SERVER] ERR 401 NOT_IN_GROUP: Vous n'êtes pas dans un groupe.\n")
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
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			room := maps[roomID]
			worldLock.Lock()
			enemy, enemyI, ok := FindEnemyInRoom(enemyName, room, world)
			worldLock.Unlock()
			if !ok {
				mapsLock.Unlock()
				LogAndFprintf(conn, "ERR 404 NPC_NOT_FOUND\n")
				continue
			}
			resp, status, majplayer, majEnemy := Attack(room.Players[playerI], enemy)
			if resp == "" {
				mapsLock.Unlock()
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				continue
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
				LogAndFprintf(conn, "%s\n", resp)
				LogAndFprintf(conn, "[SERVER] Vous êtes mort. Vous allez être téléporté à votre point de respawn.\n")
				LogAndFprintf(conn, "%s\n", ToString(respawnLoc))
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
				LogAndFprintf(conn, "%s\n", resp)
				if questMsg != "" { // ajout
					LogAndFprintf(conn, "%s\n[SERVER] Vous avez vaincu l'ennemi %s.\n", questMsg, enemyName) // ajout
				} else { // ajout
					LogAndFprintf(conn, "[SERVER] Vous avez vaincu l'ennemi %s.\n", enemyName) // ajout
				} // ajout
				LogAndFprintf(conn, "%s\n", ToString(room))
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
				LogAndFprintf(conn, "%s\n", resp)
				LogAndFprintf(conn, "[SERVER] Vous êtes en combat avec %s.\n", enemyName)
				LogAndFprintf(conn, "%s\n", ToString(room))
			}

		case strings.HasPrefix(message, "TALK"):
			room, _, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp := Talk(room, world, strings.TrimSpace(strings.TrimPrefix(message, "TALK")))
			// resp ("OK [npc]: ...") doit arriver avant l'etat de la salle,
			// meme raison que pour INVENTORY/QUEST : sinon le GUI recupere
			// l'etat de la salle au lieu du dialogue.
			LogAndFprintf(conn, "%s\n", resp)
			LogAndFprintf(conn, "%s\n", ToString(room))

		case strings.HasPrefix(message, "QUEST"):
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, player, err := ReturnQuest(world, room, room.Players[index], strings.TrimSpace(strings.TrimPrefix(message, "QUEST")))
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			} else {
				room.Players[index] = player
				mapsLock.Lock()
				maps[room.ID] = room
				mapsLock.Unlock()
				// resp est envoye avant l'etat de la salle : c'est la seule
				// ligne "OK .../ERR ..." de cette reponse, elle doit donc
				// arriver en premier pour que le GUI (qui ne recupere que la
				// toute premiere ligne OK/ERR/{ d'une reponse, cf. Session.Deliver)
				// recoive bien le resultat de la commande QUEST et pas l'etat de
				// la salle.
				LogAndFprintf(conn, "%s\n", resp)
				LogAndFprintf(conn, "%s\n", ToString(room))
			}

		case strings.HasPrefix(message, "EQUIP"):
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, updatedPlayer, err := Equip(room.Players[index], strings.TrimSpace(strings.TrimPrefix(message, "EQUIP")))
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			room.Players[index] = updatedPlayer
			mapsLock.Lock()
			maps[room.ID] = room
			mapsLock.Unlock()
			LogAndFprintf(conn, "%s\n", resp)

		case message == "UNEQUIP\n":
			room, index, ok := FindPlayerByName(player.Name, maps)
			if !ok {
				LogAndFprintf(conn, "ERR 404 PLAYER_NOT_FOUND\n")
				continue
			}
			resp, updatedPlayer, err := Unequip(room.Players[index])
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				continue
			}
			room.Players[index] = updatedPlayer
			mapsLock.Lock()
			maps[room.ID] = room
			mapsLock.Unlock()
			LogAndFprintf(conn, "%s\n", resp)

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
		}
	}
}
