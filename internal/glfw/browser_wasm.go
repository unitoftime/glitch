//go:build js && wasm
// +build js,wasm

package glfw

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"syscall/js"
	"time"
)

// Useful: https://developer.mozilla.org/en-US/docs/Web/API/WebGL_API/WebGL_best_practices

var htmlWindow = js.Global().Get("window")
var document = js.Global().Get("document")
var navigator = htmlWindow.Get("navigator")
var (
	navKeyboard js.Value
	keyboardLayoutMap js.Value
	// fnKeyboardLayoutMapGet js.Value
)

func isNilOrUndefined(val js.Value) bool {
	return val.IsNull() || val.IsUndefined()
}

var contextWatcher ContextWatcher

func Init(cw ContextWatcher) error {
	contextWatcher = cw
	return nil
}

func Terminate() error {
	return nil
}

func resolveCanvas() js.Value {
	canvas := document.Call("querySelector", "#glfw")
	if canvas.Equal(js.Null()) {
		canvas = document.Call("querySelector", "canvas")
	}
	return canvas
}

// Constructs the keyboard map based on navigator.Keyboard
func resolveNavigatorKeyboard() {
	// Notes: https://developer.mozilla.org/en-US/docs/Web/API/Keyboard
	if isNilOrUndefined(navigator) { return }

	navKeyboard = navigator.Get("keyboard")
	if isNilOrUndefined(navKeyboard) { return }

	keyboardPromise := navKeyboard.Call("getLayoutMap")
	if isNilOrUndefined(keyboardPromise) { return }

	keyboardPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) <= 0 { return nil }

		keyboardLayoutMap = args[0]
		if isNilOrUndefined(keyboardLayoutMap) { return nil }

		// // TODO: This is a nice optimization, but I'd have to bind it to something, maybe the window?
		// fnKeyboardLayoutMapGet = keyboardLayoutMap.Get("get").Call("bind", htmlWindow)
		return nil
	}))
}

func resolveIframeEmbedding() bool {
	// Notes: https://stackoverflow.com/questions/326069/how-to-identify-if-a-webpage-is-being-loaded-inside-an-iframe-or-directly-into-t
	// window.self !== window.top;
	self := htmlWindow.Get("self")
	top := htmlWindow.Get("top")
	if self.IsNull() {
		return true // Something is weird, assume we are in an iframe
	}
	if top.IsNull() {
		return true // Assume that we were blocked from accessing it do to cors
	}
	return !self.Equal(top)
}

func getDevicePixelRatio() float64 {
	devicePixelRatio := js.Global().Get("devicePixelRatio").Float()
	// if devicePixelRatio <= 0 {
	// 	devicePixelRatio = 1.0
	// } else {
	// 	devicePixelRatio = 1 / devicePixelRatio
	// }
	return devicePixelRatio
}

func CreateWindow(_, _ int, title string, monitor *Monitor, share *Window) (*Window, error) {
	// Find a canvas, preferably one with an id of glfw
	canvas := resolveCanvas()

	if canvas.Equal(js.Null()) {
		parent := document.Call("querySelector", "#glfw-container")
		canvas = document.Call("createElement", "canvas")
		canvas.Call("setAttribute", "id", "glfw")

		if parent.Equal(js.Null()) {
			parent = document.Get("body")
		}

		parent.Call("appendChild", canvas)
	}

	// HACK: Go fullscreen?
	width := js.Global().Get("innerWidth").Int()
	height := js.Global().Get("innerHeight").Int()

	devicePixelRatio := getDevicePixelRatio()
	canvas.Set("width", int((float64(width) * devicePixelRatio) + 0.5))   // Nearest non-negative int.
	canvas.Set("height", int((float64(height) * devicePixelRatio) + 0.5)) // Nearest non-negative int.
	canvas.Get("style").Call("setProperty", "width", fmt.Sprintf("%vpx", width))
	canvas.Get("style").Call("setProperty", "height", fmt.Sprintf("%vpx", height))

	document.Set("title", title)

	// Use glfw hints.
	attrs := defaultAttributes()
	attrs.Alpha = (hints[AlphaBits] > 0)
	if _, ok := hints[DepthBits]; ok {
		attrs.Depth = (hints[DepthBits] > 0)
	}
	attrs.Stencil = (hints[StencilBits] > 0)
	attrs.Antialias = (hints[Samples] > 0)
	attrs.PremultipliedAlpha = (hints[PremultipliedAlpha] > 0)
	attrs.PreserveDrawingBuffer = (hints[PreserveDrawingBuffer] > 0)
	attrs.PreferLowPowerToHighPerformance = (hints[PreferLowPowerToHighPerformance] > 0)
	attrs.FailIfMajorPerformanceCaveat = (hints[FailIfMajorPerformanceCaveat] > 0)

	// Create GL context.
	context, err := newContext(canvas, attrs)
	if err != nil {
		return nil, err
	}
	if context.Equal(js.Value{}) {
		return nil, err
	}

	w := &Window{
		canvas:           canvas,
		context:          context,
		devicePixelRatio: devicePixelRatio,
	}

	resolveNavigatorKeyboard()
	w.embeddedIframe = resolveIframeEmbedding()

	if w.canvas.Get("requestPointerLock").Equal(js.Undefined()) ||
		document.Get("exitPointerLock").Equal(js.Undefined()) {

		w.missing.pointerLock = true
	}
	if w.canvas.Get("webkitRequestFullscreen").Equal(js.Undefined()) ||
		document.Get("webkitExitFullscreen").Equal(js.Undefined()) {

		w.missing.fullscreen = true
	}

	if monitor != nil {
		if w.missing.fullscreen {
			log.Println("warning: Fullscreen API unsupported")
		} else {
			w.requestFullscreen = true
		}
	}

	SetupEventListeners(w)

	// Request first animation frame.
	// raf.Invoke(animationFrameCallback)

	// Alternative 3 RAF strategy
	// start()

	return w, nil
}

