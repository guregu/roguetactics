package main

import (
	"github.com/justinian/dice"
)

type Weapon struct {
	Name       string
	Damage     string
	Range      int
	Targeting  TargetingType
	DamageType DamageType

	// spells
	Magic      bool // ignore walls etc
	Hitbox     HitboxType
	HitboxSize int
	MPCost     int
	HitGlyph   *Glyph

	projectile func() Object
}

type DamageType int

const (
	DamageNormal DamageType = iota
	DamageHealing
)

type TargetingType int

const (
	TargetingCross TargetingType = iota
	TargetingFree
)

type HitboxType int

const (
	HitboxSingle HitboxType = iota
	HitboxCross
	HitboxBlob
)

func (w Weapon) RollDamage() int {
	res, _, _ := dice.Roll(w.Damage)
	return res.Int()
}

var weaponSword = Weapon{
	Name:   "sword",
	Damage: "2d6+5",
	Range:  1,
}

var weaponFist = Weapon{
	Name:   "fist",
	Damage: "1d3",
	Range:  1,
}

var weaponBite = Weapon{
	Name:   "bite",
	Damage: "1d2+1",
	Range:  1,
}

var weaponShank = Weapon{
	Name:   "shank",
	Damage: "1d5+1",
	Range:  1,
}

var weaponBow = Weapon{
	Name:       "bow",
	Damage:     "1d4",
	Range:      6,
	Targeting:  TargetingFree,
	projectile: projectileFunc(GlyphOf('*')),
}

var weaponStaff = Weapon{
	Name:   "staff",
	Damage: "1d2+1",
	Range:  2,
}

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

var spellHeal = Weapon{
	Name:       "heal",
	Damage:     "5d5",
	DamageType: DamageHealing,
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxCross,
	HitboxSize: 1,
	HitGlyph:   &Glyph{Rune: '✳', SGR: SGR{FG: ColorBrightGreen}},
	MPCost:     5,
}

func projectileFunc(g Glyph) func() Object {
	return func() Object {
		fx := &Effect{
			glyph: g,
			life:  -1,
		}
		return fx
	}
}
