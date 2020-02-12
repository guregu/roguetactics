package main

import (
	"io"

	"github.com/gliderlabs/ssh"
)

type Conn interface {
	io.ReadWriter
	Exit(int) error
	Pty() (ssh.Pty, <-chan ssh.Window, bool)
}
