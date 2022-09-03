package main

import (
	"bufio"
	"fmt"
	"io"
	"path"

	// "os"
	"encoding/json"
	"math/rand"
	"strings"

	"github.com/nickdavies/go-astar/astar"
)

type Map struct {
	Name    string
	Tiles   [][]*Tile
	Objects map[ID]Object

	SpawnPoints [][]Loc

	Meta MapMeta
}

type MapMeta struct {
	Name   string
	Width  int
	Height int
	Glyphs map[string]MetaGlyphDef
	BG     [][]Color256

	Teams       int
	SpawnPoints [][][2]int
	SpawnGlyphs []string
}

type MetaGlyphDef struct {
	FG      Color
	BG      Color
	Collide bool
	Replace string
}

func (gd *MetaGlyphDef) UnmarshalJSON(data []byte) error {
	convert := func(rawcolor json.RawMessage) Color {
		if len(rawcolor) == 0 {
			return nil
		}
		if rawcolor[0] == '[' {
			var rgb ColorRGB
			if err := json.Unmarshal(rawcolor, &rgb); err != nil {
				panic(err)
			}
			return rgb
		}
		var xterm Color256
		if err := json.Unmarshal(rawcolor, &xterm); err != nil {
			panic(err)
		}
		return xterm
	}

	raw := struct {
		FG, BG  json.RawMessage
		Collide bool
		Replace string
	}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	gd.FG = convert(raw.FG)
	gd.BG = convert(raw.BG)
	gd.Collide = raw.Collide
	gd.Replace = raw.Replace
	return nil
}

func (m *Map) NewTile(glyph Glyph, collides bool, x, y int) *Tile {
	return &Tile{
		Ground:   glyph,
		Objects:  make(map[ID]Object),
		Collides: collides,
		X:        x,
		Y:        y,
		Map:      m,
	}
}
func (m *Map) Reset() {
	for _, obj := range m.Objects {
		m.Remove(obj)
	}
	for y := 0; y < len(m.Tiles); y++ {
		for x := 0; x < len(m.Tiles[y]); x++ {
			m.Tiles[y][x].RemoveAll()
		}
	}
}

func (m *Map) TileAtLoc(loc Loc) *Tile {
	return m.TileAt(loc.X, loc.Y)
}

