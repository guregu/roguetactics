package main

import (
	"fmt"
)

type GameWindow struct {
	Sesh  *Sesh
	World *World
	Msgs  []string
	Team  int
	Map   *Map

	turnID int
	moved  bool
	acted  bool

	done bool
}

func (gw *GameWindow) close() {
	gw.done = true
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
			World:   gw.World,
			Sesh:    gw.Sesh,
			Char:    m,
			cursorX: -1,
			cursorY: -1,
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
	gw.World.pushBottom <- NextTurnState{}
	gw.moved = false
	gw.acted = false
	return true
}

func (gw *GameWindow) canDoSomething() bool {
	return !gw.moved || !gw.acted
}

func (gw *GameWindow) Render(scr [][]Glyph) {
	m := gw.Map
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
		y := len(scr) - 5 + i
		if n < 0 || n > len(gw.Msgs) {
			copyString(scr[y], "", true)
		} else {
			copyString(scr[y], gw.Msgs[n], true)
		}
	}
	up := gw.World.Up()
	if up != nil {
		if mob, ok := up.(*Mob); ok {
			copyGlyphs(scr[len(scr)-3], mob.StatusLine(), true)
		}
	}

	copyString(scr[len(scr)-2], "", true)

	turnInfo := fmt.Sprintf("[Turn: %d]", gw.World.turn)
	copyStringAlignRight(scr[len(scr)-3], turnInfo)

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
	return gw.done
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

	tile := gw.Map.TileAt(x, y)
	if mob, ok := tile.Top().(*Mob); ok {
		if mob.Team() != gw.Team {
			return gw.showAttack()
		}
	}

	gw.Msgs = append(gw.Msgs, fmt.Sprintf("Clicked: (%d,%d)", x, y))
	return true
}

func (gw *GameWindow) Mouseover(x, y int) bool {
	return false
}

type GameOverWindow struct {
	World *World
	Sesh  *Sesh
	done  bool
}

func (gw *GameOverWindow) Render(scr [][]Glyph) {
	lines := []string{"You were defeated!", "Game over.", "", "Press ENTER to start a new game."}
	drawCenteredBox(scr, lines, ColorDarkRed)
}

func (gw *GameOverWindow) Cursor() (x, y int) {
	return 0, 0 //TODO
}

func (gw *GameOverWindow) Input(input string) bool {
	switch input[0] {
	case 13: //ENTER
		gw.Sesh.PushWindow(&TitleWindow{World: gw.World, Sesh: gw.Sesh})
		gw.World.reset()
		gw.done = true
	}
	return true
}

func (gw *GameOverWindow) Click(x, y int) bool {
	return true
}

func (gw *GameOverWindow) ShouldRemove() bool {
	return gw.done
}

func (gw *GameOverWindow) Mouseover(x, y int) bool {
	return false
}

var (
	_ Window = (*GameWindow)(nil)
	_ Window = (*GameOverWindow)(nil)
)
