package main

import (
	"math/rand"
)

// TODO: make this its own type instead of using Weapon?

var spellFireball = Weapon{
	Name:       "fireball",
	Damage:     "2d3",
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
	Damage:     "6d4",
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
	Damage:     "4d3",
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
	Damage:     "5d5+2",
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxSingle,
	HitboxSize: 1,
	HitGlyph:   &Glyph{Rune: 'X', SGR: SGR{BG: ColorBrightYellow, FG: ColorRed}},
	MPCost:     10,
}

var spellHeal = Weapon{
	Name:       "heal",
	Damage:     "3d5+5",
	DamageType: DamageHealing,
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxCross,
	HitboxSize: 1,
	HitGlyph:   &Glyph{Rune: '✳', SGR: SGR{FG: ColorBrightGreen}},
	MPCost:     5,
}

var spellHeal2 = Weapon{
	Name:       "heal ii",
	Damage:     "4d6+8",
	DamageType: DamageHealing,
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
	Damage:    "2d10+1",
	Range:     5,
	Targeting: TargetingFree,
	Magic:     true,
	Hitbox:    HitboxSingle,
	HitGlyph:  &Glyph{Rune: '✞', SGR: SGR{FG: ColorBrightYellow}},
	MPCost:    5,
}

var spellSmite2 = Weapon{
	Name:      "smite ii",
	Damage:    "3d10+2",
	Range:     5,
	Targeting: TargetingFree,
	Magic:     true,
	Hitbox:    HitboxSingle,
	HitGlyph:  &Glyph{Rune: '✞', SGR: SGR{FG: ColorBrightYellow}},
	MPCost:    5,
}

var spellGloria = Weapon{
	Name:       "gloria",
	Damage:     "2d5+5",
	DamageType: DamageHealing,
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxBlob,
	HitboxSize: 3,
	HitGlyph:   &Glyph{Rune: '✚', SGR: SGR{FG: ColorBrightGreen}},
	MPCost:     10,
}

var spellTaunt = Weapon{
	Name:       "taunt",
	DamageType: DamageNone,
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxSingle,
	HitGlyph:   &Glyph{Rune: '!', SGR: SGR{FG: ColorDarkRed}},
	OnHit: func(w *World, source *Mob, target *Mob) {
		if source.Team() == target.Team() {
			w.Broadcast(source.Name() + " tries to taunt " + target.Name() + ", but they laugh instead.")
			return
		}
		buff := newBuff("taunt", UniqueReplace, -1, 0)
		buff.BG = 166
		buff.Apply = func(w *World, m *Mob, src *Mob) {
			m.tauntedBy = src
			w.Broadcast(src.Name() + " taunted " + m.Name() + ".")
		}
		target.ApplyBuff(w, buff, source)
	},
}

var spellCripple = Weapon{
	Name:       "aim: legs",
	DamageType: DamageNone,
	Range:      6,
	Targeting:  TargetingFree,
	// Magic:      true,
	MPCost: 5,
	Hitbox: HitboxSingle,
	// HitGlyph:   &Glyph{Rune: 'x', SGR: SGR{FG: ColorDarkRed}},
	OnHit: func(w *World, source *Mob, target *Mob) {
		life := rand.Intn(6) + 2
		buff := newBuff("crippled", Unique, life, 0.1)
		buff.BG = 237
		buff.Apply = func(w *World, m *Mob, src *Mob) {
			w.Broadcast(m.Name() + " can no longer move!")
		}
		buff.Affect = func(w *World, m *Mob, stats *Stats) {
			stats.Crippled = true
		}
		buff.Remove = func(w *World, m *Mob) {
			w.Broadcast(m.Name() + " can move again!")
		}
		target.ApplyBuff(w, buff, source)
	},
	projectile: projectileFunc(Glyph{Rune: 'x', SGR: SGR{FG: ColorRed}}),
}
