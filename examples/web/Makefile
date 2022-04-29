all:
	GOOS=js GOARCH=wasm go build -ldflags "-s" -o gophermark.wasm github.com/unitoftime/glitch/examples/gophermark
	GOOS=js GOARCH=wasm go build -ldflags "-s" -o frame.wasm github.com/unitoftime/glitch/examples/frame
	GOOS=js GOARCH=wasm go build -ldflags "-s" -o ui.wasm github.com/unitoftime/glitch/examples/ui
	GOOS=js GOARCH=wasm go build -ldflags "-s" -o 3d.wasm github.com/unitoftime/glitch/examples/3d
