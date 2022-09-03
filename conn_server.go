// +build !js

package main

import (
	"io"
	"log"
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


func consoleWrite(str string) {
	log.Println(str)
}
