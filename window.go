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

type Glyph struct {
	Rune rune
	SGR
}

func GlyphOf(r rune) Glyph {
	return Glyph{Rune: r}
}

func GlyphsOf(str string) []Glyph {
	glyphs := make([]Glyph, 0, len(str))
	for _, r := range str {
		glyphs = append(glyphs, GlyphOf(r))
	}
	return glyphs
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
		return resetSGR + string(g.Rune)
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
					buf += d.next[y][x].String()
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
	Cursor() Coords
	Input(string) bool
	Click(coords Coords) bool
	Mouseover(coords Coords) bool
	ShouldRemove() bool
}

func (gw *GameWindow) Cursor() Coords {
	up := gw.World.Up()
	m, ok := up.(*Mob)
	if !ok {
		return OriginCoords
	}
	loc := m.Loc()
	return loc.AsCoords()
}

func drawCenteredBox(scr [][]Glyph, lines []string, bgColor int) {
	linelen := len(lines[0])
	for _, line := range lines {
		if len(line) > linelen {
			linelen = len(line)
		}
	}
	xoffset := 80/2 - (linelen+2)/2
	yoffset := 20/2 - (len(lines)+1)/2
	if xoffset < 0 {
		xoffset = 0
	}
	if yoffset < 0 {
		yoffset = 0
	}
	copyStringOffset(scr[yoffset], " "+strings.Repeat(" ", linelen)+" ", xoffset)
	for n, line := range lines {
		copyStringOffset(scr[1+n+yoffset], " "+line+strings.Repeat(" ", linelen-len(line))+" ", xoffset)
	}
	copyStringOffset(scr[1+len(lines)+yoffset], " "+strings.Repeat(" ", linelen)+" ", xoffset)
	for y := 0; y < len(lines)+2; y++ {
		for x := 0; x < linelen+2; x++ {
			if y >= len(scr) {
				continue
			}
			if x >= len(scr[y+yoffset]) {
				continue
			}
			scr[y+yoffset][x+xoffset].BG = bgColor
		}
	}
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

func copyGlyphs(dst []Glyph, src []Glyph, padRight bool) {
	x := 0
	for _, r := range src {
		if x >= len(dst) {
			break
		}
		dst[x] = r
		x++
	}
	if !padRight {
		return
	}
	for ; x < len(dst); x++ {
		dst[x] = GlyphOf(' ')
	}
}

func copyStringOffset(dst []Glyph, src string, offset int) {
	x := 0
	for _, r := range src {
		if x >= len(dst) {
			break
		}
		dst[x+offset] = GlyphOf(r)
		x++
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

const (
	ColorBlack        = 0
	ColorRed          = 1
	ColorOlive        = 3
	ColorBlue         = 4
	ColorBrightRed    = 9
	ColorBrightGreen  = 10
	ColorBrightYellow = 11
	ColorBrightBlue   = 12
	ColorBrightPink   = 13
	ColorTeal         = 14
	ColorWhite        = 15
	ColorDarkRed      = 52
	ColorGray         = 237
)
