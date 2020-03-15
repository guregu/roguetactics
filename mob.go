package main

import (
	// "log"
	"fmt"
	"sort"
)

type Loc struct {
	Map     string
	X, Y, Z int
}

func (loc Loc) AsCoords() Coords {
	return Coords{loc.X, loc.Y}
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

	ct int

	class  Class
	weapon Weapon
	spells []Weapon
	armor  Armor

	hp    int
	maxHP int
	mp    int
	maxMP int

	base  Stats
	stats Stats
	bgIdx int // index of BG color to show

	moved bool
	acted bool

	tauntedBy *Mob
	buffs     map[*Buff]struct{}
}

type Stats struct {
	Move         int    // move range
	Speed        int    // how much to increment CT
	Defense      int    // physical defense
	MagicDefense int    // magical defense
	Crippled     bool   // can't move
	BGs          Colors // glyph BGs to cycle through
}

func (m *Mob) Reset(w *World) {
	m.hp = m.maxHP
	m.mp = m.maxMP
	m.ct = 0
	m.moved = false
	m.acted = false
	m.tauntedBy = nil
	m.buffs = make(map[*Buff]struct{})
	m.refreshStats(w)
	m.bgIdx = 0
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
	if len(m.stats.BGs) > 0 && m.bgIdx < len(m.stats.BGs) {
		glyph.BG = m.stats.BGs[m.bgIdx]
	}
	return glyph
}

func (m *Mob) CT() int {
	return m.ct
}

func (m *Mob) TakeTurn(w *World) {
	m.moved = false
	m.acted = false

	for buff := range m.buffs {
		buff.TurnTick()
		if buff.OnTakeTurn != nil {
			buff.OnTakeTurn(w, m)
		}
		if buff.Broken() {
			delete(m.buffs, buff)
			if buff.OnRemove != nil {
				buff.OnRemove(w, m)
			}
			continue
		}
		if buff.DoT.IsValid() && buff.DoT.Type == DamageHealing {
			dmg := m.Damage(w, buff.DoT)
			if dmg != 0 {
				w.Broadcast(fmt.Sprintf("%s was healed by %s for %d.", m.Name(), buff.Name, -dmg))
			}
		}
	}
	m.refreshStats(w)

	// TODO: move this to buff system somehow
	if m.tauntedBy != nil && m.tauntedBy.Dead() {
		m.tauntedBy = nil
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
	return m.stats.Speed
}

func (m *Mob) FinishTurn(w *World, moved, acted bool) {
	// apply DoTs
	dead := m.Dead()
	if !dead {
		for buff := range m.buffs {
			if buff.DoT.IsValid() && buff.DoT.Type != DamageHealing {
				dmg := m.Damage(w, buff.DoT)
				if dmg != 0 {
					w.Broadcast(fmt.Sprintf("%s was damaged by %s for %d.", m.Name(), buff.Name, dmg))
				}
			}
		}
	}
	if !dead && m.Dead() {
		w.Broadcast(fmt.Sprintf("%s died.", m.Name()))
	}

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
	if m.stats.Crippled {
		return 0
	}
	return m.stats.Move
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
	if !m.weapon.Damage.IsValid() {
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

func (m *Mob) Defense() int {
	return m.Armor().Defense + m.stats.Defense
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

func (m *Mob) Damage(w *World, dmg Damage) int {
	if m.Dead() {
		return 0
	}

	def := m.Defense()
	if dmg.Type == DamageMagic {
		def = m.stats.MagicDefense
	}
	hit := dmg.Dice.Roll()
	if dmg.Type == DamageHealing {
		hit = -hit
	}
	if hit > 0 {
		hit -= def
		if hit <= 0 {
			hit = 1
		}
	}

	m.hp -= hit
	if m.hp > m.maxHP {
		m.hp = m.maxHP
	}
	if m.hp <= 0 {
		m.loc.Z = 1
	}

	m.refreshStats(w)
	return hit
}

func (m *Mob) ApplyBuff(w *World, buff *Buff, src *Mob) {
	if buff.Unique() {
		for existing := range m.buffs {
			if buff.Name == existing.Name {
				switch buff.Uniqueness {
				case Unique:
					w.Broadcast("It wasn't effective.")
					return
				case UniqueReplace:
					delete(m.buffs, existing)
					if existing.OnRemove != nil {
						existing.OnRemove(w, m)
					}
				}
			}
		}
	}

	m.buffs[buff] = struct{}{}
	if buff.OnApply != nil {
		buff.OnApply(w, m, src)
	}
	m.refreshStats(w)
}

func (m *Mob) refreshStats(w *World) {
	stats := m.base
	if stats.BGs != nil {
		stats.BGs = stats.BGs[:0]
	}
	for buff := range m.buffs {
		if buff.Affect != nil {
			buff.Affect(w, m, &stats)
		}
		if buff.BG != nil {
			stats.BGs = append(stats.BGs, buff.BG)
		}
	}
	sort.Sort(stats.BGs)
	// do this after sorting so that players get immediate feedback for critical damage
	if m.HP() <= m.MaxHP()/4 {
		stats.BGs = append([]Color{ColorDarkRed}, stats.BGs...)
		m.bgIdx = 0
	}
	m.stats = stats
}

func (m *Mob) Tick(w *World, tick int64) {
	if len(m.actions) > 0 {
		m.actions[0](m, w)
		m.actions = m.actions[1:]
	}
	if tick%25 == 0 && len(m.stats.BGs) > 0 {
		m.bgIdx = (m.bgIdx + 1) % len(m.stats.BGs)
	}
}

func (m *Mob) Enqueue(action func(*Mob, *World)) {
	m.actions = append(m.actions, action)
}

func (mob *Mob) StatusLine(short bool) []Glyph {
	mobname := mob.Name()
	if mob.class != "" {
		if short {
			mobname += ", " + string(mob.Class())
		} else {
			mobname += " the " + string(mob.Class())
		}
	}
	mp := fmt.Sprintf(", MP: %d/%d", mob.MP(), mob.MaxMP())
	if mob.MaxMP() == 0 {
		mp = ""
	}
	speed := "Speed"
	if short {
		speed = "S"
	}
	status := fmt.Sprintf("[ ] %s (HP: %d/%d%s, %s: %d, CT: %d)", mobname, mob.HP(), mob.MaxHP(), mp, speed, mob.Speed(), mob.CT())
	glyphs := GlyphsOf(status)
	glyphs[1] = mob.Glyph()
	return glyphs
}

var _ Ticker = &Mob{}
var _ Object = &Mob{}
var _ Turner = &Mob{}
