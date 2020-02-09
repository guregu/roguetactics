package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nickdavies/go-astar/astar"
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

func (m *Map) Height() int {
	return len(m.Tiles)
}

func (m *Map) Width() int {
	return len(m.Tiles[0])
}

// func (m *Map) FindPath(fromX, fromY, toX, toY int, ignore ...Object) []Loc {
// 	path, dist, ok := astar.Path(m.TileAt(toX, toY), m.TileAt(fromX, fromY))
// 	fmt.Println("path", path, "dist", dist, "ok", ok)

// 	if !ok {
// 		return nil
// 	}

// 	result := make([]Loc, 0, len(path))
// 	for _, p := range path {
// 		tile := p.(*Tile)
// 		result = append(result, Loc{Map: m.Name, X: tile.X, Y: tile.Y})
// 	}
// 	return result
// 	// a := astar.NewAStar(m.Height(), m.Width())
// 	// for y := 0; y < m.Height(); y++ {
// 	// cols:
// 	// 	for x := 0; x < m.Width(); x++ {
// 	// 		tile := m.TileAt(x, y)
// 	// 		if tile.Collides {
// 	// 			a.FillTile(astar.Point{y, x}, -1)
// 	// 		}
// 	// 		if x != fromX || y != fromY {
// 	// 			top := tile.Top()
// 	// 			if top == nil {
// 	// 				continue cols
// 	// 			}
// 	// 			for _, obj := range ignore {
// 	// 				if top == obj {
// 	// 					continue cols
// 	// 				}
// 	// 			}
// 	// 			a.FillTile(astar.Point{y, x}, -1)
// 	// 		}
// 	// 	}
// 	// }
// 	// p2p := astar.NewPointToPoint()
// 	// path := a.FindPath(p2p, []astar.Point{astar.Point{fromY, fromX}}, []astar.Point{astar.Point{toY, toX}})

// 	// var locs []Loc
// 	// for path != nil {
// 	// 	if !(path.Col == fromX && path.Row == fromY) {
// 	// 		locs = append(locs, Loc{Map: m.Name, X: path.Col, Y: path.Row})
// 	// 	}
// 	// 	path = path.Parent
// 	// }
// 	// return locs
// }

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
	// a := m.astar()
	// l2p := astar.NewListToPoint(true)
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

	// // surrounding := make([]astar.Point, 0, 4)
	// surrounding := []astar.Point{
	// 	astar.Point{toLoc.Y - 1, toLoc.X},
	// 	astar.Point{toLoc.Y + 1, toLoc.X},
	// 	astar.Point{toLoc.Y, toLoc.X - 1},
	// 	astar.Point{toLoc.Y, toLoc.X + 1},
	// }
	// path := a.FindPath(l2p, surrounding, []astar.Point{astar.Point{fromLoc.Y, fromLoc.X}})

	// var locs []Loc
	// for path != nil {
	// 	if !(path.Col == fromLoc.X && path.Row == fromLoc.Y) {
	// 		locs = append(locs, Loc{Map: m.Name, X: path.Col, Y: path.Row})
	// 	}
	// 	path = path.Parent
	// }
	// return locs
}

type Tile struct {
	Ground   Glyph
	Objects  map[ID]Object
	Collides bool
	X, Y     int
	Map      *Map
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
		return top.Glyph()
	}
	return t.Ground
}

// func (t *Tile) PathNeighbors() []astar.Pather {
// 	neighbors := make([]astar.Pather, 0, 4)
// 	if t.X != 0 {
// 		other := t.Map.TileAt(t.X-1, t.Y)
// 		if other != nil {
// 			neighbors = append(neighbors, other)
// 		}
// 	}
// 	if t.X != t.Map.Width()-1 {
// 		other := t.Map.TileAt(t.X+1, t.Y)
// 		if other != nil {
// 			neighbors = append(neighbors, other)
// 		}
// 	}
// 	if t.Y != 0 {
// 		other := t.Map.TileAt(t.X, t.Y-1)
// 		if other != nil {
// 			neighbors = append(neighbors, other)
// 		}
// 	}
// 	if t.Y != t.Map.Height()-1 {
// 		other := t.Map.TileAt(t.X, t.Y+1)
// 		if other != nil {
// 			neighbors = append(neighbors, other)
// 		}
// 	}
// 	return neighbors
// }

// func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
// 	target := to.(*Tile)
// 	if target.Collides {
// 		return 1000
// 	}
// 	if target.HasCollider() {
// 		return 1000
// 	}
// 	return 1
// }

// func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
// 	target := to.(*Tile)
// 	if target.Collides {
// 		return t.ManhattanDistance(to) * 1000
// 	}
// 	if target.HasCollider() {
// 		return t.ManhattanDistance(to) * 1000
// 	}
// 	return t.ManhattanDistance(to)
// }

// func (t *Tile) ManhattanDistance(to astar.Pather) float64 {
// 	if tile, ok := to.(*Tile); ok {
// 		return float64(abs(t.X-tile.X) + abs(t.Y-tile.Y))
// 	}
// 	panic("bad tile distance")
// }

func (t *Tile) String() string {
	return fmt.Sprintf("Tile(%s:%d,%d)", t.Map.Name, t.X, t.Y)
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
	var x, y int
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
					X:        x,
					Y:        y,
					Map:      m,
				})
				x++
			}
			m.Tiles = append(m.Tiles, tline)
			y++
			x = 0
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	return m, nil
}
