package main

import(
	"net"
	"bufio"
	"fmt"
	. "the_answer_protocol/tcp/common"
	"log"
	"strings"
	"strconv"
	"slices"
)


// Selection du personnage joueur
func selectPlayer(conn net.Conn) string {
	// Lock du mutex pour le rendre indisponible pour les autres
	PersoLock.Lock()
	// Le mutex sera unlock à la fin de la fonction selectPlayer
	defer PersoLock.Unlock()
	// Initialisation d'un buffer en écoute sur le client
	reader := bufio.NewReader(conn)

	if len(PersoChoice) == 0{
		LogAndFprintf(conn, "%s| [ERROR] Errno 900: Nombre de joueurs limite atteint (CONNECTION_FAILED)\n", conn.RemoteAddr())
		return "900"
	}

	// Permet l'affichage formaté "1. nom, 2. nom, etc..."
	charactere := make([]string, len(PersoChoice))
	for i, perso := range PersoChoice {
		charactere[i] = fmt.Sprintf("%d. %s", i+1, perso)
	}
	msg := fmt.Sprintf("[INFO] Choisissez votre joueur parmi en entrant le numéro correspondant: %s\n", strings.Join(charactere, ", "))
	LogAndFprintf(conn, msg)

	// Boucle infinie tant que le joueur n'a pas effectué un bon choix
	for {
		// Recois le choix du joueur par le client
		player, err := reader.ReadString('\n')
		LogAndPrintf(conn, conn.RemoteAddr().String()+"| " + player)
		if player != "\n" {
			if err != nil {
				log.Println(conn.RemoteAddr().String() + "| [ERROR] " + err.Error())
				return ""
			}

			// Verifie qu'il s'agit bien de nombres
			if IsDigit(strings.Trim(player, "\n")) {
				// Conversion en int
				indextest, err := strconv.Atoi(strings.Trim(player, "\n"))
				if err != nil {
					return ""
				}
				// On verifie que le choix correspond à la liste de personnage
				if isInList(indextest) {
					player = PersoChoice[indextest-1]
					// On supprime le choix fait par le joueur des choix pour les prochains joueurs
					PersoChoice = slices.Delete(PersoChoice, indextest-1, indextest)
					// On retourne le personnage choisi
					return player
				}
			}
		}
		LogAndFprintf(conn, msg)
	}
}