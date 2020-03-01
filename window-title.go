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
	copyString(scr[0], "Tactics: Roguer", true)
	copyString(scr[1], " ", true)
	copyString(scr[2], " ", true)
	copyString(scr[3], "Press ENTER to start a new game!", true)
	for i := 4; i < len(scr); i++ {
		copyString(scr[i], "", true)
	}
}

func (mw *TitleWindow) Cursor() (x, y int) {
	return 0, 0 //TODO
}

func (mw *TitleWindow) Input(input string) bool {
	switch input[0] {
	case 13: //ENTER
		battle := newBattle("dojo", mw.World.player)
		mw.World.apply <- StartBattleAction{Battle: battle}
		mw.done = true
	}
	return true
}

func (mw *TitleWindow) Click(x, y int) bool {
	return true
}

func (mw *TitleWindow) ShouldRemove() bool {
	return mw.done
}
