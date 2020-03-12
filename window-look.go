package main

import (
	"fmt"
)

type FarlookWindow struct {
	World *World
	Sesh  *Sesh
	Char  Object

	cursor Coords

	done bool
}

func (mw *FarlookWindow) Render(scr [][]Glyph) {
	copyString(scr[len(scr)-1], "Query: use arrow keys or mouse to look around, ESC to exit.", true)

	const arrow = " └"
	if !mw.cursor.IsValid() {
		copyString(scr[len(scr)-2], arrow, true)
		return
	}
	m := mw.World.Map(mw.defaultLoc().Map)
	tile := m.TileAt(mw.cursor.x, mw.cursor.y)
	if target, ok := tile.Top().(*Mob); ok {
		status := append(GlyphsOf(arrow), target.StatusLine(true)...)
		copyGlyphs(scr[len(scr)-2], status, true)
	} else {
		name := "floor"
		if tile.Collides {
			name = "wall"
		}
		status := GlyphsOf(fmt.Sprintf(arrow+"[ ] %s", name))
		status[3] = tile.Glyph()
		copyGlyphs(scr[len(scr)-2], status, true)
	}
}

func (mw *FarlookWindow) defaultLoc() Loc {
	if mw.Char != nil {
		return mw.Char.Loc()
	}
	return Loc{Map: mw.World.current.Name, X: 0, Y: 0}
}

func (mw *FarlookWindow) Cursor() Coords {
	if mw.cursor.IsValid() {
		return mw.cursor
	}
	loc := mw.defaultLoc()
	return Coords{loc.X, loc.Y}
}

func (mw *FarlookWindow) Input(input string) bool {
	if len(input) == 1 {
		switch input[0] {
		case 13, 27: //ENTER
			mw.done = true
		}
	}
	switch input {
	case ArrowKeyLeft:
		mw.moveCursor(-1, 0)
	case ArrowKeyRight:
		mw.moveCursor(1, 0)
	case ArrowKeyUp:
		mw.moveCursor(0, -1)
	case ArrowKeyDown:
		mw.moveCursor(0, 1)
	}

	return true
}

func (mw *FarlookWindow) Click(_ Coords) bool {
	// TODO: popup detailed info window
	return true
}

func (mw *FarlookWindow) ShouldRemove() bool {
	return mw.done
}

func (mw *FarlookWindow) Mouseover(mouseover Coords) bool {
	m := mw.World.Map(mw.defaultLoc().Map)
	if mouseover.x >= m.Width() || mouseover.y >= m.Height() {
		return true
	}
	mw.cursor = mouseover
	return true
}

func (mw *FarlookWindow) moveCursor(dx, dy int) {
	loc := mw.defaultLoc()
	m := mw.World.Map(loc.Map)

	mw.cursor.MergeInIfInvalid(loc.AsCoords())
	mw.cursor.Add(dx, dy)
	mw.cursor.EnsureWithinBounds(m.Width(), m.Height())
}

var (
	_ Window = (*FarlookWindow)(nil)
)
