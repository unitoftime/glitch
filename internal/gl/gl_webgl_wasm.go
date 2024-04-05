// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js && wasm
// +build js,wasm

package gl

import (
	// "encoding/binary"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"syscall/js"
	"unsafe"
)

var (
	object       = js.Global().Get("Object")
	arrayBuffer  = js.Global().Get("ArrayBuffer")
	uint8Array   = js.Global().Get("Uint8Array")
	float32Array = js.Global().Get("Float32Array")
	int32Array   = js.Global().Get("Int32Array")
	uint32Array  = js.Global().Get("Uint32Array")
)

// --------------------------------------------------------------------------------

// TODO - Cache all of the .Get("").Call("")s to improve performance? https://github.com/hajimehoshi/ebiten/blob/main/internal/graphicsdriver/opengl/gl_js.go

// To prevent constantly reallocating the webgl arrays on javascript side, we just allocate one large-enough buffer and take subarrays from that when we need to copy data. Too many JS allocations was causing me to get this error: "panic: JavaScript error: Array buffer allocation failed"
var currentJsMemorySize int
var jsMemory js.Value
var jsMemoryBuffer, jsMemoryFloat32, jsMemoryInt32, jsMemoryUint32 js.Value
var jsMemoryBufferVec2, jsMemoryBufferVec3, jsMemoryBufferVec4 js.Value
var jsMemoryBufferMat4 js.Value
func init() {
	// TODO: Reasonable default?
	resizeJavascriptCopyBuffer(16*1024*1024)  // x * MB
}

func resizeJavascriptCopyBuffer(size int) {
	if currentJsMemorySize > size {
		return // Nothing to do, we are already large enough
	}

	// Else do a resize. Always remaining a multiple of 2. // TODO: Is it important to be a multiple of 2? Could also just do: currentJsMemorySize = size
	div := 1 + int(math.Log2(float64(size)))
	currentJsMemorySize = int(math.Exp2(float64(div)))
	println("Resize copybuffer: ", size, div, currentJsMemorySize)

	// TODO - Can I do something like this to avoid the extra copy?: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/WebAssembly/Memory
	jsMemory = uint8Array.New(currentJsMemorySize)

	jsMemoryBuffer = jsMemory.Get("buffer")
	jsMemoryFloat32 = float32Array.New(jsMemoryBuffer, jsMemoryBuffer.Get("byteOffset"), jsMemoryBuffer.Get("byteLength").Int()/4)
	jsMemoryInt32 = int32Array.New(jsMemoryBuffer, jsMemoryBuffer.Get("byteOffset"), jsMemoryBuffer.Get("byteLength").Int()/4)
	jsMemoryUint32 = uint32Array.New(jsMemoryBuffer, jsMemoryBuffer.Get("byteOffset"), jsMemoryBuffer.Get("byteLength").Int()/4)

	jsMemoryBufferVec2 = jsMemoryFloat32.Call("subarray", 0, 2)
	jsMemoryBufferVec3 = jsMemoryFloat32.Call("subarray", 0, 3)
	jsMemoryBufferVec4 = jsMemoryFloat32.Call("subarray", 0, 4)
	jsMemoryBufferMat4 = jsMemoryFloat32.Call("subarray", 0, 16)

	runtime.KeepAlive(jsMemory)
}

var ContextWatcher contextWatcher

type contextWatcher struct{}

var (
	fnBufferSubData js.Value
	fnBufferData js.Value
	fnBindVertexArray js.Value
	fnCreateVertexArray js.Value
	fnCreateBuffer js.Value
	fnBindBuffer js.Value
	fnBindTexture js.Value
	fnGetAttribLocation js.Value
	fnVertexAttribPointer js.Value
	fnEnableVertexAttribArray js.Value
	fnDeleteVertexArray js.Value
	fnDeleteBuffer js.Value
	fnDrawElements js.Value

	fnEnable js.Value
	fnDepthFunc js.Value
	fnDisable js.Value

	fnBindFramebuffer js.Value

	fnUniformMatrix4fv js.Value
	fnViewport js.Value

	fnClear js.Value
	fnClearColor js.Value

	fnUseProgram js.Value

	fnFinish js.Value
	fnFlush js.Value

	fnGetParameter js.Value
	fnCreateFramebuffer js.Value
	fnCreateProgram js.Value
	fnCreateTexture js.Value
	fnDeleteFramebuffer js.Value
	fnDeleteShader js.Value
	fnFramebufferTexture2D js.Value
	fnGenerateMipmap js.Value
	fnGetUniformLocation js.Value
	fnLinkProgram js.Value
	fnTexImage2D js.Value
	fnTexSubImage2D js.Value
)

