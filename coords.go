package main

type Coords struct {
	x, y int
}

var (
	OriginCoords  = Coords{0, 0}
	InvalidCoords = Coords{-1, -1}
)

func (c *Coords) IsValid() bool {
	return c.x != -1 && c.y != -1
}

func (c *Coords) MergeInIfInvalid(other Coords) {
	/**
	Replacement for the reasonably-common pattern of

	if mw.cursorX == -1 {
		mw.cursorX = loc.X
	}
	if mw.cursorY == -1 {
		mw.cursorY = loc.Y
	}
	*/
	if c.x == -1 {
		c.x = other.x
	}

	if c.y == -1 {
		c.y = other.y
	}
}

func (c *Coords) Add(dx, dy int) {
	c.x += dx
	c.y += dy
}

func (c *Coords) EnsureWithinBounds(maxWidth, maxHeight int) {
	/**
	Ensure X is within [0, maxWidth-1]
	Ensure Y is within [0, maxHeight-1]
	*/
	if c.x >= maxWidth {
		c.x = maxWidth - 1
	} else if c.x < 0 {
		c.x = 0
	}

	if c.y >= maxHeight {
		c.y = maxHeight - 1
	} else if c.y < 0 {
		c.y = 0
	}
}
