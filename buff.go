package main

import (
	"math/rand"
)

type Buff struct {
	Name       string
	Uniqueness Uniqueness
	BG         int // background color to apply to mob
	Apply      func(w *World, m *Mob, source *Mob)
	Remove     func(w *World, m *Mob)
	Affect     func(w *World, m *Mob, stats *Stats)

	BreakChance float64
	Life        int // -1 = infinite
}

type Uniqueness int

const (
	NotUnique     Uniqueness = iota // multiple buffs of same kind can overlap
	Unique                          // only first buff of kind sticks
	UniqueReplace                   // newer buffs replace older ones
)

func newBuff(name string, unique Uniqueness, life int, breakChance float64) *Buff {
	return &Buff{
		Name:        name,
		Uniqueness:  unique,
		Life:        life,
		BreakChance: breakChance,
	}
}

func (b *Buff) TurnTick() {
	if b.BreakChance != 0 {
		if rand.Float64() <= b.BreakChance {
			b.Life = 0
		}
	}
	if b.Life <= 0 {
		return
	}
	b.Life--
}

func (b *Buff) Broken() bool {
	return b.Life == 0
}

func (b *Buff) Unique() bool {
	return b.Uniqueness == Unique || b.Uniqueness == UniqueReplace
}