func (contextWatcher) OnMakeCurrent(context interface{}) {
	// context must be a WebGLRenderingContext js.Value.
	c = context.(js.Value)

	versionStr = GetString(VERSION)
	fmt.Println("Version:", versionStr)

	// Determine the webgl mode to operate in
	if strings.HasPrefix(versionStr, "WebGL 1.0") {
		fmt.Println("Switching to WebGL 1.0 mode")
		webgl1Mode = true
	} else {
		if strings.HasPrefix(versionStr, "WebGL 2.0") {
		fmt.Println("Switching to WebGL 2.0 mode")
			webgl1Mode = false
		} else {
			fmt.Println("Was unable to determine webgl mode from version string! Sticking with webgl1 mode!")
			webgl1Mode = true
		}
	}

	// TODO: Some APIs need to be unsupported in webgl1 mode

	// WebGL1
	// TODO: Some of these have webgl1 and webgl2 usages, you'd have to check where it gets used to see how we use it
	fnBufferSubData = c.Get("bufferSubData").Call("bind", c)
	fnBufferData = c.Get("bufferData").Call("bind", c)
	fnCreateBuffer = c.Get("createBuffer").Call("bind", c)
	fnBindBuffer = c.Get("bindBuffer").Call("bind", c)
	fnBindTexture = c.Get("bindTexture").Call("bind", c)
	fnGetAttribLocation = c.Get("getAttribLocation").Call("bind", c)
	fnVertexAttribPointer = c.Get("vertexAttribPointer").Call("bind", c)
	fnEnableVertexAttribArray = c.Get("enableVertexAttribArray").Call("bind", c)
	fnDeleteBuffer = c.Get("deleteBuffer").Call("bind", c)
	fnDrawElements = c.Get("drawElements").Call("bind", c)
	fnEnable = c.Get("enable").Call("bind", c)
	fnDisable = c.Get("disable").Call("bind", c)
	fnDepthFunc = c.Get("depthFunc").Call("bind", c)
	fnUniformMatrix4fv = c.Get("uniformMatrix4fv").Call("bind", c)
	fnBindFramebuffer = c.Get("bindFramebuffer").Call("bind", c)
	fnViewport = c.Get("viewport").Call("bind", c)
	fnClear = c.Get("clear").Call("bind", c)
	fnClearColor = c.Get("clearColor").Call("bind", c)
	fnFinish = c.Get("finish").Call("bind", c)
	fnFlush = c.Get("flush").Call("bind", c)
	fnUseProgram = c.Get("useProgram").Call("bind", c)
	fnGetParameter = c.Get("getParameter").Call("bind", c)
	fnCreateFramebuffer = c.Get("createFramebuffer").Call("bind", c)
	fnCreateProgram = c.Get("createProgram").Call("bind", c)
	fnCreateTexture = c.Get("createTexture").Call("bind", c)
	fnDeleteFramebuffer = c.Get("deleteFramebuffer").Call("bind", c)
	fnDeleteShader = c.Get("deleteShader").Call("bind", c)
	fnFramebufferTexture2D = c.Get("framebufferTexture2D").Call("bind", c)
	fnGenerateMipmap = c.Get("generateMipmap").Call("bind", c)
	fnGetUniformLocation = c.Get("getUniformLocation").Call("bind", c)
	fnLinkProgram = c.Get("linkProgram").Call("bind", c)
	fnTexImage2D = c.Get("texImage2D").Call("bind", c)
	fnTexSubImage2D = c.Get("texSubImage2D").Call("bind", c)

	// WebGL2 Only
	if !webgl1Mode {
		fnBindVertexArray = c.Get("bindVertexArray").Call("bind", c)
		fnCreateVertexArray = c.Get("createVertexArray").Call("bind", c)
		fnDeleteVertexArray = c.Get("deleteVertexArray").Call("bind", c)
	}
}
func (contextWatcher) OnDetach() {
	c = js.Null()
}

// c is the current WebGL context, or nil if there is no current context.
var c js.Value
var versionStr string
var webgl1Mode bool

func sliceToByteSlice(s any) []byte {
	switch s := s.(type) {
	case []int8:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		return *(*[]byte)(unsafe.Pointer(h))
	case []int16:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 2
		h.Cap *= 2
		return *(*[]byte)(unsafe.Pointer(h))
	case []int32:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 4
		h.Cap *= 4
		return *(*[]byte)(unsafe.Pointer(h))
	case []int64:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 8
		h.Cap *= 8
		return *(*[]byte)(unsafe.Pointer(h))
	case []uint8:
		return s
	case []uint16:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 2
		h.Cap *= 2
		return *(*[]byte)(unsafe.Pointer(h))
	case []uint32:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 4
		h.Cap *= 4
		return *(*[]byte)(unsafe.Pointer(h))
	case []uint64:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 8
		h.Cap *= 8
		return *(*[]byte)(unsafe.Pointer(h))
	case []float32:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 4
		h.Cap *= 4
		return *(*[]byte)(unsafe.Pointer(h))
	case []float64:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 8
		h.Cap *= 8
		return *(*[]byte)(unsafe.Pointer(h))
	default:
		panic(fmt.Sprintf("jsutil: unexpected value at sliceToBytesSlice: %T", s))
	}
}

