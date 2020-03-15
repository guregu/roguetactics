package main

import (
	"math/rand"

	"github.com/guregu/dicey"
)

// TODO: make this its own type instead of using Weapon?

var spellFireball = Weapon{
	Name:       "fireball",
	Damage:     Damage{Dice: dicey.MustParse("2d3"), Type: DamageMagic},
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxCross,
	HitboxSize: 1,
	HitGlyph:   &Glyph{Rune: 'X', SGR: SGR{BG: ColorBrightYellow, FG: ColorRed}},
	MPCost:     5,
	projectile: projectileFunc(Glyph{Rune: 'o', SGR: SGR{FG: ColorRed}}),
}

var spellFireball2 = Weapon{
	Name:       "fireball ii",
	Damage:     Damage{Dice: dicey.MustParse("6d4"), Type: DamageMagic},
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxCross,
	HitboxSize: 1,
	HitGlyph:   &Glyph{Rune: 'X', SGR: SGR{BG: ColorBrightYellow, FG: ColorRed}},
	MPCost:     5,
	projectile: projectileFunc(Glyph{Rune: 'o', SGR: SGR{FG: ColorRed}}),
}

var spellMeteor = Weapon{
	Name:       "meteor",
	Damage:     Damage{Dice: dicey.MustParse("4d3"), Type: DamageMagic},
	Range:      7,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxBlob,
	HitboxSize: 3,
	HitGlyph:   &Glyph{Rune: 'X', SGR: SGR{BG: ColorBrightYellow, FG: ColorRed}},
	MPCost:     8,
	projectile: projectileFunc(Glyph{Rune: 'O', SGR: SGR{FG: ColorRed}}),
}

var spellBolt = Weapon{
	Name:       "bolt",
	Damage:     Damage{Dice: dicey.MustParse("5d5+2"), Type: DamageMagic},
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxSingle,
	HitboxSize: 1,
	HitGlyph:   &Glyph{Rune: 'X', SGR: SGR{BG: ColorBrightYellow, FG: ColorRed}},
	MPCost:     10,
}

var spellHeal = Weapon{
	Name: "heal",
	Damage: Damage{
		Dice: dicey.MustParse("3d5+5"),
		Type: DamageHealing,
	},
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxCross,
	HitboxSize: 1,
	HitGlyph:   &Glyph{Rune: '✳', SGR: SGR{FG: ColorBrightGreen}},
	MPCost:     5,
}

var spellHeal2 = Weapon{
	Name: "heal ii",
	Damage: Damage{
		Dice: dicey.MustParse("4d6+8"),
		Type: DamageHealing,
	},
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxCross,
	HitboxSize: 1,
	HitGlyph:   &Glyph{Rune: '✳', SGR: SGR{FG: ColorBrightGreen}},
	MPCost:     8,
}

var spellSmite = Weapon{
	Name:      "smite",
	Damage:    Damage{Dice: dicey.MustParse("2d10+1"), Type: DamageMagic},
	Range:     5,
	Targeting: TargetingFree,
	Magic:     true,
	Hitbox:    HitboxSingle,
	HitGlyph:  &Glyph{Rune: '✞', SGR: SGR{FG: ColorBrightYellow}},
	MPCost:    5,
}

var spellSmite2 = Weapon{
	Name:      "smite ii",
	Damage:    Damage{Dice: dicey.MustParse("3d10+2"), Type: DamageMagic},
	Range:     5,
	Targeting: TargetingFree,
	Magic:     true,
	Hitbox:    HitboxSingle,
	HitGlyph:  &Glyph{Rune: '✞', SGR: SGR{FG: ColorBrightYellow}},
	MPCost:    5,
}

var spellGloria = Weapon{
	Name: "gloria",
	Damage: Damage{
		Dice: dicey.MustParse("2d5+5"),
		Type: DamageHealing,
	},
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxBlob,
	HitboxSize: 3,
	HitGlyph:   &Glyph{Rune: '✚', SGR: SGR{FG: ColorBrightGreen}},
	MPCost:     10,
}

