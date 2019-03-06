package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

type Map struct {
	Name    string
	Tiles   [][]*Tile
	Objects map[ID]Object
}

func (m *Map) TileAtLoc(loc Loc) *Tile {
	return m.TileAt(loc.X, loc.Y)
}

func (m *Map) TileAt(x, y int) *Tile {
	// TODO bounds check
	return m.Tiles[y][x]
}

func (m *Map) Add(obj Object) {
	m.Objects[obj.ID()] = obj
	m.TileAtLoc(obj.Loc()).Add(obj)
}

func (m *Map) Remove(obj Object) {
	delete(m.Objects, obj.ID())
	m.TileAtLoc(obj.Loc()).Remove(obj)
}

func (m *Map) Move(obj Object, x, y int) {
	if x < 0 || y < 0 {
		return
	}
	loc := obj.Loc()
	m.TileAtLoc(loc).Remove(obj)
	loc.X = x
	loc.Y = y
	obj.Move(loc)
	m.TileAtLoc(loc).Add(obj)
}

func (m *Map) Contains(id ID) bool {
	_, ok := m.Objects[id]
	return ok
}

type Tile struct {
	Ground   Glyph
	Objects  map[ID]Object
	Collides bool
}

func (t *Tile) Add(obj Object) {
	t.Objects[obj.ID()] = obj
}

func (t *Tile) Remove(obj Object) {
	delete(t.Objects, obj.ID())
}

func (t *Tile) Top() Object {
	z := -1
	var obj Object
	for _, o := range t.Objects {
		loc := o.Loc()
		if loc.Z > z {
			obj = o
			z = loc.Z
		}
	}
	return obj
}

func (t *Tile) Glyph() Glyph {
	if top := t.Top(); top != nil {
		return top.Glyph()
	}
	return t.Ground
}

func loadMap(name string) (*Map, error) {
	filename := filepath.Join("maps", name)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	m := &Map{
		Name:    "test",
		Objects: make(map[ID]Object),
	}
	for {
		line, _, err := r.ReadLine()
		var tline []*Tile
		if len(line) > 0 {
			for _, r := range string(line) {
				glyph := GlyphOf(r)
				collides := false
				if r == '.' || r == '#' {
					glyph.FG = ColorGray
				} else {
					glyph.FG = ColorWhite
					collides = true
					// glyph.Bold = true
				}
				tline = append(tline, &Tile{
					Ground:   glyph,
					Objects:  make(map[ID]Object),
					Collides: collides,
				})
			}
			m.Tiles = append(m.Tiles, tline)
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	return m, nil
}