func SliceToTypedArray(s interface{}) (js.Value, int) {
	if s == nil {
		return js.Null(), 0
	}

	// TODO - I commented out some of these b/c I didn't need them
	switch s := s.(type) {
	// case []int8:
	// 	a := jsMemory.Call("subarray", 0, len(s))
	// 	js.CopyBytesToJS(a, sliceToByteSlice(s))
	// 	runtime.KeepAlive(s)
	// 	buf := a.Get("buffer")
	// 	return js.Global().Get("Int8Array").New(buf, a.Get("byteOffset"), a.Get("byteLength"))
	// case []int16:
	// 	a := jsMemory.Call("subarray", 0, len(s) * 2)
	// 	js.CopyBytesToJS(a, sliceToByteSlice(s))
	// 	runtime.KeepAlive(s)
	// 	buf := a.Get("buffer")
	// 	return js.Global().Get("Int16Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/2)
	case []int32:
		resizeJavascriptCopyBuffer(4 * len(s))
		js.CopyBytesToJS(jsMemory, sliceToByteSlice(s))
		runtime.KeepAlive(s)
		// return jsMemoryInt32.Call("subarray", 0, len(s))
		return jsMemoryInt32, len(s)

		// a := jsMemory.Call("subarray", 0, len(s) * 4)
		// js.CopyBytesToJS(a, sliceToByteSlice(s))
		// runtime.KeepAlive(s)
		// buf := a.Get("buffer")
		// return int32Array.New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/4)
	case []uint8:
		resizeJavascriptCopyBuffer(1 * len(s))
		js.CopyBytesToJS(jsMemory, s)
		runtime.KeepAlive(s)
		// return jsMemory.Call("subarray", 0, len(s))
		return jsMemory, len(s)

	// case []uint16:
	// 	a := jsMemory.Call("subarray", 0, len(s) * 2)
	// 	js.CopyBytesToJS(a, sliceToByteSlice(s))
	// 	runtime.KeepAlive(s)
	// 	buf := a.Get("buffer")
	// 	return js.Global().Get("Uint16Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/2)
	case []uint32:
		resizeJavascriptCopyBuffer(4 * len(s))
		js.CopyBytesToJS(jsMemory, sliceToByteSlice(s))
		runtime.KeepAlive(s)
		// return jsMemoryUint32.Call("subarray", 0, len(s))
		return jsMemoryUint32, len(s)

		// a := jsMemory.Call("subarray", 0, len(s) * 4)
		// js.CopyBytesToJS(a, sliceToByteSlice(s))
		// runtime.KeepAlive(s)
		// buf := a.Get("buffer")
		// return js.Global().Get("Uint32Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/4)
	case []float32:
		resizeJavascriptCopyBuffer(4 * len(s))
		js.CopyBytesToJS(jsMemory, sliceToByteSlice(s))
		runtime.KeepAlive(s)
		// return jsMemoryFloat32.Call("subarray", 0, len(s))
		return jsMemoryFloat32, len(s)

		// a := jsMemory.Call("subarray", 0, len(s) * 4)
		// js.CopyBytesToJS(a, sliceToByteSlice(s))
		// runtime.KeepAlive(s)
		// buf := a.Get("buffer")
		// return float32Array.New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/4)
	// case []float64:
	// 	a := jsMemory.Call("subarray", 0, len(s) * 8)
	// 	js.CopyBytesToJS(a, sliceToByteSlice(s))
	// 	runtime.KeepAlive(s)
	// 	buf := a.Get("buffer")
	// 	return js.Global().Get("Float64Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/8)
	default:
		panic(fmt.Sprintf("jsutil: unexpected value at SliceToTypedArray: %T", s))
	}
}

// -----

func float32SliceToByteSlice(s []float32) []byte {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	h.Len *= 4
	h.Cap *= 4
	return *(*[]byte)(unsafe.Pointer(h))
}

func float32SliceToTypedArray(s []float32) (js.Value, int) {
	if s == nil {
		return js.Null(), 0
	}

	resizeJavascriptCopyBuffer(4 * len(s))
	js.CopyBytesToJS(jsMemory, float32SliceToByteSlice(s))
	runtime.KeepAlive(s)
	return jsMemoryFloat32, len(s)
}

// -----

func uint32SliceToByteSlice(s []uint32) []byte {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	h.Len *= 4
	h.Cap *= 4
	return *(*[]byte)(unsafe.Pointer(h))
}

func uint32SliceToTypedArray(s []uint32) (js.Value, int) {
	if s == nil {
		return js.Null(), 0
	}

	resizeJavascriptCopyBuffer(4 * len(s))
	js.CopyBytesToJS(jsMemory, uint32SliceToByteSlice(s))
	runtime.KeepAlive(s)
	return jsMemoryUint32, len(s)
}

// -----

func byteSliceToTypedArray(s []byte) (js.Value, int) {
	if s == nil {
		return js.Null(), 0
	}

	resizeJavascriptCopyBuffer(1 * len(s))
	js.CopyBytesToJS(jsMemory, s)
	runtime.KeepAlive(s)
	return jsMemory, len(s)
}

//--------------------------------------------------------------------------------

func GenVertexArrays() Buffer {
	return Buffer{Value: fnCreateVertexArray.Invoke()}
	// if webgl1Mode {
	// 	return CreateBuffer()
	// }
}

// TODO: right now I force you to make them 1 at a time
func GenBuffers() Buffer {
	return CreateBuffer()
}

func BindVertexArray(b Buffer) {
	fnBindVertexArray.Invoke(b.Value)

	// if webgl1Mode {
	// 	BindBuffer(ARRAY_BUFFER, b)
	// 	return
	// }
}

func DeleteBuffers(b Buffer) {
	fnDeleteBuffer.Invoke(b.Value)
	// c.Call("deleteBuffer", b.Value)
}

func DeleteVertexArrays(b Buffer) {
	fnDeleteVertexArray.Invoke(b.Value)
	// c.Call("deleteVertexArray", b.Value)
}

func BufferData(target Enum, size int, data any, usage Enum) {
	if data == nil {
		// Note: Webgl2 only
		fnBufferData.Invoke(int(target), size, int(usage))
		return
	}

	array, length := SliceToTypedArray(data)
	subarray := array.Call("subarray", 0, length)
	fnBufferData.Invoke(int(target), subarray, int(usage))
}

func BlitFramebuffer(srcX0 int32, srcY0 int32, srcX1 int32, srcY1 int32, dstX0 int32, dstY0 int32, dstX1 int32, dstY1 int32, mask uint32, filter uint32) {
	panic("Not supported!") // TODO
//	gl.BlitFrameBuffer(srcX0 int32, srcY0 int32, srcX1 int32, srcY1 int32, dstX0 int32, dstY0 int32, dstX1 int32, dstY1 int32, mask uint32, filter uint32)
}

// func PtrOffset(offset int) unsafe.Pointer {
// 	return gl.PtrOffset(offset)
// }

// func Ptr(data interface{}) unsafe.Pointer {
// 	return gl.Ptr(data)
// }

// func DrawBuffer(target Enum) {
// 	// TODO - revisit, for some reason drawBuffers here takes an array of ints
// 	slice := []int32{int32(target)}
// 	array, length := SliceToTypedArray(slice)
// 	subarray := array.Call("subarray", 0, length)
// 	c.Call("drawBuffers", subarray)
// 	//	gl.DrawBuffer(uint32(target))
// }

