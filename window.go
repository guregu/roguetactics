package main

import (
	"fmt"
	// "strconv"
	"strings"
)

const (
	resetScreen = "\033[2J"
	resetSGR    = "\033[0m"
	cursorTo00  = "\033[1;1H"
)

func GlyphOf(r rune) Glyph {
	return Glyph{Rune: r}
}

type Glyph struct {
	Rune rune
	SGR
}

type SGR struct {
	FG        int
	BG        int
	Bold      bool
	Underline bool
	Blink     bool
	Reverse   bool
}

func (g Glyph) args() []string {
	var args []string
	if g.FG != 0 {
		// TODO non 256 support
		args = append(args, fmt.Sprintf("38;5;%d", g.FG))
		// args = append(args, strconv.Itoa(g.FG))
	}
	if g.BG != 0 {
		args = append(args, fmt.Sprintf("48;5;%d", g.BG))
		// args = append(args, strconv.Itoa(g.BG))
	}
	if g.Bold {
		args = append(args, "1")
	}
	if g.Underline {
		args = append(args, "4")
	}
	if g.Blink {
		args = append(args, "5")
	}
	if g.Reverse {
		args = append(args, "7")
	}
	return args
}

func (g Glyph) String() string {
	if g.FG == 0 && g.BG == 0 && !g.Reverse && !g.Bold && !g.Blink && !g.Underline {
		return string(g.Rune)
	}
	return "\033[0;" + strings.Join(g.args(), ";") + "m" + string(g.Rune)
}

// func (g Glyph) StringAfter(other Glyph) string {
// 	if g.SGR == other.SGR {
// 		return string(g.Rune)
// 	}
// 	return g.String()
// }

func ansiCursorTo(x, y int) string {
	return fmt.Sprintf("\033[%d;%dH", y+1, x+1)
}

func blankScreen(w, h int) [][]Glyph {
	scr := make([][]Glyph, h)
	for y := 0; y < h; y++ {
		scr[y] = make([]Glyph, w)
		for x := 0; x < w; x++ {
			scr[y][x] = GlyphOf(' ')
		}
	}
	return scr
}

type Display struct {
	w, h int
	prev [][]Glyph
	next [][]Glyph
}

func NewDisplay(w, h int) *Display {
	return &Display{
		w:    w,
		h:    h,
		prev: blankScreen(w, h),
		next: blankScreen(w, h),
	}
}

func (d *Display) full() string {
	buf := resetScreen + cursorTo00 + resetSGR
	for y := 0; y < len(d.next); y++ {
		if y > 0 {
			buf += "\n\r"
		}
		for x := 0; x < len(d.next[y]); x++ {
			buf += d.next[y][x].String()
		}
	}
	return buf
}

func (d *Display) diff() string {
	var buf string
	sgr := SGR{FG: -1}
	for y := 0; y < len(d.next); y++ {
		for x := 0; x < len(d.next[y]); x++ {
			if d.prev[y][x] != d.next[y][x] {
				if x == 0 || d.prev[y][x-1] == d.next[y][x-1] {
					buf += ansiCursorTo(x, y)
				}
				if d.next[y][x].SGR != sgr {
					buf += resetSGR + d.next[y][x].String()
					sgr = d.next[y][x].SGR
				} else {
					buf += string(d.next[y][x].Rune)
				}
			}
		}
	}
	return buf
}

func (d *Display) nextFrame() [][]Glyph {
	d.prev, d.next = d.next, d.prev
	return d.next
}

type Window interface {
	Render(scr [][]Glyph)
	Cursor() (x, y int)
	Input(string) bool
	Click(x, y int) bool
	ShouldRemove() bool
}

type GameWindow struct {
	Sesh  *Sesh
	World *World
	Char  Object
	Msgs  []string
	Team  int

	turnID int
	moved  bool
	acted  bool
}

