package common

import (
	"fmt"
	"log"
	"net"
	"os"
)

func LogAndPrintf(conn net.Conn, format string, args ...interface{}) {
	logger := log.New(
		os.Stdout,
		"SERVER: ",
		log.Ldate | log.Ltime,
	)
	file, err := os.OpenFile("../protocol.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Erreur lors de l'ouverture du fichier de log : %v", err)
	}
	defer file.Close()
	
	// Configure le logger pour écrire UNIQUEMENT dans le fichier
	log.SetOutput(file)
	// Utilisation du logger
	logger.Printf(format, args...)
	log.Printf(format, args...)
}

func LogAndFprintf(conn net.Conn, format string, args ...interface{}) {
		file, err := os.OpenFile("../protocol.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Erreur lors de l'ouverture du fichier de log : %v", err)
	}
	defer file.Close()
	
	// Configure le logger pour écrire UNIQUEMENT dans le fichier
	log.SetOutput(file)
	log.Printf(format, args...)
	fmt.Fprintf(conn, format, args...)
}
