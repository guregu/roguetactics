package main

import (
	"fmt"
	"io"
	"log"
	"sort"
	"sync/atomic"
	"time"
)

const ctForTurn = 100
const tickTime = time.Millisecond * 40

type ID int64

type World struct {
	maps     map[string]*Map
	objects  map[ID]Object
	seshes   map[*Sesh]struct{}
	waitlist []Turner

	// battle state
	lastID ID
	tick   int64
	turn   int64
	up     Turner
	state  []StateAction
	busy   *int32

	// overall game state
	player    Team
	current   *Map
	gameOver  bool
	battleWon bool
	level     int
	battle    Battle
	score     int

	apply      chan Action
	applySync  chan Action // this exists so the shutdown hook is guaranteed to run
	push       chan StateAction
	pushBottom chan StateAction
}

type Action interface {
	Apply(*World)
}

type StateAction interface {
	Run(*World) bool
}

func newWorld() *World {
	w := &World{
		maps:    make(map[string]*Map),
		objects: make(map[ID]Object),
		seshes:  make(map[*Sesh]struct{}),

		busy: new(int32),

		player: generatePlayerTeam(),

		apply:      make(chan Action, 32),
		applySync:  make(chan Action),
		push:       make(chan StateAction, 32),
		pushBottom: make(chan StateAction, 32),
	}
	mapnames := map[string]struct{}{}
	for _, maps := range mapsByLevel {
		for _, m := range maps {
			mapnames[m] = struct{}{}
		}
	}
	n := 1
	consoleWrite("\n\r")
	for name := range mapnames {
		m, err := loadMap(name)
		if err != nil {
			panic(err)
		}
		w.maps[name] = m
		consoleWrite(fmt.Sprintf("Downloaded map: %d/%d\n\r", n, len(mapnames)))
		n++
	}
	return w
}

func (w *World) reset() {
	// log.Println("Resetting world...")
	w.state = nil
	w.up = nil
	w.gameOver = false
	w.score = 0
	w.objects = make(map[ID]Object)
	w.player = generatePlayerTeam()
	w.current = nil
	w.waitlist = nil
	w.tick = 0
	w.turn = 0
	for _, m := range w.maps {
		m.Reset()
	}
	atomic.StoreInt32(w.busy, 0)

	for sesh := range w.seshes {
		if sesh.win != nil {
			sesh.win.close()
			// sesh.win = nil
		}
	}
}

func (w *World) Map(name string) *Map {
	return w.maps[name]
}

func (w *World) NextID() ID {
	w.lastID++
	return w.lastID
}

func (w *World) Create(obj Object) {
	obj.Create(w)
	w.objects[obj.ID()] = obj
	if t, ok := obj.(Turner); ok {
		w.waitlist = append(w.waitlist, t)
		w.sortWaitlist()
	}
}

func (w *World) Delete(id ID) {
	obj, ok := w.objects[id]
	if !ok {
		return
	}

	loc := obj.Loc()
	if loc.Map != "" {
		w.Map(loc.Map).Remove(obj)
	}
	delete(w.objects, obj.ID())

	if t, ok := obj.(Turner); ok {
		for i := len(w.waitlist) - 1; i >= 0; i-- {
			if w.waitlist[i] != t {
				continue
			}
			if i < len(w.waitlist)-1 {
				copy(w.waitlist[i:], w.waitlist[i+1:])
			}
			w.waitlist[len(w.waitlist)-1] = nil
			w.waitlist = w.waitlist[:len(w.waitlist)-1]
		}
	}
}

func (w *World) Tick() {
	w.tick++
	for _, obj := range w.objects {
		if ticker, ok := obj.(Ticker); ok {
			ticker.Tick(w, w.tick)
		}
	}
}

func (w *World) Run() {
	ticker := time.NewTicker(tickTime)
	defer ticker.Stop()
	for {
		select {
		case a := <-w.apply:
			a.Apply(w)
			w.notify()
		case a := <-w.applySync:
			a.Apply(w)
			w.notify()
		case a := <-w.push:
			w.state = append(w.state, a)
		case a := <-w.pushBottom:
			w.state = append([]StateAction{a}, w.state...)
		case <-ticker.C:
			if len(w.state) > 0 {
				state := w.state[len(w.state)-1]
				if state.Run(w) {
					w.state = w.state[:len(w.state)-1]
				}
				busy := int32(0)
				if len(w.state) > 0 {
					busy = 1
				}
				atomic.StoreInt32(w.busy, busy)
			}
			w.Tick()
			if !w.gameOver {
				if w.shouldEndGame() {
					w.endGame()
				} else if !w.battleWon && w.shouldWin() {
					w.winBattle()
				}
			}
			w.notify()
		}
	}
}

