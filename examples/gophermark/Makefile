# linux:
# 	go build -ldflags "-s" -v .

# test:
# 	go test -v .

# TODO - tinyGO?
wasm:
#	cp "$(gotip env GOROOT)/misc/wasm/wasm_exec.js" .
	GOOS=js GOARCH=wasm go1.18beta1 build -ldflags "-s" -o main.wasm

windows:
	env GOOS=windows GOARCH=386 CGO_ENABLED=1 CXX=i686-w64-mingw32-g++ CC=i686-w64-mingw32-gcc go build -ldflags "-s" -v .

compress:
	ls -lh main.wasm
	gzip -f --best -c main.wasm > main.wasm.gz
	ls -lh main.wasm.gz

serve:
	~/go/src/github.com/unitoftime/go-experiments/fileserver/fileserver

run: wasm compress serve
