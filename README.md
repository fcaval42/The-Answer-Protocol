# The-Answer-Protocol

## Checklist de parsing — `world.yaml` & `entities.yaml`

### 1. Structure globale de `world.yaml`

- [ ] Parser les 4 clés racines : `world.locations`, `items`, `npcs`, `quests`.
- [ ] Attention : dans le fichier réel, `items`, `npcs` et `quests` sont **au même
      niveau que `world`**, pas imbriqués dedans. `quests` n'existe pas encore dans
      la struct `World` / `worldFile` de [parsing/world.go](parsing/world.go) → à ajouter.

### 2. Locations

- [ ] `name`, `description` (string simple, parfois multi-lignes en YAML → OK nativement).
- [ ] `exits` : map `direction → ID de location (string)`, **pas** une `Location`
      imbriquée. Le champ actuel `Exits map[string]Location` va produire un
      mauvais mapping (les exits pointent vers un ID, pas vers l'objet complet) →
      le remplacer par `map[string]string` et résoudre les IDs vers de vraies
      `Location` dans une deuxième passe, après le parsing brut.
- [ ] `items` (optionnel) : liste d'IDs d'objets présents dans la pièce au démarrage.
- [ ] `npcs` (optionnel) : liste d'IDs de PNJ présents dans la pièce.
- [ ] `spawns` (optionnel, cf. exemple du sujet) : liste de `{npc_type, count}`.
      Le champ Go `NPCType` doit avoir le tag `yaml:"npc_type"`, sinon le mapping
      implicite de yaml.v3 (nom du champ tout en minuscules, sans underscore)
      ne matchera jamais `npc_type`.

### 3. Items

- [ ] `name`, `description`, `obtainable` (bool).
- [ ] Vérifier que les items marqués `obtainable: true` sont bien référencés dans
      au moins une location (contrainte du sujet : 2 items obtenables minimum).

### 4. NPCs

- [ ] `name`, `role` (`dialogue` / `quest_giver` / `enemy`), `description`,
      `dialogue` (`[]string`).
- [ ] `stats.hp` optionnel dans `world.yaml`. Les vraies stats de combat
      (`attack`, `defense`, `speed`, `difficulty`) vivent dans `entities.yaml` →
      il faudra fusionner les deux sources après parsing (merge par nom/ID).

### 5. Quests

- [ ] Absentes du struct actuel → à créer entièrement.
- [ ] `name`, `giver` (référence à un NPC ID), `description`.
- [ ] `objective` : polymorphe selon `type` (`fetch` → `item` + `deliver_to` ;
      `defeat` → `npc`). Deux options : une struct avec tous les champs en
      pointeurs optionnels + un champ `Type` discriminant, ou une méthode
      `UnmarshalYAML(value *yaml.Node) error` custom qui valide selon le type.
- [ ] `reward` : `type` (`skill` / `item`) + `value` (string) — même logique
      polymorphe que `objective`.

### 6. `entities.yaml`

- [ ] Aucun struct Go n'existe encore pour ce fichier → créer un `entities.go`
      dédié (pendant à `world.go`).
- [ ] `entities.players.<id>` : `name`, `hp`, `attack`, `defense`, `speed`,
      `respawn.{hp, location}` (`location` = ID de location, à valider contre
      `world.yaml`).
- [ ] `entities.enemies.<id>` : `name`, `hp`, `attack`, `defense`, `speed`,
      `difficulty` (string : `easy` / `medium` / `hard`...).
- [ ] Le commentaire en tête du fichier dit explicitement que `enemies` doit
      rester aligné avec les NPCs `role: "enemy"` de `world.yaml` → valider ça
      au chargement (erreur ou warning en cas de désync).

### 7. Validations croisées (exigées par le sujet : "valider les exits et
      références")

- [ ] Chaque `exits.<dir>` pointe vers une location existante.
- [ ] Chaque ID dans `location.items` existe dans `items`.
- [ ] Chaque ID dans `location.npcs` existe dans `npcs`.
- [ ] Chaque `quest.giver` existe dans `npcs`.
- [ ] Chaque `quest.objective.item` existe dans `items`, chaque
      `quest.objective.npc` existe dans `npcs`.
- [ ] Chaque `quest.objective.deliver_to` existe dans `locations`.
- [ ] Chaque enemy de `entities.yaml` correspond à un NPC `role: "enemy"` dans
      `world.yaml` (et vice versa).
- [ ] Le graphe de locations contient au moins une boucle + une branche
      optionnelle (8 rooms mini, pas de carte "en ligne") → un petit algo de
      détection de cycle sur le graphe fait un bon test unitaire.
- [ ] Cohérence bidirectionnelle des exits (recommandé, pas obligatoire) : si
      `A.east == B`, vérifier si `B.west == A`.

### Tips Go / yaml.v3

- [ ] Mettre des tags yaml explicites partout (`yaml:"npc_type"`,
      `yaml:"deliver_to"`, `yaml:"obtainable"`...) plutôt que de compter sur le
      mapping implicite : yaml.v3 met le nom du champ Go en minuscules mais ne
      convertit pas le camelCase en snake_case.
- [ ] `parsing_world` doit retourner `(worldFile, error)`, pas juste
      `worldFile`. Actuellement le code retourne `world_yaml` (le `[]byte`
      brut) alors que la signature attend un `worldFile` : ça ne compile pas
      en l'état. Il faut faire `yaml.Unmarshal(world_yaml, &wf)` puis
      `return wf, err`.
- [ ] Séparer clairement "parsing" (désérialisation brute) et "validation"
      (vérification des références croisées) en deux fonctions/fichiers
      distincts — plus facile à tester unitairement l'un sans l'autre.
- [ ] Écrire des tests avec `t.TempDir()` + petits fichiers YAML de test (cas
      valides et cas invalides : exit vers location inexistante, item
      dupliqué, quest sans giver...) plutôt que de tout vérifier à la main sur
      le vrai `world.yaml`.
- [ ] Dans [parsing/parse_main.go](parsing/parse_main.go), `os.ReadFile(world_file)`
      référence une variable `world_file` qui n'existe pas (le paramètre
      s'appelle `world_path`), et `entities_path` n'est jamais utilisé : à
      corriger, en faisant remonter l'erreur (`return err` / `os.Exit(1)`) au
      lieu d'un simple `fmt.Println`.
- [ ] Les accents dans les clés (`Lumière`, `Vallées_fleuries`,
      `Océan_suspendu`...) sont gérés nativement en UTF-8 par yaml.v3, mais si
      vous faites de la recherche insensible à la casse/accents sur les IDs,
      pensez à normaliser (`strings.ToLower` ne suffit pas pour les accents).
- [ ] yaml.v3 expose des erreurs typées (`*yaml.TypeError`) avec le détail des
      champs en échec : les logguer proprement plutôt que de les avaler,
      conformément à l'exigence "Server Logging" du sujet.
