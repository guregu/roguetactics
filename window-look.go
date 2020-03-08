package main

import (
	"fmt"
)

type FarlookWindow struct {
	World *World
	Sesh  *Sesh
	Char  Object

	cursorX int
	cursorY int

	done bool
}

func (mw *FarlookWindow) Render(scr [][]Glyph) {
	copyString(scr[len(scr)-1], "Query: use arrow keys or mouse to look around, ESC to exit.", true)

	const arrow = " └"
	if mw.cursorX == -1 || mw.cursorY == -1 {
		copyString(scr[len(scr)-2], arrow, true)
		return
	}
	m := mw.World.Map(mw.defaultLoc().Map)
	tile := m.TileAt(mw.cursorX, mw.cursorY)
	if target, ok := tile.Top().(*Mob); ok {
		status := append(GlyphsOf(arrow), target.StatusLine()...)
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

func (mw *FarlookWindow) Cursor() (x, y int) {
	if mw.cursorX != -1 && mw.cursorY != -1 {
		return mw.cursorX, mw.cursorY
	}
	loc := mw.defaultLoc()
	return loc.X, loc.Y
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

func (mw *FarlookWindow) Click(x, y int) bool {
	// TODO: popup detailed info window
	return true
}

func (mw *FarlookWindow) ShouldRemove() bool {
	return mw.done
}

func (mw *FarlookWindow) Mouseover(x, y int) bool {
	m := mw.World.Map(mw.defaultLoc().Map)
	if x >= m.Width() || y >= m.Height() {
		return true
	}
	mw.cursorX = x
	mw.cursorY = y
	return true
}

func (mw *FarlookWindow) moveCursor(dx, dy int) {
	loc := mw.defaultLoc()
	m := mw.World.Map(loc.Map)
	if mw.cursorX == -1 {
		mw.cursorX = loc.X
	}
	if mw.cursorY == -1 {
		mw.cursorY = loc.Y
	}
	mw.cursorX += dx
	if mw.cursorX >= m.Width() {
		mw.cursorX = m.Width() - 1
	}
	mw.cursorY += dy
	if mw.cursorY >= m.Height() {
		mw.cursorY = m.Height() - 1
	}
}

var (
	_ Window = (*FarlookWindow)(nil)
)
