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

var playerTeamColors = []Color{
	Color256(27),
	Color256(44),
	Color256(6),
	Color256(105),
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
		unit.glyph.FG = playerTeamColors[i]

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
		class: "Knight",
		glyph: GlyphOf('@'),
		base: Stats{
			Speed: 4,
			Move:  5,
		},
		maxHP:  25,
		weapon: weaponShortsword,
		armor:  armorLeather,
		spells: []Weapon{
			spellTaunt,
			spellCharge,
		},
	},
	"Archer": Mob{
		class: "Archer",
		glyph: GlyphOf('@'),
		base: Stats{
			Speed: 6,
			Move:  7,
		},
		maxHP:  15,
		maxMP:  10,
		weapon: weaponBow,
		armor:  armorLeather,
		spells: []Weapon{
			spellCripple,
			spellPoisonShot,
		},
	},
	"Wizard": Mob{
		class: "Wizard",
		glyph: GlyphOf('@'),
		base: Stats{
			Speed: 5,
			Move:  6,
		},
		maxHP:  15,
		maxMP:  35,
		weapon: weaponStaff,
		spells: []Weapon{
			spellFireball,
		},
		armor: armorRobe,
	},
	"Priest": Mob{
		class: "Priest",
		glyph: GlyphOf('@'),
		base: Stats{
			Speed: 6,
			Move:  5,
		},
		hp:     15,
		maxHP:  20,
		maxMP:  25,
		weapon: weaponStaff,
		spells: []Weapon{
			spellHeal,
			spellRenew,
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
		// "oneroom",
		"courtyard",
		// "islandfort",
	},
	{
		"split",
	},
	{
		"dojo",
	},
	{
		"chambers",
	},
	{
		"forest",
	},
	{
		"mountain",
	},
	{
		"throne",
	},
}

var monstersByLevel = [][]Mob{
	{
		Mob{
			name:  "cute blob",
			glyph: GlyphOf('o'),
			base: Stats{
				Speed: 6,
				Move:  3,
			},
			maxHP:  10,
			weapon: weaponLick,
		},
		Mob{
			name:  "rabbit",
			glyph: GlyphOf('w'),
			base: Stats{
				Speed: 8,
				Move:  4,
			},
			maxHP:  8,
			weapon: weaponBite,
		},
		Mob{
			name:  "little bird",
			glyph: GlyphOf('b'),
			base: Stats{
				Speed: 8,
				Move:  5,
			},
			maxHP:  6,
			weapon: weaponPeck,
		},
		Mob{
			name:  "pig",
			glyph: GlyphOf('p'),
			base: Stats{
				Speed: 10,
				Move:  3,
			},
			maxHP:  8,
			weapon: weaponScratch,
		},
	},
	{
		Mob{
			name:  "little Kobold",
			glyph: GlyphOf('k'),
			base: Stats{
				Speed: 5,
				Move:  5,
			},
			maxHP:  15,
			weapon: weaponShank,
		},
		Mob{
			name:  "big Kobold",
			glyph: GlyphOf('K'),
			base: Stats{
				Speed: 4,
				Move:  6,
			},
			maxHP:  20,
			weapon: weaponShank,
		},
		Mob{
			name:  "jackal",
			glyph: GlyphOf('d'),
			base: Stats{
				Speed: 8,
				Move:  6,
			},
			maxHP:  10,
			weapon: weaponBite,
		},
		Mob{
			name:  "sewer rat",
			glyph: GlyphOf('r'),
			base: Stats{
				Speed: 10,
				Move:  4,
			},
			maxHP:  8,
			weapon: weaponBite,
		},
	},
	{
		Mob{
			name:  "archer",
			glyph: GlyphOf('@'),
			base: Stats{
				Speed: 5,
				Move:  5,
			},
			maxHP:  15,
			weapon: weaponBow,
		},
		Mob{
			name:  "ninja",
			glyph: GlyphOf('@'),
			base: Stats{
				Speed: 6,
				Move:  6,
			},
			maxHP:  15,
			weapon: weaponSword,
		},
		Mob{
			name:  "samurai",
			glyph: GlyphOf('@'),
			base: Stats{
				Speed: 4,
				Move:  4,
			},
			maxHP:  20,
			weapon: weaponSpear,
		},
		Mob{
			name:  "fox",
			glyph: GlyphOf('f'),
			base: Stats{
				Speed: 12,
				Move:  4,
			},
			maxHP:  10,
			weapon: weaponBite,
		},
	},
	{
		Mob{
			name:  "dwarf",
			glyph: GlyphOf('h'),
			base: Stats{
				Speed: 5,
				Move:  5,
			},
			maxHP:  20,
			weapon: weaponPick,
		},
		Mob{
			name:  "gnome",
			glyph: GlyphOf('g'),
			base: Stats{
				Speed: 8,
				Move:  5,
			},
			maxHP:  15,
			weapon: weaponShank,
		},
		Mob{
			name:  "gnome lord",
			glyph: GlyphOf('G'),
			base: Stats{
				Speed: 8,
				Move:  4,
			},
			maxHP:  20,
			weapon: weaponSword,
		},
		Mob{
			name:  "horse",
			glyph: GlyphOf('H'),
			base: Stats{
				Speed: 5,
				Move:  8,
			},
			maxHP:  12,
			weapon: weaponKick,
		},
	},
	{
		Mob{
			name:  "bear",
			glyph: GlyphOf('B'),
			base: Stats{
				Speed: 4,
				Move:  4,
			},
			maxHP:  26,
			weapon: weaponSwipe,
		},
		Mob{
			name:  "hunter",
			glyph: GlyphOf('@'),
			base: Stats{
				Speed: 5,
				Move:  6,
			},
			maxHP:  20,
			weapon: weaponLongbow,
		},
		Mob{
			name:  "foxhound",
			glyph: GlyphOf('d'),
			base: Stats{
				Speed: 8,
				Move:  6,
			},
			maxHP:  12,
			weapon: weaponBite,
		},
	},
	{
		Mob{
			name:  "yeti",
			glyph: GlyphOf('Y'),
			base: Stats{
				Speed: 6,
				Move:  4,
			},
			maxHP:  25,
			weapon: weaponYetiFist,
		},
		Mob{
			name:  "polar bear",
			glyph: GlyphOf('P'),
			base: Stats{
				Speed: 3,
				Move:  3,
			},
			maxHP:  35,
			weapon: weaponSwipe,
		},
		Mob{
			name:  "snow fox",
			glyph: GlyphOf('S'),
			base: Stats{
				Speed: 7,
				Move:  7,
			},
			maxHP:  15,
			weapon: weaponSnowFoxBite,
		},
	},
	{
		Mob{
			name:  "golem",
			glyph: GlyphOf('&'),
			base: Stats{
				Speed: 3,
				Move:  4,
			},
			maxHP:  40,
			weapon: weaponCrush,
		},
		Mob{
			name:  "dragon",
			glyph: GlyphOf('D'),
			base: Stats{
				Speed: 5,
				Move:  5,
			},
			maxHP:  32,
			weapon: weaponFirebreathing,
		},
		Mob{
			name:  "archon",
			glyph: GlyphOf('A'),
			base: Stats{
				Speed: 4,
				Move:  6,
			},
			maxHP:  30,
			maxMP:  100,
			weapon: spellSmite,
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
	"Meepches",
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
	"Cordelia",
}
