package main

import (
	"embed"
	"io"
)

//go:embed maps
var embedded embed.FS

func open(name string) (io.ReadCloser, error) {
	return embedded.Open(name)
}