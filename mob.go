package main

import (
	// "log"
	"fmt"
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

	weapon *Weapon

	hp    int
	maxHP int
	mp    int
	maxMP int

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
	if m.Dead() {
		corpse := m.glyph
		corpse.Rune = '%'
		return corpse
	}
	return m.glyph
}

func (m *Mob) CT() int {
	return m.ct
}

func (m *Mob) TakeTurn(w *World) {
	m.moved = false
	m.acted = false

	// TODO: friendly AI
	if m.Team() != 0 && !w.gameOver {
		w.push <- &EnemyAIState{
			self: m,
		}
	}
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
	return !m.Dead()
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

func (m *Mob) CanAttack(other *Mob) bool {
	myloc := m.Loc()
	otherloc := other.Loc()
	if abs(myloc.X-otherloc.X)+abs(myloc.Y-otherloc.Y) <= m.Weapon().Range {
		return true
	}
	return false
}

func (m *Mob) Weapon() Weapon {
	if m.weapon != nil {
		return *m.weapon
	}
	return weaponFist
}

func (m *Mob) HP() int {
	return m.hp
}

func (m *Mob) MaxHP() int {
	return m.maxHP
}

func (m *Mob) MP() int {
	return m.mp
}

func (m *Mob) MaxMP() int {
	return m.maxMP
}

func (m *Mob) Dead() bool {
	return m.hp <= 0
}

func (m *Mob) CanAct() bool {
	return !m.Dead()
}

func (m *Mob) CanMove() bool {
	return !m.Dead()
}

func (m *Mob) Attackable() bool {
	return !m.Dead()
}

func (m *Mob) Damage(dmg int) int {
	m.hp -= dmg
	if m.hp <= 0 {
		m.loc.Z = 1
	}
	return dmg
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

func (mob *Mob) StatusLine() []Glyph {
	status := fmt.Sprintf("[ ] %s (HP: %d, MP: %d, Speed: %d, CT: %d)", mob.Name(), mob.HP(), mob.MP(), mob.Speed(), mob.CT())
	glyphs := GlyphsOf(status)
	glyphs[1] = mob.Glyph()
	return glyphs
}

var _ Ticker = &Mob{}
var _ Object = &Mob{}
var _ Turner = &Mob{}
