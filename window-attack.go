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

	*cursorHandler
}

func (mw *AttackWindow) Render(scr [][]Glyph) {
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	wep := mw.Weapon
	attackRange := wep.Range
	highlightRange(scr, loc, m, mw.Self, attackRange, wep.Targeting, 130)

	helpheader := "Attack: "
	if mw.Weapon.Magic {
		helpheader = "Cast " + mw.Weapon.Name + ": "
	}
	help := helpheader + "click or arrow keys to target; ESC to cancel"
	if wep.Targeting == TargetingFree {
		help = helpheader + "click or arrow keys to target, then . to confirm; ESC to cancel"
	}
	copyString(scr[len(scr)-1], help, true)

	cursor := mw.cursor
	if !cursor.IsValid() {
		cursor = mw.origin
	}

	switch wep.Hitbox {
	case HitboxSingle:
		scr[cursor.y][cursor.x].BG = ColorOlive
	case HitboxCross:
		highlightRange(scr, Loc{Map: loc.Map, X: cursor.x, Y: cursor.y}, m, true, wep.HitboxSize, TargetingCross, ColorOlive)
	case HitboxBlob:
		highlightRange(scr, Loc{Map: loc.Map, X: cursor.x, Y: cursor.y}, m, true, wep.HitboxSize, TargetingFree, ColorOlive)
	}

	if target, ok := m.TileAt(cursor.x, cursor.y).Top().(*Mob); ok {
		var dmginfo string
		if wep.Damage.IsValid() {
			dmgname := "damage"
			if wep.Damage.Type == DamageHealing {
				dmgname = "heal"
			}
			dmginfo = fmt.Sprintf(" (%s: %s)", dmgname, wep.Damage.Dice.String())
		}
		status := append(append(GlyphsOf(" └"), target.StatusLine(true)...), GlyphsOf(dmginfo)...)
		copyGlyphs(scr[len(scr)-2], status, true)
	} else {
		copyString(scr[len(scr)-2], " └", true)
	}
}

func (mw *AttackWindow) Input(input string) bool {
	if mw.Readonly {
		return false
	}
	if len(input) == 1 {
		switch input[0] {
		case EscKey:
			if mw.callback != nil {
				mw.callback(false)
			}
			mw.done = true
			return true
		case '.', EnterKey:
			if mw.cursor.IsValid() {
				return mw.Click(mw.cursor)
			} else if mw.Self {
				loc := mw.Char.Loc()
				return mw.Click(Coords{loc.X, loc.Y})
			}
		}
	}
	loc := mw.Char.Loc()
	wep := mw.Weapon

	if wep.Targeting == TargetingFree {
		return mw.cursorInput(input)
	}

	switch input {
	case ArrowKeyLeft:
		mw.Click(Coords{loc.X - wep.Range, loc.Y})
	case ArrowKeyRight:
		mw.Click(Coords{loc.X + wep.Range, loc.Y})
	case ArrowKeyUp:
		mw.Click(Coords{loc.X, loc.Y - wep.Range})
	case ArrowKeyDown:
		mw.Click(Coords{loc.X, loc.Y + wep.Range})
	case ">":
		mw.Click(Coords{loc.X, loc.Y})
	}
	return true
}

func (mw *AttackWindow) Click(click Coords) bool {
	if mw.Readonly {
		return true
	}
	// fmt.Println("attack click", x, y)
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	wep := mw.Weapon
	targetLoc := Loc{Map: m.Name, X: click.x, Y: click.y}
	canAttack := wep.Magic
	var targets []*Mob
	var projpath []Loc
	var hitlocs []Loc
	if wep.Magic {
		if !withinRange(loc, m, true, wep.Range, wep.Targeting, click.x, click.y) {
			mw.Sesh.Bell()
			return true
		}
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
		if (target == nil && !blocked) || (target != nil && !target.Attackable()) {
			mw.Sesh.Bell()
			return true
		}
		if len(path) > wep.Range ||
			(wep.Targeting == TargetingCross && ((loc.X != click.x) && (loc.Y != click.y))) {
			// out of range
			mw.Sesh.Bell()
			return true
		}
		if blocked {
			mw.Sesh.Send(fmt.Sprintf("%s's attack is obstructed.", mw.Char.Name()))
		} else {
			canAttack = true
			targets = []*Mob{target}
			projpath = path
		}
	}

	if !canAttack {
		return true
	}

	mw.World.push <- &AttackState{
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

func withinRange(loc Loc, m *Map, selfOK bool, size int, targeting TargetingType, targetX, targetY int) bool {
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
				return false
			}
			if abs(loc.X-x)+abs(loc.Y-y) > size {
				continue
			}
			if targeting == TargetingCross {
				if (loc.X != x) && (loc.Y != y) {
					continue
				}
			}

			if x == targetX && y == targetY {
				return true
			}
		}
	}
	return false
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
				if mob, ok := obj.(*Mob); ok && !mob.Dead() {
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