func ReadBuffer(target Enum) {
	c.Call("readBuffer", int(target))
	//	gl.ReadBuffer(uint32(target))
}

func PolygonMode(face, mode Enum) {
	//	fmt.Println("Error: PolygonMode not supported in webgl")
}


func ActiveTexture(texture Enum) {
	c.Call("activeTexture", int(texture))
}

func AttachShader(p Program, s Shader) {
	c.Call("attachShader", p.Value, s.Value)
}

func BindAttribLocation(p Program, a Attrib, name string) {
	c.Call("bindAttribLocation", p.Value, a.Value, name)
}

func BindBuffer(target Enum, b Buffer) {
	fnBindBuffer.Invoke(int(target), b.Value)
	// c.Call("bindBuffer", int(target), b.Value)
}

func BindFramebuffer(target Enum, fb Framebuffer) {
	fnBindFramebuffer.Invoke(int(target), fb.Value)
	// c.Call("bindFramebuffer", int(target), fb.Value)
}

func BindRenderbuffer(target Enum, rb Renderbuffer) {
	c.Call("bindRenderbuffer", int(target), rb.Value)
}

func BindTexture(target Enum, t Texture) {
	fnBindTexture.Invoke(int(target), t.Value)
	// c.Call("bindTexture", int(target), t.Value)
}

func BlendColor(red, green, blue, alpha float32) {
	c.Call("blendColor", red, green, blue, alpha)
}

func BlendEquation(mode Enum) {
	c.Call("blendEquation", int(mode))
}

func BlendEquationSeparate(modeRGB, modeAlpha Enum) {
	c.Call("blendEquationSeparate", modeRGB, modeAlpha)
}

func BlendFunc(sfactor, dfactor Enum) {
	c.Call("blendFunc", int(sfactor), int(dfactor))
}

func BlendFuncSeparate(sfactorRGB, dfactorRGB, sfactorAlpha, dfactorAlpha Enum) {
	c.Call("blendFuncSeparate", int(sfactorRGB), int(dfactorRGB), int(sfactorAlpha), int(dfactorAlpha))
}

// func BufferData(target Enum, data interface{}, usage Enum) {
// 	c.Call("bufferData", int(target), SliceToTypedArray(data), int(usage))
// }

// func BufferInit(target Enum, size int, usage Enum) {
// 	c.Call("bufferData", int(target), size, int(usage))
// }

func BufferSubDataByte(target Enum, offset int, data []byte) {
	array, length := byteSliceToTypedArray(data)
	fnBufferSubData.Invoke(int(target), offset, array, 0, length)
}

func BufferSubDataUint32(target Enum, offset int, data []uint32) {
	array, length := uint32SliceToTypedArray(data)
	fnBufferSubData.Invoke(int(target), offset, array, 0, length)
}

// Note: I removed this because it requires me to do interface-based type switches which causes allocs
// func BufferSubData(target Enum, offset int, data any) {
// 	array, length := SliceToTypedArray(data)
// 	fnBufferSubData.Invoke(int(target), offset, array, 0, length)
// }

// func GetBufferSubData(target Enum, offset int, data interface{}) {
// 	array, length := SliceToTypedArray(data)
// 	subarray := array.Call("subarray", 0, length)
// 	c.Call("getBufferSubData", int(target), offset, subarray)
// 	// size := 0
// 	// // TODO - other types
// 	// switch t := data.(type) {
// 	// case *[]float32:
// 	// 	size = len(*t) * 4
// 	// 	c.Call("getBufferSubData", int(target), offset, SliceToTypedArray(data))
// 	// 	// gl.GetBufferSubData(uint32(target), offset, size, gl.Ptr(*t))
// 	// case *[]byte:
// 	// 	size = len(*t)
// 	// 	gl.GetBufferSubData(uint32(target), offset, size, gl.Ptr(*t))
// 	// default:
// 	// 	panic("Invalid data type!")
// 	// }
// }

func CheckFramebufferStatus(target Enum) Enum {
	return Enum(c.Call("checkFramebufferStatus", int(target)).Int())
}

func Clear(mask Enum) {
	fnClear.Invoke(int(mask))
	// c.Call("clear", int(mask))
}

func ClearColor(red, green, blue, alpha float32) {
	fnClearColor.Invoke(red, green, blue, alpha)
	// c.Call("clearColor", red, green, blue, alpha)
}

// func ClearDepthf(d float32) {
// 	c.Call("clearDepth", d)
// }

// func ClearStencil(s int) {
// 	c.Call("clearStencil", s)
// }

// func ColorMask(red, green, blue, alpha bool) {
// 	c.Call("colorMask", red, green, blue, alpha)
// }

func CompileShader(s Shader) {
	c.Call("compileShader", s.Value)
}

// func CompressedTexImage2D(target Enum, level int, internalformat Enum, width, height, border int, data interface{}) {
// 	array, length := SliceToTypedArray(data)
// 	subarray := array.Call("subarray", 0, length)
// 	c.Call("compressedTexImage2D", int(target), level, internalformat, width, height, border, subarray)
// }

// func CompressedTexSubImage2D(target Enum, level, xoffset, yoffset, width, height int, format Enum, data interface{}) {
// 	array, length := SliceToTypedArray(data)
// 	subarray := array.Call("subarray", 0, length)
// 	c.Call("compressedTexSubImage2D", int(target), level, xoffset, yoffset, width, height, format, subarray)
// }

// func CopyTexImage2D(target Enum, level int, internalformat Enum, x, y, width, height, border int) {
// 	c.Call("copyTexImage2D", int(target), level, internalformat, x, y, width, height, border)
// }

// func CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
// 	c.Call("copyTexSubImage2D", int(target), level, xoffset, yoffset, x, y, width, height)
// }