func (gw *GameWindow) Input(in string) bool {
	if in[0] == 13 {
		// enter key
		gw.Sesh.PushWindow(&ChatWindow{prompt: "Chat: "})
		return true
	}
	switch in {
	case "Q":
		gw.Sesh.ssh.Exit(0)
		return true
	case "R":
		gw.Sesh.redraw()
		return true
	}

	if !gw.myTurn() {
		return true
	}

	switch in {
	case "m":
		return gw.showMove()
	case "a":
		return gw.showAttack()
	case "n":
		return gw.nextTurn()
	}

	// var x, y int
	// switch in {
	// case ArrowKeyUp:
	// 	y--
	// case ArrowKeyDown:
	// 	y++
	// case ArrowKeyLeft:
	// 	x--
	// case ArrowKeyRight:
	// 	x++
	// }
	// if x != 0 || y != 0 {
	// 	gw.World.apply <- EnqueueAction{ID: gw.Char.ID(), Action: func(mob *Mob, world *World) {
	// 		loc := mob.Loc()
	// 		m := world.Map(loc.Map)
	// 		loc.X += x
	// 		loc.Y += y
	// 		target := m.TileAtLoc(loc)
	// 		if target.Collides {
	// 			gw.Sesh.Send("Ouch! You bumped into a wall.")
	// 			return
	// 		}
	// 		if top := target.Top(); top != nil {
	// 			if col, ok := top.(Collider); ok && col.Collides(world, mob.ID()) {
	// 				gw.Sesh.Send("You're blocked by " + col.Name() + ".")
	// 				return
	// 			}
	// 		}
	// 		m.Move(mob, loc.X, loc.Y)
	// 		// go func() {
	// 		// 	gw.World.apply <- PlaceAction{ID: gw.Char.ID(), Loc: loc, Src: gw.Sesh, Collide: true}
	// 		// }()
	// 	}}
	// }
	return true
}

func (gw *GameWindow) myTurn() bool {
	if gw.World.Busy() {
		return false
	}

	if m, ok := gw.World.Up().(*Mob); ok {
		if m.Team() != gw.Team {
			return false
		}
	}

	return true
}

func (gw *GameWindow) showMove() bool {
	if gw.moved {
		return true
	}
	up := gw.World.Up()
	if m, ok := up.(*Mob); ok {
		gw.Sesh.PushWindow(&MoveWindow{
			World:   gw.World,
			Sesh:    gw.Sesh,
			Char:    m,
			Range:   m.MoveRange(),
			cursorX: -1,
			cursorY: -1,
			callback: func(moved bool) {
				if moved {
					gw.moved = true
					if !gw.canDoSomething() {
						gw.nextTurn()
					}
				}
			}})
	}
	return true
}

func (gw *GameWindow) showAttack() bool {
	if gw.acted {
		return true
	}
	up := gw.World.Up()
	if m, ok := up.(*Mob); ok {
		gw.Sesh.PushWindow(&AttackWindow{
			World: gw.World,
			Sesh:  gw.Sesh,
			Char:  m,
			callback: func(acted bool) {
				if acted {
					gw.acted = true
					if !gw.canDoSomething() {
						gw.nextTurn()
					}
				}
			}})
	}
	return true
}

func (gw *GameWindow) nextTurn() bool {
	up := gw.World.Up()
	if m, ok := up.(*Mob); ok {
		if m.Team() != gw.Team {
			return true
		}
		fmt.Println("finish turn", gw.moved, gw.acted)
		m.FinishTurn(gw.moved, gw.acted)
	}
	gw.World.NextTurn()
	gw.moved = false
	gw.acted = false
	return true
}

func (gw *GameWindow) canDoSomething() bool {
	return !gw.moved || !gw.acted
}

