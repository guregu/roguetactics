package main

import (
	"math/rand"
)

type Buff struct {
	Name   string
	Unique bool
	Apply  func(w *World, m *Mob, source *Mob)
	Remove func(w *World, m *Mob)

	BreakChance float64
	Life        int // -1 = infinite
}

func newBuff(name string, unique bool, life int, breakChance float64) *Buff {
	return &Buff{
		Name:        name,
		Unique:      unique,
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
