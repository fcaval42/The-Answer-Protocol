package main

import (
	"the_answer_protocol/tcp/commands"
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
					maps[loc.Name] = loc
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

func Move(conn net.Conn, player Player, loc string, message string) Location {
	// Récupère la direction du mouvement à partir du message
	direction := strings.TrimSpace(strings.TrimPrefix(message, "MOVE"))
	mapsLock.Lock()
	locate := maps[loc]
	mapsLock.Unlock()

	if locate.Exits[direction] == ""{
		LogAndFprintf(conn, "[SERVER] ERRNO 301: Vous ne pouvez pas vous déplacer dans cette direction.\n")
		return locate
	}
	deletePlayerInMap(player.Name)
	direction = locate.Exits[direction]
	mapsLock.Lock()
	locate = maps[direction]
	player.Location = locate
	locate.Players = append(locate.Players, player)
	maps[direction] = locate
	mapsLock.Unlock()
	LogAndFprintf(conn, "%s\n", ToString(locate))
	LogAndFprintf(conn, "Vous êtes maintenant dans %s\n", locate.Name)
	return locate
}

func takeItem(conn net.Conn, player Player, loc Location, message string) Location {
	// Récupère le nom de l'objet à prendre à partir du message
	itemName := strings.TrimSpace(strings.TrimPrefix(message, "TAKE"))
	mapsLock.Lock()
	defer mapsLock.Unlock()
	for i, p := range loc.Players {
		if p.Name == player.Name {
			for j, item := range loc.Items {
				if item == itemName {
					worldLock.Lock()
					for _, it := range world.Items {
						if it.Name == itemName && !it.Obtainable {
							LogAndFprintf(conn, "[SERVER] L'objet %s n'est pas obtenable.\n", itemName)
							worldLock.Unlock()
							return loc
						}
						if it.Name == itemName{
						// Ajoute l'objet à l'inventaire du joueur
						loc.Players[i].Inventory = append(loc.Players[i].Inventory, it)
						// Supprime l'objet de la liste des objets de la location
						loc.Items = append(loc.Items[:j], loc.Items[j+1:]...)
						maps[loc.Name] = loc
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
	for {
		message := Recept(conn)
		pendingInvitesLock.Lock()
		ch, hasPendingInvite := pendingInvites[player.Name]
		pendingInvitesLock.Unlock()
		switch message{
		case "":
			return
		case "WHO\n":
			nb_player := 5 - len(PersoChoice)
			LogAndFprintf(conn, "[SERVER] Il y a actuellement %d membre.s d'expédition\n", nb_player)
		case "LOOK\n":
			worldLock.Lock()
			resp, err := commands.Look(world, player.Name)
			worldLock.Unlock()
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
			} else {
				LogAndFprintf(conn, "%s\n", resp)
			}
		case "INVENTORY\n":
			worldLock.Lock()
			resp, err := commands.Inventory(world, player.Name)
			worldLock.Unlock()
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
			} else {
				LogAndFprintf(conn, "%s\n", resp)
			}
		case "QUESTS\n":
			worldLock.Lock()
			resp, err := commands.Quests(world, player.Name)
			worldLock.Unlock()
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
			} else {
				LogAndFprintf(conn, "%s\n", resp)
			}
		case "STATUS\n":
			worldLock.Lock()
			resp, err := commands.Status(world, player.Name)
			worldLock.Unlock()
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
			} else {
				LogAndFprintf(conn, "%s\n", resp)
			}
		case "ACCEPT\n", "DECLINE\n":
			if hasPendingInvite{
				ch <- message
				LogAndFprintf(conn, "[SERVER] Vous avez répondu à l'invitation avec: %s\n", strings.TrimSpace(message))
				continue
			}
		case "DEBUG LOC\n":
			mapsLock.Lock()
			mainloop:
			for _, l := range maps {
				for _, p := range l.Players {
					if p.Name == player.Name {
						LogAndFprintf(conn, "%s\n", ToString(l))
						break mainloop
					}
				}
			}
			mapsLock.Unlock()
		case "CHAT\n":
			chat:
			for {
				LogAndFprintf(conn, "[SERVER] À qui souhaitez vous envoyer un message (GLOBAL, ROOM, GROUP)?\n")
				message := Recept(conn)
				if message == "" {
					return
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
		case "GROUP CREATE\n":
			GroupCreate(conn, player.Name)
		case "GROUP INVITE\n":
			GroupInvite(conn, player.Name)
		case "GROUP JOIN\n":
			if DisplayGroup(conn){
				groupID := Recept(conn)
				groupID = strings.TrimSpace(groupID)
				GroupJoin(conn, player.Name, groupID)
			}
		case "GROUP LEAVE\n":
			if GroupLeave(player.Name) {
				LogAndFprintf(conn, "[SERVER] Vous avez quitté le groupe.\n")
			} else {
				LogAndFprintf(conn, "[SERVER] Vous n'êtes pas dans un groupe.\n")
			}
		case "QUIT\n":
			return
		default:
			if npcName, ok := strings.CutPrefix(message, "QUEST "); ok {
				npcName = strings.TrimSuffix(npcName, "\n")
				worldLock.Lock()
				resp, err := commands.Quest(world, player.Name, npcName)
				worldLock.Unlock()
				if err != nil {
					LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				} else {
					LogAndFprintf(conn, "%s\n", resp)
				}
			} else if npcName, ok := strings.CutPrefix(message, "TALK "); ok {
				npcName = strings.TrimSuffix(npcName, "\n")
				worldLock.Lock()
				resp, err := commands.Talk(world, player.Name, npcName)
				worldLock.Unlock()
				if err != nil {
					LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				} else {
					LogAndFprintf(conn, "%s\n", resp)
				}
			} else if npcName, ok := strings.CutPrefix(message, "ATTACK "); ok {
				npcName = strings.TrimSuffix(npcName, "\n")
				worldLock.Lock()
				resp, err := commands.Attack(world, player.Name, npcName)
				worldLock.Unlock()
				if err != nil {
					LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
				} else {
					LogAndFprintf(conn, "%s\n", resp)
				}
			}
		}
		if strings.HasPrefix(message, "MOVE"){
			loc = Move(conn, player, loc.Name, message)
		}
		if strings.HasPrefix(message, "TAKE"){
			loc = takeItem(conn, player, loc, message)
		}
		if strings.HasPrefix(message, "DROP"){
			loc = dropItem(conn, player, loc, message)
		}
	}
	}