var spellRenew = Weapon{
	Name:      "renew",
	Damage:    Damage{Type: DamageNone},
	Range:     6,
	Targeting: TargetingFree,
	Magic:     true,
	MPCost:    5,
	Hitbox:    HitboxSingle,
	HitGlyph:  &Glyph{Rune: '✚', SGR: SGR{FG: ColorBrightGreen}},
	OnHit: func(w *World, source *Mob, target *Mob) {
		life := rand.Intn(3) + 4
		buff := newBuff("renew", NotUnique, life, 0)
		buff.BG = ColorDarkGreen
		buff.DoT = Damage{
			Dice: dicey.MustParse("1d4+1"),
			Type: DamageHealing,
		}
		buff.OnApply = func(w *World, m *Mob, src *Mob) {
			// w.Broadcast(m.Name() + " is ")
		}
		buff.OnRemove = func(w *World, m *Mob) {
			// w.Broadcast(m.Name() + " ")
		}
		target.ApplyBuff(w, buff, source)
	},
}

var spellTaunt = Weapon{
	Name:      "taunt",
	Damage:    Damage{Type: DamageNone},
	Range:     6,
	Targeting: TargetingFree,
	Magic:     true,
	Hitbox:    HitboxSingle,
	HitGlyph:  &Glyph{Rune: '!', SGR: SGR{FG: ColorDarkRed}},
	OnHit: func(w *World, source *Mob, target *Mob) {
		if source.Team() == target.Team() {
			w.Broadcast(
				source.NameColored(),
				" tries to taunt ",
				target.NameColored(),
				", but they laugh instead.",
			)
			return
		}
		buff := newBuff("taunt", UniqueReplace, -1, 0)
		buff.BG = Color256(166)
		buff.OnApply = func(w *World, m *Mob, src *Mob) {
			m.tauntedBy = src
			w.Broadcast(
				src.NameColored(),
				" taunted ",
				m.NameColored(),
				".",
			)
		}
		buff.OnTakeTurn = func(w *World, m *Mob) {
			if m.tauntedBy != nil && m.tauntedBy.Dead() {
				buff.Life = 0
			}
		}
		buff.OnRemove = func(w *World, m *Mob) {
			m.tauntedBy = nil
		}
		target.ApplyBuff(w, buff, source)
	},
}

var spellCripple = Weapon{
	Name:      "aim: legs",
	Damage:    Damage{Type: DamageNone},
	Range:     6,
	Targeting: TargetingFree,
	// Magic:      true,
	MPCost: 5,
	Hitbox: HitboxSingle,
	// HitGlyph:   &Glyph{Rune: 'x', SGR: SGR{FG: ColorDarkRed}},
	OnHit: func(w *World, source *Mob, target *Mob) {
		life := rand.Intn(6) + 2
		buff := newBuff("crippled", Unique, life, 0.1)
		buff.BG = Color256(237)
		buff.OnApply = func(w *World, m *Mob, src *Mob) {
			w.Broadcast(
				m.NameColored(),
				" can no longer move!",
			)
		}
		buff.Affect = func(w *World, m *Mob, stats *Stats) {
			stats.Crippled = true
		}
		buff.OnRemove = func(w *World, m *Mob) {
			w.Broadcast(
				m.NameColored(),
				" can move again!",
			)
		}
		target.ApplyBuff(w, buff, source)
	},
	projectile: projectileFunc(Glyph{Rune: 'x', SGR: SGR{FG: ColorRed}}),
}

var spellPoisonShot = Weapon{
	Name:      "poison shot",
	Damage:    Damage{Type: DamageNone},
	Range:     6,
	Targeting: TargetingFree,
	// Magic:      true,
	MPCost: 5,
	Hitbox: HitboxSingle,
	// HitGlyph:   &Glyph{Rune: 'x', SGR: SGR{FG: ColorDarkRed}},
	OnHit: func(w *World, source *Mob, target *Mob) {
		life := rand.Intn(3) + 4
		buff := newBuff("poison", NotUnique, life, 0.1)
		buff.BG = ColorDiarrhea
		buff.DoT = Damage{
			Dice: dicey.MustParse("1d4+1"),
		}
		buff.OnApply = func(w *World, m *Mob, src *Mob) {
			w.Broadcast(
				m.NameColored(),
				" is ",
				GlyphsOf("poisoned", StyleFG(ColorDiarrhea)),
				"!",
			)
		}
		buff.OnRemove = func(w *World, m *Mob) {
			// w.Broadcast(m.Name() + " ")
		}
		target.ApplyBuff(w, buff, source)
	},
	projectile: projectileFunc(Glyph{Rune: '*', SGR: SGR{FG: ColorDiarrhea}}),
}
