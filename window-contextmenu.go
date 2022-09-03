package main

import (
	"strings"
)

type ContextMenu struct {
	world    *World
	sesh     *Sesh
	anchor   Coords
	options  []MenuItem
	selected int

	width   int
	height  int
	topLeft Coords

	done bool
}

type MenuItem struct {
	text   string
	action func()
}

func newContextMenu(world *World, sesh *Sesh, anchor Coords, options []MenuItem) *ContextMenu {
	options = append(options, MenuItem{
		text:   "Cancel",
		action: func() {},
	})

	cm := &ContextMenu{
		world:    world,
		sesh:     sesh,
		anchor:   anchor,
		options:  options,
		selected: 1,
	}

	cm.height = len(cm.options) + 2
	for _, opt := range cm.options {
		if len(opt.text) > cm.width {
			cm.width = len(opt.text)
		}
	}
	cm.topLeft = Coords{x: cm.anchor.x + 1, y: cm.anchor.y - 1}
	cm.topLeft.EnsureWithinBounds(sesh.disp.w, sesh.disp.h)
	if cm.topLeft.x+cm.width+2 >= sesh.disp.w {
		cm.topLeft.x = cm.anchor.x - cm.width - 2
	}
	if cm.topLeft.y+cm.height >= sesh.disp.h {
		cm.topLeft.y = sesh.disp.h - cm.height
	}
	return cm
}

func (cm *ContextMenu) Render(scr [][]Glyph) {
	const (
		boxH  = "═"
		boxV  = "║"
		boxNW = "╔"
		boxNE = "╗"
		boxSW = "╚"
		boxSE = "╝"
		bg    = Color256(234)
	)

	// target status bar
	if cm.world.current != nil {
		tile := cm.world.current.TileAt(cm.anchor.x, cm.anchor.y)
		if !tile.IsValid() {
			return
		}
		target, ok := tile.Top().(*Mob)
		if !ok {
			return
		}
		status := append(GlyphsOf(" └"), target.StatusLine(true)...)
		copyGlyphs(scr[len(scr)-2], status, true)
	}

	// menu
	copyStringOffset(scr[cm.topLeft.y], boxNW+strings.Repeat(boxH, cm.width)+boxNE, cm.topLeft.x)
	ApplyStyle(scr[cm.topLeft.y][cm.topLeft.x:cm.topLeft.x+cm.width+2], StyleBG(bg))
	for i, opt := range cm.options {
		text := boxV + opt.text + strings.Repeat(" ", cm.width-len(opt.text)) + boxV
		copyStringOffset(scr[cm.topLeft.y+i+1], text, cm.topLeft.x)
		if cm.selected == i+1 {
			ApplyStyle(scr[cm.topLeft.y+i+1][cm.topLeft.x+1:cm.topLeft.x+cm.width+1], StyleReverse)
		} else {
			ApplyStyle(scr[cm.topLeft.y+i+1][cm.topLeft.x:cm.topLeft.x+cm.width+2], StyleBG(bg))
		}
	}
	copyStringOffset(scr[cm.topLeft.y+cm.height-1], boxSW+strings.Repeat(boxH, cm.width)+boxSE, cm.topLeft.x)
	ApplyStyle(scr[cm.topLeft.y+cm.height-1][cm.topLeft.x:cm.topLeft.x+cm.width+2], StyleBG(bg))
}

func (cm *ContextMenu) Cursor() Coords {
	return cm.anchor
}

func (cm *ContextMenu) Input(in string) bool {
	switch in {
	case string(EscKey):
		cm.done = true
	case string(EnterKey):
		cm.activate()
	case ArrowKeyUp:
		cm.selected--
	case ArrowKeyDown:
		cm.selected++
	default:
		return true
	}
	if cm.selected <= 0 {
		cm.selected = len(cm.options)
	}
	if cm.selected > len(cm.options) {
		cm.selected = 1
	}
	return true
}

func (cm *ContextMenu) Click(coords Coords) bool {
	// oob click = dismiss
	if coords.y < cm.topLeft.y || coords.y >= cm.topLeft.y+cm.height-1 ||
		coords.x < cm.topLeft.x || coords.x > cm.topLeft.x+cm.width+1 {
		cm.done = true
		return true
	}
	if cm.selected == 0 {
		return true
	}
	cm.activate()
	return true
}

func (cm *ContextMenu) activate() {
	opt := cm.options[cm.selected-1]
	if opt.action != nil {
		opt.action()
		cm.done = true
	}
}

func (cm *ContextMenu) Mouseover(coords Coords) bool {
	if coords.y < cm.topLeft.y || coords.y >= cm.topLeft.y+cm.height-1 ||
		coords.x < cm.topLeft.x || coords.x > cm.topLeft.x+cm.width+1 {
		cm.selected = 0
		return true
	}

	cm.selected = coords.y - cm.topLeft.y
	return true
}

func (cm *ContextMenu) ShouldRemove() bool {
	return cm.done
}
