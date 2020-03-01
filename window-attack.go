package main

import (
	"fmt"
)

type AttackWindow struct {
	World    *World
	Sesh     *Sesh
	Char     *Mob
	Range    int
	Self     bool
	Readonly bool
	done     bool
	callback func(moved bool)
}

func (mw *AttackWindow) Render(scr [][]Glyph) {
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	attackRange := mw.Char.Weapon().Range
	for y := loc.Y - attackRange; y <= loc.Y+attackRange; y++ {
		if y < 0 {
			continue
		}
		if y >= m.Height() {
			break
		}
		for x := loc.X - attackRange; x <= loc.X+attackRange; x++ {
			if x < 0 {
				continue
			}
			if x >= m.Width() {
				break
			}
			if !mw.Self && loc.X == x && loc.Y == y {
				continue
			}
			if abs(loc.X-x)+abs(loc.Y-y) > attackRange {
				continue
			}
			scr[y][x].BG = ColorOlive
			// tile := m.TileAt(x, y)
			// top := tile.Top()
			// if !tile.Collides && top == nil {
			// 	scr[y][x].BG = ColorOlive
			// }
		}
	}
	copyString(scr[len(scr)-1], "Attack: click or arrow keys to target, ESC to cancel", true)
}

func (mw *AttackWindow) Cursor() (x, y int) {
	loc := mw.Char.Loc()
	return loc.X, loc.Y
}

func (mw *AttackWindow) Input(input string) bool {
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
		}
	}
	loc := mw.Char.Loc()
	switch input {
	case ArrowKeyLeft:
		mw.Click(loc.X-1, loc.Y)
	case ArrowKeyRight:
		mw.Click(loc.X+1, loc.Y)
	case ArrowKeyUp:
		mw.Click(loc.X, loc.Y-1)
	case ArrowKeyDown:
		mw.Click(loc.X, loc.Y+1)
	case ">":
		mw.Click(loc.X, loc.Y)
	}
	return true
}

func (mw *AttackWindow) Click(x, y int) bool {
	if mw.Readonly {
		return true
	}
	fmt.Println("attack click", x, y)
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	targetTile := m.TileAt(x, y)
	target := targetTile.Top()
	if mob, ok := target.(*Mob); ok {
		if !mob.Attackable() {
			mw.Sesh.Bell()
			return true
		}
		mw.World.apply <- AttackAction{
			Source: mw.Char,
			Target: mob,
		}
		mw.done = true
		mw.callback(true)
		return true
	} else {
		mw.Sesh.Bell()
	}
	return true
}

func (mw *AttackWindow) ShouldRemove() bool {
	return mw.done
}

func (mw *AttackWindow) close() {
	mw.done = true
}
