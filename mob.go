package main

import (
	// "log"
	"fmt"
	"math/rand"
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

	class  Class
	weapon Weapon
	spells []Weapon

	armor Armor

	hp    int
	maxHP int
	mp    int
	maxMP int

	moved bool
	acted bool

	tauntedBy *Mob
	crippled  bool
}

func (m *Mob) Reset() {
	m.hp = m.maxHP
	m.mp = m.maxMP
	m.ct = 0
	m.moved = false
	m.acted = false
	m.tauntedBy = nil
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
	loc := m.loc
	if m.Dead() {
		loc.Z = 1
	} else {
		loc.Z = 100
	}
	return loc
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
	glyph := m.glyph
	if m.HP() <= m.MaxHP()/4 {
		glyph.BG = ColorDarkRed
	} else if m.crippled {
		glyph.BG = 242
	} else if m.tauntedBy != nil {
		glyph.BG = 166
	}
	return glyph
}

func (m *Mob) CT() int {
	return m.ct
}

func (m *Mob) TakeTurn(w *World) {
	m.moved = false
	m.acted = false

	if m.tauntedBy != nil && m.tauntedBy.Dead() {
		m.tauntedBy = nil
	}
	if m.crippled && rand.Intn(6) == 0 {
		m.crippled = false
		w.Broadcast(m.Name() + " can move again.")
	}

	m.AddMP(m.Armor().MPRecovery + 1)

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
	if m.crippled {
		return 0
	}
	return m.move
}

func (m *Mob) CanAttack(world *World, other *Mob, weapon Weapon) bool {
	myloc := m.Loc()
	return m.CanAttackFrom(world, myloc, other, weapon)
}

func (m *Mob) CanAttackFrom(world *World, loc Loc, other *Mob, weapon Weapon) bool {
	myloc := loc
	otherloc := other.Loc()
	mymap := world.Map(myloc.Map)
	if abs(myloc.X-otherloc.X)+abs(myloc.Y-otherloc.Y) <= weapon.Range {
		_, blocked, _ := mymap.Raycast(myloc, otherloc, weapon.Magic)
		return !blocked
	}
	return false
}

func (m *Mob) Class() Class {
	if m.class == "" {
		return "Monster"
	}
	return m.class
}

func (m *Mob) Weapon() Weapon {
	if m.weapon.Damage == "" {
		return weaponFist
	}
	return m.weapon
}

func (m *Mob) Spells() []Weapon {
	return m.spells
}

func (m *Mob) Armor() Armor {
	return m.armor
}

func (m *Mob) HP() int {
	if m.hp < 0 {
		return 0
	}
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

func (m *Mob) AddMP(n int) {
	m.mp += n
	if m.mp > m.maxMP {
		m.mp = m.maxMP
	}
	if m.mp < 0 {
		m.mp = 0
	}
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

func (m *Mob) Damage(dmg int, src Weapon) int {
	if dmg > 0 && !src.Magic {
		dmg -= m.Armor().Defense
		if dmg <= 0 {
			dmg = 1
		}
	}

	m.hp -= dmg
	if m.hp > m.maxHP {
		m.hp = m.maxHP
	}
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
	mobname := mob.Name()
	if mob.class != "" {
		mobname += " the " + string(mob.Class())
	}
	status := fmt.Sprintf("[ ] %s (HP: %d, MP: %d, Speed: %d, CT: %d)", mobname, mob.HP(), mob.MP(), mob.Speed(), mob.CT())
	glyphs := GlyphsOf(status)
	glyphs[1] = mob.Glyph()
	return glyphs
}

var _ Ticker = &Mob{}
var _ Object = &Mob{}
var _ Turner = &Mob{}