func (w *World) Busy() bool {
	busy := atomic.LoadInt32(w.busy)
	return busy != 0
}

func (w *World) NextTurn() {
	w.turn++
	if len(w.waitlist) == 0 {
		return
	}
	allDead := true
	for _, t := range w.waitlist {
		if m, ok := t.(*Mob); ok {
			if m.CanAct() || m.CanMove() {
				allDead = false
				break
			}
		}
	}
	if allDead {
		w.up = nil
		return
	}

	for _, turner := range w.waitlist {
		turner.TurnTick(w)
	}
	w.sortWaitlist()
	top := w.waitlist[0]
	if top.CT() >= ctForTurn {
		w.up = top
		top.TakeTurn(w)
		if m, ok := top.(*Mob); ok {
			if !m.CanAct() && !m.CanMove() {
				m.FinishTurn(w, false, false)
				w.NextTurn()
			}
		}
	} else {
		w.NextTurn()
	}
}

func (w *World) Up() Turner {
	return w.up
}

func (w *World) sortWaitlist() {
	sort.Slice(w.waitlist, func(i, j int) bool {
		t0, t1 := w.waitlist[i], w.waitlist[j]
		ct0, ct1 := t0.CT(), t1.CT()
		if ct0 == ct1 {
			s0, s1 := t0.Speed(), t1.Speed()
			if s0 == s1 {
				return t0.ID() < t1.ID()
			}
			return s0 > s1
		}
		return ct0 > ct1
	})
}

func (w *World) Broadcast(msg string) {
	for sesh := range w.seshes {
		sesh.Send(msg)
	}
}

func (w *World) notify() {
	for sesh := range w.seshes {
		sesh.refresh()
	}
}

func (w *World) shouldEndGame() bool {
	if w.current == nil {
		return false
	}
	for _, obj := range w.current.Objects {
		if mob, ok := obj.(*Mob); ok && mob.Team() == PlayerTeam {
			if !mob.Dead() {
				return false
			}
		}
	}
	return true
}

func (w *World) shouldWin() bool {
	if w.current == nil {
		return false
	}
	for _, obj := range w.current.Objects {
		if mob, ok := obj.(*Mob); ok && mob.Team() != PlayerTeam {
			if !mob.Dead() {
				return false
			}
		}
	}
	return true
}

func (w *World) endGame() {
	w.push <- GameOverState{}
	for sesh := range w.seshes {
		sesh.PushWindow(&GameOverWindow{World: w, Sesh: sesh})
	}
	w.gameOver = true
}

func (w *World) winBattle() {
	w.battleWon = true
	for sesh := range w.seshes {
		sesh.PushWindow(&VictoryWindow{World: w, Sesh: sesh})
	}
}

type ListenAction struct {
	listener *Sesh
}

func (la ListenAction) Apply(w *World) {
	w.seshes[la.listener] = struct{}{}
	la.listener.redraw()
}

type PartAction struct {
	listener *Sesh
}

func (pa PartAction) Apply(w *World) {
	log.Println("parting", pa.listener)
	delete(w.seshes, pa.listener)
}

func (w *World) StartBattle(level int) {
	if level > 0 {
		w.score += 1000
		w.score += max(0, (500-int(w.turn))*2)
		for i := 0; i < len(w.player.Units); i++ {
			if !w.player.Units[i].Dead() {
				continue
			}
			w.score -= 500
			fmt.Println("Dead:", w.player.Units[i].Name())
			if i < len(w.player.Units)-1 {
				copy(w.player.Units[i:], w.player.Units[i+1:])
			}
			w.player.Units[len(w.player.Units)-1] = nil
			w.player.Units = w.player.Units[:len(w.player.Units)-1]
		}
	}

	battle := newBattle(level, w.player)

	m := w.Map(battle.Map)
	m.Reset()
	w.level = level
	w.current = m
	w.turn = 0
	w.waitlist = nil
	w.battleWon = false
	w.battle = battle

	n := 0
	for teamID, team := range battle.Teams {
		for i, unit := range team.Units {
			unit.loc = m.SpawnPoints[teamID][i]
			unit.Reset(w)
			w.Add(unit)
			n++
		}
	}

	for sesh := range w.seshes {
		gw := &GameWindow{World: w, Map: m, Team: PlayerTeam, Sesh: sesh}
		sesh.PushWindow(gw)
		sesh.win = gw
	}

	w.NextTurn()
}

