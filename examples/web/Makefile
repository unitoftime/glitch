all:
	GOOS=js GOARCH=wasm gotip build -ldflags "-s" -o gophermark.wasm github.com/unitoftime/glitch/examples/gophermark
	GOOS=js GOARCH=wasm gotip build -ldflags "-s" -o ui.wasm github.com/unitoftime/glitch/examples/ui
	GOOS=js GOARCH=wasm gotip build -ldflags "-s" -o 3d.wasm github.com/unitoftime/glitch/examples/3d
