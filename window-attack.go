package main

import (
	"fmt"
)

type AttackWindow struct {
	World    *World
	Sesh     *Sesh
	Char     *Mob
	Weapon   Weapon
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
	wep := mw.Weapon
	attackRange := wep.Range
	highlightRange(scr, loc, m, mw.Self, attackRange, wep.Targeting, 130)

	help := "Attack: click or arrow keys to target; ESC to cancel"
	if wep.Targeting == TargetingFree {
		help = "Attack: click or arrow keys to target, then . to confirm; ESC to cancel"
	}
	copyString(scr[len(scr)-1], help, true)

	if mw.cursorX != -1 && mw.cursorY != -1 {
		switch wep.Hitbox {
		case HitboxSingle:
			scr[mw.cursorY][mw.cursorX].BG = ColorOlive
		case HitboxCross:
			highlightRange(scr, Loc{Map: loc.Map, X: mw.cursorX, Y: mw.cursorY}, m, true, wep.HitboxSize, TargetingCross, ColorOlive)
		case HitboxBlob:
			highlightRange(scr, Loc{Map: loc.Map, X: mw.cursorX, Y: mw.cursorY}, m, true, wep.HitboxSize, TargetingFree, ColorOlive)
		}

		if target, ok := m.TileAt(mw.cursorX, mw.cursorY).Top().(*Mob); ok {
			status := append(GlyphsOf(" ↪︎"), target.StatusLine()...)
			copyGlyphs(scr[len(scr)-2], status, true)
		} else {
			copyString(scr[len(scr)-2], " ↪︎", true)
		}
	} else {
		copyString(scr[len(scr)-2], " ↪︎", true)
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
	wep := mw.Weapon

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
	wep := mw.Weapon
	targetLoc := Loc{Map: m.Name, X: x, Y: y}
	canAttack := wep.Magic
	var targets []*Mob
	var projpath []Loc
	var hitlocs []Loc
	if wep.Magic {
		t, hit := findTargets(targetLoc, m, true, wep.HitboxSize, wep.Hitbox)
		if len(t) == 0 {
			mw.Sesh.Bell()
			return true
		}
		targets = t
		hitlocs = hit
		_, _, projpath = m.Raycast(loc, targetLoc, true)
	} else {
		target, blocked, path := m.Raycast(loc, targetLoc, false)
		fmt.Println("TARGET", target, "BLOCKED", blocked, "PATH", path)
		if (target == nil && !blocked) || (target != nil && !target.Attackable()) {
			mw.Sesh.Bell()
			return true
		}
		if len(path) > wep.Range ||
			(wep.Targeting == TargetingCross && ((loc.X != x) && (loc.Y != y))) {
			// out of range
			mw.Sesh.Bell()
			return true
		}
		if blocked {
			mw.Sesh.Send(fmt.Sprintf("%s's attack was obstructed.", mw.Char.Name()))
		} else {
			canAttack = true
			targets = []*Mob{target}
			projpath = path
		}
	}

	if !canAttack {
		return true
	}

	mw.World.push <- AttackState{
		Char:     mw.Char,
		Targets:  targets,
		Weapon:   wep,
		ProjPath: projpath,
		HitLocs:  hitlocs,
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

func (mw *AttackWindow) Mouseover(x, y int) bool {
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	if x >= m.Width() || y >= m.Height() {
		return true
	}
	mw.cursorX = x
	mw.cursorY = y
	return true
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

func highlightRange(scr [][]Glyph, loc Loc, m *Map, selfOK bool, size int, targeting TargetingType, bgColor int) {
	for y := loc.Y - size; y <= loc.Y+size; y++ {
		if y < 0 {
			continue
		}
		if y >= m.Height() {
			break
		}
		for x := loc.X - size; x <= loc.X+size; x++ {
			if x < 0 {
				continue
			}
			if x >= m.Width() {
				break
			}
			if !selfOK && loc.X == x && loc.Y == y {
				continue
			}
			if abs(loc.X-x)+abs(loc.Y-y) > size {
				continue
			}
			if targeting == TargetingCross {
				if (loc.X != x) && (loc.Y != y) {
					continue
				}
			}
			scr[y][x].BG = bgColor
			// tile := m.TileAt(x, y)
			// top := tile.Top()
			// if !tile.Collides && top == nil {
			// 	scr[y][x].BG = ColorOlive
			// }
		}
	}
}

func findTargets(loc Loc, m *Map, selfOK bool, size int, hitbox HitboxType) (targets []*Mob, aoe []Loc) {
	if hitbox == HitboxSingle {
		for _, obj := range m.TileAtLoc(loc).Objects {
			if mob, ok := obj.(*Mob); ok {
				return []*Mob{mob}, []Loc{loc}
			}
		}
		return nil, []Loc{loc}
	}

	for y := loc.Y - size; y <= loc.Y+size; y++ {
		if y < 0 {
			continue
		}
		if y >= m.Height() {
			break
		}
		for x := loc.X - size; x <= loc.X+size; x++ {
			if x < 0 {
				continue
			}
			if x >= m.Width() {
				break
			}
			if !selfOK && loc.X == x && loc.Y == y {
				continue
			}
			if abs(loc.X-x)+abs(loc.Y-y) > size {
				continue
			}
			if hitbox == HitboxCross {
				if (loc.X != x) && (loc.Y != y) {
					continue
				}
			}

			aoe = append(aoe, Loc{Map: loc.Map, X: x, Y: y})
			for _, obj := range m.TileAt(x, y).Objects {
				if mob, ok := obj.(*Mob); ok {
					targets = append(targets, mob)
				}
			}
		}
	}
	return targets, aoe
}

var (
	_ Window = (*AttackWindow)(nil)
)
