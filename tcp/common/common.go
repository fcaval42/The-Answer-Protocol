package common

import "sync"

// Variable globale commune avec un mutex
var (
	PersoChoice = []string{"Gustave", "Maelle", "Sciel", "Lune", "Verso"}
	PersoLock sync.Mutex
)
