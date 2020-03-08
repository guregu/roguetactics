package main

import (
	"math/rand"
)

const (
	PlayerTeam = 0
	AITeam     = 1
)

type Team struct {
	ID    int
	Units []*Mob
}

type Battle struct {
	Map   string
	Teams []Team
}

type Class string

var PlayerClasses = []Class{
	"Knight",
	"Archer",
	"Wizard",
	"Priest",
}

func newBattle(level int, playerTeam Team) Battle {
	return Battle{
		Map:   randomMap(level),
		Teams: []Team{playerTeam, generateEnemyTeam(level)},
	}
}

func generatePlayerTeam() Team {
	team := Team{
		ID: PlayerTeam,
	}
	classes := make(map[Class]bool)
	names := rand.Perm(len(PlayerNames))
	for i := 0; i < 4; i++ {
		class := randomClass()
		if classes[class] {
			for {
				if rand.Float32() < 0.4 {
					break
				}
				class = randomClass()
				if !classes[class] {
					break
				}
			}
		}
		classes[class] = true

		unit := generateUnit(class)
		unit.name = PlayerNames[names[i]]
		unit.glyph.FG = ColorBlue

		team.Units = append(team.Units, unit)
	}

	return team
}

func generateEnemyTeam(level int) Team {
	const teamSize = 4
	monsters := monstersByLevel[level]

	team := Team{
		ID: AITeam,
	}

	for i := 0; i < teamSize; i++ {
		mob := monsters[rand.Intn(len(monsters))]
		mob.team = AITeam
		mob.glyph.FG = ColorRed
		team.Units = append(team.Units, &mob)
	}

	return team
}

func randomClass() Class {
	return PlayerClasses[rand.Intn(len(PlayerClasses))]
}

func generateUnit(class Class) *Mob {
	unit := classBase[class]
	return &unit
}

var classBase = map[Class]Mob{
	"Knight": Mob{
		class:  "Knight",
		glyph:  GlyphOf('@'),
		speed:  4,
		move:   5,
		maxHP:  25,
		weapon: weaponShortsword,
		armor:  armorLeather,
		spells: []Weapon{
			spellTaunt,
		},
	},
	"Archer": Mob{
		class:  "Archer",
		glyph:  GlyphOf('@'),
		speed:  6,
		move:   7,
		maxHP:  15,
		maxMP:  5,
		weapon: weaponBow,
		armor:  armorLeather,
		spells: []Weapon{
			spellCripple,
		},
	},
	"Wizard": Mob{
		class:  "Wizard",
		glyph:  GlyphOf('@'),
		speed:  5,
		move:   6,
		maxHP:  15,
		maxMP:  35,
		weapon: weaponStaff,
		spells: []Weapon{
			spellFireball,
		},
		armor: armorRobe,
	},
	"Priest": Mob{
		class:  "Priest",
		glyph:  GlyphOf('@'),
		speed:  6,
		move:   5,
		hp:     15,
		maxHP:  20,
		maxMP:  25,
		weapon: weaponStaff,
		spells: []Weapon{
			spellHeal,
		},
		armor: armorRobe,
	},
}

func randomMap(level int) string {
	maps := mapsByLevel[level]
	return maps[rand.Intn(len(maps))]
}

var mapsByLevel = [][]string{
	{
		"oneroom",
	},
	{
		"chambers",
	},
	{
		"dojo",
	},
	{
		"dojo",
	},
}

