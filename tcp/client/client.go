package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// Creation de la connexion en tcp au localhost
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}
	// Fermeture de la connexion a la fin du programme
	defer conn.Close()

	// Buffer de l'entrée clavier
	stdin := bufio.NewReader(os.Stdin)

	// Buffer des messages de la connexion
	connReader := bufio.NewReader(conn)

	// On recupere le message envoyé par le serveur pour la selection de personnages
	select_player, err := connReader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}

	// Affichage du message
	fmt.Print(select_player)
	fmt.Print("> ")

	// On récupère le choix de l'utilisateur
	player, err := stdin.ReadString('\n')
	_, err = fmt.Fprintf(conn, "%s", player)
	player = strings.Trim(player, "\n")

	go func() {
		connRoutine := bufio.NewReader(conn)
		for {

			response, err := connRoutine.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Println("Erreur de lecture :", err)
				return
			}

			if response != "\n" {
				fmt.Print("\nServeur : " + response)
				fmt.Print("> ")
			}
		}
	}()

	for {
		fmt.Print("> ")
		line, err := stdin.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		if line == "exit\n" {
			fmt.Println("Fermeture de la connexion...")
			return
		}

		_, err = fmt.Fprintf(conn, "%s", line)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
