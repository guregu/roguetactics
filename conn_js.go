// +build js

package main

import (
	"io"
	"log"
	"net/http"
	// "strings"
	"syscall/js"

	"github.com/gliderlabs/ssh"
)

func handleSSH(world *World) {
	term := js.Global().Get("term")
	term.Call("write", "hello from go")

	sesh := NewSesh(newConn(), world)
	sesh.Run()
	// io.WriteString(sesh.ssh, resetScreen+cursorTo00)
}

func listenAndWait() {
	select {}
}

func runSesh(sesh *Sesh) {
	term := sesh.ssh.(*xtermConn).term
	cb := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		evt := args[0]
		key := evt.Get("key").String()
		log.Println("got key", key)
		sesh.do(key)
		return nil
	})
	term.Call("onKey", cb)
	mouse := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		evt := args[0].String()
		log.Println("got mouse", evt)
		sesh.do(evt)
		return nil
	})
	term.Call("onBinary", mouse)

	sesh.setup()
	sesh.redraw()
	io.WriteString(sesh.ssh, EnableMouseReporting)
}

type xtermConn struct {
	term js.Value
}

func newConn() Conn {
	term := js.Global().Get("term")
	xterm := &xtermConn{
		term: term,
	}
	return xterm
}

func (xt *xtermConn) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (xt *xtermConn) Write(p []byte) (n int, err error) {
	arr := js.Global().Get("Uint8Array")
	buf := arr.New(len(p))
	js.CopyBytesToJS(buf, p)
	xt.term.Call("write", buf)
	return len(p), nil
}

func (xt *xtermConn) Exit(code int) error {
	return nil
}

func (xt *xtermConn) Pty() (ssh.Pty, <-chan ssh.Window, bool) {
	return ssh.Pty{}, nil, true
}

func open(name string) (io.ReadCloser, error) {
	resp, err := http.Get(name)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
