package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"the_answer_protocol/tcp/common"
)

var (
	clients     = make(map[net.Conn]bool) // Map des connexions actives
	clientsLock sync.Mutex                // Mutex pour synchroniser l'accès à la map
)


func isInList(num int) bool {

    // 2. Vérifier que num est entre 1 et len(common.PersoChoice)
    persoLen := len(common.PersoChoice)
    return num >= 1 && num <= persoLen
}


func removeUnordered(s []string, i int) []string {
    s[i] = s[len(s)-1] // remplace l'élément i par le dernier
    return s[:len(s)-1] // réduit la taille du slice
}


func selectPlayer(conn net.Conn) string{
	common.PersoLock.Lock()
	defer common.PersoLock.Unlock()
	reader := bufio.NewReader(conn)

	charactere := make([]string, len(common.PersoChoice))
	for i, perso := range common.PersoChoice{
		charactere[i] = fmt.Sprintf("%d. %s", i+1, perso)
	}

	fmt.Fprintf(conn, "Choisissez votre joueur parmi en entrant le numéro correspondant: %s\n", strings.Join(charactere, ", "))


	for{
		player, err := reader.ReadString('\n')
		if player != "\n"{
			if err != nil {
				log.Println(err)
				return ""
			}
			indextest, err := strconv.Atoi(strings.Trim(player, "\n"))
			if err != nil {
			return ""
			}
			if isInList(indextest){
				player = common.PersoChoice[indextest - 1]
				common.PersoChoice = removeUnordered(common.PersoChoice, indextest - 1)
				return player
			}
		}
		fmt.Fprintf(conn, "Choisissez votre joueur parmi en entrant le numéro correspondant: %s\n", strings.Join(charactere, ", "))
		
	}
}


func broadcast(message string, sender net.Conn, player string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	// Préparer le message à envoyer (avec le pseudo du client expéditeur)
	fullMessage := fmt.Sprintf("[%s] : %s\n", player, message)

	// Envoyer à tous les clients sauf l'expéditeur
	for client := range clients {
		if client == sender {
			continue
		}
		_, err := client.Write([]byte(fullMessage))
		if err != nil {
			log.Printf("Erreur lors de l'envoi à %s : %v\n", client.RemoteAddr(), err)
			// Supprimer le client en erreur
			delete(clients, client)
			client.Close()
		}
	}
}

func handleConnection(conn net.Conn) {
	var player string
	defer func() {
		// Supprimer le client de la liste à la fin
		clientsLock.Lock()
		delete(clients, conn)
		clientsLock.Unlock()
		if player != ""{
			common.PersoLock.Lock()
			common.PersoChoice = append(common.PersoChoice, player)
			common.PersoLock.Unlock()
		}
		conn.Close()
		fmt.Printf("Client déconnecté : %s\n", conn.RemoteAddr())
	}()

	reader := bufio.NewReader(conn)
	player = selectPlayer(conn)
	player = strings.Trim(player, "\n")
	if player == "" {
		log.Println("Error: Player doesn't exist")
		return
	}
	
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("Recu de %s : %s", player, message)
		broadcast(message, conn, player)
	}
}

func main() {
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		clientsLock.Lock()
		clients[conn] = true
		clientsLock.Unlock()

		fmt.Printf("Nouveau client connecté : %s\n", conn.RemoteAddr())
		go handleConnection(conn)
	}
}
