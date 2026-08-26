package common

import (
	"sync"
	"net"
	"the_answer_protocol/parsing"
)


// Création du'une map (dictionnaire) pour gérer tous les clients
// Protégé par un mutex
var (
	Clients     = make(map[net.Conn]bool) // Map des connexions actives
	ClientsLock sync.Mutex                // Mutex pour synchroniser l'accès à la map
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
func IsInList(num int) bool {
	persoLen := len(PersoChoice)
	return num >= 1 && num <= persoLen
}


// Variable globale commune avec un mutex
var (
	PersoChoice = []string{"Gustave", "Maëlle", "Lune", "Sciel", "Verso"}
	PersoLock   sync.Mutex
)

// Monde partagé, parsé une seule fois au lancement du serveur (voir
// tcp/server/server.go). Plusieurs clients pouvant lire/modifier des joueurs
// en même temps (connexion, déconnexion...), tout accès à World doit passer
// par WorldLock.
var (
	World     parsing.WorldFile
	WorldLock sync.Mutex
)




