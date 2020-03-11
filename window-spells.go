package main

import (
	"fmt"
)

type SpellsWindow struct {
	World    *World
	Sesh     *Sesh
	Char     *Mob
	callback func(int)
	done     bool
}

func (gw *SpellsWindow) Render(scr [][]Glyph) {
	spells := gw.Char.Spells()
	var lines = []string{"Which spell to cast? (ESC to cancel)", ""}
	for i, spell := range spells {
		opt := string(rune('a' + i))
		lines = append(lines, fmt.Sprintf("%s) %s (%d MP)", opt, spell.Name, spell.MPCost))
	}
	drawCenteredBox(scr, lines, 53)
}

func (gw *SpellsWindow) Cursor() Coords {
	return Coords{0, 0} //TODO
}

func (gw *SpellsWindow) Input(input string) bool {
	if len(input) == 1 {
		switch input[0] {
		case 27: // ESC
			gw.done = true
		default:
			i := int(input[0] - 'a')
			spells := gw.Char.Spells()
			if i >= 0 && i < len(spells) {
				if spells[i].MPCost > gw.Char.MP() {
					gw.Sesh.Bell()
					gw.Sesh.Send(fmt.Sprintf("Not enough MP to cast %s.", spells[i].Name))
					return true
				}
				gw.callback(i)
				gw.done = true
			}
		}
	}
	return true
}

func (gw *SpellsWindow) Click(_ Coords) bool {
	return true
}

func (gw *SpellsWindow) ShouldRemove() bool {
	return gw.done
}

func (gw *SpellsWindow) Mouseover(_ Coords) bool {
	return false
}
