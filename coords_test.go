package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsValid(t *testing.T) {
	r := require.New(t)

	c := Coords{5, 10}
	r.True(c.IsValid())

	c = Coords{-1, 10}
	r.False(c.IsValid())

	c = Coords{5, -1}
	r.False(c.IsValid())
}

func TestMergeinIfInvalid(t *testing.T) {
	r := require.New(t)

	// only y
	base := Coords{5, -1}
	merging := Coords{0, 0}
	base.MergeInIfInvalid(merging)
	r.Equal(Coords{5, 0}, base)

	// only x
	base = Coords{-1, 5}
	merging = Coords{0, 0}
	base.MergeInIfInvalid(merging)
	r.Equal(Coords{0, 5}, base)
}

func TestEnsureWithinBounds(t *testing.T) {
	r := require.New(t)

	// both under
	base := Coords{-1, -1}
	base.EnsureWithinBounds(64, 48)
	r.Equal(Coords{0, 0}, base)

	// both over
	base = Coords{640, 480}
	base.EnsureWithinBounds(64, 48)
	r.Equal(Coords{63, 47}, base)

	// both ok
	base = Coords{30, 30}
	base.EnsureWithinBounds(64, 48)
	r.Equal(Coords{30, 30}, base)
}
