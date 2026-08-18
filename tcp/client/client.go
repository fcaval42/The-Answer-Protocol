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
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	stdin := bufio.NewReader(os.Stdin)
	connReader := bufio.NewReader(conn)

	select_player, err := connReader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Print(select_player)
	fmt.Print("> ")
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
