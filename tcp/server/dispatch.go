package main

import (
	"encoding/json"
	"the_answer_protocol/tcp/commands"
	. "the_answer_protocol/tcp/common"
	. "the_answer_protocol/tcp"
	. "the_answer_protocol/parsing"
	"net"
	"strings"
)

// dropItem traite DROP <item> : itemName peut etre l'ID ou le nom affiche
// de l'item (recherche dans l'inventaire du joueur), noms multi-mots geres
// tels quels puisque itemName est tout le reste de la ligne, non redecoupe.
func dropItem(conn net.Conn, player Player, loc Location, message string) Location {
	// Récupère le nom de l'objet à déposer à partir du message
	itemName := strings.TrimSpace(strings.TrimPrefix(message, "DROP"))
	mapsLock.Lock()
	defer mapsLock.Unlock()
	worldLock.Lock()
	defer worldLock.Unlock()

	for j, it := range player.Inventory {
		if it.ID != itemName && it.Name != itemName {
			continue
		}
		// Retire l'objet de l'inventaire "officiel" (world.Players, relu par
		// INVENTORY/STATUS/...), pas seulement de la copie locale a loc.
		player.Inventory = append(player.Inventory[:j], player.Inventory[j+1:]...)
		world.Players[player.Name] = player
		for i, p := range loc.Players {
			if p.Name == player.Name {
				loc.Players[i] = player
				break
			}
		}
		// loc.Items contient des ID, jamais le texte tape par le joueur.
		loc.Items = append(loc.Items, it.ID)
		// maps est indexee par ID, pas par nom affiche (cf. Move) : ecrire
		// sous loc.Name créait une entree fantome au lieu de mettre a jour
		// la vraie salle, ce qui dupliquait/perdait les objets.
		maps[loc.ID] = loc

		// Une seule ligne de reponse, comme toutes les autres commandes (le
		// texte d'info "Vous avez depose..." qui precedait en plus faisait de
		// DROP une reponse a deux lignes - Deliver ne peut pas savoir de
		// facon fiable laquelle des deux est "la" reponse selon leur ordre
		// d'arrivee, ce qui causait des timeouts intermittents cote GUI).
		LogAndFprintf(conn, "%s\n", ToString(loc))
		return loc
	}
	// "ERR ..." comme toutes les autres reponses (cf. le commentaire dans
	// Move ci-dessous) : sans ca, Send() cote GUI attend le timeout de 5s.
	LogAndFprintf(conn, "ERR 404 ITEM_NOT_OWNED\n")
	return loc
}