func (w *World) Add(obj Object) {
	w.Create(obj)
	loc := obj.Loc()
	if loc.Map != "" {
		w.Map(loc.Map).TileAtLoc(loc).Add(obj)
	}
}

func (w *World) Attack(target *Mob, source *Mob, weapon Weapon) {
	if weapon.Damage.Type != DamageNone {
		dmg := target.Damage(w, weapon.Damage)
		msg := fmt.Sprintf("%s attacked %s with %s for %d damage!", source.Name(), target.Name(), weapon.Name, dmg)
		if dmg < 0 {
			msg = fmt.Sprintf("%s healed %s with %s for %d HP!", source.Name(), target.Name(), weapon.Name, -dmg)
		}
		w.Broadcast(msg)
	}
	if weapon.OnHit != nil {
		weapon.OnHit(w, source, target)
	}
	if target.Dead() {
		w.Broadcast(fmt.Sprintf("%s died.", target.Name()))
	}
}

type EnqueueAction struct {
	ID     ID
	Action func(*Mob, *World)
}

func (eq EnqueueAction) Apply(w *World) {
	obj, ok := w.objects[eq.ID]
	if !ok {
		fmt.Println("enqueue no ID", eq)
	}
	if mob, ok := obj.(*Mob); ok {
		mob.Enqueue(eq.Action)
	} else {
		fmt.Println("NOT A MOB")
	}
}

type InputAction struct {
	UI    []Window
	Input string
	Sesh  *Sesh
}

func (ia InputAction) Apply(_ *World) {
	for i := len(ia.UI) - 1; i >= 0; i-- {
		win := ia.UI[i]
		if win.Input(ia.Input) {
			return
		}
	}
	ia.Sesh.removeWindows()
}

type ClickAction struct {
	UI   []Window
	X, Y int
	Sesh *Sesh
}

func (ca ClickAction) Apply(_ *World) {
	for i := len(ca.UI) - 1; i >= 0; i-- {
		win := ca.UI[i]
		if win.Click(Coords{ca.X, ca.Y}) {
			return
		}
	}
	ca.Sesh.removeWindows()
}

type MouseoverAction struct {
	UI   []Window
	X, Y int
	Sesh *Sesh
}

func (ca MouseoverAction) Apply(_ *World) {
	for i := len(ca.UI) - 1; i >= 0; i-- {
		win := ca.UI[i]
		if win.Mouseover(Coords{ca.X, ca.Y}) {
			return
		}
	}
	ca.Sesh.removeWindows()
}

// ShutdownAction resets the terminal (useful when debugging).
// It's the only thing that should be used with applySync.
type ShutdownAction struct{}

func (ShutdownAction) Apply(w *World) {
	for sesh := range w.seshes {
		io.WriteString(sesh.ssh, resetScreen+resetSGR+"\033[?1003l")
	}
}

func (*World) ApplyBonus(bonus Bonus, mob *Mob) {
	bonus.Apply(mob)
}

// NextTurnState goes to the next turn.
// It's usually called with pushBottom so that other states,
// like MoveState will end before the next turn.
type NextTurnState struct{}

func (NextTurnState) Run(w *World) bool {
	w.NextTurn()
	return true
}

type MoveState struct {
	Obj    Object
	Path   []Loc
	Delete bool
	Speed  int
	OnEnd  func(w *World)

	i    int
	wait int
}

func (ms *MoveState) Run(w *World) bool {
	if len(ms.Path) == 0 {
		if ms.OnEnd != nil {
			ms.OnEnd(w)
		}
		return true
	}
	if ms.wait < ms.Speed {
		ms.wait++
		return false
	}
	ms.wait = 0
	loc := ms.Path[ms.i]
	m := w.Map(loc.Map)
	m.Move(ms.Obj, loc.X, loc.Y)
	ms.i++
	if ms.i != len(ms.Path) {
		return false
	}
	if ms.Delete {
		w.Delete(ms.Obj.ID())
	}
	if ms.OnEnd != nil {
		ms.OnEnd(w)
	}
	return true
}

