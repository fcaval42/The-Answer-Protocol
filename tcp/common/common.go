package common

import "sync"

var (
	PersoChoice = []string{"Gustave", "Maelle", "Sciel", "Lune", "Verso"}
	PersoLock sync.Mutex
)
