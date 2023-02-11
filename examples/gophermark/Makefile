# linux:
# 	go build -ldflags "-s" -v .

# test:
# 	go test -v .

# TODO - tinyGO?
tinygo:
	tinygo build -o main.wasm -target wasm ./main.go
wasm:
#	cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
	GOOS=js GOARCH=wasm go build -ldflags "-s" -o main.wasm

windows:
	env GOOS=windows GOARCH=386 CGO_ENABLED=1 CXX=i686-w64-mingw32-g++ CC=i686-w64-mingw32-gcc go build -ldflags "-s" -v .

compress:
	ls -lh main.wasm
	gzip -f --best -c main.wasm > main.wasm.gz
	ls -lh main.wasm.gz

serve:
	serve

run: wasm compress serve
