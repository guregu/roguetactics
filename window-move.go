package main

import (
	"fmt"
)

type MoveWindow struct {
	World    *World
	Sesh     *Sesh
	Char     Object
	Range    int
	Self     bool
	Readonly bool
	done     bool
	callback func(moved bool)

	pathcache map[*Tile][]Loc
	cursorX   int
	cursorY   int
}

func (mw *MoveWindow) Render(scr [][]Glyph) {
	if mw.pathcache == nil {
		mw.pathcache = make(map[*Tile][]Loc)
	}
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	for y := loc.Y - mw.Range; y <= loc.Y+mw.Range; y++ {
		if y < 0 {
			continue
		}
		if y >= m.Height() {
			break
		}
		for x := loc.X - mw.Range; x <= loc.X+mw.Range; x++ {
			if x < 0 {
				continue
			}
			if x >= m.Width() {
				break
			}
			if !mw.Self && loc.X == x && loc.Y == y {
				continue
			}
			if abs(loc.X-x)+abs(loc.Y-y) > mw.Range {
				continue
			}
			tile := m.TileAt(x, y)
			if !tile.Collides && !tile.HasCollider(mw.Char) {
				var path []Loc
				if p, ok := mw.pathcache[tile]; ok {
					path = p
				} else {
					path = m.FindPath(loc.X, loc.Y, x, y, mw.Char)
					mw.pathcache[tile] = path
				}
				if path != nil && len(path) <= mw.Range {
					scr[y][x].BG = 17
				}
			}
		}
	}
	if mw.cursorX != -1 && mw.cursorY != -1 {
		glyph := mw.Char.Glyph()
		scr[mw.cursorY][mw.cursorX].Rune = 'X'
		scr[mw.cursorY][mw.cursorX].FG = glyph.FG
		scr[mw.cursorY][mw.cursorX].BG = ColorBlack
	}
	copyString(scr[len(scr)-1], "Move: click, or arrow keys then . to move; ESC to cancel", true)
}

func (mw *MoveWindow) Cursor() (x, y int) {
	if mw.cursorX != -1 && mw.cursorY != -1 {
		return mw.cursorX, mw.cursorY
	}
	loc := mw.Char.Loc()
	return loc.X, loc.Y
}

func (mw *MoveWindow) Input(input string) bool {
	if mw.Readonly {
		return false
	}
	if len(input) == 1 {
		switch input[0] {
		case 27: // ESC
			if mw.callback != nil {
				mw.callback(false)
			}
			mw.done = true
			return true
		case '.':
			if mw.cursorX != -1 && mw.cursorY != -1 {
				return mw.Click(mw.cursorX, mw.cursorY)
			}
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

func (mw *MoveWindow) moveCursor(dx, dy int) {
	loc := mw.Char.Loc()
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

func (mw *MoveWindow) Click(x, y int) bool {
	if mw.Readonly {
		return false
	}
	fmt.Println("move click", x, y)
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	path := m.FindPath(loc.X, loc.Y, x, y, mw.Char)
	fmt.Println("path:", path)
	if len(path) > mw.Range {
		fmt.Println("too far:", len(path), mw.Range)
		mw.Sesh.Bell()
		return true
	}
	if len(path) == 0 {
		mw.Sesh.Bell()
		return true
	}
	// for _, loc := range path {
	// enqueueMove(mw.World, mw.Char.(*Mob), loc.X, loc.Y)
	// }
	mw.World.push <- &MoveState{Obj: mw.Char, Path: path}
	if mw.callback != nil {
		mw.callback(true)
	}
	mw.done = true
	return true
}

func (mw *MoveWindow) ShouldRemove() bool {
	return mw.done
}

func (mw *MoveWindow) Mouseover(x, y int) bool {
	return false
}

func (mw *MoveWindow) close() {
	mw.done = true
}

var (
	_ Window = (*MoveWindow)(nil)
)
