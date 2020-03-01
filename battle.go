package main

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

func newBattle(mapname string, playerTeam Team) Battle {
	return Battle{
		Map:   mapname,
		Teams: []Team{playerTeam, generateEnemyTeam()},
	}
}

func generatePlayerTeam() Team {
	glyph := GlyphOf('@')
	glyph.FG = ColorBlue

	// glyph2 := GlyphOf('d')
	// glyph2.FG = ColorBlue

	team := Team{
		ID: PlayerTeam,
		Units: []*Mob{
			&Mob{
				name:   "Knight",
				glyph:  glyph,
				speed:  3,
				move:   5,
				hp:     20,
				maxHP:  20,
				weapon: &weaponSword,
			},
			&Mob{
				name:   "Archer",
				glyph:  glyph,
				speed:  5,
				move:   10,
				hp:     15,
				maxMP:  15,
				weapon: &weaponBow,
			},
			&Mob{
				name:   "Wizard",
				glyph:  glyph,
				speed:  3,
				move:   10,
				hp:     15,
				maxMP:  15,
				weapon: &weaponStaff,
			},
			&Mob{
				name:   "Priest",
				glyph:  glyph,
				speed:  4,
				move:   10,
				hp:     15,
				maxMP:  15,
				weapon: &weaponStaff,
			},
		},
	}
	return team
}

func generateEnemyTeam() Team {
	koboldGlyph := GlyphOf('k')
	kobold2Glyph := GlyphOf('K')
	koboldGlyph.FG = ColorRed
	kobold2Glyph.FG = ColorRed

	enemies := Team{
		ID: AITeam,
		Units: []*Mob{
			&Mob{
				name:   "little Kobold",
				glyph:  koboldGlyph,
				speed:  5,
				move:   5,
				hp:     10,
				maxHP:  10,
				weapon: &weaponShank,
				team:   1,
			},
			&Mob{
				name:   "big Kobold",
				glyph:  kobold2Glyph,
				speed:  4,
				move:   4,
				hp:     20,
				maxMP:  20,
				weapon: &weaponShank,
				team:   1,
			},
			&Mob{
				name:   "big Kobold",
				glyph:  kobold2Glyph,
				speed:  4,
				move:   4,
				hp:     20,
				maxMP:  20,
				weapon: &weaponShank,
				team:   1,
			},
			&Mob{
				name:   "big Kobold",
				glyph:  kobold2Glyph,
				speed:  4,
				move:   4,
				hp:     20,
				maxMP:  20,
				weapon: &weaponShank,
				team:   1,
			},
		},
	}

	return enemies
}
