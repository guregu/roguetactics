# Bitesize Tactics

This is a roguelite game inspired by Final Fantasy Tactics made for [7DRL 2020](https://itch.io/jam/7drl-challenge-2020/rate/583702).

You can [play the game on itch.io](https://kawaiisolutions.itch.io/bitesize-tactics).

## Background

This is written in Go, and runs locally over SSH and online via xterm.js+WebAssembly. All the display code and ANSI wrangling is custom. This project might be interesting if you want to see how to do a roguelike-style text game "the hard way" without using a fake terminal or curses-style library.

Originally, this code derives from a realtime multiplayer arena shooter game. The turn-based system was bolted on afterwards, so it's basically a realtime game pretending to be turn-based. The gameplay code was mostly hastily written over 7 days.

## Build & Run

### Install

(this uses modules, don't put it in $GOPATH):
```
# make sure you have latest go
go get
```

### Run locally (via SSH):
```
# run server
go build && ./roguetactics  

# attach client (other tab)
./ssh.sh  
```

### Web version
```
GOOS=js GOARCH=wasm go build -o web/main.wasm
# symlink web/maps to ./maps
# run python -m SimpleHTTPServer etc in web
```

## Contributing

This is BSD licensed (xterm.js is MIT). You're free to fork it and do whatever.

If you'd like to contribute something, please open an issue first.

There are a couple cool maps that are drawn out but unimplemented, and lots of gameplay mechanics that could be more fleshed out.

## Thanks

Shout outs to:

- meepches and Eric for some of the maps
- @wittekm for some code contributions
- the 7DRL organizers for the fun competition
- everyone who played the game
