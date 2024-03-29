package main

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
	// "git.sr.ht/~mna/zzterm" // TODO: use this instead of parsing ansi seqs manually
)

const (
	TabKey        = 9
	EnterKey      = 13
	EscKey        = 27
	BackspaceKey  = 127
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
	world *World
	ui    []Window
	win   *GameWindow
	ssh   Conn
	disp  *Display

	cursor Coords
}

func NewSesh(s Conn, w *World) *Sesh {
	return &Sesh{
		world: w,
		ssh:   s,
		disp:  NewDisplay(80, 27),
	}
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
		cursor := sesh.ui[len(sesh.ui)-1].Cursor()
		if sesh.cursor != cursor {
			sesh.renderCursor(cursor)
			sesh.cursor = cursor
		}
		return
	}
	// fmt.Println("Render: ", strings.Replace(render, "\033", "ESC", -1))
	io.WriteString(sesh.ssh, render)
	sesh.renderCursor(sesh.ui[len(sesh.ui)-1].Cursor())
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
	sesh.renderCursor(sesh.ui[len(sesh.ui)-1].Cursor())
}

func (sesh *Sesh) renderCursor(coords Coords) {
	io.WriteString(sesh.ssh, fmt.Sprintf("\033[%d;%dH", coords.y+1, coords.x+1))
}

func (sesh *Sesh) Send(msg []Glyph) {
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
	// fmt.Println("disconnex")
}

func (sesh *Sesh) Run() {
	defer sesh.cleanup()
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
	rand.Seed(time.Now().UnixNano())

	world := newWorld()
	go world.Run()

	handleSSH(world)
	listenAndWait()
}