func Move(conn net.Conn, player Player, loc string, message string) Location {
	// Récupère la direction du mouvement à partir du message
	direction := strings.TrimSpace(strings.TrimPrefix(message, "MOVE"))
	mapsLock.Lock()
	locate := maps[loc]
	mapsLock.Unlock()

	if locate.Exits[direction] == ""{
		// Doit commencer par "ERR" comme toutes les autres commandes : le GUI
		// (gui/session.go, Deliver) ne reconnait une reponse que si elle
		// commence par "OK ", "ERR" ou "{". Avec l'ancien texte libre
		// "[SERVER] ERRNO...", Send() ne recevait jamais cette reponse et
		// attendait bêtement le timeout de 5s avant d'afficher une erreur.
		LogAndFprintf(conn, "ERR 301 EXIT_NOT_FOUND\n")
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

	// Sans ça, world.Players (source consultee par Attack/Talk/Quest/Status/
	// Inventory pour savoir dans quelle salle se trouve le joueur) reste
	// bloque sur la salle de depart : maps est mis a jour ci-dessus mais
	// c'est une structure distincte, non relue par tcp/commands.
	worldLock.Lock()
	world.Players[player.Name] = player
	worldLock.Unlock()

	// Une seule ligne de reponse, comme toutes les autres commandes (cf. le
	// commentaire equivalent dans dropItem : deux lignes rendait la
	// correlation cote GUI dependante du timing d'arrivee, timeouts
	// intermittents a la cle).
	LogAndFprintf(conn, "%s\n", ToString(locate))
	return locate
}

// takeItem traite TAKE <item> : itemName peut etre l'ID ou le nom affiche
// de l'item (recherche dans loc.Items), noms multi-mots geres tels quels
// puisque itemName est tout le reste de la ligne, non redecoupe.
func takeItem(conn net.Conn, player Player, loc Location, message string) Location {
	// Récupère le nom de l'objet à prendre à partir du message
	itemName := strings.TrimSpace(strings.TrimPrefix(message, "TAKE"))
	mapsLock.Lock()
	defer mapsLock.Unlock()
	worldLock.Lock()
	defer worldLock.Unlock()

	for j, itemID := range loc.Items {
		it, exists := world.Items[itemID]
		if !exists || (it.ID != itemName && it.Name != itemName) {
			continue
		}
		if !it.Obtainable {
			LogAndFprintf(conn, "ERR 409 ITEM_NOT_OBTAINABLE\n")
			return loc
		}
		// Retire l'objet de la salle par son ID (celui reellement stocke
		// dans loc.Items), pas par le texte tape par le joueur.
		loc.Items = append(loc.Items[:j], loc.Items[j+1:]...)
		// maps est indexee par ID, pas par nom affiche (cf. Move) : ecrire
		// sous loc.Name créait une entree fantome au lieu de mettre a jour
		// la vraie salle, ce qui permettait a l'objet d'etre repris par
		// plusieurs joueurs (duplication).
		maps[loc.ID] = loc

		// Ajoute l'objet a l'inventaire "officiel" du joueur (world.Players,
		// relu par INVENTORY/STATUS/...), pas seulement a la copie locale
		// dans loc.Players.
		player.Inventory = append(player.Inventory, it)
		world.Players[player.Name] = player
		for i, p := range loc.Players {
			if p.Name == player.Name {
				loc.Players[i] = player
				break
			}
		}

		// Une seule ligne de reponse, comme toutes les autres commandes (cf.
		// le commentaire equivalent dans dropItem).
		LogAndFprintf(conn, "%s\n", ToString(loc))
		return loc
	}
	// "ERR ..." comme toutes les autres reponses (cf. le commentaire dans
	// Move ci-dessous) : sans ca, Send() cote GUI attend le timeout de 5s.
	LogAndFprintf(conn, "ERR 404 ITEM_NOT_FOUND\n")
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
		LogAndPrintf(conn, "%s| %s", conn.RemoteAddr().String(), message)
		switch message{
		case "":
			return
		case "WHO\n":
			worldLock.Lock()
			room := make([]string, 0)
			serverCount := 0
			for _, p := range world.Players {
				if !p.Connected {
					continue
				}
				serverCount++
				if p.Location.ID == loc.ID {
					room = append(room, p.Name)
				}
			}
			worldLock.Unlock()
			data, err := json.Marshal(struct {
				Room   []string `json:"room"`
				Server int      `json:"server"`
			}{Room: room, Server: serverCount})
			if err != nil {
				LogAndFprintf(conn, "ERR 900 CONNECTION_FAILED\n")
			} else {
				LogAndFprintf(conn, "OK %s\n", data)
			}
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
			// maps est indexee par ID, pas par nom affiche (ex: nom "Vallées
			// fleuries" mais ID "Vallées_fleuries") : passer loc.Name faisait
			// echouer la recherche pour toute salle dont le nom differe de
			// son ID, et MOVE renvoyait alors une erreur silencieuse.
			loc = Move(conn, player, loc.ID, message)
		}
		if strings.HasPrefix(message, "TAKE"){
			loc = takeItem(conn, player, loc, message)
		}
		if strings.HasPrefix(message, "DROP"){
			loc = dropItem(conn, player, loc, message)
		}
		if strings.HasPrefix(message, "MOVE") || strings.HasPrefix(message, "TAKE") || strings.HasPrefix(message, "DROP") {
			// Move/takeItem/dropItem ecrivent la version a jour dans
			// world.Players, mais chacune ne mute que sa PROPRE copie locale
			// de player (parametre passe par valeur) : sans cette relecture,
			// la variable player de Dispatch restait figee sur l'etat du
			// debut de connexion pour toute la suite de la boucle - un DROP
			// juste apres un TAKE ne voyait donc jamais l'objet pris, et une
			// deuxieme prise pouvait ecraser world.Players en perdant la
			// premiere (elle repartait de cette meme copie obsolete).
			worldLock.Lock()
			if p, ok := world.Players[player.Name]; ok {
				player = p
			}
			worldLock.Unlock()
		}
	}
	}