func CreateBuffer() Buffer {
	// return Buffer{Value: c.Call("createBuffer")}
	return Buffer{Value: fnCreateBuffer.Invoke()}
}

func CreateFramebuffer() Framebuffer {
	return Framebuffer{Value: fnCreateFramebuffer.Invoke()}
	// return Framebuffer{Value: c.Call("createFramebuffer")}
}

func CreateProgram() Program {
	return Program{Value: fnCreateProgram.Invoke("createProgram")}
	// return Program{Value: c.Call("createProgram")}
}

// func CreateRenderbuffer() Renderbuffer {
// 	return Renderbuffer{Value: c.Call("createRenderbuffer")}
// }

func CreateShader(ty Enum) Shader {
	return Shader{Value: c.Call("createShader", int(ty))}
}

func CreateTexture() Texture {
	return Texture{Value: fnCreateTexture.Invoke()}
	// return Texture{Value: c.Call("createTexture")}
}

// func CullFace(mode Enum) {
// 	c.Call("cullFace", int(mode))
// }

// func DeleteBuffer(v Buffer) {
// 	c.Call("deleteBuffer", v.Value)
// }

func DeleteFramebuffer(v Framebuffer) {
	fnDeleteFramebuffer.Invoke(v.Value)
	// c.Call("deleteFramebuffer", v.Value)
}

func DeleteProgram(p Program) {
	c.Call("deleteProgram", p.Value)
}

// func DeleteRenderbuffer(v Renderbuffer) {
// 	c.Call("deleteRenderbuffer", v.Value)
// }

func DeleteShader(s Shader) {
	fnDeleteShader.Invoke(s.Value)
	// c.Call("deleteShader", s.Value)
}

func DeleteTexture(v Texture) {
	c.Call("deleteTexture", v.Value)
}

func DepthFunc(fn Enum) {
	fnDepthFunc.Invoke(int(fn))
	// c.Call("depthFunc", int(fn))
}

// func DepthMask(flag bool) {
// 	c.Call("depthMask", flag)
// }

// func DepthRangef(n, f float32) {
// 	c.Call("depthRange", n, f)
// }

// func DetachShader(p Program, s Shader) {
// 	c.Call("detachShader", p.Value, s.Value)
// }

func Disable(cap Enum) {
	fnDisable.Invoke(int(cap))
	// c.Call("disable", int(cap))
}

// func DisableVertexAttribArray(a Attrib) {
// 	c.Call("disableVertexAttribArray", a.Value)
// }

// func DrawArrays(mode Enum, first, count int) {
// 	c.Call("drawArrays", int(mode), first, count)
// }

// TODO - webgl1 won't work until I change all calls to this to use UNSIGNED_BYTE or UNSIGNED_SHORT as the type (ty Enum). https://registry.khronos.org/OpenGL-Refpages/es2.0/xhtml/glDrawElements.xml
func DrawElements(mode Enum, count int, ty Enum, offset int) {
	fnDrawElements.Invoke(int(mode), count, int(ty), offset)
	// c.Call("drawElements", int(mode), count, int(ty), offset)
}

func Enable(cap Enum) {
	fnEnable.Invoke(int(cap))
	// c.Call("enable", int(cap))
}

func EnableVertexAttribArray(a Attrib) {
	fnEnableVertexAttribArray.Invoke(a.Value)
	// c.Call("enableVertexAttribArray", a.Value)
}

func Finish() {
	fnFinish.Invoke()
}

func Flush() {
	fnFlush.Invoke()
}

// func FramebufferRenderbuffer(target, attachment, rbTarget Enum, rb Renderbuffer) {
// 	c.Call("framebufferRenderbuffer", target, attachment, int(rbTarget), rb.Value)
// }

func FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	fnFramebufferTexture2D.Invoke(int(target), int(attachment), int(texTarget), t.Value, level)
	// c.Call("framebufferTexture2D", int(target), int(attachment), int(texTarget), t.Value, level)
}

// func FrontFace(mode Enum) {
// 	c.Call("frontFace", int(mode))
// }

func GenerateMipmap(target Enum) {
	fnGenerateMipmap.Invoke(int(target))
}

// func GetActiveAttrib(p Program, index uint32) (name string, size int, ty Enum) {
// 	ai := c.Call("getActiveAttrib", p.Value, index)
// 	return ai.Get("name").String(), ai.Get("size").Int(), Enum(ai.Get("type").Int())
// }

// func GetActiveUniform(p Program, index uint32) (name string, size int, ty Enum) {
// 	ai := c.Call("getActiveUniform", p.Value, index)
// 	return ai.Get("name").String(), ai.Get("size").Int(), Enum(ai.Get("type").Int())
// }

// func GetAttachedShaders(p Program) []Shader {
// 	objs := c.Call("getAttachedShaders", p.Value)
// 	shaders := make([]Shader, objs.Length())
// 	for i := 0; i < objs.Length(); i++ {
// 		shaders[i] = Shader{Value: objs.Index(i)}
// 	}
// 	return shaders
// }

func GetAttribLocation(p Program, name string) Attrib {
	return Attrib{Value: fnGetAttribLocation.Invoke(p.Value, name).Int()}
	// return Attrib{Value: c.Call("getAttribLocation", p.Value, name).Int()}
}

// func GetBooleanv(dst []bool, pname Enum) {
// 	println("GetBooleanv: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	result := c.Call("getParameter", int(pname))
// 	length := result.Length()
// 	for i := 0; i < length; i++ {
// 		dst[i] = result.Index(i).Bool()
// 	}
// }

// func GetFloatv(dst []float32, pname Enum) {
// 	println("GetFloatv: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	result := c.Call("getParameter", int(pname))
// 	length := result.Length()
// 	for i := 0; i < length; i++ {
// 		dst[i] = float32(result.Index(i).Float())
// 	}
// }

