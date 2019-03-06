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

type Mob struct {
	id      ID
	name    string
	loc     Loc
	glyph   Glyph
	actions []func(*Mob, *World)
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

func (m *Mob) Glyph() Glyph {
	return m.glyph
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

func (m *Mob) Tick(w *World, tick int64) {
	if len(m.actions) > 0 {
		m.actions[0](m, w)
		m.actions = m.actions[1:]
	}
}

func (m *Mob) Enqueue(action func(*Mob, *World)) {
	m.actions = append(m.actions, action)
}

var _ Ticker = &Mob{}
var _ Object = &Mob{}
