package main

type OverworldWindow struct {
	World *World
	Sesh  *Sesh

	done bool
}

func (mw *OverworldWindow) Render(scr [][]Glyph) {
	copyString(scr[0], "TODO: world map", true)
	copyString(scr[1], " ", true)
	copyString(scr[2], " ", true)
	copyString(scr[3], "Press ENTER to start a new game!", true)
}

func (mw *OverworldWindow) Cursor() (x, y int) {
	return 0, 0 //TODO
}

func (mw *OverworldWindow) Input(input string) bool {
	switch input[0] {
	case 13: //ENTER
		battle := newBattle("dojo", mw.World.player)
		mw.World.apply <- StartBattleAction{Battle: battle}
		mw.done = true
	}
	return true
}

func (mw *OverworldWindow) Click(x, y int) bool {
	return true
}

func (mw *OverworldWindow) ShouldRemove() bool {
	return mw.done
}
