package main

import (
	"math/rand"
)

type Buff struct {
	Name       string
	Uniqueness Uniqueness
	BG         int // background color to apply to mob

	DoT Damage // damage or healing over time. HoTs are applied when turn starts, DoTs when turn ends

	// OnApply is called once, when this buff is applied to m.
	OnApply func(w *World, m *Mob, source *Mob)
	// OnRemove is called once, when this buff is removed (breaks, purged, etc).
	OnRemove func(w *World, m *Mob)
	// OnTakeTurn is called at the start of the affected mob's turn.
	OnTakeTurn func(w *World, m *Mob)
	// Affect is called every time m's stats can change.
	// Use it for temporarily modifying a unit's stats.
	Affect func(w *World, m *Mob, stats *Stats)

	BreakChance float64 // chance to break when unit starts turn: 0 = never, 0.1 = 10%
	Life        int     // turns until this buff will guaranteed break: -1 = infinite
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
