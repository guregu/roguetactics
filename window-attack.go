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

	cursorX int
	cursorY int
}

func (mw *AttackWindow) Render(scr [][]Glyph) {
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	wep := mw.Char.Weapon()
	attackRange := wep.Range
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
			if wep.Targeting == TargetingCross {
				if (loc.X != x) && (loc.Y != y) {
					continue
				}
			}
			scr[y][x].BG = ColorOlive
			// tile := m.TileAt(x, y)
			// top := tile.Top()
			// if !tile.Collides && top == nil {
			// 	scr[y][x].BG = ColorOlive
			// }
		}
	}

	help := "Attack: click or arrow keys to target; ESC to cancel"
	if wep.Targeting == TargetingFree {
		help = "Attack: click or arrow keys to target, then . to confirm; ESC to cancel"
	}
	copyString(scr[len(scr)-1], help, true)

	if mw.cursorX != -1 && mw.cursorY != -1 {
		glyph := mw.Char.Glyph()
		scr[mw.cursorY][mw.cursorX].Rune = 'X'
		scr[mw.cursorY][mw.cursorX].FG = glyph.FG
		scr[mw.cursorY][mw.cursorX].BG = ColorBlack
	}
}

func (mw *AttackWindow) Cursor() (x, y int) {
	if mw.cursorX != -1 && mw.cursorY != -1 {
		return mw.cursorX, mw.cursorY
	}
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
		case '.', 13:
			if mw.cursorX != -1 && mw.cursorY != -1 {
				return mw.Click(mw.cursorX, mw.cursorY)
			}
		}
	}
	loc := mw.Char.Loc()
	wep := mw.Char.Weapon()

	if wep.Targeting == TargetingFree {
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

	switch input {
	case ArrowKeyLeft:
		mw.Click(loc.X-wep.Range, loc.Y)
	case ArrowKeyRight:
		mw.Click(loc.X+wep.Range, loc.Y)
	case ArrowKeyUp:
		mw.Click(loc.X, loc.Y-wep.Range)
	case ArrowKeyDown:
		mw.Click(loc.X, loc.Y+wep.Range)
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
	// TODO: check valid range
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	target, blocked, path := m.Raycast(loc, Loc{Map: m.Name, X: x, Y: y})
	fmt.Println("TARGET", target, "BLOCKED", blocked, "PATH", path)
	if (target == nil && !blocked) || (target != nil && !target.Attackable()) {
		mw.Sesh.Bell()
		return true
	}
	if blocked {
		mw.Sesh.Send(fmt.Sprintf("%s's attack was obstructed.", mw.Char.Name()))
	} else {
		mw.World.apply <- AttackAction{
			Source: mw.Char,
			Target: target,
		}
	}
	if wep := mw.Char.Weapon(); wep.projectile != nil && len(path) > 0 {
		proj := wep.projectile()
		proj.Move(path[0])
		mw.World.Add(proj)
		mw.World.push <- &MoveState{
			Obj:    proj,
			Path:   path,
			Delete: true,
			Speed:  1,
		}
	}
	mw.done = true
	mw.callback(true)
	return true
}

func (mw *AttackWindow) ShouldRemove() bool {
	return mw.done
}

func (mw *AttackWindow) close() {
	mw.done = true
}

func (mw *AttackWindow) moveCursor(dx, dy int) {
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
