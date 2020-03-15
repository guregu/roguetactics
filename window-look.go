package main

import (
	"fmt"
)

type FarlookWindow struct {
	World *World
	Sesh  *Sesh
	Char  Object

	*cursorHandler

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
			if tile.Ground.Rune == ' ' {
				name = "empty space"
			} else {
				name = "wall"
			}
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

func (mw *FarlookWindow) Input(input string) bool {
	if len(input) == 1 {
		switch input[0] {
		case EscKey, EnterKey:
			mw.done = true
		}
	}

	return mw.cursorInput(input)
}

func (mw *FarlookWindow) Click(_ Coords) bool {
	// TODO: popup detailed info window
	return true
}

func (mw *FarlookWindow) ShouldRemove() bool {
	return mw.done
}

var (
	_ Window = (*FarlookWindow)(nil)
)
