// +build js

package main

import (
	"syscall/js"
)

func handleSSH(world *World) {
	term := js.Global().Get("term")
	sesh := NewSesh(newConn(), world)
	cb := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		evt := args[0]
		key := evt.Get("key").String()
		sesh.ssh.(*xtermConn).in <- key
		return nil
	})
	term.Call("onKey", cb)
	mouse := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		evt := args[0].String()
		sesh.ssh.(*xtermConn).in <- evt
		return nil
	})
	term.Call("onBinary", mouse)
	sesh.Run()
}

func listenAndWait() {
	select {}
}

type xtermConn struct {
	term js.Value
	in   chan string
	buf  []byte
}

func newConn() Conn {
	term := js.Global().Get("term")
	xterm := &xtermConn{
		term: term,
		in:   make(chan string, 8),
	}
	return xterm
}

func (xt *xtermConn) Read(p []byte) (n int, err error) {
start:
	if l := len(xt.buf); l > 0 {
		if len(p) < l {
			l = len(p)
		}
		copy(p, xt.buf[:l])
		xt.buf = xt.buf[l:]
		return l, nil
	}
	input := []byte(<-xt.in)
	xt.buf = input
	goto start
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

func consoleWrite(str string) {
	term := js.Global().Get("term")
	term.Call("write", str)
}
