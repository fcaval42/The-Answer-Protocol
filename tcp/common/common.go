package common

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sync"
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
	logMutex   sync.Mutex
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

func LogAndPrintf(player string, conn net.Conn, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[SERVER] %s", msg)
	LogInFile("[%s] %s", player, msg)
}

func LogAndFprintf(player string, conn net.Conn, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(conn, format, args...)
	LogInFile("[%s] %s", player, msg)
}

func LogServFprintf(conn net.Conn, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	LogInFile("%s", msg)
	fmt.Fprintf(conn, format, args...)
}

func LogServPrintf(conn net.Conn, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	LogInFile("%s", msg)
	log.Printf("[SERVER] %s", msg)
}
