package main

import (
	"github.com/justinian/dice"
)

type Weapon struct {
	Name   string
	Damage string
	Range  int
}

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