func (gw *GameWindow) Render(scr [][]Glyph) {
	if gw.Char == nil {
		return
	}
	loc := gw.Char.Loc()
	m := gw.World.Map(loc.Map)
nextline:
	for y := 0; y < len(m.Tiles); y++ {
		if y >= len(scr) {
			break
		}
		for x := 0; x < len(m.Tiles[y]); x++ {
			if x >= len(scr[y]) {
				continue nextline
			}
			tile := m.TileAt(x, y)
			scr[y][x] = tile.Glyph()
		}
	}
	for i := 0; i < 2; i++ {
		n := len(gw.Msgs) - 2 + i
		y := len(scr) - 4 + i
		if n < 0 || n > len(gw.Msgs) {
			copyString(scr[y], "", true)
		} else {
			copyString(scr[y], gw.Msgs[n], true)
		}
	}
	statusBar := ""
	up := gw.World.Up()
	if up != nil {
		if mob, ok := up.(*Mob); ok {
			statusBar += fmt.Sprintf("[ ] %s (HP: %d, MP: %d, Speed: %d, CT: %d)", mob.Name(), mob.HP(), mob.MP(), mob.Speed(), mob.CT())
		}
	}

	// copyString(scr[len(scr)-1], "Guest (HP: 42/42, MP: 100/100)", true)
	copyString(scr[len(scr)-2], statusBar, true)
	if mob, ok := gw.World.Up().(*Mob); ok {
		scr[len(scr)-2][1] = mob.Glyph()
	}

	turnInfo := fmt.Sprintf("[Turn: %d]", gw.World.turn)
	copyStringAlignRight(scr[len(scr)-2], turnInfo)

	if gw.World.Busy() {
		helpBar := "Busy..."
		copyString(scr[len(scr)-1], helpBar, true)
		return
	}

	helpBar := ""
	if !gw.moved {
		helpBar += "m) Move"
	}
	if len(helpBar) > 0 {
		helpBar += " "
	}
	if !gw.acted {
		helpBar += "a) Attack"
	}
	if len(helpBar) > 0 {
		helpBar += " "
	}
	helpBar += "n) Next turn"
	copyString(scr[len(scr)-1], helpBar, true)
}

func (gw *GameWindow) ShouldRemove() bool {
	return false
}

func (gw *GameWindow) Click(x, y int) bool {
	if !gw.myTurn() {
		return true
	}

	up := gw.World.Up()
	uploc := up.Loc()
	if uploc.X == x && uploc.Y == y {
		return gw.showMove()
	}

	m := gw.Char.Loc().Map
	tile := gw.World.Map(m).TileAt(x, y)
	if mob, ok := tile.Top().(*Mob); ok {
		if mob.Team() != gw.Team {
			return gw.showAttack()
		}
	}

	gw.Msgs = append(gw.Msgs, fmt.Sprintf("Clicked: (%d,%d)", x, y))
	return true
}

func copyString(dst []Glyph, src string, padRight bool) {
	x := 0
	for _, r := range src {
		if x >= len(dst) {
			break
		}
		dst[x] = GlyphOf(r)
		x++
	}
	if !padRight {
		return
	}
	for ; x < len(dst); x++ {
		dst[x] = GlyphOf(' ')
	}
}

func copyStringAlignRight(dst []Glyph, src string) {
	x := len(dst) - len(src)
	for _, r := range src {
		if x >= len(dst) {
			break
		}
		dst[x] = GlyphOf(r)
		x++
	}
}

func (gw *GameWindow) Cursor() (x, y int) {
	up := gw.World.Up()
	m, ok := up.(*Mob)
	if !ok {
		return 0, 0
	}
	loc := m.Loc()
	return loc.X, loc.Y
}

type ChatWindow struct {
	prompt string
	input  string
	done   bool
}

func (cw *ChatWindow) Render(scr [][]Glyph) {
	bottom := scr[len(scr)-1]
	text := cw.prompt + cw.input
	copyString(bottom, text, true)
}

func (cw *ChatWindow) Cursor() (x, y int) {
	return len(cw.prompt) + len(cw.input), 24
}

func (cw *ChatWindow) Input(input string) bool {
	switch input[0] {
	case 13: // ENTER
		cw.done = true
		fmt.Println("Chat:", cw.input)
		return true
	case 27: // ESC
		cw.done = true
		return true
	case 127: // BS
		if len(cw.input) > 0 {
			cw.input = cw.input[:len(cw.input)-1]
		}
		return true
	}
	if input == ArrowKeyLeft || input == ArrowKeyRight || input == ArrowKeyUp || input == ArrowKeyDown {
		return true
	}
	cw.input += input
	return true
}

func (cw *ChatWindow) Click(x, y int) bool {
	return false
}

func (cw *ChatWindow) ShouldRemove() bool {
	return cw.done
}

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
					scr[y][x].BG = ColorBlue
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
	mw.World.push <- &MoveState{Mob: mw.Char.(*Mob), Path: path}
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

const (
	ColorBlack = 0
	ColorRed   = 1
	ColorOlive = 3
	ColorBlue  = 4
	ColorWhite = 15
	ColorGray  = 237
)

var (
	_ Window = (*GameWindow)(nil)
	_ Window = (*ChatWindow)(nil)
	_ Window = (*MoveWindow)(nil)
	_ Window = (*AttackWindow)(nil)
)
