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
		speed:  3,
		move:   5,
		maxHP:  25,
		weapon: weaponSword,
		armor:  armorLeather,
	},
	"Archer": Mob{
		class:  "Archer",
		glyph:  GlyphOf('@'),
		speed:  5,
		move:   6,
		maxHP:  15,
		maxMP:  15,
		weapon: weaponBow,
		armor:  armorLeather,
	},
	"Wizard": Mob{
		class:  "Wizard",
		glyph:  GlyphOf('@'),
		speed:  4,
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
			move:   4,
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
			move:   6,
			maxHP:  6,
			weapon: weaponPeck,
		},
		Mob{
			name:   "pig",
			glyph:  GlyphOf('p'),
			speed:  10,
			move:   4,
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
}

var PlayerNames = []string{
	"Kelladros",
	"Cid",
	"Papi",
	"Franklin",
	"Ella",
	"Kiffe",
	"Stewart",
	"Meep",
	"Wes",
	"Aaron",
	"Grayson",
	"Ernest",
	"Kim",
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
}