// func GetIntegerv(pname Enum, data []int32) {
// 	result := c.Call("getParameter", int(pname))
// 	length := result.Length()
// 	for i := 0; i < length; i++ {
// 		data[i] = int32(result.Index(i).Int())
// 	}
// }

// // func GetInteger(pname Enum) int {
// // 	return c.Call("getParameter", int(pname)).Int()
// // }
// func GetInteger(pname Enum) Object {
// 	return Object{c.Call("getParameter", int(pname))}
// }


// func GetBufferParameteri(target, pname Enum) int {
// 	return c.Call("getBufferParameter", int(target), int(pname)).Int()
// }

func GetError() Enum {
	return Enum(c.Call("getError").Int())
}

// func GetBoundFramebuffer() Framebuffer {
// 	return Framebuffer{Value: c.Call("getParameter", FRAMEBUFFER_BINDING)}
// }

// func GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
// 	return c.Call("getFramebufferAttachmentParameter", int(target), int(attachment), int(pname)).Int()
// }

func GetProgrami(p Program, pname Enum) int {
	switch pname {
	case DELETE_STATUS, LINK_STATUS, VALIDATE_STATUS:
		if c.Call("getProgramParameter", p.Value, int(pname)).Bool() {
			return TRUE
		}
		return FALSE
	default:
		return c.Call("getProgramParameter", p.Value, int(pname)).Int()
	}
}

func GetProgramInfoLog(p Program) string {
	return c.Call("getProgramInfoLog", p.Value).String()
}

// func GetRenderbufferParameteri(target, pname Enum) int {
// 	return c.Call("getRenderbufferParameter", int(target), int(pname)).Int()
// }

func GetShaderi(s Shader, pname Enum) int {
	val := c.Call("getShaderParameter", s.Value, int(pname))
	if val.IsNull() {
		return FALSE
	}

	switch pname {
	case DELETE_STATUS, COMPILE_STATUS:
		if val.Bool() {
			return TRUE
		}
		return FALSE
	default:
		return val.Int()
	}

	// Bug: syscall/js: call of Value.Bool on null (Line: 	// 	if c.Call("getShaderParameter", s.Value, int(pname)).Bool() {)
	// switch pname {
	// case DELETE_STATUS, COMPILE_STATUS:
	// 	if c.Call("getShaderParameter", s.Value, int(pname)).Bool() {
	// 		return TRUE
	// 	}
	// 	return FALSE
	// default:
	// 	return c.Call("getShaderParameter", s.Value, int(pname)).Int()
	// }
}

func GetShaderInfoLog(s Shader) string {
	return c.Call("getShaderInfoLog", s.Value).String()
}

// func GetShaderPrecisionFormat(shadertype, precisiontype Enum) (rangeMin, rangeMax, precision int) {
// 	println("GetShaderPrecisionFormat: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	format := c.Call("getShaderPrecisionFormat", shadertype, precisiontype)
// 	rangeMin = format.Get("rangeMin").Int()
// 	rangeMax = format.Get("rangeMax").Int()
// 	precision = format.Get("precision").Int()
// 	return
// }

func GetShaderSource(s Shader) string {
	return c.Call("getShaderSource", s.Value).String()
}

func GetString(pname Enum) string {
	// return fnGetParameter.Invoke(int(pname)).String()
	return c.Call("getParameter", int(pname)).String()
}

// func GetTexParameterfv(dst []float32, target, pname Enum) {
// 	dst[0] = float32(c.Call("getTexParameter", int(pname)).Float())
// }

// func GetTexParameteriv(dst []int32, target, pname Enum) {
// 	dst[0] = int32(c.Call("getTexParameter", int(pname)).Int())
// }

// func GetUniformfv(dst []float32, src Uniform, p Program) {
// 	println("GetUniformfv: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	result := c.Call("getUniform")
// 	length := result.Length()
// 	for i := 0; i < length; i++ {
// 		dst[i] = float32(result.Index(i).Float())
// 	}
// }

// func GetUniformiv(dst []int32, src Uniform, p Program) {
// 	println("GetUniformiv: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	result := c.Call("getUniform")
// 	length := result.Length()
// 	for i := 0; i < length; i++ {
// 		dst[i] = int32(result.Index(i).Int())
// 	}
// }

func GetUniformLocation(p Program, name string) Uniform {
	return Uniform{Value: fnGetUniformLocation.Invoke(p.Value, name)}
	// return Uniform{Value: c.Call("getUniformLocation", p.Value, name)}
}

// func GetVertexAttribf(src Attrib, pname Enum) float32 {
// 	return float32(c.Call("getVertexAttrib", src.Value, int(pname)).Float())
// }

// func GetVertexAttribfv(dst []float32, src Attrib, pname Enum) {
// 	println("GetVertexAttribfv: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	result := c.Call("getVertexAttrib")
// 	length := result.Length()
// 	for i := 0; i < length; i++ {
// 		dst[i] = float32(result.Index(i).Float())
// 	}
// }

// func GetVertexAttribi(src Attrib, pname Enum) int32 {
// 	return int32(c.Call("getVertexAttrib", src.Value, int(pname)).Int())
// }

// func GetVertexAttribiv(dst []int32, src Attrib, pname Enum) {
// 	println("GetVertexAttribiv: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	result := c.Call("getVertexAttrib")
// 	length := result.Length()
// 	for i := 0; i < length; i++ {
// 		dst[i] = int32(result.Index(i).Int())
// 	}
// }

// func Hint(target, mode Enum) {
// 	c.Call("hint", int(target), int(mode))
// }

// func IsBuffer(b Buffer) bool {
// 	return c.Call("isBuffer", b.Value).Bool()
// }

