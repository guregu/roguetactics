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

	apply      chan Action
	applySync  chan Action
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

		apply:      make(chan Action, 1024),
		applySync:  make(chan Action, 1024),
		push:       make(chan StateAction, 32),
		pushBottom: make(chan StateAction, 32),
	}
	testMap, err := loadMap("dojo")
	if err != nil {
		panic(err)
	}
	w.maps["dojo"] = testMap
	return w
}

func (w *World) reset() {
	log.Println("Resetting world...")
	w.state = nil
	w.up = nil
	w.gameOver = false
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

// func (w *World) Place(obj *Object, mapName string) {
// 	m := w.maps[mapName]
// 	obj.Map = m
// 	m.Add(obj)
// }

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
			// if w.gameOver {
			// 	w.notify()
			// 	return
			// }
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
				m.FinishTurn(false, false)
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
	fmt.Println("broadcast:", msg)
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
	log.Println("listening", la.listener)
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

type StartBattleAction struct {
	Level int
}

func (sba StartBattleAction) Apply(w *World) {
	fmt.Println("START BATTLE", sba.Level)
	battle := newBattle(sba.Level, w.player)

	m := w.Map(battle.Map)
	m.Reset()
	w.level = sba.Level
	w.current = m
	w.turn = 0
	w.waitlist = nil
	w.battleWon = false
	w.battle = battle

	n := 0
	for teamID, team := range battle.Teams {
		for i, unit := range team.Units {
			unit.loc = m.SpawnPoints[teamID][i]
			unit.hp = unit.maxHP
			unit.mp = unit.maxMP
			w.Add(unit)
			n++
		}
	}

	for sesh := range w.seshes {
		gw := &GameWindow{World: w, Map: m, Team: PlayerTeam, Sesh: sesh}
		sesh.PushWindow(gw)
		sesh.win = gw
	}

	w.apply <- NextTurnAction{}
}

type ForceWinActionDEBUG struct{}

func (ForceWinActionDEBUG) Apply(w *World) {
	w.winBattle()
}

type AddAction struct {
	Obj Object
}

func (w *World) Add(obj Object) {
	w.Create(obj)
	loc := obj.Loc()
	if loc.Map != "" {
		w.Map(loc.Map).TileAtLoc(loc).Add(obj)
	}
}

func (aa AddAction) Apply(w *World) {
	log.Println("create action", aa.Obj)
	w.Add(aa.Obj)
}

type PlaceAction struct {
	ID      ID
	Loc     Loc
	Src     *Sesh
	Collide bool
}

func (pa PlaceAction) Apply(w *World) {
	log.Println("place action", pa)
	obj, ok := w.objects[pa.ID]
	if !ok {
		log.Println("Can't place ID", pa.ID, pa.Loc)
		return
	}

	m := w.Map(pa.Loc.Map)
	if !m.Contains(pa.ID) {
		m.Add(obj)
		return
	}
	if pa.Collide && m.TileAtLoc(pa.Loc).Collides {
		pa.Src.Send("Ouch!")
		return
	}
	m.Move(obj, pa.Loc.X, pa.Loc.Y)

	if pa.Src != nil {
		loc := obj.Loc()
		pa.Src.Send("Moved to " + fmt.Sprintf("(%d,%d)", loc.X, loc.Y))
	}
}

type RemoveAction ID

func (ra RemoveAction) Apply(w *World) {
	log.Println("remove action", ra)
	w.Delete(ID(ra))
}

type AttackAction struct {
	Source *Mob
	Weapon Weapon
	Target *Mob
}

func (aa AttackAction) Apply(w *World) {
	weapon := aa.Weapon
	dmg := weapon.RollDamage()
	if weapon.DamageType == DamageHealing {
		dmg = -dmg
	}
	dmg = aa.Target.Damage(dmg, weapon)
	msg := fmt.Sprintf("%s attacked %s with %s for %d damage!", aa.Source.Name(), aa.Target.Name(), weapon.Name, dmg)
	if dmg < 0 {
		msg = fmt.Sprintf("%s heals %s with %s for %d HP!", aa.Source.Name(), aa.Target.Name(), weapon.Name, -dmg)
	}
	w.Broadcast(msg)
	if aa.Target.Dead() {
		w.Broadcast(fmt.Sprintf("%s died.", aa.Target.Name()))
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
		if win.Click(ca.X, ca.Y) {
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
		if win.Mouseover(ca.X, ca.Y) {
			return
		}
	}
	ca.Sesh.removeWindows()
}

type ShutdownAction struct{}

func (ShutdownAction) Apply(w *World) {
	for sesh := range w.seshes {
		io.WriteString(sesh.ssh, resetScreen+resetSGR+"\033[?1003l")
	}
}

type NextTurnAction struct{}

func (na NextTurnAction) Apply(w *World) {
	w.NextTurn()
}

type ApplyBonusAction struct {
	Mob   *Mob
	Bonus Bonus
}

func (ab ApplyBonusAction) Apply(w *World) {
	ab.Bonus.Apply(ab.Mob)
}

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
			w.apply <- AttackAction{
				Source: as.Char,
				Weapon: wep,
				Target: t,
			}
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
		w.apply <- NextTurnAction{}
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

			if ai.self.CanAttack(mob) {
				ai.target = mob
				return false
			}

			// otherloc := mob.Loc()
			// newpath := m.FindPath(loc.X, loc.Y, otherloc.X, otherloc.Y, ai.self, mob)
			newpath := m.FindPathNextTo(ai.self, mob)
			fmt.Println("AI NEWPATH", newpath)
			if len(newpath) == 0 {
				continue
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
			fmt.Println("AI moving to", path)
			ai.moved = true
			w.push <- &MoveState{Obj: ai.self, Path: path}
			return false
		}
		ai.self.FinishTurn(ai.moved, ai.acted)
		w.apply <- NextTurnAction{}
		return true
	}

	if ai.self.CanAttack(ai.target) {
		w.apply <- AttackAction{
			Source: ai.self,
			Weapon: ai.self.Weapon(),
			Target: ai.target,
		}
		ai.acted = true
	}

	fmt.Println("AI done")
	ai.self.FinishTurn(ai.moved, ai.acted)
	w.apply <- NextTurnAction{}
	return true
}

type GameOverState struct{}

func (gos GameOverState) Run(w *World) bool {
	return false
}
