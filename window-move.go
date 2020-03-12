package main

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

	*cursorHandler
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
	if mw.cursor.IsValid() {
		glyph := mw.Char.Glyph()
		scr[mw.cursor.y][mw.cursor.x].Rune = 'X'
		scr[mw.cursor.y][mw.cursor.x].FG = glyph.FG
		scr[mw.cursor.y][mw.cursor.x].BG = ColorBlack
	}
	copyString(scr[len(scr)-1], "Move: click, or arrow keys then . or enter to move; ESC to cancel", true)
}

func (mw *MoveWindow) Input(input string) bool {
	if mw.Readonly {
		return false
	}
	// Handle single-char inputs, like keypress
	if len(input) == 1 {
		switch input[0] {
		case EscKey:
			if mw.callback != nil {
				mw.callback(false)
			}
			mw.done = true
			return true
		case '.', 13:
			if mw.cursor.IsValid() {
				return mw.Click(mw.cursor)
			}
		}
	}

	return mw.cursorInput(input)
}

func (mw *MoveWindow) Click(coords Coords) bool {
	if mw.Readonly {
		return false
	}
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	path := m.FindPath(loc.X, loc.Y, coords.x, coords.y, mw.Char)
	if len(path) > mw.Range {
		mw.Sesh.Bell()
		return true
	}
	if len(path) == 0 {
		mw.Sesh.Bell()
		return true
	}
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

func (mw *MoveWindow) close() {
	mw.done = true
}

var (
	_ Window = (*MoveWindow)(nil)
)
