package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"strings"
)

var (
	clients     = make(map[net.Conn]bool) // Map des connexions actives
	clientsLock sync.Mutex                // Mutex pour synchroniser l'accès à la map
)


func broadcast(message string, sender net.Conn, username string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	// Préparer le message à envoyer (avec l'adresse du client expéditeur)
	fullMessage := fmt.Sprintf("[%s] : %s\n", username, message)

	// Envoyer à tous les clients sauf l'expéditeur (optionnel)
	for client := range clients {
		if client == sender {
			continue // Sauter l'expéditeur (optionnel)
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
	defer func() {
		// Supprimer le client de la liste à la fin
		clientsLock.Lock()
		delete(clients, conn)
		clientsLock.Unlock()
		conn.Close()
		fmt.Printf("Client déconnecté : %s\n", conn.RemoteAddr())
	}()

	reader := bufio.NewReader(conn)

	username, err := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if err != nil {
		log.Println(err)
		return
	}

	for {
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("Recu de %s : %s", username, message)
			broadcast(message, conn, username)
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