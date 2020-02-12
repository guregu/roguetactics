package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"

	// "github.com/davecgh/go-spew/spew"
	"github.com/gliderlabs/ssh"
	// "github.com/kr/pty"
)

const (
	ArrowKeyUp    = "\033[A"
	ArrowKeyDown  = "\033[B"
	ArrowKeyRight = "\033[C"
	ArrowKeyLeft  = "\033[D"
	MousePrefix   = "\033[M"

	// https://www.xfree86.org/current/ctlseqs.html#Mouse%20Tracking
	EnableMouseReporting = "\033[?1000h"
)

var mainMap *Map

type Sesh struct {
	x, y  int
	world *World
	ui    []Window
	win   *GameWindow
	ssh   Conn
	disp  *Display

	cursorX, cursorY int
}

func NewSesh(s Conn, w *World) *Sesh {
	return &Sesh{
		world: w,
		ssh:   s,
		disp:  NewDisplay(80, 24),
	}
}

func (sesh *Sesh) resize(win ssh.Window) {
	fmt.Println("resize:", win.Width, win.Height)
}

func (sesh *Sesh) do(input string) {
	if strings.HasPrefix(input, MousePrefix) && len(input) >= 6 && input[3] == 35 {
		x := int(input[4] - 32 - 1)
		y := int(input[5] - 32 - 1)
		// for i := len(sesh.ui) - 1; i >= 0; i-- {
		// 	win := sesh.ui[i]
		// 	if win.Click(x, y) {
		// 		return
		// 	}
		// }
		sesh.world.apply <- ClickAction{UI: sesh.ui, X: x, Y: y, Sesh: sesh}
		return
	}

	sesh.world.apply <- InputAction{UI: sesh.ui, Input: input, Sesh: sesh}
	// for i := len(sesh.ui) - 1; i >= 0; i-- {
	// 	win := sesh.ui[i]
	// 	if win.Input(input) {
	// 		return
	// 	}
	// }
}

func (sesh *Sesh) removeWindows() {
	for i := len(sesh.ui) - 1; i >= 0; i-- {
		if !sesh.ui[i].ShouldRemove() {
			continue
		}
		if i < len(sesh.ui)-1 {
			copy(sesh.ui[i:], sesh.ui[i+1:])
		}
		sesh.ui[len(sesh.ui)-1] = nil
		sesh.ui = sesh.ui[:len(sesh.ui)-1]
	}
}

func (sesh *Sesh) refresh() {
	sesh.removeWindows()
	scr := sesh.disp.nextFrame()
	for i := 0; i < len(sesh.ui); i++ {
		sesh.ui[i].Render(scr)
	}
	render := sesh.disp.diff()
	if render == "" {
		x, y := sesh.ui[len(sesh.ui)-1].Cursor()
		if sesh.cursorX != x || sesh.cursorY != y {
			sesh.setCursor(x, y)
			sesh.cursorX, sesh.cursorY = x, y
		}
		return
	}
	fmt.Println("Render: ", strings.Replace(render, "\033", "ESC", -1))
	io.WriteString(sesh.ssh, render)
	sesh.setCursor(sesh.ui[len(sesh.ui)-1].Cursor())
}

func (sesh *Sesh) redraw() {
	sesh.removeWindows()
	scr := sesh.disp.nextFrame()
	for i := 0; i < len(sesh.ui); i++ {
		sesh.ui[i].Render(scr)
	}
	io.WriteString(sesh.ssh, sesh.disp.full())
	sesh.setCursor(sesh.ui[len(sesh.ui)-1].Cursor())
}

func (sesh *Sesh) setCursor(x, y int) {
	io.WriteString(sesh.ssh, fmt.Sprintf("\033[%d;%dH", y+1, x+1))
}

func (sesh *Sesh) Send(msg string) {
	sesh.win.Msgs = append(sesh.win.Msgs, msg)
}

func (sesh *Sesh) Bell() {
	io.WriteString(sesh.ssh, string(7))
}

func (sesh *Sesh) setup() {
	glyph := GlyphOf('@')
	glyph.FG = ColorBlue

	glyph2 := GlyphOf('d')
	glyph2.FG = ColorBlue

	koboldGlyph := GlyphOf('k')
	kobold2Glyph := GlyphOf('K')
	koboldGlyph.FG = ColorRed
	kobold2Glyph.FG = ColorRed

	team := Team{
		ID: 0,
		Units: []*Mob{
			&Mob{
				name:   "Guy",
				glyph:  glyph,
				loc:    Loc{Map: "test", X: 10, Y: 10, Z: 10},
				speed:  3,
				move:   5,
				hp:     20,
				maxHP:  20,
				weapon: &weaponSword,
			},
			&Mob{
				name:   "Dog",
				glyph:  glyph2,
				loc:    Loc{Map: "test", X: 11, Y: 10, Z: 10},
				speed:  5,
				move:   10,
				hp:     15,
				maxMP:  15,
				weapon: &weaponBite,
			},
		},
	}
	enemies := Team{
		ID: 1,
		Units: []*Mob{
			&Mob{
				name:   "little Kobold",
				glyph:  koboldGlyph,
				loc:    Loc{Map: "test", X: 20, Y: 10, Z: 10},
				speed:  5,
				move:   5,
				hp:     10,
				maxHP:  10,
				weapon: &weaponShank,
				team:   1,
			},
			&Mob{
				name:   "big Kobold",
				glyph:  kobold2Glyph,
				loc:    Loc{Map: "test", X: 21, Y: 10, Z: 10},
				speed:  4,
				move:   4,
				hp:     20,
				maxMP:  20,
				weapon: &weaponShank,
				team:   1,
			},
		},
	}
	player := team.Units[0]
	game := &GameWindow{World: sesh.world, Char: player, Sesh: sesh}
	sesh.win = game
	sesh.PushWindow(game)
	sesh.world.apply <- ListenAction{listener: sesh}
	for _, mob := range team.Units {
		sesh.world.apply <- AddAction{Obj: mob}
	}
	for _, mob := range enemies.Units {
		sesh.world.apply <- AddAction{Obj: mob}
	}
	sesh.world.apply <- NextTurnAction{}

	io.WriteString(sesh.ssh, EnableMouseReporting)
}

func (sesh *Sesh) PushWindow(win Window) {
	sesh.ui = append(sesh.ui, win)
}

func (sesh *Sesh) cleanup() {
	sesh.world.apply <- PartAction{listener: sesh}
	sesh.world.apply <- RemoveAction(sesh.win.Char.ID())
	fmt.Println("disconnex")
}

func (sesh *Sesh) Run() {
	runSesh(sesh)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	world := newWorld()
	go world.Run()

	handleSSH(world)

	log.Println("starting ssh server on port 2222...")
	log.Fatal(ssh.ListenAndServe(":2222", nil))
}
