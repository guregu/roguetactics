package main

type cursorHandler struct {
	cursor Coords
	origin Coords
	at     *Map
}

func newCursorHandler(origin Coords, at *Map) *cursorHandler {
	return &cursorHandler{
		cursor: InvalidCoords,
		origin: origin,
		at:     at,
	}
}

func newCursorHandlerOn(world *World, obj Object) *cursorHandler {
	if obj == nil {
		return newCursorHandler(OriginCoords, world.current)
	}
	loc := obj.Loc()
	return newCursorHandler(loc.AsCoords(), world.Map(loc.Map))
}

func (ch *cursorHandler) cursorInput(input string) bool {
	switch input {
	case ArrowKeyLeft, "4":
		ch.moveCursor(-1, 0)
	case ArrowKeyRight, "6":
		ch.moveCursor(1, 0)
	case ArrowKeyUp, "8":
		ch.moveCursor(0, -1)
	case ArrowKeyDown, "2":
		ch.moveCursor(0, 1)
	case "7":
		ch.moveCursor(-1, -1)
	case "9":
		ch.moveCursor(1, -1)
	case "1":
		ch.moveCursor(-1, 1)
	case "3":
		ch.moveCursor(1, 1)
	}

	return true
}

func (ch *cursorHandler) moveCursor(dx, dy int) {
	ch.cursor.MergeInIfInvalid(ch.origin)
	ch.cursor.Add(dx, dy)
	ch.cursor.EnsureWithinBounds(ch.at.Width(), ch.at.Height())
}

func (ch *cursorHandler) Cursor() (coords Coords) {
	if ch.cursor.IsValid() {
		return ch.cursor
	}
	return ch.origin
}

func (ch *cursorHandler) Mouseover(mouseover Coords) bool {
	ch.cursor = mouseover
	ch.cursor.EnsureWithinBounds(ch.at.Width(), ch.at.Height())
	return true
}
