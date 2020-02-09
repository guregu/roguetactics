package main

import (
	"fmt"
	"log"
	"sort"
	"sync/atomic"
	"time"
)

const ctForTurn = 100

type ID int64

type World struct {
	maps     map[string]*Map
	objects  map[ID]Object
	seshes   map[*Sesh]struct{}
	waitlist []Turner

	lastID ID
	tick   int64
	turn   int64
	up     Turner
	state  []StateAction
	busy   *int32

	apply chan Action
	push  chan StateAction
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

		apply: make(chan Action, 32),
		push:  make(chan StateAction, 32),
	}
	testMap, err := loadMap("test.map")
	if err != nil {
		panic(err)
	}
	w.maps["test"] = testMap
	return w
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
	ticker := time.NewTicker(time.Millisecond * 40)
	defer ticker.Stop()
	for {
		select {
		case a := <-w.apply:
			a.Apply(w)
			w.notify()
		case a := <-w.push:
			w.state = append(w.state, a)
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

	for _, turner := range w.waitlist {
		turner.TurnTick(w)
	}
	w.sortWaitlist()
	top := w.waitlist[0]
	if top.CT() >= ctForTurn {
		w.up = top
		top.TakeTurn(w)
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

func (w *World) notify() {
	for sesh := range w.seshes {
		sesh.refresh()
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

type AddAction struct {
	Obj Object
}

func (aa AddAction) Apply(w *World) {
	log.Println("create action", aa.Obj)
	w.Create(aa.Obj)
	loc := aa.Obj.Loc()
	if loc.Map != "" {
		w.Map(loc.Map).TileAtLoc(loc).Add(aa.Obj)
	}
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
}

func (ia InputAction) Apply(_ *World) {
	for i := len(ia.UI) - 1; i >= 0; i-- {
		win := ia.UI[i]
		if win.Input(ia.Input) {
			return
		}
	}
}

type ClickAction struct {
	UI   []Window
	X, Y int
}

func (ca ClickAction) Apply(_ *World) {
	for i := len(ca.UI) - 1; i >= 0; i-- {
		win := ca.UI[i]
		if win.Click(ca.X, ca.Y) {
			return
		}
	}
}

type NextTurnAction struct{}

func (na NextTurnAction) Apply(w *World) {
	w.NextTurn()
}

type MoveState struct {
	Mob  *Mob
	Path []Loc
	i    int
}

func (ms *MoveState) Run(w *World) bool {
	if len(ms.Path) == 0 {
		return true
	}
	loc := ms.Path[ms.i]
	m := w.Map(loc.Map)
	m.Move(ms.Mob, loc.X, loc.Y)
	ms.i++
	return ms.i == len(ms.Path)
}
