package common

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// Variable globale commune avec un mutex
var (
	PersoChoice = []string{"Gustave", "Maelle", "Sciel", "Lune", "Verso"}
	PersoLock   sync.Mutex
)

func LogAndPrintf(conn net.Conn, format string, args ...interface{}) {
	fmt.Printf(format, args...)
	log.Printf(format, args...)
}

func LogAndFprintf(conn net.Conn, format string, args ...interface{}) {
	fmt.Fprintf(conn, format, args...)
	log.Printf(format, args...)
}
