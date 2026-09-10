package commands

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	. "the_answer_protocol/tcp/common"
)

func ConnectCom(conn net.Conn) string {
	PersoLock.Lock()
	file, err := os.OpenFile("protocol.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Erreur lors de l'ouverture du fichier de log : %v", err)
	}
	defer file.Close()
	log.SetOutput(file)
	defer PersoLock.Unlock()
	reader := bufio.NewReader(conn)

	if len(PersoChoice) == 0 {
		LogServFprintf(conn, "%s| [ERROR] Errno 900: Nombre de joueurs limite atteint (CONNECTION_FAILED)\n", conn.RemoteAddr())
		return "900"
	}

	charactere := make([]string, len(PersoChoice))
	for i, perso := range PersoChoice {
		charactere[i] = fmt.Sprintf("%d. %s", i+1, perso)
	}
	msg := fmt.Sprintf("[INFO] Choisissez votre joueur parmi la liste en entrant le numéro correspondant: %s\n", strings.Join(charactere, ", "))
	LogServFprintf(conn, "%s", msg)

	for {
		player, err := reader.ReadString('\n')
		LogServFprintf(conn, "%s| %s", conn.RemoteAddr().String(), player)
		if player != "\n" {
			if err != nil {
				log.Println(conn.RemoteAddr().String() + "| [ERROR] " + err.Error())
				return ""
			}

			if IsDigit(strings.Trim(player, "\n")) {
				indextest, err := strconv.Atoi(strings.Trim(player, "\n"))
				if err != nil {
					return ""
				}
				if IsInList(indextest) {
					player = PersoChoice[indextest-1]
					PersoChoice = slices.Delete(PersoChoice, indextest-1, indextest)
					LogAndFprintf(player, conn, "%s| Connection réussie ! Bienvenue %s !\n", conn.RemoteAddr().String(), player)
					return player
				}
			}
		}
		LogAndFprintf(player, conn, "%s", msg)
	}
}
