## roguelike tactics game

install:
```
go get
```

run locally (ssh ver):
```
go build && ./roguetactics
./ssh.sh # run the server; make sure your terminal is 80x27

# and in another shell session
ssh -o StrictHostKeyChecking=no localhost -p 2222 # connect to the server as a client
```


web version (see deploy.sh)
```
GOOS=js GOARCH=wasm go build -o web/main.wasm
# symlink web/maps to ./maps
# run pythom -m SimpleHTTPServer etc in web
```
