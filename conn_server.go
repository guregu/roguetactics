// +build !js

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/gliderlabs/ssh"
	"github.com/ztrue/shutdown"
)

func handleSSH(world *World) {
	ssh.Handle(func(s ssh.Session) {
		sesh := NewSesh(s, world)
		io.WriteString(sesh.ssh, resetScreen+cursorTo00)
		sesh.Run()
	})
	shutdown.Add(func() {
		world.applySync <- ShutdownAction{}
	})
}

func listenAndWait() {
	log.Println("starting ssh server on port 2222...")
	go func() {
		log.Fatal(ssh.ListenAndServe(":2222", nil))
	}()
	shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)
}

func open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func runSesh(sesh *Sesh) {
	log.Println("running")
	ptyReq, winCh, isPty := sesh.ssh.Pty()
	_ = ptyReq

	if !isPty {
		io.WriteString(sesh.ssh, "No PTY requested.\n")
		sesh.ssh.Exit(1)
		return
	}

	defer sesh.cleanup()

	go func() {
		for win := range winCh {
			sesh.resize(win)
		}
	}()
	sesh.setup()
	buf := make([]byte, 256)
	for {
		n, err := sesh.ssh.Read(buf[:])
		if err != nil {
			fmt.Println("Error: 1", err)
			sesh.ssh.Exit(1)
			break
		}
		if n > 0 {
			sesh.do(string(buf[:n]))
			fmt.Println("GOT:", buf[:n], ">>>", strings.ReplaceAll(string(buf[:n]), "\033", "ESC"))
		}
	}
}

func consoleWrite(str string) {
	log.Println(str)
}
