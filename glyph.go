package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

type Glyph struct {
	Rune rune
	SGR
}

func GlyphOf(r rune, styles ...Style) Glyph {
	g := Glyph{Rune: r}
	for _, style := range styles {
		style(&g)
	}
	return g
}

func GlyphsOf(str string, styles ...Style) []Glyph {
	glyphs := make([]Glyph, 0, len(str))
	for _, r := range str {
		g := GlyphOf(r, styles...)
		glyphs = append(glyphs, g)
	}
	return glyphs
}

type SGR struct {
	FG        Color
	BG        Color
	Bold      bool
	Underline bool
	Blink     bool
	Reverse   bool
}

func (g Glyph) args() []string {
	var args []string
	if g.FG != nil {
		args = append(args, fmt.Sprintf("38;%d;%s", g.FG.Colorspace(), g.FG.Params()))
	}
	if g.BG != nil {
		args = append(args, fmt.Sprintf("48;%d;%s", g.BG.Colorspace(), g.BG.Params()))
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
	if g.FG == nil && g.BG == nil && !g.Reverse && !g.Bold && !g.Blink && !g.Underline {
		return resetSGR + string(g.Rune)
	}
	return "\033[0;" + strings.Join(g.args(), ";") + "m" + string(g.Rune)
}

type Color interface {
	Params() string
	Colorspace() byte
}

// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
type Color256 byte

func (c Color256) Params() string {
	return strconv.Itoa(int(c))
}

func (Color256) Colorspace() byte {
	return 5
}

// https://en.wikipedia.org/wiki/ANSI_escape_code#24-bit
type ColorRGB [3]byte

func (c ColorRGB) Params() string {
	return fmt.Sprintf("%d;%d;%d", c[0], c[1], c[2])
}

func (ColorRGB) Colorspace() byte {
	return 2
}

// InvalidColor is kind of a hack, see Display.diff(); don't use elsewhere
type InvalidColor struct{}

func (InvalidColor) Params() string {
	return ""
}

func (InvalidColor) Colorspace() byte {
	return 0
}

type Style func(*Glyph)

func StyleFG(c Color) Style {
	return func(g *Glyph) {
		g.FG = c
	}
}

func StyleBG(c Color) Style {
	return func(g *Glyph) {
		g.BG = c
	}
}

func StyleBold(g *Glyph) {
	g.Bold = true
}

func StyleUnderline(g *Glyph) {
	g.Underline = true
}

type Colors []Color

func (c Colors) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Colors) Less(i, j int) bool {
	return c[i].Params() < c[j].Params()
}

func (c Colors) Len() int {
	return len(c)
}

func ConcatGlyphs(strs ...[]Glyph) []Glyph {
	var result []Glyph
	for _, str := range strs {
		result = append(result, str...)
	}
	return result
}

// Concat concatenates args into a glyph string.
// Arguments must be []Glyph, Glyph, string, rune, or int.
func Concat(args ...interface{}) []Glyph {
	var result []Glyph
	for _, arg := range args {
		switch x := arg.(type) {
		case []Glyph:
			result = append(result, x...)
		case Glyph:
			result = append(result, x)
		case string:
			result = append(result, GlyphsOf(x)...)
		case rune:
			result = append(result, GlyphOf(x))
		case int:
			result = append(result, GlyphsOf(strconv.Itoa(x))...)
		default:
			log.Printf("invalid type for concat: %T (%v)", x, x)
		}
	}
	return result
}

func ColorDamage(dmg int) []Glyph {
	color := ColorDarkOrange
	if dmg < 0 {
		color = ColorBrightGreen
		dmg = -dmg
	}
	return GlyphsOf(strconv.Itoa(dmg), StyleFG(color))
}

const (
	ColorBlack        Color256 = 0
	ColorRed          Color256 = 1
	ColorGreen        Color256 = 2
	ColorOlive        Color256 = 3
	ColorBlue         Color256 = 4
	ColorBrightRed    Color256 = 9
	ColorBrightGreen  Color256 = 10
	ColorBrightYellow Color256 = 11
	ColorBrightBlue   Color256 = 12
	ColorBrightPink   Color256 = 13
	ColorTeal         Color256 = 14
	ColorWhite        Color256 = 15
	ColorNavy         Color256 = 17
	ColorDarkGreen    Color256 = 22
	ColorDarkRed      Color256 = 52
	ColorDiarrhea     Color256 = 58
	ColorDarkOrange   Color256 = 166
	ColorGray         Color256 = 237
)

var (
	_ Color          = Color256(0)
	_ Color          = ColorRGB{}
	_ sort.Interface = (Colors)(nil)
)
