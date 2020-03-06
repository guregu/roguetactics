package main

import (
	"fmt"
	"math/rand"
)

type Bonus struct {
	Name  string
	Apply func(*Mob)
}

func generateBonuses(team Team, level int) []Bonus {
	bonuses := make([]Bonus, 0, len(team.Units))
	for _, unit := range team.Units {
		bonuses = append(bonuses, randomBonus(level, unit))
	}
	return bonuses
}

func randomBonus(level int, unit *Mob) Bonus {
	if rand.Float64() >= 0.4 {
		bonuses := classBonuses[unit.Class()]
		if len(bonuses) > 0 {
			return bonuses[rand.Intn(len(bonuses))](level, unit)
		}
	}

	return genericBonuses[rand.Intn(len(genericBonuses))](level)
}

var genericBonuses = []func(level int) Bonus{
	func(level int) Bonus {
		hp := (level + 1) * 5
		return Bonus{
			Name: fmt.Sprintf("+%d HP", hp),
			Apply: func(m *Mob) {
				m.maxHP += hp
			},
		}
	},
	func(level int) Bonus {
		speed := rand.Intn(2) + 1
		return Bonus{
			Name: fmt.Sprintf("+%d Speed", speed),
			Apply: func(m *Mob) {
				m.speed += speed
			},
		}
	},
}

var classBonuses = map[Class][]func(level int, unit *Mob) Bonus{
	"Wizard": []func(level int, unit *Mob) Bonus{
		learnSpellBonus("Wizard"),
	},
	"Priest": []func(level int, unit *Mob) Bonus{
		learnSpellBonus("Priest"),
	},
}

func learnSpellBonus(class Class) func(int, *Mob) Bonus {
	spells := classSpells[class]
	return func(level int, unit *Mob) Bonus {
		perm := rand.Perm(len(spells))
	next:
		for _, i := range perm {
			spell := spells[i]
			if spell.level > level {
				continue
			}
			for _, learned := range unit.spells {
				if learned.Name == spell.spell.Name {
					continue next
				}
			}
			return Bonus{
				Name: "☆ " + spell.spell.Name,
				Apply: func(mob *Mob) {
					mob.spells = append(mob.spells, spell.spell)
				},
			}
		}
		mp := (level + 1) * 5
		return Bonus{
			Name: fmt.Sprintf("+%d MP", mp),
			Apply: func(mob *Mob) {
				mob.maxMP += mp
			},
		}
	}
}

type spellProgression struct {
	spell Weapon
	level int
}

var classSpells = map[Class][]spellProgression{
	"Wizard": []spellProgression{
		{
			spell: spellMeteor,
		},
	},
	"Priest": []spellProgression{
		{
			spell: spellSmite,
		},
		{
			spell: spellGloria,
		},
	},
}