// func IsEnabled(cap Enum) bool {
// 	return c.Call("isEnabled", int(cap)).Bool()
// }

// func IsFramebuffer(fb Framebuffer) bool {
// 	return c.Call("isFramebuffer", fb.Value).Bool()
// }

// func IsProgram(p Program) bool {
// 	return c.Call("isProgram", p.Value).Bool()
// }

// func IsRenderbuffer(rb Renderbuffer) bool {
// 	return c.Call("isRenderbuffer", rb.Value).Bool()
// }

// func IsShader(s Shader) bool {
// 	return c.Call("isShader", s.Value).Bool()
// }

// func IsTexture(t Texture) bool {
// 	return c.Call("isTexture", t.Value).Bool()
// }

// func LineWidth(width float32) {
// 	c.Call("lineWidth", width)
// }

func LinkProgram(p Program) {
	fnLinkProgram.Invoke(p.Value)
	// c.Call("linkProgram", p.Value)
}

// func PixelStorei(pname Enum, param int32) {
// 	c.Call("pixelStorei", int(pname), param)
// }

// func PolygonOffset(factor, units float32) {
// 	c.Call("polygonOffset", factor, units)
// }

// func ReadPixels(dst []byte, x, y, width, height int, format, ty Enum) {
// 	println("ReadPixels: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	if ty == Enum(UNSIGNED_BYTE) {
// 		c.Call("readPixels", x, y, width, height, format, int(ty), dst)
// 	} else {
// 		tmpDst := make([]float32, len(dst)/4)
// 		c.Call("readPixels", x, y, width, height, format, int(ty), tmpDst)
// 		for i, f := range tmpDst {
// 			binary.LittleEndian.PutUint32(dst[i*4:], math.Float32bits(f))
// 		}
// 	}
// }

func ReleaseShaderCompiler() {
	// do nothing
}

// func RenderbufferStorage(target, internalFormat Enum, width, height int) {
// 	c.Call("renderbufferStorage", target, internalFormat, width, height)
// }

// func SampleCoverage(value float32, invert bool) {
// 	c.Call("sampleCoverage", value, invert)
// }

// func Scissor(x, y, width, height int32) {
// 	c.Call("scissor", x, y, width, height)
// }

func ShaderSource(s Shader, src string) {
	c.Call("shaderSource", s.Value, src)
}

// func StencilFunc(fn Enum, ref int, mask uint32) {
// 	c.Call("stencilFunc", fn, ref, mask)
// }

// func StencilFuncSeparate(face, fn Enum, ref int, mask uint32) {
// 	c.Call("stencilFuncSeparate", face, fn, ref, mask)
// }

// func StencilMask(mask uint32) {
// 	c.Call("stencilMask", mask)
// }

// func StencilMaskSeparate(face Enum, mask uint32) {
// 	c.Call("stencilMaskSeparate", face, mask)
// }

// func StencilOp(fail, zfail, zpass Enum) {
// 	c.Call("stencilOp", fail, zfail, zpass)
// }

// func StencilOpSeparate(face, sfail, dpfail, dppass Enum) {
// 	c.Call("stencilOpSeparate", face, sfail, dpfail, dppass)
// }

// Note: I removed this because it requires me to do interface-based type switches which causes allocs
// func TexImage2D(target Enum, level int, width, height int, format Enum, ty Enum, data interface{}) {
// 	array, length := SliceToTypedArray(data)
// 	if !array.IsNull() {
// 		subarray := array.Call("subarray", 0, length)
// 		fnTexImage2D.Invoke(int(target), level, int(format), width, height, 0, int(format), int(ty), subarray)
// 	} else {
// 		fnTexImage2D.Invoke(int(target), level, int(format), width, height, 0, int(format), int(ty), nil)
// 	}
// }

func TexImage2DFull(target Enum, level int, format1 Enum, width, height int, format Enum, ty Enum, data []byte) {
	array, length := byteSliceToTypedArray(data)
	if !array.IsNull() {
		subarray := array.Call("subarray", 0, length)
		fnTexImage2D.Invoke(int(target), level, int(format1), width, height, 0, int(format), int(ty), subarray)
	} else {
		fnTexImage2D.Invoke(int(target), level, int(format1), width, height, 0, int(format), int(ty), nil)
	}
}


func TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	array, length := byteSliceToTypedArray(data)
	if !array.IsNull() {
		subarray := array.Call("subarray", 0, length)
		fnTexSubImage2D.Invoke(int(target), level, x, y, width, height, int(format), int(ty), subarray)
	} else {
		// TODO: is this the correct behavior?
		fnTexSubImage2D.Invoke(int(target), level, x, y, width, height, int(format), int(ty), nil)
	}
	// subarray := array.Call("subarray", 0, length)
	// c.Call("texSubImage2D", int(target), level, x, y, width, height, format, int(ty), subarray)
}

// func TexParameterf(target, pname Enum, param float32) {
// 	c.Call("texParameterf", int(target), int(pname), param)
// }

// func TexParameterfv(target, pname Enum, params []float32) {
// 	println("TexParameterfv: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	for _, param := range params {
// 		c.Call("texParameterf", int(target), int(pname), param)
// 	}
// }

func TexParameteri(target, pname Enum, param int) {
	c.Call("texParameteri", int(target), int(pname), param)
}

// func TexParameteriv(target, pname Enum, params []int32) {
// 	println("TexParameteriv: not yet tested (TODO: remove this after it's confirmed to work. Your feedback is welcome.)")
// 	for _, param := range params {
// 		c.Call("texParameteri", int(target), int(pname), param)
// 	}
// }

// func Uniform1f(dst Uniform, v float32) {
// 	c.Call("uniform1f", dst.Value, v)
// }

