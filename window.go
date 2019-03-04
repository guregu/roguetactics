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
	buf += resetSGR
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
	if buf != "" {
		buf += resetSGR
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
	Input(string)
}

type GameWindow struct {
	Sesh  *Sesh
	World *World
	Char  Object
	Msgs  []string
}

func (gw *GameWindow) Input(in string) {
	loc := gw.Char.Loc()
	oldLoc := loc
	// m := gw.World.Map(loc.Map)
	switch in {
	case ArrowKeyUp:
		loc.Y--
	case ArrowKeyDown:
		loc.Y++
	case ArrowKeyLeft:
		loc.X--
	case ArrowKeyRight:
		loc.X++
	}
	if loc != oldLoc {
		gw.World.apply <- PlaceAction{ID: gw.Char.ID(), Loc: loc, Src: gw.Sesh}
	}
}

func (gw *GameWindow) Render(scr [][]Glyph) {
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

const (
	ColorRed   = 1
	ColorGray  = 237
	ColorWhite = 15
)
