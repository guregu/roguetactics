package main

import (
// "fmt"
)

type TitleWindow struct {
	World *World
	Sesh  *Sesh

	done bool
}

func (mw *TitleWindow) Render(scr [][]Glyph) {
	for i := 0; i < len(scr); i++ {
		copyString(scr[i], "", true)
	}
	copyString(scr[0], "  Bitesize Tactics", true)
	for i := 0; i < len("Bitesize Tactics"); i++ {
		scr[0][i+2].Underline = true
	}
	copyString(scr[1], "      a 7DRL by Kawaii Solutions", true)
	copyString(scr[2], " ", true)

	copyString(scr[4], "       A group of four daring adventurers descends into the dungeon...", true)
	copyString(scr[6], "       Danger lurks within.", true)
	copyString(scr[5], "       Will they claim the golden throne?", true)
	copyString(scr[7], "       Lead them to victory... or a crushing defeat.", true)

	copyString(scr[14], " * Click on this screen to focus it.", true)
	copyString(scr[15], " * Then press ENTER to start a new game!", true)
	copyString(scr[16], " ↓ Read the guide on this page below to learn how to play.", true)

	copyString(scr[18], " (Note to 7DRL judges: see description for original 7DRL version)", true)

	copyStringAlignRight(scr[len(scr)-2], "Twitter: @kawaiisolutions ")
	copyString(scr[len(scr)-1], "Press ENTER to start!", false)
	for i := 0; i < len("Press ENTER to start!"); i++ {
		scr[len(scr)-1][i].Blink = true
	}
	copyStringAlignRight(scr[len(scr)-1], "© Kawaii Solutions 2020")
}

func (mw *TitleWindow) Cursor() Coords {
	return OriginCoords //TODO
}

func (mw *TitleWindow) Input(input string) bool {
	switch input[0] {
	case EnterKey:
		mw.World.StartBattle(0)
		mw.done = true
	}
	return true
}

func (mw *TitleWindow) Click(_ Coords) bool {
	return true
}

func (mw *TitleWindow) ShouldRemove() bool {
	return mw.done
}

func (gw *TitleWindow) Mouseover(_ Coords) bool {
	return false
}

var (
	_ Window = (*TitleWindow)(nil)
)