type AttackState struct {
	Char     *Mob
	Targets  []*Mob
	Weapon   Weapon
	ProjPath []Loc
	HitLocs  []Loc

	done bool
}

func (as *AttackState) Run(w *World) bool {
	wep := as.Weapon
	if as.done {
		for _, t := range as.Targets {
			w.Attack(t, as.Char, wep)
		}
		return true
	}
	var onend func(*World)
	if wep.HitGlyph != nil && len(as.HitLocs) > 0 {
		onend = func(w *World) {
			for _, loc := range as.HitLocs {
				loc := loc
				loc.Z = 999
				fx := &Effect{
					loc:   loc,
					glyph: *wep.HitGlyph,
					life:  15,
				}
				w.Add(fx)
			}
		}
	}
	if wep.projectile != nil && len(as.ProjPath) > 0 {
		proj := wep.projectile()
		proj.Move(as.ProjPath[0])
		w.Add(proj)
		w.push <- &MoveState{
			Obj:    proj,
			Path:   as.ProjPath,
			Delete: true,
			Speed:  1,
			OnEnd:  onend,
		}
	} else if onend != nil {
		onend(w)
	}

	if wep.MPCost > 0 {
		as.Char.AddMP(-wep.MPCost)
	}
	as.done = true
	return false
}

type EnemyAIState struct {
	self   *Mob
	target *Mob
	moved  bool
	acted  bool

	// path  []Loc
	// pause int
}

func (ai *EnemyAIState) Run(w *World) bool {
	loc := ai.self.Loc()
	m := w.Map(loc.Map)

	if ai.self.Dead() {
		w.NextTurn()
		return true
	}

	if ai.target == nil {
		var path []Loc
		for _, obj := range m.Objects {
			mob, ok := obj.(*Mob)
			if !ok {
				continue
			}
			if mob.Team() == ai.self.Team() {
				continue
			}
			if mob.Dead() {
				continue
			}
			if ai.self.tauntedBy != nil && !ai.self.tauntedBy.Dead() && mob.ID() != ai.self.tauntedBy.ID() {
				continue
			}

			if ai.self.CanAttack(w, mob, ai.self.Weapon()) {
				ai.target = mob
				// TODO: maybe run away when too close
				return false
			}

			// otherloc := mob.Loc()
			// newpath := m.FindPath(loc.X, loc.Y, otherloc.X, otherloc.Y, ai.self, mob)
			newpath := m.FindPathNextTo(ai.self, mob)
			if len(newpath) == 0 {
				continue
			}
			for i := 0; i < len(newpath)-1; i++ {
				if ai.self.CanAttackFrom(w, newpath[i], mob, ai.self.Weapon()) {
					newpath = newpath[:i+1]
					fmt.Println("AI shorter path:", newpath)
					break
				}
			}
			if path == nil || len(newpath) < len(path) {
				path = newpath
				ai.target = mob
			}
		}
		if path != nil {
			if len(path) == 0 {
				// already close
				return false
			}
			if len(path) > ai.self.MoveRange() {
				path = path[:ai.self.MoveRange()]
			}
			ai.moved = true
			w.push <- &MoveState{Obj: ai.self, Path: path}
			return false
		}
		ai.self.FinishTurn(w, ai.moved, ai.acted)
		w.NextTurn()
		return true
	}

	if ai.self.CanAttack(w, ai.target, ai.self.Weapon()) {
		wep := ai.self.Weapon()
		var hitlocs, projpath []Loc
		var targets []*Mob
		if wep.Magic {
			targets, hitlocs = findTargets(ai.target.Loc(), m, true, wep.HitboxSize, wep.Hitbox)
			_, _, projpath = m.Raycast(loc, ai.target.Loc(), true)
		} else {
			t, _, path := m.Raycast(ai.self.Loc(), ai.target.Loc(), false)
			targets = []*Mob{t}
			projpath = path
		}
		w.push <- &AttackState{
			Char:     ai.self,
			Weapon:   wep,
			HitLocs:  hitlocs,
			ProjPath: projpath,
			Targets:  targets,
		}
		ai.acted = true
	}

	ai.self.FinishTurn(w, ai.moved, ai.acted)
	w.pushBottom <- NextTurnState{}
	return true
}

type GameOverState struct{}

func (gos GameOverState) Run(w *World) bool {
	return false
}
