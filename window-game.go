package main

import (
	"fmt"
)

type GameWindow struct {
	Sesh  *Sesh
	World *World
	Msgs  [][]Glyph
	Team  int
	Map   *Map

	turnID   int
	moved    bool
	acted    bool
	startLoc Loc

	done bool
}

func (gw *GameWindow) close() {
	gw.done = true
}

func (gw *GameWindow) Input(in string) bool {
	// if in[0] == 13 {
	// 	// enter key
	// 	gw.Sesh.PushWindow(&ChatWindow{prompt: "Chat: "})
	// 	return true
	// }
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
	case "c":
		return gw.showCast()
	case "q", ";":
		gw.Sesh.PushWindow(&FarlookWindow{
			World:         gw.World,
			Sesh:          gw.Sesh,
			Char:          gw.World.Up(),
			cursorHandler: newCursorHandlerOn(gw.World, gw.World.Up()),
		})
	case "r":
		if !gw.moved {
			return true
		}
		if gw.acted {
			return true
		}
		if mob, ok := gw.World.Up().(*Mob); ok {
			m := gw.World.Map(mob.Loc().Map)
			m.Move(mob, gw.startLoc.X, gw.startLoc.Y)
			gw.moved = false
		}
	case "t", "i", "\t":
		gw.Sesh.PushWindow(&TeamWindow{
			World: gw.World,
			Sesh:  gw.Sesh,
			Team:  gw.World.battle.Teams[PlayerTeam],
		})
	case "W":
		// gw.World.winBattle()
	}

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
		gw.startLoc = m.Loc()
		gw.Sesh.PushWindow(&MoveWindow{
			World:         gw.World,
			Sesh:          gw.Sesh,
			Char:          m,
			Range:         m.MoveRange(),
			cursorHandler: newCursorHandlerOn(gw.World, m),
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
			World:         gw.World,
			Sesh:          gw.Sesh,
			Char:          m,
			Weapon:        m.Weapon(),
			cursorHandler: newCursorHandlerOn(gw.World, m),
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

func (gw *GameWindow) showCast() bool {
	if gw.acted {
		return true
	}
	up := gw.World.Up()
	if m, ok := up.(*Mob); ok {
		spells := m.Spells()
		if len(spells) == 0 {
			return true
		}
		gw.Sesh.PushWindow(&SpellsWindow{
			World: gw.World,
			Sesh:  gw.Sesh,
			Char:  m,
			callback: func(i int) {
				loc := m.Loc()

				gw.Sesh.PushWindow(&AttackWindow{
					World:         gw.World,
					Sesh:          gw.Sesh,
					Char:          m,
					Weapon:        spells[i],
					Self:          true,
					cursorHandler: newCursorHandler(loc.AsCoords(), gw.World.Map(loc.Map)),
					callback: func(acted bool) {
						if acted {
							gw.acted = true
							if !gw.canDoSomething() {
								gw.nextTurn()
							}
						}
					}})
			},
		})
	}
	return true
}

func (gw *GameWindow) nextTurn() bool {
	up := gw.World.Up()
	if m, ok := up.(*Mob); ok {
		if m.Team() != gw.Team {
			return true
		}
		m.FinishTurn(gw.World, gw.moved, gw.acted)
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
	// render map
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

	// render party status
	for i := 0; i < len(gw.World.player.Units); i++ {
		unit := gw.World.player.Units[i]
		name := unit.NameColored()
		if gw.World.Up() == unit {
			ApplyStyle(name, StyleUnderline)
		}
		copyGlyphs(scr[1+i*4], name, false)
		copyString(scr[1+i*4+1], string(unit.Class()), false)
		copyGlyphs(scr[1+i*4+2], Concat("HP: ", unit.HPText()), false)
	}

	// render combat log
	const chatLines = 4
	const bottomUILines = 3 // target info, help etc
	for i := 0; i < chatLines; i++ {
		n := len(gw.Msgs) - chatLines + i
		y := len(scr) - chatLines - bottomUILines + i
		if n < 0 || n > len(gw.Msgs) {
			copyString(scr[y], "", true)
		} else {
			copyGlyphs(scr[y], gw.Msgs[n], true)
		}
	}

	// render current unit status
	up := gw.World.Up()
	if up != nil {
		if mob, ok := up.(*Mob); ok {
			copyGlyphs(scr[len(scr)-3], mob.StatusLine(false), true)
		}
	}

	copyString(scr[len(scr)-2], "", true)

	turnInfo := fmt.Sprintf("[Turn: %d]", gw.World.turn)
	copyStringAlignRight(scr[0], turnInfo)

	if gw.World.Busy() {
		helpBar := "Busy..."
		copyString(scr[len(scr)-1], helpBar, true)
		return
	}

	// render help bar
	var helpBar string
	pushHelp := func(str string) {
		if len(helpBar) > 0 {
			helpBar += " "
		}
		helpBar += str
	}
	if !gw.moved {
		pushHelp("m) Move")
	} else if !gw.acted {
		pushHelp("r) Reset move")
	}
	if !gw.acted {
		pushHelp("a) Attack")
		if mob, ok := up.(*Mob); ok && len(mob.Spells()) > 0 {
			pushHelp("c) Cast spell")
		}
	}
	pushHelp("q) Query t) Team info")
	pushHelp("n) Next turn")
	copyString(scr[len(scr)-1], helpBar, true)
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

func (gw *GameWindow) ShouldRemove() bool {
	return gw.done
}

func (gw *GameWindow) Click(click Coords) bool {
	if !gw.myTurn() {
		return true
	}

	up := gw.World.Up()
	uploc := up.Loc()
	if uploc.X == click.x && uploc.Y == click.y {
		return gw.showMove()
	}

	tile := gw.Map.TileAt(click.x, click.y)
	if mob, ok := tile.Top().(*Mob); ok {
		if mob.Team() != gw.Team {
			return gw.showAttack()
		}
	}
	return true
}

func (gw *GameWindow) Mouseover(_ Coords) bool {
	return false
}

type GameOverWindow struct {
	World *World
	Sesh  *Sesh
	done  bool
}

func (gw *GameOverWindow) Render(scr [][]Glyph) {
	score := fmt.Sprintf("Score: %d", gw.World.score)
	lines := []string{"You were defeated!", "Game over.", score, "", "Press ENTER to return to the title screen."}
	drawCenteredBox(scr, lines, ColorDarkRed)
}

func (gw *GameOverWindow) Cursor() Coords {
	return OriginCoords //TODO
}

func (gw *GameOverWindow) Input(input string) bool {
	switch input[0] {
	case EnterKey:
		gw.Sesh.PushWindow(&TitleWindow{World: gw.World, Sesh: gw.Sesh})
		gw.World.reset()
		gw.done = true
	}
	return true
}

func (gw *GameOverWindow) Click(_ Coords) bool {
	return true
}

func (gw *GameOverWindow) ShouldRemove() bool {
	return gw.done
}

func (gw *GameOverWindow) Mouseover(_ Coords) bool {
	return false
}

type VictoryWindow struct {
	World *World
	Sesh  *Sesh
	done  bool
}

func (gw *VictoryWindow) Render(scr [][]Glyph) {
	var lines []string
	if gw.World.level+1 >= len(mapsByLevel) {
		lines = []string{"Victory!", "You seize the golden throne!", "The dungeon is now yours...", "", "Press ENTER to continue."}
	} else {
		lines = []string{"Victory!", "You defeated the enemies and descend deeper...", "", "Press ENTER to continue."}
	}
	drawCenteredBox(scr, lines, ColorNavy)
}

func (gw *VictoryWindow) Cursor() Coords {
	return OriginCoords //TODO
}

func (gw *VictoryWindow) Input(input string) bool {
	if gw.done {
		return false
	}

	switch input[0] {
	case EnterKey:
		if gw.World.level+1 >= len(mapsByLevel) {
			gw.Sesh.PushWindow(&GameWonWindow{
				World: gw.World,
				Sesh:  gw.Sesh,
			})
			gw.done = true
			return true
		}

		bonuses := generateBonuses(gw.World.player, gw.World.level)
		gw.Sesh.PushWindow(&BonusWindow{
			World:   gw.World,
			Sesh:    gw.Sesh,
			Team:    gw.World.player,
			Bonuses: bonuses,
			choice:  -1,
		})
		gw.done = true
	}
	return true
}

func (gw *VictoryWindow) Click(_ Coords) bool {
	return true
}

func (gw *VictoryWindow) ShouldRemove() bool {
	return gw.done
}

func (gw *VictoryWindow) Mouseover(_ Coords) bool {
	return false
}

type GameWonWindow struct {
	World *World
	Sesh  *Sesh
	done  bool
}

func (gw *GameWonWindow) Render(scr [][]Glyph) {
	score := fmt.Sprintf("Score: %d", gw.World.score)
	lines := []string{"You win!", "Congratulations, you won the game.", score, "", "Press ENTER to see your final stats."}
	drawCenteredBox(scr, lines, ColorNavy)
}

func (gw *GameWonWindow) Cursor() Coords {
	return OriginCoords //TODO
}

func (gw *GameWonWindow) Input(input string) bool {
	if gw.done {
		return false
	}

	switch input[0] {
	case EnterKey:
		gw.Sesh.PushWindow(&TeamWindow{World: gw.World, Sesh: gw.Sesh, Win: true, Team: gw.World.player})
		gw.done = true
	}
	return true
}

func (gw *GameWonWindow) Click(_ Coords) bool {
	return true
}

func (gw *GameWonWindow) ShouldRemove() bool {
	return gw.done
}

func (gw *GameWonWindow) Mouseover(_ Coords) bool {
	return false
}

var (
	_ Window = (*GameWindow)(nil)
	_ Window = (*GameOverWindow)(nil)
)

var (
	_ Window = (*GameWindow)(nil)
	_ Window = (*GameOverWindow)(nil)
)
