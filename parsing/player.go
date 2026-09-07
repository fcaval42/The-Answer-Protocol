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
}
