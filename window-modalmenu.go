package main

import (
	"strings"
)

// TODO: this (WIP)

type ModalMenu struct {
	world *World
	sesh  *Sesh

	prompt string

	options  []MenuItem
	selected int

	width        int
	height       int
	topLeft      Coords
	promptHeight int

	done bool
}

func newModalMenu(world *World, sesh *Sesh, prompt string, options []MenuItem) *ModalMenu {
	cm := &ModalMenu{
		world:   world,
		sesh:    sesh,
		prompt:  prompt,
		options: options,
		// selected: 1,
	}
	cm.promptHeight = strings.Count(prompt, "\n") + 1
	cm.height = cm.promptHeight + len(cm.options) + 2
	const extraWidth = len("a) ")
	// cm.width = len("ESC) Cancel")
	for _, line := range strings.Split(prompt, "\n") {
		if len(line) > cm.width {
			cm.width = len(line)
		}
	}
	for _, opt := range cm.options {
		if len(opt.text)+extraWidth > cm.width {
			cm.width = len(opt.text) + extraWidth
		}
	}
	dw, dh := sesh.disp.w, sesh.disp.h
	cm.topLeft = Coords{x: dw/2 - cm.width/2, y: dh/2 - cm.height/2}
	cm.topLeft.EnsureWithinBounds(dw, dh)
	// if cm.topLeft.x+cm.width+2 >= sesh.disp.w {
	// 	cm.topLeft.x = cm.anchor.x - cm.width - 2
	// }
	// if cm.topLeft.y+cm.height >= sesh.disp.h {
	// 	cm.topLeft.y = sesh.disp.h - cm.height
	// }
	return cm
}

func (cm *ModalMenu) Render(scr [][]Glyph) {
	const (
		boxH  = "═"
		boxV  = "║"
		boxNW = "╔"
		boxNE = "╗"
		boxSW = "╚"
		boxSE = "╝"
		bg    = Color256(53)
	)

	// menu
	y := cm.topLeft.y
	copyStringOffset(scr[y], boxNW+strings.Repeat(boxH, cm.width)+boxNE, cm.topLeft.x)
	ApplyStyle(scr[y][cm.topLeft.x:cm.topLeft.x+cm.width+2], StyleBG(bg))
	y++
	for _, line := range strings.Split(cm.prompt, "\n") {
		text := boxV + line + strings.Repeat(" ", cm.width-len(line)) + boxV
		copyStringOffset(scr[y], text, cm.topLeft.x)
		ApplyStyle(scr[y][cm.topLeft.x:cm.topLeft.x+cm.width+2], StyleBG(bg))
		y++
	}
	// ApplyStyle(scr[cm.topLeft.y][cm.topLeft.x:cm.topLeft.x+cm.width+2], StyleBG(bg))
	for i, opt := range cm.options {
		text := boxV + string('a'+i) + ") " + opt.text + strings.Repeat(" ", cm.width-len(opt.text)-len("a) ")) + boxV
		copyStringOffset(scr[y], text, cm.topLeft.x)
		if cm.selected == i+1 {
			ApplyStyle(scr[y][cm.topLeft.x+1:cm.topLeft.x+cm.width+1], StyleReverse)
		} else {
			ApplyStyle(scr[y][cm.topLeft.x:cm.topLeft.x+cm.width+2], StyleBG(bg))
		}
		y++
	}
	copyStringOffset(scr[y], boxSW+strings.Repeat(boxH, cm.width)+boxSE, cm.topLeft.x)
	ApplyStyle(scr[y][cm.topLeft.x:cm.topLeft.x+cm.width+2], StyleBG(bg))
}

func (cm *ModalMenu) Cursor() Coords {
	return OriginCoords
}

func (cm *ModalMenu) Input(in string) bool {
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

func (cm *ModalMenu) Click(coords Coords) bool {
	if cm.selected == 0 {
		return true
	}
	cm.activate()
	return true
}

func (cm *ModalMenu) activate() {
	opt := cm.options[cm.selected-1]
	if opt.action != nil {
		opt.action()
		cm.done = true
	}
}

func (cm *ModalMenu) Mouseover(coords Coords) bool {
	if coords.y < cm.topLeft.y || coords.y >= cm.topLeft.y+cm.height-1 ||
		coords.x < cm.topLeft.x || coords.x > cm.topLeft.x+cm.width+1 {
		cm.selected = 0
		return true
	}

	cm.selected = coords.y - cm.topLeft.y - cm.promptHeight
	return true
}

func (cm *ModalMenu) ShouldRemove() bool {
	return cm.done
}
