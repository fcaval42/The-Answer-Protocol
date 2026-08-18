package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
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

	// On envoi le choix au serveur
	conn.Write([]byte(player))

	// Goroutine qui va chercher en continue ce qu'envoie le serveur
	go func() {
		// Buffer en ecoute du serveur, unique a la goroutine pour eviter
		// que plusieurs thread utilisent le meme buffer
		connRoutine := bufio.NewReader(conn)

		// Boucle infinie
		for {
			// Recupère les messages envoyés par le serveur
			response, err := connRoutine.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Println("Erreur de lecture :", err)
				return
			}
			// On affiche le message s'il n'est pas vide
			if response != "\n" {
				fmt.Print("\nServeur : " + response)
				fmt.Print("> ")
			}
		}
	}()

	// Boucle infinie pour les inputs
	for {
		fmt.Print("> ")
		// Récupèration de l'input
		line, err := stdin.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		// Si l'input est "exit", gestion de la fermeture de la connexion
		if line == "exit\n" {
			fmt.Println("Fermeture de la connexion...")
			return
		}

		// On envoie l'input au serveur
		conn.Write([]byte(line))
	}
}
