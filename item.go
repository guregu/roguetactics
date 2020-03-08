package main

import (
	"fmt"

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
		target.tauntedBy = source
		w.Broadcast(source.Name() + " taunted " + target.Name() + ".")
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
		if target.crippled {
			return
		}
		target.crippled = true
		w.Broadcast(target.Name() + " can no longer move!")
	},
	projectile: projectileFunc(Glyph{Rune: 'x', SGR: SGR{FG: ColorRed}}),
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
	Name       string
	Defense    int
	MPRecovery int
	Value      int
}

var armorLeather = Armor{
	Name:    "tunic",
	Defense: 1,
}
var armorLeather2 = Armor{
	Name:    "jerkin",
	Defense: 2,
	Value:   1,
}

var armorRobe = Armor{
	Name:       "robe",
	MPRecovery: 1,
	Defense:    0,
}
var armorFineRobe = Armor{
	Name:       "fine robe",
	MPRecovery: 3,
	Defense:    1,
	Value:      1,
}
var armorHolyRobe = Armor{
	Name:       "holy robe",
	MPRecovery: 5,
	Defense:    2,
	Value:      2,
}
var armorWizardHat = Armor{
	Name:       "pointy hat",
	MPRecovery: 8,
	Defense:    0,
	Value:      2,
}

var armorChainmail = Armor{
	Name:    "chainmail",
	Defense: 3,
	Value:   2,
}
var armorPlate = Armor{
	Name:    "platemail",
	Defense: 5,
	Value:   3,
}

func (a Armor) String() string {
	var info string
	if a.Defense != 0 {
		info = fmt.Sprintf("%dAC", a.Defense)
	}
	if a.MPRecovery > 0 {
		if info != "" {
			info += " "
		}
		info += fmt.Sprintf("%dMP/t", a.MPRecovery)
	}
	return fmt.Sprintf("%s (%s)", a.Name, info)
}
