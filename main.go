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
)

var mainMap *Map

type Sesh struct {
	x, y  int
	world *World
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
	switch input {
	case "Q":
		sesh.ssh.Exit(0)
		return
	case "R":
		sesh.redraw()
		return
	}
	sesh.win.Input(input)
}

func (sesh *Sesh) refresh() {
	sesh.win.Render(sesh.disp.nextFrame())
	render := sesh.disp.diff()
	if render == "" {
		return
	}
	fmt.Println("Render: ", strings.Replace(render, "\033", "ESC", -1))
	io.WriteString(sesh.ssh, render)
	sesh.setCursor(sesh.win.Cursor())
}

func (sesh *Sesh) redraw() {
	sesh.win.Render(sesh.disp.nextFrame())
	io.WriteString(sesh.ssh, sesh.disp.full())
	sesh.setCursor(sesh.win.Cursor())
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
	// io.WriteString(sesh.ssh, "\033[?9h") // mouse on
	sesh.world.apply <- ListenAction{listener: sesh}
	sesh.world.apply <- AddAction{Obj: player}
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
	buf := make([]byte, 3)
	for {
		n, err := sesh.ssh.Read(buf)
		if err != nil {
			fmt.Println("Error: 1", err)
			sesh.ssh.Exit(1)
			break
		}
		if n > 0 {
			sesh.do(string(buf[:n]))
			fmt.Println("GOT:", buf[:n], ">>>", string(buf[:n]))
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
