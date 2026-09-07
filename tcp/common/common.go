package common

import (
	"fmt"
	"log"
	"net"
	"sync"
	"os"
	"runtime"
	"os/exec"
)


func ClearScreen() {
    if runtime.GOOS == "windows" {
        cmd := exec.Command("cmd", "/c", "cls")
        cmd.Stdout = os.Stdout
        cmd.Run()
    } else {
        cmd := exec.Command("clear")
        cmd.Stdout = os.Stdout
        cmd.Run()
    }
}

var (
	logMutex  sync.Mutex
	fileLogger *log.Logger
)

func init() {
	file, err := os.OpenFile("protocol.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Erreur lors de l'ouverture du fichier de log : %v", err)
	}

	fileLogger = log.New(file, "", log.Ldate|log.Ltime)
}

func LogInFile(format string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	fileLogger.Printf(format, args...)
}

func LogAndPrintf(conn net.Conn, format string, args ...interface{}) {
	LogInFile(format, args...)
	log.Printf(format, args...)
}

func LogAndFprintf(conn net.Conn, format string, args ...interface{}) {
	LogInFile(format, args...)
	fmt.Fprintf(conn, format, args...)
}
