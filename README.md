## roguelike tactics game

install:
```
go get
```

run locally (ssh ver):
```
go build && ./roguetactics
./ssh.sh # make sure your terminal is 80x27
```

web version (see deploy.sh)
```
GOOS=js GOARCH=wasm go build -o web/main.wasm
# symlink web/maps to ./maps
# run pythom -m SimpleHTTPServer etc in web
```