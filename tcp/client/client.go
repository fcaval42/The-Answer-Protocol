package main

import (
	"bufio"
	"fmt"
	"github.com/chzyer/readline"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	. "the_answer_protocol/parsing"
	. "the_answer_protocol/tcp"
	. "the_answer_protocol/tcp/common"
)

var (
	locMu sync.Mutex
	loc   Location
)

var (
	worldMapMu sync.Mutex
	worldMap   WorldFile
)

func main() {
	file, err := os.OpenFile("protocol.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Erreur lors de l'ouverture du fichier de log : %v", err)
	}
	defer file.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	rl, err := readline.New(">>> ")
	if err != nil {
		log.Fatalf("[ERROR] Erreur readline : %v", err)
	}
	defer rl.Close()
	log.SetOutput(file)

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
	go func() {
		connRoutine := bufio.NewReader(conn)
		defer wg.Done()

		for {
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
		for {
			line, err := rl.Readline()
			if err != nil {
				if err == readline.ErrInterrupt || err == io.EOF {
					log.Println("[INFO] Interruption demandée par l'utilisateur")
					fmt.Print("Au revoir, Mon ami !\n")
				} else {
					log.Println("[ERROR] Erreur de lecture :", err)
				}
				conn.Close()
				return
			}
			line = strings.TrimSpace(line)
			if line == "DEBUG WORLD" {
				fmt.Print(worldMap)
			}
			if line == "DEBUG LOC" {
				fmt.Print(loc)
			}
			if line == "QUIT" {
				fmt.Fprintf(conn, "%s", line)
				fmt.Fprint(conn, "[INFO] Fermeture de la connexion...\n")
				fmt.Print("Au revoir, Mon ami !\n")
				os.Exit(0)
			}
			fmt.Fprintf(conn, "%s\n", line)
		}
	}()
	wg.Wait()
}
