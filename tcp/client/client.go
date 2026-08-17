package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	stdin := bufio.NewReader(os.Stdin)
	connReader := bufio.NewReader(conn)

	fmt.Print("Entrez votre pseudonyme : ")
	username, err := stdin.ReadString('\n')
	_, err = fmt.Fprintf(conn, "%s", username)
 
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			// Définir un timeout pour éviter les blocages infinis
			err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				log.Println("Erreur SetReadDeadline :", err)
				return
			}

			// Lire la réponse du serveur
			response, err := connReader.ReadString('\n')
			if err != nil {
				// Vérifier si c'est un timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Ignorer les timeouts (pas de message reçu)
				} else {
					log.Println("Erreur de lecture :", err)
					return
				}
			}
			// Afficher la réponse
			if response != "\n"{
				fmt.Print("\nServeur : " + response)
				fmt.Print("> ")
			}
		}
	}()

	for {
		var line string
		fmt.Print("> ")
		line, err = stdin.ReadString('\n')
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