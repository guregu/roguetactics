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
	Damage: "2d3+2",
	Range:  1,
}

var weaponFist = Weapon{
	Name:   "fist",
	Damage: "1d3",
	Range:  1,
}

var weaponBite = Weapon{
	Name:   "bite",
	Damage: "1d3+1",
	Range:  1,
}

var weaponScratch = Weapon{
	Name:   "scratch",
	Damage: "2d2",
	Range:  1,
}

var weaponLick = Weapon{
	Name:   "lick",
	Damage: "1d2",
	Range:  1,
}

var weaponPeck = Weapon{
	Name:   "peck",
	Damage: "1d3",
	Range:  1,
}

var weaponShank = Weapon{
	Name:   "shank",
	Damage: "1d5+1",
	Range:  1,
}

var weaponBow = Weapon{
	Name:       "bow",
	Damage:     "2d2+1",
	Range:      6,
	Targeting:  TargetingFree,
	projectile: projectileFunc(GlyphOf('*')),
}

var weaponStaff = Weapon{
	Name:   "staff",
	Damage: "1d4",
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

var spellMeteor = Weapon{
	Name:       "meteor",
	Damage:     "4d3",
	Range:      6,
	Targeting:  TargetingFree,
	Magic:      true,
	Hitbox:     HitboxBlob,
	HitboxSize: 4,
	HitGlyph:   &Glyph{Rune: 'X', SGR: SGR{BG: ColorBrightYellow, FG: ColorRed}},
	MPCost:     5,
	projectile: projectileFunc(Glyph{Rune: 'O', SGR: SGR{FG: ColorRed}}),
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

var spellSmite = Weapon{
	Name:      "smite",
	Damage:    "2d5+1",
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

func projectileFunc(g Glyph) func() Object {
	return func() Object {
		fx := &Effect{
			glyph: g,
			life:  -1,
		}
		return fx
	}
}

type Armor struct {
	Name    string
	Defense int
}

var armorLeather = Armor{
	Name:    "jerkin",
	Defense: 1,
}

var armorRobe = Armor{
	Name:    "robe",
	Defense: 0,
}
