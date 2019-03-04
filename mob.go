package main

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
	Tick(*World, int64)
}

type Collider interface {
	Collides(*World, ID) bool
	OnCollide(*World, ID) bool
}

type Mob struct {
	id    ID
	name  string
	loc   Loc
	glyph Glyph
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