func SetupEventListeners(w *Window) {
	history := htmlWindow.Get("history")
	history.Call("pushState", nil, nil)
	history.Call("pushState", nil, nil)
	history.Call("pushState", nil, nil)

	htmlWindow.Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// e := args[0]
		// e.Call("preventDefault")
		history.Call("pushState", nil, nil)
		return nil
	}))

	js.Global().Call("addEventListener", "resize", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// HACK: Go fullscreen?
		width := js.Global().Get("innerWidth").Int()
		height := js.Global().Get("innerHeight").Int()

		w.devicePixelRatio = getDevicePixelRatio()
		// fmt.Println("DevicePixelRatio:", w.devicePixelRatio)
		w.canvas.Set("width", int((float64(width) * w.devicePixelRatio) + 0.5))   // Nearest non-negative int.
		w.canvas.Set("height", int((float64(height) * w.devicePixelRatio) + 0.5)) // Nearest non-negative int.
		w.canvas.Get("style").Call("setProperty", "width", fmt.Sprintf("%vpx", width))
		w.canvas.Get("style").Call("setProperty", "height", fmt.Sprintf("%vpx", height))

		if w.framebufferSizeCallback != nil {
			// TODO: Callbacks may be blocking so they need to happen asyncronously. However,
			//       GLFW API promises the callbacks will occur from one thread (i.e., sequentially), so may want to do that.

			go w.framebufferSizeCallback(w, w.canvas.Get("width").Int(), w.canvas.Get("height").Int())
			// go w.framebufferSizeCallback(w, width, height) // TODO: Is it just this?
		}
		if w.sizeCallback != nil {
			boundingW, boundingH := w.GetSize()
			go w.sizeCallback(w, boundingW, boundingH)
		}
		return nil
	}))

	document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ke := args[0]
		w.goFullscreenIfRequested()

		action := Press
		if ke.Get("repeat").Bool() {
			action = Repeat
		}

		key := toKey(ke)

		if key != KeyUnknown {
			// Extend slice if needed.
			neededSize := int(key) + 1
			if neededSize > len(w.keys) {
				w.keys = append(w.keys, make([]Action, neededSize-len(w.keys))...)
			}
			w.keys[key] = action
		}

		if w.keyCallback != nil {
			mods := toModifierKey(ke)

			go w.keyCallback(w, key, -1, action, mods)
		}

		if w.charCallback != nil {
			keyStr := ke.Get("key").String()
			if len(keyStr) == 1 {
				keyRune := []rune(keyStr)
				go w.charCallback(w, keyRune[0])
			}
		}

		// Dont prevent default on F11. That's the fullscreen hotkey
		// TODO: make this configurable? Maybe also include KeyF12?
		if key != KeyF11 {
			ke.Call("preventDefault")
		}
		return nil
	}))
	document.Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ke := args[0]

		w.goFullscreenIfRequested()

		key := toKey(ke)

		if key != KeyUnknown {
			// Extend slice if needed.
			neededSize := int(key) + 1
			if neededSize > len(w.keys) {
				w.keys = append(w.keys, make([]Action, neededSize-len(w.keys))...)
			}
			w.keys[key] = Release
		}

		if w.keyCallback != nil {
			mods := toModifierKey(ke)

			go w.keyCallback(w, key, -1, Release, mods)
		}

		ke.Call("preventDefault")
		return nil
	}))
	document.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		me := args[0]
		w.goFullscreenIfRequested()

		button := me.Get("button").Int()
		if !(button >= 0 && button < mouseButtonMax) {
			return nil
		}

		w.mouseButton[button] = Press
		if w.mouseButtonCallback != nil {
			go w.mouseButtonCallback(w, MouseButton(button), Press, 0)
		}

		// Note: I commented out the preventDefault here, because if you are running your game inside an iframe, when the user clicks the canvas, I guess this preventDefault will cause the focus to never get set on the canvas. And that will cause keyboard events to not register properly. You might rethink how this works in the future though.
		// me.Call("preventDefault")

		return nil
	}))
	document.Call("addEventListener", "mouseup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		me := args[0]
		w.goFullscreenIfRequested()

		button := me.Get("button").Int()
		if !(button >= 0 && button < mouseButtonMax) {
			return nil
		}

		w.mouseButton[button] = Release
		if w.mouseButtonCallback != nil {
			go w.mouseButtonCallback(w, MouseButton(button), Release, 0)
		}

		me.Call("preventDefault")
		return nil
	}))
	document.Call("addEventListener", "contextmenu", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		me := args[0]
		me.Call("preventDefault")
		return nil
	}))

	document.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		me := args[0]
		var movementX, movementY float64
		if !w.missing.pointerLock {
			movementX = me.Get("movementX").Float()
			movementY = me.Get("movementY").Float()
		} else {
			movementX = me.Get("clientX").Float() - w.cursorPos[0]
			movementY = me.Get("clientY").Float() - w.cursorPos[1]
		}

		w.cursorPos[0], w.cursorPos[1] = me.Get("clientX").Float(), me.Get("clientY").Float()
		if w.cursorPosCallback != nil {
			go w.cursorPosCallback(w, w.cursorPos[0], w.cursorPos[1])
		}
		if w.mouseMovementCallback != nil {
			go w.mouseMovementCallback(w, w.cursorPos[0], w.cursorPos[1], movementX, movementY)
		}

		me.Call("preventDefault")
		return nil
	}))
	document.Call("addEventListener", "wheel", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		we := args[0]
		deltaX := we.Get("deltaX").Float()
		deltaY := we.Get("deltaY").Float()

		// var multiplier float64
		// /*
		// 	switch we.DeltaMode {
		// 	case dom.DeltaPixel:
		// 		multiplier = 0.1
		// 	case dom.DeltaLine:
		// 		multiplier = 1
		// 	default:
		// 		log.Println("unsupported WheelEvent.DeltaMode:", we.DeltaMode)
		// 		multiplier = 1
		// 	}*/
		// multiplier = 1

		// if w.scrollCallback != nil {
		// 	go w.scrollCallback(w, -deltaX*multiplier, -deltaY*multiplier)
		// }

		// TODO: Snap scroll to individual ticks. This isn't exactly correct, but browsers return larger values that dont really match what GLFW typically returns
		if deltaX > 0 {
			deltaX = 1
		} else if deltaX < 0 {
			deltaX = -1
		}
		if deltaY > 0 {
			deltaY = 1
		} else if deltaY < 0 {
			deltaY = -1
		}

		if w.scrollCallback != nil {
			go w.scrollCallback(w, -deltaX, -deltaY)
		}

		we.Call("preventDefault")
		return nil
	}),
		map[string]any{"passive": false}, // Note: Lets us preventDefault on wheel events to prevent "ctrl-zoom"
	)

	htmlWindow.Call("addEventListener", "focus", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// fmt.Println("FOCUS")
		if w.focusCallback != nil {
			inFocus := true
			go w.focusCallback(w, inFocus)
		}

		return nil
	}))

	htmlWindow.Call("addEventListener", "blur", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// fmt.Println("BLUR")

		// Attempt to clear keys
		for key := range w.keys {
			w.keys[key] = Release
		}
		// animationFrameChan <- struct{}{}

		if w.focusCallback != nil {
			inFocus := false
			go w.focusCallback(w, inFocus)
		}

		return nil
	}))

	// Detect window losing focus: https://developer.mozilla.org/en-US/docs/Web/API/Document/visibilitychange_event
	document.Call("addEventListener", "visibilitychange", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			// event := args[0]
			state := document.Get("visibilityState").String()
			// fmt.Println("VISCHANGE:", state)

			// If they are leaving the page, clear all the inputs
			if state == "hidden" {
				w.hidden = true

				// TODO - clear mouse input too?
				for key := range w.keys {
					w.keys[key] = Release
				}
				// animationFrameChan <- struct{}{}
			} else if state == "visible" {
				w.hidden = false
			}
		}()
		return nil
	}))

	// TODO: Maybe in the future I'll allow people to set this. It kinda doesn't work well b/c it freezes the window. so the game locks up
	// htmlWindow.Call("addEventListener", "beforeunload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 	fmt.Println("Catching BeforeUnload!")
	// 	we := args[0]
	// 	we.Call("preventDefault")

	// 	return js.ValueOf("Sure?")
	// }))

	/*
		// Hacky mouse-emulation-via-touch.
		touchHandler := func(event dom.Event) {
			w.goFullscreenIfRequested()

			te := event.(*dom.TouchEvent)

			touches := te.Get("touches")
			if touches.Length() > 0 {
				t := touches.Index(0)

				if w.touches != nil && w.touches.Length() > 0 { // This event is a movement only if we previously had > 0 touch points.
					if w.mouseMovementCallback != nil {
						go w.mouseMovementCallback(w, t.Get("clientX").Float(), t.Get("clientY").Float(), t.Get("clientX").Float()-w.cursorPos[0], t.Get("clientY").Float()-w.cursorPos[1])
					}
				}

				w.cursorPos[0], w.cursorPos[1] = t.Get("clientX").Float(), t.Get("clientY").Float()
				if w.cursorPosCallback != nil {
					go w.cursorPosCallback(w, w.cursorPos[0], w.cursorPos[1])
				}
			}
			w.touches = touches

			te.PreventDefault()
		}
		document.AddEventListener("touchstart", false, touchHandler)
		document.AddEventListener("touchmove", false, touchHandler)
		document.AddEventListener("touchend", false, touchHandler)
	*/

}

