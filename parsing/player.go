package parsing

type Weapon struct {
	Name     string  `yaml:"name"`
	Critical float64 `yaml:"critical"`
}

type Respawn struct {
	HP       int    `yaml:"hp"`
	Location string `yaml:"location"`
}

type Player struct {
	Name    string  `yaml:"name"`
	HP      int     `yaml:"hp"`
	MaxHP   int     `yaml:"-"`
	Weapon  Weapon  `yaml:"weapon"`
	Attack  int     `yaml:"attack"`
	Defense int     `yaml:"defense"`
	Respawn Respawn `yaml:"respawn"`
	Inventory []Item
	Quests []PlayerQuest
	Connected bool
	// BaseWeapon est l'arme de depart du joueur (copiee depuis Weapon au
	// chargement du monde) : UNEQUIP y revient. EquippedItem est l'ID de
	// l'objet d'inventaire actuellement equipe ("" = arme de base), ce qui
	// permet de n'autoriser qu'un seul objet equipe a la fois et d'empecher
	// de re-equiper un objet deja porte.
	BaseWeapon   Weapon `yaml:"-"`
	EquippedItem string `yaml:"-"`
}
