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
	// Utilisation du logger
	logger.Printf(format, args...)
	log.Printf(format, args...)
}

func LogAndFprintf(conn net.Conn, format string, args ...interface{}) {
	log.Printf(format, args...)
	fmt.Fprintf(conn, format, args...)
}
