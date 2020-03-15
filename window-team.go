package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

type TeamWindow struct {
	World *World
	Sesh  *Sesh
	Team  Team
	Win   bool
	done  bool
}

func (gw *TeamWindow) Render(scr [][]Glyph) {
	if gw.Win {
		copyString(scr[len(scr)-3], "", true)
		copyString(scr[len(scr)-2], "Congratulations and thank you for playing!", true)
		copyString(scr[len(scr)-1], "Press ENTER to return to the title screen.", true)
	} else {
		copyString(scr[len(scr)-1], "Team summary: press TAB to switch teams, or ESC to exit.", true)
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 8, 4, 1, ' ', 0)
	if gw.Win {
		fmt.Fprintln(w, "You win! Final score:", gw.World.score)
	}
	spellCount := 0
	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprintf(w, "%s", unit.Name())

		if spells := len(unit.Spells()); spells > spellCount {
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
		fmt.Fprintf(w, "HP: %d/%d", unit.HP(), unit.MaxHP())
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if unit.MaxMP() > 0 {
			fmt.Fprintf(w, "MP: %d/%d", unit.MP(), unit.MaxMP())
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
		fmt.Fprintf(w, "Speed: %d", unit.Speed())
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprintf(w, "CT: %d", unit.CT())
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprintf(w, "%s (%s)", unit.Weapon().Name, unit.Weapon().Damage.Dice.String())
	}
	fmt.Fprintln(w)

	for i := 0; i < len(gw.Team.Units); i++ {
		unit := gw.Team.Units[i]
		if i == 0 {
			// fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, "\t")
		}
		if unit.Armor().Name != "" {
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
			if spell < len(unit.spells) {
				fmt.Fprint(w, "☆ ", unit.spells[spell].Name)
			}
		}
		fmt.Fprintln(w)
	}

	w.Flush()
	lines := strings.Split(buf.String(), "\n")
	lines = lines[:len(lines)-1]
	color := ColorNavy
	if gw.Team.ID == AITeam {
		color = Color256(52)
	}
	drawCenteredBox(scr, lines, color)
}

func (gw *TeamWindow) Cursor() Coords {
	return OriginCoords //TODO
}

func (gw *TeamWindow) Input(input string) bool {
	if gw.Win {
		if input[0] == 13 {
			gw.Sesh.PushWindow(&TitleWindow{World: gw.World, Sesh: gw.Sesh})
			gw.World.reset()
			gw.done = true
		}
		return true
	}

	if len(input) == 1 {
		switch input[0] {
		case EscKey, EnterKey:
			gw.done = true
		case TabKey: // tab
			if gw.Team.ID == PlayerTeam {
				gw.Team = gw.World.battle.Teams[AITeam]
			} else {
				gw.Team = gw.World.battle.Teams[PlayerTeam]
			}
		}
	}

	return true
}

func (gw *TeamWindow) Click(_ Coords) bool {
	return true
}

func (gw *TeamWindow) ShouldRemove() bool {
	return gw.done
}

func (gw *TeamWindow) Mouseover(_ Coords) bool {
	return false
}