func SwapInterval(interval int) error {
	// TODO: Implement.
	return nil
}

type Window struct {
	canvas            js.Value
	context           js.Value
	requestFullscreen bool // requestFullscreen is set to true when fullscreen should be entered as soon as possible (in a user input handler).
	fullscreen        bool // fullscreen is true if we're currently in fullscreen mode.
	embeddedIframe    bool // true if the window is embedded in an iframe

	// Unavailable browser APIs.
	missing struct {
		pointerLock bool // Pointer Lock API.
		fullscreen  bool // Fullscreen API.
	}

	devicePixelRatio float64

	cursorMode  int
	cursorPos   [2]float64
	mouseButton [mouseButtonMax]Action

	keys []Action

	cursorPosCallback       CursorPosCallback
	mouseMovementCallback   MouseMovementCallback
	mouseButtonCallback     MouseButtonCallback
	keyCallback             KeyCallback
	scrollCallback          ScrollCallback
	charCallback            CharCallback
	framebufferSizeCallback FramebufferSizeCallback
	sizeCallback            SizeCallback
	focusCallback           FocusCallback

	hidden bool // Used to track if the window is hidden or visible
	rafOnce sync.Once

	touches js.Value // Hacky mouse-emulation-via-touch.
}

func (w *Window) SetPos(xpos, ypos int) {
	fmt.Println("not implemented: SetPos:", xpos, ypos)
}

