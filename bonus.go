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
		hp := (level + 2) * 3
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
				m.base.Speed += speed
			},
		}
	},
}

var classBonuses = map[Class][]func(level int, unit *Mob) Bonus{
	"Wizard": []func(level int, unit *Mob) Bonus{
		learnSpellBonus("Wizard", true),
		itemBonus("Wizard", true, true),
	},
	"Priest": []func(level int, unit *Mob) Bonus{
		learnSpellBonus("Priest", true),
		itemBonus("Priest", true, true),
	},
	"Knight": []func(level int, unit *Mob) Bonus{
		itemBonus("Knight", false, false),
	},
	"Archer": []func(level int, unit *Mob) Bonus{
		itemBonus("Archer", true, false),
	},
}

func learnSpellBonus(class Class, magicUser bool) func(int, *Mob) Bonus {
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
		mp := (level + 2) * 3
		if magicUser {
			return Bonus{
				Name: fmt.Sprintf("+%d MP", mp),
				Apply: func(mob *Mob) {
					mob.maxMP += mp
				},
			}
		} else {
			return Bonus{
				Name: fmt.Sprintf("+%d HP", mp),
				Apply: func(mob *Mob) {
					mob.maxHP += mp
				},
			}
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
		{
			spell: spellBolt,
			level: 2,
		},
	},
	"Priest": []spellProgression{
		{
			spell: spellSmite,
		},
		{
			spell: spellGloria,
		},
		{
			spell: spellHeal2,
		},
		{
			spell: spellSmite2,
			level: 3,
		},
	},
}

func itemBonus(class Class, magicUser, reroll bool) func(int, *Mob) Bonus {
	spells := classItems[class]
	return func(level int, unit *Mob) Bonus {
		if magicUser && reroll && rand.Float64() < 0.5 {
			return learnSpellBonus(class, magicUser)(level, unit)
		}
		perm := rand.Perm(len(spells))
		for _, i := range perm {
			spell := spells[i]
			if spell.level > level {
				continue
			}
			if spell.weapon != nil && (unit.Weapon().Value > spell.weapon.Value || spell.weapon.Name == unit.Weapon().Name) {
				continue
			}
			if spell.armor != nil && (unit.Armor().Value > spell.armor.Value || spell.armor.Name == unit.Armor().Name) {
				continue
			}

			if spell.weapon != nil {
				return Bonus{
					Name: fmt.Sprintf("%s (%s)", spell.weapon.Name, spell.weapon.Damage),
					Apply: func(mob *Mob) {
						mob.weapon = *spell.weapon
					},
				}
			}
			if spell.armor != nil {
				return Bonus{
					Name: spell.armor.String(),
					Apply: func(mob *Mob) {
						mob.armor = *spell.armor
					},
				}
			}
		}
		if magicUser {
			return learnSpellBonus(class, magicUser)(level, unit)
		}
		mp := (level + 2) * 3
		if magicUser {
			return Bonus{
				Name: fmt.Sprintf("+%d MP", mp),
				Apply: func(mob *Mob) {
					mob.maxMP += mp
				},
			}
		} else {
			return Bonus{
				Name: fmt.Sprintf("+%d HP", mp),
				Apply: func(mob *Mob) {
					mob.maxHP += mp
				},
			}
		}
	}
}

type itemProgression struct {
	weapon *Weapon
	armor  *Armor
	level  int
}

var classItems = map[Class][]itemProgression{
	"Knight": []itemProgression{
		{
			armor: &armorChainmail,
		},
		{
			weapon: &weaponSword,
		},
		{
			armor: &armorPlate,
			level: 2,
		},
		{
			weapon: &weaponGreatsword,
			level:  2,
		},
	},
	"Archer": []itemProgression{
		{
			weapon: &weaponLongbow,
		},
		{
			armor: &armorLeather2,
		},
		{
			armor: &armorChainmail,
			level: 2,
		},
		{
			weapon: &weaponCrossbow,
			level:  2,
		},
	},
	"Priest": []itemProgression{
		{
			weapon: &weaponHealingStaff,
		},
		{
			weapon: &weaponBeatstick,
		},
		{
			armor: &armorFineRobe,
		},
		{
			armor: &armorHolyRobe,
			level: 2,
		},
	},
	"Wizard": []itemProgression{
		{
			armor: &armorFineRobe,
		},
		{
			weapon: &weaponBeatstick,
		},
		{
			armor: &armorWizardHat,
			level: 2,
		},
	},
}
