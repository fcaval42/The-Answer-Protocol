package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	. "the_answer_protocol/tcp/common"
	"strings"
)

func main() {
	// Ouvre (ou crée) un fichier pour les logs
	file, err := os.OpenFile("../protocol.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Erreur lors de l'ouverture du fichier de log : %v", err)
	}
	defer file.Close()
	
	// Configure le logger pour écrire UNIQUEMENT dans le fichier
	log.SetOutput(file)
	
	// Creation de la connexion en tcp au localhost
	conn, err := net.Dial("tcp", "localhost:8080")
	if conn == nil {
		log.Print("[ERROR] Errno 900: CONNECTION_FAILED")
		fmt.Print("[ERROR] Errno 900: CONNECTION_FAILED\n")
		return
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
		log.Println(conn.RemoteAddr().String() + "| [ERROR] " + err.Error())
		return
	}
	fmt.Print(select_player) 
	if strings.Contains(select_player, "[ERROR]"){
		return
	}
	// Affichage du message
	fmt.Print(">>> ")

	// On récupère le choix de l'utilisateur
	player, err := stdin.ReadString('\n')

	// On envoi le choix au serveur
	fmt.Fprint(conn, player)

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
				log.Println("[ERROR] Erreur de lecture :", err)
				return
			}
			// On affiche le message s'il n'est pas vide
			if response != "\n" {
				fmt.Print("\n" + response)
				fmt.Print(">>> ")
			}
		}
	}()

	// Boucle infinie pour les inputs
	for {
		fmt.Print(">>> ")
		// Récupèration de l'input
		line, err := stdin.ReadString('\n')
		if err != nil {
			log.Println(conn.RemoteAddr().String() + "| [ERROR] " + err.Error())
			return
		}

		// Si l'input est "exit", gestion de la fermeture de la connexion
		if line == "exit\n" {
			fmt.Fprint(conn, "[INFO] Fermeture de la connexion...\n")
			return
		}

		// On envoie l'input au serveur
		LogAndFprintf(conn, "%s", line)
	}
}