func (w *Window) SetSize(width, height int) {
	fmt.Println("not implemented: SetSize:", width, height)
}

func (w *Window) BrowserHidden() bool {
	return w.hidden
}

// goFullscreenIfRequested performs webkitRequestFullscreen if it was scheduled. It is called only from
// user events, because that API will fail if called at any other time.
func (w *Window) goFullscreenIfRequested() {
	if !w.requestFullscreen {
		return
	}
	w.requestFullscreen = false
	// https://developer.mozilla.org/en-US/docs/Web/API/Element/requestFullscreen
	w.canvas.Call("requestFullscreen")
	w.fullscreen = true
}

func (w *Window) ScreenMode() ScreenModeType {
	fullscreenElem := document.Get("fullscreenElement")
	canvasIsFull := fullscreenElem.Equal(w.canvas)

	if canvasIsFull {
		return ScreenModeFull
	} else {
		return ScreenModeWindowed
	}
}

func (w *Window) SetScreenMode(smt ScreenModeType) {
	if smt == ScreenModeFull {
		w.requestFullscreen = true
	} else if smt == ScreenModeWindowed {
		current := w.ScreenMode()
		if current != ScreenModeWindowed {
			document.Call("exitFullscreen")
			w.fullscreen = false
		}
	}
}

func (w *Window) EmbeddedIframe() bool {
	return w.embeddedIframe
}

type Monitor struct{}

func (m *Monitor) GetVideoMode() *VidMode {
	return &VidMode{
		// HACK: Hardcoded sample values.
		// TODO: Try to get real values from browser via some API, if possible.
		Width:       1680,
		Height:      1050,
		RedBits:     8,
		GreenBits:   8,
		BlueBits:    8,
		RefreshRate: 60,
	}
}

func GetPrimaryMonitor() *Monitor {
	// TODO: Implement real functionality.
	return &Monitor{}
}

func PollEvents() error {
	return nil
}

func (w *Window) MakeContextCurrent() {
	contextWatcher.OnMakeCurrent(w.context)
}

func DetachCurrentContext() {
	contextWatcher.OnDetach()
}

func GetCurrentContext() *Window {
	panic("not implemented")
}

type CursorPosCallback func(w *Window, xpos float64, ypos float64)

