package main

import (
	"fmt"
	"io"
	"log"
	"strings"

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
	ssh   ssh.Session
	disp  *Display
}

func NewSesh(s ssh.Session, w *World) *Sesh {
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
		for i := len(sesh.ui) - 1; i >= 0; i-- {
			win := sesh.ui[i]
			if win.Click(x, y) {
				return
			}
		}
	}

	for i := len(sesh.ui) - 1; i >= 0; i-- {
		win := sesh.ui[i]
		if win.Input(input) {
			return
		}
	}
	// sesh.win.Input(input)
}

func (sesh *Sesh) removeWindows() {
	for i := len(sesh.ui) - 1; i >= 0; i-- {
		if !sesh.ui[i].ShouldRemove() {
			continue
		}
		if i < len(sesh.ui)-1 {
			copy(sesh.ui[i:], sesh.ui[i+1:])
		}
		sesh.ui[len(sesh.ui)-1] = nil // or the zero vsesh.uilue of T
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
		sesh.setCursor(sesh.ui[len(sesh.ui)-1].Cursor())
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

func (sesh *Sesh) setup() {
	glyph := GlyphOf('@')
	glyph.FG = ColorRed

	player := &Mob{
		name:  sesh.ssh.User(),
		glyph: glyph,
		loc:   Loc{Map: "test", X: 10, Y: 10, Z: 10},
	}
	game := &GameWindow{World: sesh.world, Char: player, Sesh: sesh}
	sesh.win = game
	sesh.PushWindow(game)
	// io.WriteString(sesh.ssh, "\033[?9h") // mouse on
	sesh.world.apply <- ListenAction{listener: sesh}
	sesh.world.apply <- AddAction{Obj: player}

	io.WriteString(sesh.ssh, EnableMouseReporting)
	sesh.redraw()
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
	ptyReq, winCh, isPty := sesh.ssh.Pty()
	_ = ptyReq

	if !isPty {
		io.WriteString(sesh.ssh, "No PTY requested.\n")
		sesh.ssh.Exit(1)
		return
	}

	defer sesh.cleanup()

	go func() {
		for win := range winCh {
			sesh.resize(win)
		}
	}()
	sesh.setup()
	buf := make([]byte, 256)
	for {
		n, err := sesh.ssh.Read(buf[:])
		if err != nil {
			fmt.Println("Error: 1", err)
			sesh.ssh.Exit(1)
			break
		}
		if n > 0 {
			sesh.do(string(buf[:n]))
			fmt.Println("GOT:", buf[:n], ">>>", strings.ReplaceAll(string(buf[:n]), "\033", "ESC"))
		}
	}
}

func main() {
	world := newWorld()
	go world.Run()

	ssh.Handle(func(s ssh.Session) {
		sesh := NewSesh(s, world)
		io.WriteString(sesh.ssh, resetScreen+cursorTo00)
		sesh.Run()
	})

	log.Println("starting ssh server on port 2222...")
	log.Fatal(ssh.ListenAndServe(":2222", nil))
}
