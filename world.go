package main

import (
	"fmt"
	"log"
	"time"
)

type ID int64

type World struct {
	maps    map[string]*Map
	objects map[ID]Object
	seshes  map[*Sesh]struct{}

	lastID ID
	tick   int64

	apply chan Action
}

type Action interface {
	Apply(*World)
}

func newWorld() *World {
	w := &World{
		maps:    make(map[string]*Map),
		objects: make(map[ID]Object),
		seshes:  make(map[*Sesh]struct{}),

		apply: make(chan Action),
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
}

func (w *World) Tick() {
	w.tick++
}

func (w *World) Run() {
	ticker := time.NewTicker(time.Millisecond * 40)
	defer ticker.Stop()
	for {
		select {
		case a := <-w.apply:
			a.Apply(w)
			w.notify()
		case <-ticker.C:
			w.Tick()
		}
	}
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
}

type PlaceAction struct {
	ID  ID
	Loc Loc
	Src *Sesh
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
