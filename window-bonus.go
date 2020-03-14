package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

type BonusWindow struct {
	World    *World
	Sesh     *Sesh
	Team     Team
	Bonuses  []Bonus
	choice   int
	callback func(int)
	done     bool
}

func (gw *BonusWindow) Render(scr [][]Glyph) {
	for i := 1; i <= 7; i++ {
		copyString(scr[len(scr)-i], "", true)
	}
	if gw.choice != -1 {
		copyString(scr[len(scr)-1], fmt.Sprintf(`Apply "%s" to %s? ENTER to confirm, ESC to cancel.`, gw.Bonuses[gw.choice].Name, gw.Team.Units[gw.choice].Name()), true)
		for i := 0; i < len(scr[len(scr)-1]); i++ {
			scr[len(scr)-1][i].BG = 53
		}
	} else {
		copyString(scr[len(scr)-1], "Press a, b, c, or d to pick a bonus.", true)
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 8, 4, 1, ' ', 0)
	fmt.Fprintln(w, "Which unit would you like to receive a bonus?")
	fmt.Fprintln(w) // stop vet from complaining about a trailing \n
	spellCount := 0
	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprintf(w, "%s", unit.Name())

		if spells := len(unit.Spells()); spells > spellCount && !unit.Dead() {
			spellCount = spells
		}
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprintf(w, "%s", unit.Class())
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if unit.Dead() {
			fmt.Fprint(w, "Dead")
		} else {
			fmt.Fprintf(w, "HP: %d", unit.MaxHP())
		}
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if unit.MaxMP() > 0 && !unit.Dead() {
			fmt.Fprintf(w, "MP: %d", unit.MaxMP())
		}
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if !unit.Dead() {
			fmt.Fprintf(w, "Speed: %d", unit.Speed())
		}
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if !unit.Dead() {
			fmt.Fprintf(w, "%s (%s)", unit.Weapon().Name, unit.Weapon().Damage.Dice.String())
		}
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if !unit.Dead() {
			fmt.Fprint(w, unit.Armor().String())
		}
	}
	fmt.Fprintln(w)

	for spell := 0; spell < spellCount; spell++ {
		for i := 0; i < len(gw.Team.Units); i++ {
			unit := gw.Team.Units[i]
			if i == 0 {
				// fmt.Fprint(w, " ")
			} else {
				fmt.Fprint(w, "\t")
			}
			if unit.Dead() {
				continue
			}
			if spell < len(unit.spells) {
				fmt.Fprint(w, "☆ ", unit.spells[spell].Name)
			}
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, strings.Repeat("\t", len(gw.Team.Units)-1))
	for i := 0; i < len(gw.Team.Units); i++ {
		// unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if gw.Team.Units[i].Dead() {
			continue
		}
		fmt.Fprintf(w, "%c) Bonus:", 'a'+i)
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		// unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if gw.Team.Units[i].Dead() {
			continue
		}
		fmt.Fprintf(w, "%s", gw.Bonuses[i].Name)
	}
	fmt.Fprintln(w)

	w.Flush()
	lines := strings.Split(buf.String(), "\n")
	lines = lines[:len(lines)-1]
	drawCenteredBox(scr, lines, 17)
}

func (gw *BonusWindow) Cursor() Coords {
	return OriginCoords // TODO
}

func (gw *BonusWindow) Input(input string) bool {
	if gw.done {
		return true
	}

	if len(input) == 1 {
		switch input[0] {
		case EscKey, 'n':
			gw.choice = -1
		case EnterKey, 'y':
			if gw.choice >= 0 && gw.choice < len(gw.Bonuses) {
				gw.World.ApplyBonus(gw.Bonuses[gw.choice], gw.Team.Units[gw.choice])
				gw.World.StartBattle(gw.World.level + 1)
				// TODO: show new stats?
				gw.done = true
			}
		default:
			i := int(input[0] - 'a')
			if i >= 0 && i < len(gw.Bonuses) {
				if !gw.Team.Units[i].Dead() {
					gw.choice = i
				}
			}
		}
	}

	return true
}

func (gw *BonusWindow) Click(_ Coords) bool {
	return true
}

func (gw *BonusWindow) ShouldRemove() bool {
	return gw.done
}

func (gw *BonusWindow) Mouseover(_ Coords) bool {
	return false
}
