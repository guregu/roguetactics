package main

import (
	"fmt"
)

type Armor struct {
	Name       string
	Defense    int
	MPRecovery int
	Value      int
}

var armorLeather = Armor{
	Name:    "tunic",
	Defense: 1,
}
var armorLeather2 = Armor{
	Name:    "jerkin",
	Defense: 2,
	Value:   1,
}

var armorRobe = Armor{
	Name:       "robe",
	MPRecovery: 1,
	Defense:    0,
}
var armorFineRobe = Armor{
	Name:       "fine robe",
	MPRecovery: 3,
	Defense:    1,
	Value:      1,
}
var armorHolyRobe = Armor{
	Name:       "holy robe",
	MPRecovery: 5,
	Defense:    2,
	Value:      2,
}
var armorWizardHat = Armor{
	Name:       "pointy hat",
	MPRecovery: 8,
	Defense:    0,
	Value:      2,
}

var armorChainmail = Armor{
	Name:    "chainmail",
	Defense: 3,
	Value:   2,
}
var armorPlate = Armor{
	Name:    "platemail",
	Defense: 5,
	Value:   3,
}

func (a Armor) String() string {
	var info string
	if a.Defense != 0 {
		info = fmt.Sprintf("%dAC", a.Defense)
	}
	if a.MPRecovery > 0 {
		if info != "" {
			info += " "
		}
		info += fmt.Sprintf("%dMP/t", a.MPRecovery)
	}
	return fmt.Sprintf("%s (%s)", a.Name, info)
}
