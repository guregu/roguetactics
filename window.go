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
			buf += "\n"
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
	case "m":
		gw.Sesh.PushWindow(&MoveWindow{World: gw.World, Char: gw.Char, Range: 3})
		return true
	}

	var x, y int
	switch in {
	case ArrowKeyUp:
		y--
	case ArrowKeyDown:
		y++
	case ArrowKeyLeft:
		x--
	case ArrowKeyRight:
		x++
	}
	if x != 0 || y != 0 {
		gw.World.apply <- EnqueueAction{ID: gw.Char.ID(), Action: func(mob *Mob, world *World) {
			loc := mob.Loc()
			m := world.Map(loc.Map)
			loc.X += x
			loc.Y += y
			target := m.TileAtLoc(loc)
			if target.Collides {
				gw.Sesh.Send("Ouch! You bumped into a wall.")
				return
			}
			if top := target.Top(); top != nil {
				if col, ok := top.(Collider); ok && col.Collides(world, mob.ID()) {
					gw.Sesh.Send("You're blocked by " + col.Name() + ".")
					return
				}
			}
			m.Move(mob, loc.X, loc.Y)
			// go func() {
			// 	gw.World.apply <- PlaceAction{ID: gw.Char.ID(), Loc: loc, Src: gw.Sesh, Collide: true}
			// }()
		}}
	}
	return true
}

func enqueueMove(world *World, mob *Mob, x, y int) {
	world.apply <- EnqueueAction{ID: mob.ID(), Action: func(mob *Mob, world *World) {
		loc := mob.Loc()
		m := world.Map(loc.Map)
		loc.X = x
		loc.Y = y
		target := m.TileAtLoc(loc)
		if target.Collides {
			// gw.Sesh.Send("Ouch! You bumped into a wall.")
			return
		}
		if top := target.Top(); top != nil {
			if col, ok := top.(Collider); ok && col.Collides(world, mob.ID()) {
				// gw.Sesh.Send("You're blocked by " + col.Name() + ".")
				return
			}
		}
		m.Move(mob, loc.X, loc.Y)
	}}
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
	for i := 0; i < 3; i++ {
		n := len(gw.Msgs) - 3 + i
		y := len(scr) - 4 + i
		if n < 0 || n > len(gw.Msgs) {
			copyString(scr[y], "", true)
		} else {
			copyString(scr[y], gw.Msgs[n], true)
		}
	}
	copyString(scr[len(scr)-1], "Guest (HP: 42/42, MP: 100/100)", true)
}

func (gw *GameWindow) ShouldRemove() bool {
	return false
}

func (gw *GameWindow) Click(x, y int) bool {
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

func (gw *GameWindow) Cursor() (x, y int) {
	if gw.Char == nil {
		return 0, 0
	}
	loc := gw.Char.Loc()
	return loc.X, loc.Y
}

type ChatWindow struct {
	prompt string
	input  string
	done   bool
}

/*
type Window interface {
	Render(scr [][]Glyph)
	Cursor() (x, y int)
	Input(string) bool
	Click(x, y int) bool
	ShouldRemove() bool
}
*/

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
	World *World
	Char  Object
	Range int
	Self  bool
	done  bool
}

func (mw *MoveWindow) Render(scr [][]Glyph) {
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
			top := tile.Top()
			if !tile.Collides && top == nil {
				scr[y][x].BG = ColorBlue
			}
		}
	}
	copyString(scr[len(scr)-1], "Move", true)
}

func (mw *MoveWindow) Cursor() (x, y int) {
	loc := mw.Char.Loc()
	return loc.X, loc.Y
}

func (mw *MoveWindow) Input(input string) bool {
	if len(input) == 1 {
		switch input[0] {
		case 27: // ESC
			mw.done = true
			return true
		}
	}
	return true
}

func (mw *MoveWindow) Click(x, y int) bool {
	fmt.Println("move click", x, y)
	loc := mw.Char.Loc()
	m := mw.World.Map(loc.Map)
	path := m.FindPath(loc.X, loc.Y, x, y)
	fmt.Println("path:", path)
	for _, loc := range path {
		enqueueMove(mw.World, mw.Char.(*Mob), loc.X, loc.Y)
	}
	mw.done = true
	return true
}

func (mw *MoveWindow) ShouldRemove() bool {
	return mw.done
}

const (
	ColorRed   = 1
	ColorBlue  = 4
	ColorWhite = 15
	ColorGray  = 237
)

var (
	_ Window = (*GameWindow)(nil)
	_ Window = (*ChatWindow)(nil)
	_ Window = (*MoveWindow)(nil)
)
