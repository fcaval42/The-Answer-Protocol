package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"os"
	"strings"
	"io"
	"github.com/chzyer/readline"
	. "the_answer_protocol/tcp/common"
	. "the_answer_protocol/tcp"
	. "the_answer_protocol/parsing"
	// "strings"
)


var (
    locMu	sync.Mutex
    loc Location
)

var (
    worldMapMu sync.Mutex
    worldMap WorldFile
)

func main() {
	// Ouvre un fichier pour les logs
	file, err := os.OpenFile("protocol.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Erreur lors de l'ouverture du fichier de log : %v", err)
	}
	defer file.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	// reader := bufio.NewReader(os.Stdin)
	rl, err := readline.New(">>> ")
	if err != nil {
		log.Fatalf("[ERROR] Erreur readline : %v", err)
	}
	defer rl.Close()
	// Configure le logger pour ecrire dans le fichier
	log.SetOutput(file)

	// Creation de la connexion en tcp au localhost
	conn, err := net.Dial("tcp", "localhost:8080")
	if conn == nil {
		log.Print("[ERROR] Errno 900: CONNECTION_FAILED")
		fmt.Print("[ERROR] Errno 900: CONNECTION_FAILED\n")
		return
	}

	// Fermeture de la connexion a la fin du programme
	defer func() {
		fmt.Fprint(conn, "QUIT\n")
		conn.Close()
	}()
	// Goroutine qui va chercher en continue ce qu'envoie le serveur
	go func() {
		// Buffer en ecoute du serveur, unique a la goroutine pour eviter
		// que plusieurs thread utilisent le meme buffer
		connRoutine := bufio.NewReader(conn)
		defer wg.Done()

		// Boucle infinie
		for {
			// Recupere les messages envoyes par le serveur
			response, err := connRoutine.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Println("[ERROR] Fin de connexion :", err)
				fmt.Print("La Peintresse a déchirée la toile !\n")
				os.Exit(0)
			}
			if strings.Contains(response, "\\") {
				response = strings.ReplaceAll(response, "\\", "\n")
			}
			if strings.HasPrefix(response, "{\"ID\":") {
				response = strings.TrimSpace(response)
				location := FromString[Location](response)
				locMu.Lock()
				loc = location
				locMu.Unlock()
				continue
			}
			if strings.HasPrefix(response, "{\"World\":{") {
				response = strings.TrimSpace(response)
				world := FromString[WorldFile](response)
				worldMapMu.Lock()
				worldMap = world
				worldMapMu.Unlock()
				continue
			}
			// On affiche le message s'il n'est pas vide
			if response != "\n" {
				response = strings.TrimRight(response, "\r\n")
				if response == "" {
					continue
				}
				if strings.HasPrefix(response, "[EVT") {
					fmt.Fprint(rl.Stdout(), response+"\n")
				} else {
					ClearScreen()
					fmt.Fprint(rl.Stdout(), response+"\n")
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		// Boucle infinie pour les inputs
		for {
			// Recuperation de l'input
			line, err := rl.Readline()
			if err != nil {
				if err == readline.ErrInterrupt || err == io.EOF {
				log.Println("[INFO] Interruption demandée par l'utilisateur")
				fmt.Print("Au revoir, Mon ami !\n")
			} else {
				log.Println("[ERROR] Erreur de lecture :", err)
			}
			conn.Close()   // débloque la goroutine réseau bloquée sur ReadString
			return
			}
			line = strings.TrimSpace(line)
			if line == "DEBUG WORLD" {
				fmt.Print(worldMap)
			}
			if line == "DEBUG LOC" {
				fmt.Print(loc)
			}
			// Si l'input est "exit", gestion de la fermeture de la connexion
			if line == "QUIT" {
				fmt.Fprintf(conn, "%s", line)
				fmt.Fprint(conn, "[INFO] Fermeture de la connexion...\n")
				fmt.Print("Au revoir, Mon ami !\n")
				os.Exit(0)
			}
			// On envoie l'input au serveur
			fmt.Fprintf(conn, "%s\n", line)
		}
	}()
	wg.Wait()
}
