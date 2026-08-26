package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"the_answer_protocol/gui"
	"the_answer_protocol/parsing"
	. "the_answer_protocol/tcp/common"

	"github.com/chzyer/readline"
)

// Capture le nom du personnage dans le message de confirmation envoyé par
// le serveur une fois la sélection terminée (cf. tcp/commands/session.go).
var welcomeRe = regexp.MustCompile(`Bienvenue (\S+) !`)

// readServerMessages écoute en continu ce qu'envoie le serveur et l'affiche
// au client. world est passé en paramètre pour que cette fonction (et celles
// qui suivront) puisse s'y référer, comme le font déjà les commandes du
// package tcp/commands.
func readServerMessages(conn net.Conn, rl *readline.Instance, world parsing.WorldFile, withGUI bool, playerChosen chan<- string) {
	connRoutine := bufio.NewReader(conn)
	for {
		response, err := connRoutine.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Println("[ERROR] Erreur de lecture :", err)
			return
		}
		if strings.Contains(response, "\\") {
			response = strings.ReplaceAll(response, "\\", "\n")
		}
		if response != "\n" {
			ClearScreen()
			fmt.Fprint(rl.Stdout(), "\n"+response)
		}
		if withGUI {
			if m := welcomeRe.FindStringSubmatch(response); m != nil {
				select {
				case playerChosen <- m[1]:
				default:
				}
			}
		}
	}
}

// runInputLoop lit les commandes tapées au clavier et les envoie au serveur.
// world est transmis pour les mêmes raisons que dans readServerMessages.
func runInputLoop(conn net.Conn, rl *readline.Instance, world parsing.WorldFile) {
	for {
		line, err := rl.Readline()
		if err != nil {
			log.Println("[ERROR] Erreur de lecture :", err)
			return
		}
		line = strings.TrimSpace(line)

		if line == "QUIT" {
			fmt.Fprintf(conn, "%s", line)
			fmt.Fprint(conn, "[INFO] Fermeture de la connexion...\n")
			fmt.Print("Au revoir, Mon ami !\n")
			return
		}
		fmt.Fprintf(conn, "%s\n", line)
	}
}

func main() {

	world, worldErrs := parsing.Rally("yaml/world.yaml")
	if len(worldErrs) > 0 {
		fmt.Println("[ERROR] Impossible de charger le monde :", errors.Join(worldErrs...))
		return
	}

	withGUI := len(os.Args) > 1 && strings.EqualFold(os.Args[1], "gui")

	file, err := os.OpenFile("../protocol.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Erreur lors de l'ouverture du fichier de log : %v", err)
	}
	defer file.Close()
	log.SetOutput(file)

	rl, err := readline.New(">>> ")
	if err != nil {
		log.Fatalf("[ERROR] Erreur readline : %v", err)
	}
	defer rl.Close()

	conn, err := net.Dial("tcp", "localhost:8080")
	if conn == nil {
		log.Print("[ERROR] Errno 900: CONNECTION_FAILED")
		fmt.Print("[ERROR] Errno 900: CONNECTION_FAILED\n")
		return
	}
	defer func() {
		fmt.Fprint(conn, "QUIT\n")
		conn.Close()
	}()

	// Déclenche immédiatement le menu de sélection de personnage côté serveur
	// (cf. tcp/server/server.go, qui attend "CONNECT\n" avant d'afficher quoi
	// que ce soit) : sans ça, le joueur se retrouve devant un prompt vide sans
	// savoir qu'il doit taper CONNECT lui-même.
	fmt.Fprint(conn, "CONNECT\n")

	playerChosen := make(chan string, 1)

	// Goroutine qui va chercher en continue ce qu'envoie le serveur
	go readServerMessages(conn, rl, world, withGUI, playerChosen)

	if !withGUI {
		runInputLoop(conn, rl, world)
		return
	}

	// Le choix du personnage se fait au clavier ; une fois confirmé par le
	// serveur, on ouvre le GUI (sur la goroutine principale, requis par Fyne)
	// pendant que le client textuel continue de tourner en tâche de fond.
	go runInputLoop(conn, rl, world)
	player := <-playerChosen
	gui.Run(player, world)
}