func Uniform1fv(dst Uniform, src []float32) {
	// TODO: invoke
	array, length := float32SliceToTypedArray(src)
	subarray := array.Call("subarray", 0, length)
	c.Call("uniform1fv", dst.Value, subarray)
}

// func Uniform1i(dst Uniform, v int) {
// 	c.Call("uniform1i", dst.Value, v)
// }

// func Uniform1iv(dst Uniform, src []int32) {
// 	c.Call("uniform1iv", dst.Value, src)
// }

// func Uniform2f(dst Uniform, v0, v1 float32) {
// 	c.Call("uniform2f", dst.Value, v0, v1)
// }

// func Uniform2fv(dst Uniform, src []float32) {
// 	// c.Call("uniform2fv", dst.Value, src)

// 	SliceToTypedArray(src)
// 	c.Call("uniform3fv", dst.Value, jsMemoryBufferVec2)
// }

// func Uniform2i(dst Uniform, v0, v1 int) {
// 	c.Call("uniform2i", dst.Value, v0, v1)
// }

// func Uniform2iv(dst Uniform, src []int32) {
// 	c.Call("uniform2iv", dst.Value, src)
// }

// func Uniform3f(dst Uniform, v0, v1, v2 float32) {
// 	c.Call("uniform3f", dst.Value, v0, v1, v2)
// }

// func Uniform3fv(dst Uniform, src []float32) {
// 	// c.Call("uniform3fv", dst.Value, src)

// 	SliceToTypedArray(src)
// 	c.Call("uniform3fv", dst.Value, jsMemoryBufferVec3)

// 	// array, length := SliceToTypedArray(src)
// 	// subarray := array.Call("subarray", 0, length)
// 	// c.Call("uniform3fv", dst.Value, subarray)
// }

// func Uniform3i(dst Uniform, v0, v1, v2 int32) {
// 	c.Call("uniform3i", dst.Value, v0, v1, v2)
// }

// func Uniform3iv(dst Uniform, src []int32) {
// 	c.Call("uniform3iv", dst.Value, src)
// }

// func Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
// 	c.Call("uniform4f", dst.Value, v0, v1, v2, v3)
// }

// func Uniform4fv(dst Uniform, src []float32) {
// 	// c.Call("uniform4fv", dst.Value, src)

// 	SliceToTypedArray(src)
// 	c.Call("uniform3fv", dst.Value, jsMemoryBufferVec4)

// 	// array, length := SliceToTypedArray(src)
// 	// subarray := array.Call("subarray", 0, length)
// 	// c.Call("uniform4fv", dst.Value, subarray) // TODO - I think probably most uniforms need this
// }

// func Uniform4i(dst Uniform, v0, v1, v2, v3 int32) {
// 	c.Call("uniform4i", dst.Value, v0, v1, v2, v3)
// }

// func Uniform4iv(dst Uniform, src []int32) {
// 	c.Call("uniform4iv", dst.Value, src)
// }

// func UniformMatrix2fv(dst Uniform, src []float32) {
// 	array, length := SliceToTypedArray(src)
// 	subarray := array.Call("subarray", 0, length)
// 	c.Call("uniformMatrix2fv", dst.Value, false, subarray)
// }

// func UniformMatrix3fv(dst Uniform, src []float32) {
// 	array, length := SliceToTypedArray(src)
// 	subarray := array.Call("subarray", 0, length)
// 	c.Call("uniformMatrix3fv", dst.Value, false, subarray)
// }

func UniformMatrix4fv(dst Uniform, src []float32) {
	float32SliceToTypedArray(src)
	fnUniformMatrix4fv.Invoke(dst.Value, false, jsMemoryBufferMat4)
}

func UseProgram(p Program) {
	// Workaround for js.Value zero value.
	if p.Value.Equal(js.Value{}) {
		p.Value = js.Null()
	}
	fnUseProgram.Invoke(p.Value)
	// c.Call("useProgram", p.Value)
}

// func ValidateProgram(p Program) {
// 	c.Call("validateProgram", p.Value)
// }

// func VertexAttrib1f(dst Attrib, x float32) {
// 	c.Call("vertexAttrib1f", dst.Value, x)
// }

// func VertexAttrib1fv(dst Attrib, src []float32) {
// 	c.Call("vertexAttrib1fv", dst.Value, src)
// }

// func VertexAttrib2f(dst Attrib, x, y float32) {
// 	c.Call("vertexAttrib2f", dst.Value, x, y)
// }

// func VertexAttrib2fv(dst Attrib, src []float32) {
// 	c.Call("vertexAttrib2fv", dst.Value, src)
// }

// func VertexAttrib3f(dst Attrib, x, y, z float32) {
// 	c.Call("vertexAttrib3f", dst.Value, x, y, z)
// }

// func VertexAttrib3fv(dst Attrib, src []float32) {
// 	c.Call("vertexAttrib3fv", dst.Value, src)
// }

// func VertexAttrib4f(dst Attrib, x, y, z, w float32) {
// 	c.Call("vertexAttrib4f", dst.Value, x, y, z, w)
// }

// func VertexAttrib4fv(dst Attrib, src []float32) {
// 	c.Call("vertexAttrib4fv", dst.Value, src)
// }

func VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	fnVertexAttribPointer.Invoke(dst.Value, size, int(ty), normalized, stride, offset)
	// c.Call("vertexAttribPointer", dst.Value, size, int(ty), normalized, stride, offset)
}

// func VertexAttribIPointer(dst Attrib, size int, ty Enum, stride, offset int) {
// 	c.Call("vertexAttribIPointer", dst.Value, size, int(ty), stride, offset)
// }

func Viewport(x, y, width, height int) {
	// c.Call("viewport", x, y, width, height)
	fnViewport.Invoke(x, y, width, height)
}