func (w *Window) SetCursorPosCallback(cbfun CursorPosCallback) (previous CursorPosCallback) {
	w.cursorPosCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type MouseMovementCallback func(w *Window, xpos float64, ypos float64, xdelta float64, ydelta float64)

func (w *Window) SetMouseMovementCallback(cbfun MouseMovementCallback) (previous MouseMovementCallback) {
	w.mouseMovementCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type KeyCallback func(w *Window, key Key, scancode int, action Action, mods ModifierKey)

func (w *Window) SetKeyCallback(cbfun KeyCallback) (previous KeyCallback) {
	w.keyCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type CharCallback func(w *Window, char rune)

func (w *Window) SetCharCallback(cbfun CharCallback) (previous CharCallback) {
	w.charCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type ScrollCallback func(w *Window, xoff float64, yoff float64)

func (w *Window) SetScrollCallback(cbfun ScrollCallback) (previous ScrollCallback) {
	w.scrollCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type MouseButtonCallback func(w *Window, button MouseButton, action Action, mods ModifierKey)

func (w *Window) SetMouseButtonCallback(cbfun MouseButtonCallback) (previous MouseButtonCallback) {
	w.mouseButtonCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type FramebufferSizeCallback func(w *Window, width int, height int)

func (w *Window) SetFramebufferSizeCallback(cbfun FramebufferSizeCallback) (previous FramebufferSizeCallback) {
	w.framebufferSizeCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type FocusCallback func(w *Window, focused bool)

func (w *Window) SetFocusCallback(cbfun FocusCallback) (previous FocusCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

func (w *Window) GetSize() (width, height int) {
	width = int(w.canvas.Call("getBoundingClientRect").Get("width").Float()*w.devicePixelRatio + 0.5)
	height = int(w.canvas.Call("getBoundingClientRect").Get("height").Float()*w.devicePixelRatio + 0.5)

	return width, height
}

func (w *Window) GetFramebufferSize() (width, height int) {
	return w.canvas.Get("width").Int(), w.canvas.Get("height").Int()
}

// TODO - is it possible for these to differ?
func (w *Window) GetContentScale() (float32, float32) {
	return float32(w.devicePixelRatio), float32(w.devicePixelRatio)
}

func (w *Window) GetPos() (x, y int) {
	// Not implemented.
	return
}

func (w *Window) ShouldClose() bool {
	return false
}

func (w *Window) SetShouldClose(value bool) {
	// TODO: Implement.
	// THINK: What should happen in the browser if we're told to "close" the window. Do we destroy/remove the canvas? Or nothing?
	//        Perhaps https://developer.mozilla.org/en-US/docs/Web/API/Window.close is relevant.
}

func (w *Window) SwapBuffers() error {
	// How this works (because its kind of complicated):
	// 1. RAF is invoked once, and once the raf is consumed (by reading from aimationFrameChan), the w.rafOnce object is reset so it can be invoked again
	// 2. If the browser window is hidden, usually RAF gets halted. To keep the game running, we have a 100ms timeout that executes. Eventually the game is visibile again, causing the animationFrameChan to get consumed, and `rafOnce` to be reset
	// Note: One thing I tried is calling `raf.Invoke(animationFrameCallback)` *after* reading from the channel (this effectively synchronizes in the opposite direction, causing the raf to be blocked for until the next swapbuffers call). This seemed to cause a lot of weird stability isuess in wasm, so I moved away from it.
	// Note: Its *very* important that we dont accidentally start multiple "raf loops"
	w.rafOnce.Do(func() {
		raf.Invoke(animationFrameCallback)
	})

	select {
	case <-animationFrameChan:
		w.rafOnce = sync.Once{} // Reset rafOnce
	case <-time.After(100 * time.Millisecond): // TODO: would be nice to make this timeout configurable
	}

	return nil
}

var raf = js.Global().Get("requestAnimationFrame")
var animationFrameChan = make(chan struct{})
var lastFrame float64
var animationFrameCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	newFrame := args[0].Float()
	// if (newFrame - lastFrame) > 18 {
	// 	fmt.Println("Warning: Possible Dropped Frame: ", newFrame - lastFrame)
	// }
	lastFrame = newFrame

	animationFrameChan <- struct{}{}

	return nil
})

func (w *Window) GetCursorPos() (x, y float64) {
	return w.cursorPos[0], w.cursorPos[1]
}

var keyWarnings = 10

func (w *Window) GetKey(key Key) Action {
	if key == -1 && keyWarnings > 0 {
		// TODO: Implement all keys, get rid of this.
		keyWarnings--
		log.Println("GetKey: key not implemented.")
		return Release
	}
	if int(key) >= len(w.keys) {
		return Release
	}
	return w.keys[key]
}

func (w *Window) GetMouseButton(button MouseButton) Action {
	if !(button >= 0 && button < mouseButtonMax) {
		panic(fmt.Errorf("button is out of range: %v", button))
	}

	// Hacky mouse-emulation-via-touch.
	if !w.touches.Equal(js.Value{}) {
		switch button {
		case MouseButton1:
			if w.touches.Length() == 1 || w.touches.Length() == 3 {
				return Press
			}
		case MouseButton2:
			if w.touches.Length() == 2 || w.touches.Length() == 3 {
				return Press
			}
		}

		return Release
	}

	return w.mouseButton[button]
}

func (w *Window) GetInputMode(mode InputMode) int {
	switch mode {
	case CursorMode:
		return w.cursorMode
	default:
		panic(errors.New("not implemented"))
	}
}

var ErrInvalidParameter = errors.New("invalid parameter")
var ErrInvalidValue = errors.New("invalid value")

func (w *Window) SetInputMode(mode InputMode, value int) {
	switch mode {
	case CursorMode:
		if w.missing.pointerLock {
			log.Println("warning: Pointer Lock API unsupported")
			return
		}
		switch value {
		case CursorNormal:
			w.cursorMode = value
			document.Call("exitPointerLock")
			w.canvas.Get("style").Call("setProperty", "cursor", "initial")
			return
		case CursorHidden:
			w.cursorMode = value
			document.Call("exitPointerLock")
			w.canvas.Get("style").Call("setProperty", "cursor", "none")
			return
		case CursorDisabled:
			w.cursorMode = value
			w.canvas.Call("requestPointerLock")
			return
		default:
			panic(ErrInvalidValue)
		}
	case StickyKeysMode:
		panic(errors.New("not implemented"))
	case StickyMouseButtonsMode:
		panic(errors.New("not implemented"))
	default:
		panic(ErrInvalidParameter)
	}
}

type Key int

// Docs: https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/code
const (
	KeySpace Key = Key(iota) + 32 // TODO: Do more research, +32 is to shift it up above the mouse positions because that's how glfw does it
	KeyApostrophe
	KeyComma
	KeyMinus
	KeyPeriod
	KeySlash
	Key0
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
	KeySemicolon
	KeyEqual
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyLeftBracket
	KeyBackslash
	KeyRightBracket
	KeyGraveAccent
	KeyWorld1
	KeyWorld2
	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyInsert
	KeyDelete
	KeyRight
	KeyLeft
	KeyDown
	KeyUp
	KeyPageUp
	KeyPageDown
	KeyHome
	KeyEnd
	KeyCapsLock
	KeyScrollLock
	KeyNumLock
	KeyPrintScreen
	KeyPause
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
	KeyF25
	KeyKP0
	KeyKP1
	KeyKP2
	KeyKP3
	KeyKP4
	KeyKP5
	KeyKP6
	KeyKP7
	KeyKP8
	KeyKP9
	KeyKPDecimal
	KeyKPDivide
	KeyKPMultiply
	KeyKPSubtract
	KeyKPAdd
	KeyKPEnter
	KeyKPEqual
	KeyLeftShift
	KeyLeftControl
	KeyLeftAlt
	KeyLeftSuper
	KeyRightShift
	KeyRightControl
	KeyRightAlt
	KeyRightSuper
	KeyMenu

	KeyUnknown Key = -1
	KeyLast    Key = KeyMenu
)

// Experimental: Get String based on KeyCode
// https://developer.mozilla.org/en-US/docs/Web/API/Keyboard

// KeyboardEvent.key: The Key's text
// https://developer.mozilla.org/en-US/docs/Web/API/UI_Events/Keyboard_event_key_values

// KeyCode: Physical Positions
// https://developer.mozilla.org/en-US/docs/Web/API/UI_Events/Keyboard_event_code_values

// Contains a mapping from javascript
// TODO - some of these I wasn't sure about
var keycodeMap = map[string]Key{
	"Space":        KeySpace,
	"Quote":        KeyApostrophe, //???
	"Comma":        KeyComma,
	"Minus":        KeyMinus,
	"Period":       KeyPeriod,
	"Slash":        KeySlash,
	"Digit0":       Key0,
	"Digit1":       Key1,
	"Digit2":       Key2,
	"Digit3":       Key3,
	"Digit4":       Key4,
	"Digit5":       Key5,
	"Digit6":       Key6,
	"Digit7":       Key7,
	"Digit8":       Key8,
	"Digit9":       Key9,
	"Semicolon":    KeySemicolon,
	"Equal":        KeyEqual,
	"KeyA":         KeyA,
	"KeyB":         KeyB,
	"KeyC":         KeyC,
	"KeyD":         KeyD,
	"KeyE":         KeyE,
	"KeyF":         KeyF,
	"KeyG":         KeyG,
	"KeyH":         KeyH,
	"KeyI":         KeyI,
	"KeyJ":         KeyJ,
	"KeyK":         KeyK,
	"KeyL":         KeyL,
	"KeyM":         KeyM,
	"KeyN":         KeyN,
	"KeyO":         KeyO,
	"KeyP":         KeyP,
	"KeyQ":         KeyQ,
	"KeyR":         KeyR,
	"KeyS":         KeyS,
	"KeyT":         KeyT,
	"KeyU":         KeyU,
	"KeyV":         KeyV,
	"KeyW":         KeyW,
	"KeyX":         KeyX,
	"KeyY":         KeyY,
	"KeyZ":         KeyZ,
	"BracketLeft":  KeyLeftBracket,
	"Backslash":    KeyBackslash,
	"BracketRight": KeyRightBracket,
	//	"KeyGraveAccent": KeyGraveAccent,
	// "KeyWorld1": KeyWorld1,
	// "KeyWorld2": KeyWorld2,
	"Escape":      KeyEscape,
	"Enter":       KeyEnter,
	"Tab":         KeyTab,
	"Backspace":   KeyBackspace,
	"Insert":      KeyInsert,
	"Delete":      KeyDelete,
	"ArrowRight":  KeyRight,
	"ArrowLeft":   KeyLeft,
	"ArrowDown":   KeyDown,
	"ArrowUp":     KeyUp,
	"PageUp":      KeyPageUp,
	"PageDown":    KeyPageDown,
	"Home":        KeyHome,
	"End":         KeyEnd,
	"CapsLock":    KeyCapsLock,
	"ScrollLock":  KeyScrollLock,
	"NumLock":     KeyNumLock,
	"PrintScreen": KeyPrintScreen,
	"Pause":       KeyPause,
	"F1":          KeyF1,
	"F2":          KeyF2,
	"F3":          KeyF3,
	"F4":          KeyF4,
	"F5":          KeyF5,
	"F6":          KeyF6,
	"F7":          KeyF7,
	"F8":          KeyF8,
	"F9":          KeyF9,
	"F10":         KeyF10,
	"F11":         KeyF11,
	"F12":         KeyF12,
	"F13":         KeyF13,
	"F14":         KeyF14,
	"F15":         KeyF15,
	"F16":         KeyF16,
	"F17":         KeyF17,
	"F18":         KeyF18,
	"F19":         KeyF19,
	"F20":         KeyF20,
	"F21":         KeyF21,
	"F22":         KeyF22,
	"F23":         KeyF23,
	"F24":         KeyF24,
	// "F25": KeyF25,
	"Numpad0":        KeyKP0,
	"Numpad1":        KeyKP1,
	"Numpad2":        KeyKP2,
	"Numpad3":        KeyKP3,
	"Numpad4":        KeyKP4,
	"Numpad5":        KeyKP5,
	"Numpad6":        KeyKP6,
	"Numpad7":        KeyKP7,
	"Numpad8":        KeyKP8,
	"Numpad9":        KeyKP9,
	"NumpadDecimal":  KeyKPDecimal,
	"NumpadDivide":   KeyKPDivide,
	"NumpadMultiply": KeyKPMultiply,
	"NumpadSubtract": KeyKPSubtract,
	"NumpadAdd":      KeyKPAdd,
	"NumpadEnter":    KeyKPEnter,
	"NumpadEqual":    KeyKPEqual,
	"ShiftLeft":      KeyLeftShift,
	"ControlLeft":    KeyLeftControl,
	"AltLeft":        KeyLeftAlt,
	"OSLeft":         KeyLeftSuper,
	"MetaLeft":       KeyLeftSuper,
	"ShiftRight":     KeyRightShift,
	"ControlRight":   KeyRightControl,
	"AltRight":       KeyRightAlt,
	"OSRight":        KeyRightSuper,
	"MetaRight":      KeyRightSuper,
	"ContextMenu":    KeyMenu, // ????
}

// Contains the reverse mapping of the keycodeMap
var reverseKeycodeMap = make(map[Key]string)
func init() {
	for s, k := range keycodeMap {
		reverseKeycodeMap[k] = s
	}
}

func GetKeyScanCode(key Key) int {
	// TODO - this is wrong
	return int(key)
}

// TODO: scancode doesn't work
var keynameCache = make(map[Key]string)
func GetKeyName(key Key, scancode int) string {
	name, has := keynameCache[key]
	if has {
		return name
	}

	if !isNilOrUndefined(keyboardLayoutMap) {
		str := reverseKeycodeMap[key]
		val := keyboardLayoutMap.Call("get", str)

		if !isNilOrUndefined(val) {
			ret := val.String()
			if ret != "" {
				keynameCache[key] = ret
				return ret
			}
		}
	}

	// Fallback to qwerty defined map
	name, ok := qwertyKeyNameMap[key]
	if !ok {
		// TODO: Use scancode to lookup
		name = "Unknown"
	}

	keynameCache[key] = name
	return name
}

// toKey extracts Key from given KeyboardEvent.
func toKey(ke js.Value) Key {
	keyStr := ke.Get("code").String()
	key, ok := keycodeMap[keyStr]
	if !ok {
		return KeyUnknown
	}
	return key
}

// toModifierKey extracts ModifierKey from given KeyboardEvent.
func toModifierKey(ke js.Value) ModifierKey {
	mods := ModifierKey(0)
	if ke.Get("shiftKey").Bool() {
		mods += ModShift
	}
	if ke.Get("ctrlKey").Bool() {
		mods += ModControl
	}
	if ke.Get("altKey").Bool() {
		mods += ModAlt
	}
	if ke.Get("metaKey").Bool() {
		mods += ModSuper
	}
	return mods
}

type MouseButton int

// Documentation: https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/button#value
const (
	MouseButton1 MouseButton = 0
	MouseButton2 MouseButton = 2 // Web MouseEvent has middle and right mouse buttons in reverse order.
	MouseButton3 MouseButton = 1 // Web MouseEvent has middle and right mouse buttons in reverse order.

	MouseButtonLeft   = MouseButton1
	MouseButtonRight  = MouseButton2
	MouseButtonMiddle = MouseButton3

	MouseButton4    = 3 // Typically Browser Back
	MouseButton5    = 4 // Typically Browser Forward

	mouseButtonMax  = 5 // This is for checking buttons to see if they are mouse buttons

	// TODO - everything below this is wrong
	MouseButton6    = 5
	MouseButton7    = 6
	MouseButton8    = 7
	MouseButtonLast = 8
)

type Action int

const (
	Release Action = 0
	Press   Action = 1
	Repeat  Action = 2
)

type InputMode int

const (
	CursorMode InputMode = iota
	StickyKeysMode
	StickyMouseButtonsMode
)

const (
	CursorNormal = iota
	CursorHidden
	CursorDisabled
)

type ModifierKey int

const (
	ModShift ModifierKey = (1 << iota)
	ModControl
	ModAlt
	ModSuper
)

func WaitEvents() {
	// TODO.

	runtime.Gosched()
}

func PostEmptyEvent() {
	// TODO: Implement.
}

func DefaultWindowHints() {
	// TODO: Implement.
}

func (w *Window) SetClipboardString(str string) {
	// TODO: Implement.
}
func (w *Window) GetClipboardString() string {
	// TODO: Implement.
	return "GetClipboardString not implemented"
}

func (w *Window) SetTitle(title string) {
	document.Set("title", title)
}

func (w *Window) Show() {
	// TODO: Implement.
}

func (w *Window) Hide() {
	// TODO: Implement.
}

func (w *Window) Destroy() {
	document.Get("body").Call("removeChild", w.canvas)
	if w.fullscreen {
		if w.missing.fullscreen {
			log.Println("warning: Fullscreen API unsupported")
		} else {
			document.Call("webkitExitFullscreen")
			w.fullscreen = false
		}
	}
}


func (w *Window) SetMonitor(monitor *Monitor, xpos, ypos, width, height, refreshRate int) {
	// TODO: Not sure?
}

type CloseCallback func(w *Window)

func (w *Window) SetCloseCallback(cbfun CloseCallback) (previous CloseCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type RefreshCallback func(w *Window)

func (w *Window) SetRefreshCallback(cbfun RefreshCallback) (previous RefreshCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type SizeCallback func(w *Window, width int, height int)

func (w *Window) SetSizeCallback(cbfun SizeCallback) (previous SizeCallback) {
	w.sizeCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type CursorEnterCallback func(w *Window, entered bool)

func (w *Window) SetCursorEnterCallback(cbfun CursorEnterCallback) (previous CursorEnterCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type CharModsCallback func(w *Window, char rune, mods ModifierKey)

func (w *Window) SetCharModsCallback(cbfun CharModsCallback) (previous CharModsCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type PosCallback func(w *Window, xpos int, ypos int)

func (w *Window) SetPosCallback(cbfun PosCallback) (previous PosCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type IconifyCallback func(w *Window, iconified bool)

func (w *Window) SetIconifyCallback(cbfun IconifyCallback) (previous IconifyCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type DropCallback func(w *Window, names []string)

func (w *Window) SetDropCallback(cbfun DropCallback) (previous DropCallback) {
	// TODO: Implement. Can use HTML5 file drag and drop API?

	// TODO: Handle previous.
	return nil
}

//------------------------------------------------------------------------------
// TODO - everything below here is wrong
// Use: https://developer.mozilla.org/en-US/docs/Web/API/Gamepad_API
// ------------------------------------------------------------------------------
// TODO - these are all wrong

type Joystick int

// List all of the joysticks.
const (
	Joystick1 = iota
	Joystick2
	Joystick3
	Joystick4
	Joystick5
	Joystick6
	Joystick7
	Joystick8
	Joystick9
	Joystick10
	Joystick11
	Joystick12
	Joystick13
	Joystick14
	Joystick15
	Joystick16

	JoystickLast
)

type GamepadAxis int

const (
	AxisLeftX = iota
	AxisLeftY
	AxisRightX
	AxisRightY
	AxisLeftTrigger
	AxisRightTrigger
	AxisLast
)

type GamepadButton int

// Gamepad button IDs.
const (
	ButtonA = iota
	ButtonB
	ButtonX
	ButtonY
	ButtonLeftBumper
	ButtonRightBumper
	ButtonBack
	ButtonStart
	ButtonGuide
	ButtonLeftThumb
	ButtonRightThumb
	ButtonDpadUp
	ButtonDpadRight
	ButtonDpadDown
	ButtonDpadLeft
	ButtonLast
	ButtonCross
	ButtonCircle
	ButtonSquare
	ButtonTriangle
)

// TODO: Some of these might be wrong
var qwertyKeyNameMap map[Key]string = map[Key]string{
	KeySpace:        " ",
	KeyApostrophe:        "'", //???
	KeyComma:        ",",
	KeyMinus:        "-",
	KeyPeriod:       ".",
	KeySlash:        "/",
	Key0:       "0",
	Key1:       "1",
	Key2:       "2",
	Key3:       "3",
	Key4:       "4",
	Key5:       "5",
	Key6:       "6",
	Key7:       "7",
	Key8:       "8",
	Key9:       "9",
	KeySemicolon:    ";",
	KeyEqual:        "=",
	KeyA:         "a",
	KeyB:         "b",
	KeyC:         "c",
	KeyD:         "d",
	KeyE:         "e",
	KeyF:         "f",
	KeyG:         "g",
	KeyH:         "h",
	KeyI:         "i",
	KeyJ:         "j",
	KeyK:         "k",
	KeyL:         "l",
	KeyM:         "m",
	KeyN:         "n",
	KeyO:         "o",
	KeyP:         "p",
	KeyQ:         "q",
	KeyR:         "r",
	KeyS:         "s",
	KeyT:         "t",
	KeyU:         "u",
	KeyV:         "v",
	KeyW:         "w",
	KeyX:         "x",
	KeyY:         "y",
	KeyZ:         "z",
	KeyLeftBracket:  "[",
	KeyBackslash:    "\\",
	KeyRightBracket: "[",
	//	"KeyGraveAccent": KeyGraveAccent,
	// "KeyWorld1": KeyWorld1,
	// "KeyWorld2": KeyWorld2,
	KeyEscape:      "Esc",
	KeyEnter:       "Enter",
	KeyTab:         "Tab",
	KeyBackspace:   "Backspace",
	KeyInsert:      "Insert",
	KeyDelete:      "Delete",
	KeyRight:  "ArrowRight",
	KeyLeft:   "ArrowLeft",
	KeyDown:   "ArrowDown",
	KeyUp:     "ArrowUp",
	KeyPageUp:      "PageUp",
	KeyPageDown:    "PageDown",
	KeyHome:        "Home",
	KeyEnd:         "End",
	KeyCapsLock:    "CapsLock",
	KeyScrollLock:  "ScrollLock",
	KeyNumLock:     "NumLock",
	KeyPrintScreen: "PrintScreen",
	KeyPause:       "Pause",
	KeyF1:          "F1",
	KeyF2:          "F2",
	KeyF3:          "F3",
	KeyF4:          "F4",
	KeyF5:          "F5",
	KeyF6:          "F6",
	KeyF7:          "F7",
	KeyF8:          "F8",
	KeyF9:          "F9",
	KeyF10:         "F10",
	KeyF11:         "F11",
	KeyF12:         "F12",
	KeyF13:         "F13",
	KeyF14:         "F14",
	KeyF15:         "F15",
	KeyF16:         "F16",
	KeyF17:         "F17",
	KeyF18:         "F18",
	KeyF19:         "F19",
	KeyF20:         "F20",
	KeyF21:         "F21",
	KeyF22:         "F22",
	KeyF23:         "F23",
	KeyF24:         "F24",
	// "F25": KeyF25,
	KeyKP0:        "Numpad0",
	KeyKP1:        "Numpad1",
	KeyKP2:        "Numpad2",
	KeyKP3:        "Numpad3",
	KeyKP4:        "Numpad4",
	KeyKP5:        "Numpad5",
	KeyKP6:        "Numpad6",
	KeyKP7:        "Numpad7",
	KeyKP8:        "Numpad8",
	KeyKP9:        "Numpad9",
	KeyKPDecimal:  "NumpadDecimal",
	KeyKPDivide:   "NumpadDivide",
	KeyKPMultiply: "NumpadMultiply",
	KeyKPSubtract: "NumpadSubtract",
	KeyKPAdd:      "NumpadAdd",
	KeyKPEnter:    "NumpadEnter",
	KeyKPEqual:    "NumpadEqual",
	KeyLeftShift:      "ShiftLeft",
	KeyLeftControl:    "ControlLeft",
	KeyLeftAlt:        "AltLeft",
	// KeyLeftSuper:         "OSLeft",
	KeyLeftSuper:       "MetaLeft",
	KeyRightShift:     "ShiftRight",
	KeyRightControl:   "ControlRight",
	KeyRightAlt:       "AltRight",
	// KeyRightSuper:        "OSRight",
	KeyRightSuper:      "MetaRight",
	KeyMenu:    "ContextMenu", // ????
}
