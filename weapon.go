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
	Value      int

	// spells
	Magic      bool // ignore walls etc
	Hitbox     HitboxType
	HitboxSize int
	MPCost     int
	HitGlyph   *Glyph

	OnHit      func(w *World, caster *Mob, target *Mob)
	projectile func() Object
}

type DamageType int

const (
	DamageNormal DamageType = iota
	DamageHealing
	DamageNone
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
	if w.Damage == "" {
		return 0
	}
	res, _, _ := dice.Roll(w.Damage)
	return res.Int()
}

var weaponShortsword = Weapon{
	Name:   "shortsword",
	Damage: "2d3+2",
	Range:  1,
}

var weaponSword = Weapon{
	Name:   "sword",
	Damage: "2d4+3",
	Range:  1,
	Value:  1,
}

var weaponGreatsword = Weapon{
	Name:   "greatsword",
	Damage: "3d5+4",
	Range:  1,
	Value:  2,
}

var weaponFist = Weapon{
	Name:   "fist",
	Damage: "1d3",
	Range:  1,
}

var weaponYetiFist = Weapon{
	Name:   "yetifist",
	Damage: "3d3",
	Range:  1,
}

var weaponBite = Weapon{
	Name:   "bite",
	Damage: "1d3+1",
	Range:  1,
}

var weaponSnowFoxBite = Weapon{
	Name:   "snow fox bite",
	Damage: "2d3+4",
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

var weaponSwipe = Weapon{
	Name:   "swipe",
	Damage: "3d4+1",
	Range:  1,
}

var weaponShank = Weapon{
	Name:   "shank",
	Damage: "1d5+1",
	Range:  1,
}

var weaponFirebreathing = Weapon{
	Name:       "firebreathing",
	Damage:     "2d8+4",
	Range:      3,
	projectile: projectileFunc(Glyph{Rune: '#', SGR: SGR{FG: ColorBrightRed}}),
}

var weaponPick = Weapon{
	Name:   "mattock",
	Damage: "2d6+3",
	Range:  1,
}

var weaponSpear = Weapon{
	Name:   "spear",
	Damage: "2d8",
	Range:  2,
}

var weaponKick = Weapon{
	Name:   "kick",
	Damage: "3d6",
}

var weaponCrush = Weapon{
	Name:   "kick",
	Damage: "2d10",
}

var weaponBow = Weapon{
	Name:       "bow",
	Damage:     "2d2+1",
	Range:      6,
	Targeting:  TargetingFree,
	projectile: projectileFunc(GlyphOf('*')),
}

var weaponLongbow = Weapon{
	Name:       "longbow",
	Damage:     "2d5+2",
	Range:      7,
	Targeting:  TargetingFree,
	projectile: projectileFunc(GlyphOf('*')),
	Value:      1,
}

var weaponCrossbow = Weapon{
	Name:       "crossbow",
	Damage:     "3d4+4",
	Range:      6,
	Targeting:  TargetingFree,
	projectile: projectileFunc(GlyphOf('*')),
	Value:      2,
}

var weaponStaff = Weapon{
	Name:   "staff",
	Damage: "1d4",
	Range:  2,
}

var weaponBeatstick = Weapon{
	Name:   "beatstick",
	Damage: "2d10",
	Range:  2,
}

var weaponHealingStaff = Weapon{
	Name:       "healing rod",
	Damage:     "3d8",
	DamageType: DamageHealing,
	Range:      2,
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
