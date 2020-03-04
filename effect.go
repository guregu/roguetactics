package main

type Effect struct {
	id    ID
	loc   Loc
	glyph Glyph
	name  string

	life int
}

func (e *Effect) Create(w *World) {
	e.id = w.NextID()
}

func (e *Effect) ID() ID {
	return e.id
}

func (e *Effect) Name() string {
	return e.name
}

func (e *Effect) Glyph() Glyph {
	return e.glyph
}

func (e *Effect) Loc() Loc {
	// loc := e.loc
	// loc.Z = 999
	return e.loc
}

func (e *Effect) Move(loc Loc) {
	loc.Z = 999
	e.loc = loc
}

func (e *Effect) Tick(w *World, tick int64) {
	if e.life == -1 {
		return
	}

	e.life--
	if e.life <= 0 {
		w.Delete(e.ID())
	}
}
