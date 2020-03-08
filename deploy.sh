GOOS=js GOARCH=wasm go build -o web/main.wasm && gzip -kf web/main.wasm && 
rsync -avv web/* ubuntu@18.237.249.17:/var/www/html/bitesize
rsync -avv maps ubuntu@18.237.249.17:/var/www/html/bitesize
