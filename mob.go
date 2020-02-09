package main

import (
// "log"
)

type Loc struct {
	Map     string
	X, Y, Z int
}

type Pos struct {
	X, Y, Z int
}

type Object interface {
	Create(*World)
	ID() ID
	Name() string
	Glyph() Glyph
	Loc() Loc
	Move(Loc)
}

type Ticker interface {
	Object
	Tick(*World, int64)
}

type Collider interface {
	Object
	Collides(*World, ID) bool
	OnCollide(*World, ID) bool
}

type Turner interface {
	Object
	CT() int
	Speed() int
	TakeTurn(*World)
	TurnTick(*World)
}

type Mob struct {
	id      ID
	name    string
	team    int
	loc     Loc
	glyph   Glyph
	actions []func(*Mob, *World)

	ct    int
	speed int
	move  int // move range

	HP    int
	MaxHP int
	MP    int
	MaxMP int

	moved bool
	acted bool
}

func (m *Mob) Create(w *World) {
	m.id = w.NextID()
	if m.loc.Map != "" {
		w.Map(m.loc.Map).Add(m)
	}
}

func (m *Mob) ID() ID {
	return m.id
}

func (m *Mob) Loc() Loc {
	return m.loc
}

func (m *Mob) Name() string {
	return m.name
}

func (m *Mob) Team() int {
	return m.team
}

func (m *Mob) Glyph() Glyph {
	return m.glyph
}

func (m *Mob) CT() int {
	return m.ct
}

func (m *Mob) TakeTurn(_ *World) {
	m.moved = false
	m.acted = false
}

func (m *Mob) TurnTick(*World) {
	m.ct += m.Speed()
}

func (m *Mob) Speed() int {
	return m.speed
}

func (m *Mob) FinishTurn(moved, acted bool) {
	if moved && acted {
		m.ct -= 100
		return
	}
	if moved || acted {
		m.ct -= 80
		return
	}
	m.ct -= 60
	if m.ct > 60 {
		m.ct = 60
	}
}

func (m *Mob) Collides(_ *World, _ ID) bool {
	return true
}

func (m *Mob) OnCollide(_ *World, _ ID) bool {
	return true
}

func (m *Mob) Move(loc Loc) {
	m.loc = loc
}

func (m *Mob) MoveRange() int {
	return m.move
}

func (m *Mob) Tick(w *World, tick int64) {
	if len(m.actions) > 0 {
		m.actions[0](m, w)
		m.actions = m.actions[1:]
	}
}

func (m *Mob) Enqueue(action func(*Mob, *World)) {
	m.actions = append(m.actions, action)
}

const (
	PlayerTeam = 0
	AITeam     = 1
)

type Team struct {
	ID    int
	Units []*Mob
}

var _ Ticker = &Mob{}
var _ Object = &Mob{}
var _ Turner = &Mob{}
