package main

import (
	"sync"
	"net"
)


// Création du'une map (dictionnaire) pour gérer tous les clients
// Protégé par un mutex
var (
	clients     = make(map[net.Conn]bool) // Map des connexions actives
	clientsLock sync.Mutex                // Mutex pour synchroniser l'accès à la map
)

// Fonction isdigit pour eviter de faire un atoi sur des caractères
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

// Permet de vérifier que le choix de l'utilisateur est dans
// la liste de personnage
func isInList(num int) bool {
	persoLen := len(PersoChoice)
	return num >= 1 && num <= persoLen
}


// Variable globale commune avec un mutex
var (
	PersoChoice = []string{"Gustave", "Maëlle", "Lune", "Sciel", "Verso"}
	PersoLock   sync.Mutex
)