var monstersByLevel = [][]Mob{
	{
		Mob{
			name:   "cute blob",
			glyph:  GlyphOf('o'),
			speed:  6,
			move:   3,
			maxHP:  10,
			weapon: weaponLick,
		},
		Mob{
			name:   "rabbit",
			glyph:  GlyphOf('w'),
			speed:  8,
			move:   4,
			maxHP:  8,
			weapon: weaponBite,
		},
		Mob{
			name:   "little bird",
			glyph:  GlyphOf('b'),
			speed:  8,
			move:   5,
			maxHP:  6,
			weapon: weaponPeck,
		},
		Mob{
			name:   "pig",
			glyph:  GlyphOf('p'),
			speed:  10,
			move:   3,
			maxHP:  8,
			weapon: weaponScratch,
		},
	},
	{
		Mob{
			name:   "little Kobold",
			glyph:  GlyphOf('k'),
			speed:  5,
			move:   5,
			maxHP:  15,
			weapon: weaponShank,
		},
		Mob{
			name:   "big Kobold",
			glyph:  GlyphOf('K'),
			speed:  4,
			move:   6,
			maxHP:  20,
			weapon: weaponShank,
		},
		Mob{
			name:   "jackal",
			glyph:  GlyphOf('d'),
			speed:  8,
			move:   6,
			maxHP:  10,
			weapon: weaponBite,
		},
		Mob{
			name:   "sewer rat",
			glyph:  GlyphOf('r'),
			speed:  10,
			move:   4,
			maxHP:  8,
			weapon: weaponBite,
		},
	},
	{
		Mob{
			name:   "archer",
			glyph:  GlyphOf('@'),
			speed:  5,
			move:   5,
			maxHP:  15,
			weapon: weaponBow,
		},
		Mob{
			name:   "ninja",
			glyph:  GlyphOf('@'),
			speed:  6,
			move:   6,
			maxHP:  15,
			weapon: weaponSword,
		},
		Mob{
			name:   "bear",
			glyph:  GlyphOf('B'),
			speed:  4,
			move:   4,
			maxHP:  20,
			weapon: weaponSwipe,
		},
		Mob{
			name:   "fox",
			glyph:  GlyphOf('f'),
			speed:  12,
			move:   4,
			maxHP:  10,
			weapon: weaponBite,
		},
	},
	{
		Mob{
			name:   "dwarf",
			glyph:  GlyphOf('h'),
			speed:  5,
			move:   5,
			maxHP:  20,
			weapon: weaponPick,
		},
		Mob{
			name:   "gnome",
			glyph:  GlyphOf('g'),
			speed:  8,
			move:   5,
			maxHP:  15,
			weapon: weaponShank,
		},
		Mob{
			name:   "gnome lord",
			glyph:  GlyphOf('G'),
			speed:  8,
			move:   4,
			maxHP:  20,
			weapon: weaponSword,
		},
	},
}

var PlayerNames = []string{
	"Kelladros",
	"Cid",
	"Papi",
	"Franklin",
	"Ella",
	"Kiffe",
	"Stewart",
	"Miepches",
	"Wes",
	"Aaron",
	"Grayson",
	"Ernest",
	"Eugene",
	"Kim",
	"Benjamin",
	"Eric",
	"Sophie",
	"Constance",
	"Maria",
	"Taro",
	"Aoi",
	"Murton",
	"Bron",
	"Dixon",
	"Laius",
	"Juan",
	"Modesty",
	"Alice",
	"Leroy",
	"Guy",
	"Cecelia",
	"Mimi",
	"Mildred",
	"Clarence",
	"Doris",
	"Cyril",
	"Nigel",
	"Horace",
	"Ray",
	"Leonard",
	"Elaine",
	"Jerry",
	"George",
	"Kramer",
	"Newman",
	"Arthur",
	"Carles",
	"Ronaldo",
	"Quintus",
	"Jelani",
	"Rose",
	"Bernard",
	"Claud",
	"Benor",
	"Eris",
	"Roscoe",
	"Jebuiz",
	"Saley",
	"Willach",
	"Tiko",
	"Chris",
	"Lela",
	"Mercy",
	"Maxwell",
	"Michael",
	"Dorian",
	"Conrad",
	"Samio",
	"Victorio",
	"Tomothy",
	"Bustopher",
	"Narry",
	"Eldrac",
	"Rob",
	"Violet",
	"Lucy",
	"Beatrice",
	"Edith",
	"Vera",
	"Oscar",
	"Vahe",
	"Kagami",
	"Rico",
	"Satoshi",
	"Marcel",
	"Enef",
	"Erika",
	"Miyao",
	"Zafod",
	"Colette",
	"Nephele",
	"Adelle",
	"Ethel",
	"Ollie",
	"Dustin",
	"Antigone",
	"Carson",
	"Alamar",
	"Arvin",
	"Harquad",
	"Jan",
	"Philoketes",
	"Ramzuh",
	"Deleta",
	"Mareo",
	"Rikasa",
	"Bossie",
	"Steen",
	"Asuka",
	"Chimi",
	"Jojo",
	"Horo",
	"Lawrence",
	"Pipo",
	"Luie",
	"Reginald",
	"Rygel",
	"Balindria",
	"Sven",
	"Dorado",
	"Abraham",
	"Titus",
	"Wolfgang",
	"Leslie",
	"Gertrude",
	"Carmine",
	"Kytel",
	"Wilbur",
}
