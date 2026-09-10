package common

import (
	"net"
	"sync"
	"the_answer_protocol/parsing"
)

var (
	Clients     = make(map[net.Conn]bool)
	ClientsLock sync.Mutex
)

func IsDigit(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func IsInList(num int) bool {
	persoLen := len(PersoChoice)
	return num >= 1 && num <= persoLen
}

var (
	PersoChoice = []string{"Gustave", "Maëlle", "Lune", "Sciel", "Verso"}
	PersoLock   sync.Mutex
)

var (
	World     parsing.WorldFile
	WorldLock sync.Mutex
)
