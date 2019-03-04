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
	Ground  Glyph
	Objects map[ID]Object
}

func (t *Tile) Add(obj Object) {
	t.Objects[obj.ID()] = obj
}

func (t *Tile) Remove(obj Object) {
	delete(t.Objects, obj.ID())
}

func (t *Tile) Glyph() Glyph {
	g := t.Ground
	z := -1
	for _, obj := range t.Objects {
		loc := obj.Loc()
		if loc.Z > z {
			g = obj.Glyph()
			z = loc.Z
		}
	}
	return g
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
				if r == '.' || r == '#' {
					glyph.FG = ColorGray
					if r == '.' {
						glyph.Blink = true
					}
				} else {
					glyph.FG = ColorWhite
					// glyph.Bold = true
				}
				tline = append(tline, &Tile{
					Ground:  glyph,
					Objects: make(map[ID]Object),
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