func (m *Map) TileAt(x, y int) *Tile {
	// TODO bounds check
	if y >= len(m.Tiles) || x >= len(m.Tiles[y]) {
		// fmt.Println("invalid tile access", x, y)
		return &Tile{
			X:        x,
			Y:        y,
			Map:      m,
			Collides: true,
			Ground:   GlyphOf(' '),
			Objects:  make(map[ID]Object),
			invalid:  true,
		}
	}
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

func (m *Map) Height() int {
	return len(m.Tiles)
}

func (m *Map) Width() int {
	return len(m.Tiles[0])
}

func (m *Map) FindPath(fromX, fromY, toX, toY int, ignore ...Object) []Loc {
	a := m.astar(ignore...)
	p2p := astar.NewPointToPoint()
	path := a.FindPath(p2p, []astar.Point{astar.Point{fromY, fromX}}, []astar.Point{astar.Point{toY, toX}})

	var locs []Loc
	for path != nil {
		if !(path.Col == fromX && path.Row == fromY) {
			locs = append(locs, Loc{Map: m.Name, X: path.Col, Y: path.Row})
		}
		path = path.Parent
	}
	return locs
}

func (m *Map) astar(ignore ...Object) astar.AStar {
	a := astar.NewAStar(m.Height(), m.Width())
	for y := 0; y < m.Height(); y++ {
		for x := 0; x < m.Width(); x++ {
			tile := m.TileAt(x, y)
			if tile.Collides {
				a.FillTile(astar.Point{y, x}, -1)
			} else if tile.HasCollider(ignore...) {
				a.FillTile(astar.Point{y, x}, -1)
			}
		}
	}
	return a
}

func (m *Map) FindPathNextTo(from *Mob, to *Mob) []Loc {
	fromLoc := from.Loc()
	toLoc := to.Loc()

	n := m.FindPath(fromLoc.X, fromLoc.Y, toLoc.X, toLoc.Y-1, from)
	e := m.FindPath(fromLoc.X, fromLoc.Y, toLoc.X+1, toLoc.Y, from)
	s := m.FindPath(fromLoc.X, fromLoc.Y, toLoc.X, toLoc.Y+1, from)
	w := m.FindPath(fromLoc.X, fromLoc.Y, toLoc.X-1, toLoc.Y, from)

	path := n
	if e != nil && (path == nil || len(e) < len(path)) {
		path = e
	}
	if s != nil && (path == nil || len(s) < len(path)) {
		path = s
	}
	if w != nil && (path == nil || len(w) < len(path)) {
		path = w
	}

	return path
}

func (m *Map) Raycast(from, to Loc, ignoreObstacles bool) (hit *Mob, blocked bool, path []Loc) {
	x0 := from.X
	y0 := from.Y
	x1 := to.X
	y1 := to.Y

	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	for {
		if !(x0 == from.X && y0 == from.Y) {
			path = append(path, Loc{Map: m.Name, X: x0, Y: y0})
			tile := m.TileAt(x0, y0)
			if !ignoreObstacles {
				for _, obj := range tile.Objects {
					if mob, ok := obj.(*Mob); ok && !mob.Dead() {
						return mob, false, path
					}
				}
			}
			if !ignoreObstacles && tile.Collides {
				return nil, true, path
			}
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
	return nil, false, path
}

type Tile struct {
	Ground   Glyph
	Objects  map[ID]Object
	Collides bool
	X, Y     int
	Map      *Map
	invalid  bool
}

func (t *Tile) Add(obj Object) {
	t.Objects[obj.ID()] = obj
}

func (t *Tile) Remove(obj Object) {
	delete(t.Objects, obj.ID())
}

func (t *Tile) RemoveAll() {
	t.Objects = make(map[ID]Object)
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

func (t *Tile) HasCollider(ignore ...Object) bool {
loop:
	for _, obj := range t.Objects {
		for _, ig := range ignore {
			if ig == obj {
				continue loop
			}
		}
		if col, ok := obj.(Collider); ok {
			if col.Collides(nil, 0) {
				return true
			}
		}
	}
	return false
}

func (t *Tile) Glyph() Glyph {
	if top := t.Top(); top != nil {
		g := top.Glyph()
		if g.BG == nil && t.Ground.BG != nil {
			g.BG = t.Ground.BG
		}
		return g
	}
	return t.Ground
}

func (t *Tile) String() string {
	return fmt.Sprintf("Tile(%s:%d,%d)", t.Map.Name, t.X, t.Y)
}

func (t *Tile) IsValid() bool {
	return !t.invalid
}

func loadMap(name string) (*Map, error) {
	filename := path.Join("maps", name+".map")
	f, err := open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var meta MapMeta
	metaf, err := open(strings.Replace(filename, ".map", ".json", 1))
	if err == nil {
		defer metaf.Close()
		if err := json.NewDecoder(metaf).Decode(&meta); err != nil {
			return nil, err
		}
	}
	// log.Printf("Loading map: %s %+v", name, meta)

	if meta.Height == 0 {
		meta.Height = 20
	}
	if meta.Width == 0 {
		meta.Width = 80
	}

	r := bufio.NewReader(f)
	m := &Map{
		Name:    name,
		Objects: make(map[ID]Object),
		Meta:    meta,
	}
	m.SpawnPoints = make([][]Loc, meta.Teams)
	var x, y int
	for {
		line, _, err := r.ReadLine()
		var tline []*Tile
		if len(line) > 0 {
			for _, r := range string(line) {
				// if x > meta.Width {
				// 	continue
				// }
				switch r {
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					n := int(r - '0')
					m.SpawnPoints[n] = append(m.SpawnPoints[n], Loc{Map: m.Name, X: x, Y: y})
					if n < len(meta.SpawnGlyphs) {
						opts := []rune(meta.SpawnGlyphs[n])
						r = opts[rand.Intn(len(opts))]
					} else {
						r = '.'
					}
				}
				glyph := GlyphOf(r)
				collides := false
				var replace string
				for glyphs, info := range meta.Glyphs {
					if !strings.ContainsRune(glyphs, r) {
						continue
					}
					if info.Collide {
						collides = info.Collide
					}
					if info.FG != nil {
						glyph.FG = info.FG
					}
					if info.BG != nil {
						glyph.BG = info.BG
					}
					if info.Replace != "" {
						fmt.Println("REPPP", info.Replace)
						replace += info.Replace
					}
				}
				if meta.BG != nil {
					glyph.BG = meta.BG[y][x]
				}
				if len(replace) > 0 {
					fmt.Println("REPLACE", replace)
					runes := []rune(replace)
					glyph.Rune = runes[rand.Intn(len(runes))]
				}
				tile := m.NewTile(glyph, collides, x, y)
				tline = append(tline, tile)
				x++
			}
			for len(tline) < meta.Width {
				tline = append(tline, m.NewTile(GlyphOf(' '), true, x, y))
				x++
			}
			m.Tiles = append(m.Tiles, tline)
			y++
			x = 0
			// if y > meta.Height {
			// 	break
			// }
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	for y := len(m.Tiles); y < meta.Height; y++ {
		var tline []*Tile
		for x := 0; x < meta.Width; x++ {
			tline = append(tline, m.NewTile(GlyphOf(' '), true, x, y))
		}
		m.Tiles = append(m.Tiles, tline)
	}
	for i, spawns := range meta.SpawnPoints {
		for _, spawn := range spawns {
			m.SpawnPoints[i] = append(m.SpawnPoints[i], Loc{Map: m.Name, X: spawn[0], Y: spawn[1]})
		}
	}
	return m, nil
}
