package main

import (
	"fmt"
	"io"
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
	EnableMouseReporting = "\033[?1003h"
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
	if strings.HasPrefix(input, MousePrefix) && len(input) >= 6 {
		x := int(input[4] - 32 - 1)
		y := int(input[5] - 32 - 1)
		switch input[3] {
		case 35:
			sesh.world.apply <- ClickAction{UI: sesh.ui, X: x, Y: y, Sesh: sesh}
		case 67:
			sesh.world.apply <- MouseoverAction{UI: sesh.ui, X: x, Y: y, Sesh: sesh}
		}
		// for i := len(sesh.ui) - 1; i >= 0; i-- {
		// 	win := sesh.ui[i]
		// 	if win.Click(x, y) {
		// 		return
		// 	}
		// }
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
	if len(sesh.ui) == 0 {
		return
	}

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
	if len(sesh.ui) == 0 {
		return
	}

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
	sesh.PushWindow(&TitleWindow{World: sesh.world, Sesh: sesh})
	io.WriteString(sesh.ssh, EnableMouseReporting)
	sesh.world.apply <- ListenAction{listener: sesh}
}

func (sesh *Sesh) PushWindow(win Window) {
	sesh.ui = append(sesh.ui, win)
}

func (sesh *Sesh) cleanup() {
	sesh.world.apply <- PartAction{listener: sesh}
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
	listenAndWait()
}
