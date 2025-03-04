package glitch

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/mainthread"
	"github.com/unitoftime/glitch/shaders"
)

const sof int = 4 // SizeOf(Float)

// TODO - rename
type ISubBuffer interface {
	Clear()
	Buffer() []byte
	Offset() int
	Len() int
	Cap() int
	SetData(any)
}

type SupportedSubBuffers interface {
	glVec4 | glVec3 | glVec2 | float32
}

type SubBuffer[T SupportedSubBuffers] struct {
	attr        shaders.Attr
	maxVerts    int
	offset      int
	vertexCount int
	buffer      []T
	sliceScale  int
	byteBuffer  []byte // This is just a byte header that points to the buffer
}

type SubSubBuffer[T SupportedSubBuffers] struct {
	Buffer []T
}

func (b *SubBuffer[T]) Clear() {
	b.buffer = b.buffer[:0]
	b.vertexCount = 0
}

func (b *SubBuffer[T]) Len() int {
	return len(b.buffer)
}

func (b *SubBuffer[T]) Cap() int {
	return cap(b.buffer)
}

// func (b *SubBuffer[T]) VertexCount() uint32 {
// 	return uint32(b.vertexCount)
// }

func (b *SubBuffer[T]) Buffer() []byte {
	copiedHeader := b.buffer
	h := (*reflect.SliceHeader)(unsafe.Pointer(&copiedHeader))
	h.Len = h.Len * b.sliceScale
	h.Cap = h.Len * b.sliceScale

	b.byteBuffer = *(*[]byte)(unsafe.Pointer(h))
	return b.byteBuffer

	// copiedHeader := b.buffer
	// h := (*reflect.SliceHeader)(unsafe.Pointer(&copiedHeader))
	// h.Len *= b.sliceScale
	// h.Cap *= b.sliceScale
	// return *(*[]byte)(unsafe.Pointer(h))
}

// func (b *SubBuffer[T]) BufferSubData(offset int) {
// 	if b.attrSize == AttrVec2 {
// 		gl.BufferSubData(gl.ARRAY_BUFFER, offset, [][3]float32(b.buffer))
// 	} else if b.attrSize == AttrVec3 {
// 		gl.BufferSubData(gl.ARRAY_BUFFER, offset, [][2]float32(b.buffer))
// 	} else {
// 		panic("Unknown")
// 	}
// }

func (b *SubBuffer[T]) Offset() int {
	return sof * int(b.attr.Size()) * b.maxVerts
}

func (b *SubBuffer[T]) SetData(data any) {
	b.buffer = data.([]T)
}

func (b *SubBuffer[T]) Reserve(count int) []T {
	start := len(b.buffer)
	b.buffer = b.buffer[:len(b.buffer)+count]
	end := len(b.buffer)
	b.vertexCount += count

	return b.buffer[start:end]
}

type bufferData struct {
	format   shaders.VertexFormat
	stride   int
	numVerts int // The maximum number of verts we can buffer

	buffers []ISubBuffer
	indices []uint32
}

type VertexBuffer struct {
	vao, vbo, ebo gl.Buffer

	data bufferData

	numVerts           uint32 // The number of vertices we currently have buffered
	numIndicesToDraw   int    // The number of indices we are currently drawing
	bufferedToGPU      bool   // Tracks whether the data has been written to the GPU
	deallocAfterBuffer bool   // If set true, once we write data to the GPU we deallocate CPU buffers
	deleted            bool   // If true, we've already deleted this
}

func NewSubBuffers(shader *Shader, numVerts, numIndices int) bufferData {
	format := shader.attrFmt
	b := bufferData{
		format:   format,
		numVerts: numVerts,
		buffers:  make([]ISubBuffer, len(format)),
		indices:  make([]uint32, numIndices),
	}

	b.stride = 0
	offset := 0
	for i := range format {
		b.stride += (format[i].Size() * sof)

		if format[i].Type == shaders.AttrVec4 {
			b.buffers[i] = &SubBuffer[glVec4]{
				attr:        format[i].Attr,
				maxVerts:    numVerts,
				vertexCount: 0,
				offset:      offset,
				buffer:      make([]glVec4, numVerts),
				sliceScale:  format[i].Size() * sof,
			}
		} else if format[i].Type == shaders.AttrVec3 {
			b.buffers[i] = &SubBuffer[glVec3]{
				attr:        format[i].Attr,
				maxVerts:    numVerts,
				vertexCount: 0,
				offset:      offset,
				buffer:      make([]glVec3, numVerts),
				sliceScale:  format[i].Size() * sof,
			}
		} else if format[i].Type == shaders.AttrVec2 {
			b.buffers[i] = &SubBuffer[glVec2]{
				attr:        format[i].Attr,
				maxVerts:    numVerts,
				vertexCount: 0,
				offset:      offset,
				buffer:      make([]glVec2, numVerts),
				sliceScale:  format[i].Size() * sof,
			}
		} else if format[i].Type == shaders.AttrFloat {
			b.buffers[i] = &SubBuffer[float32]{
				attr:        format[i].Attr,
				maxVerts:    numVerts,
				vertexCount: 0,
				offset:      offset,
				buffer:      make([]float32, numVerts),
				sliceScale:  format[i].Size() * sof,
			}
		} else {
			panic(fmt.Sprintf("Unknown format: %v", format[i]))
		}

		offset += sof * int(format[i].Size()) * numVerts
	}

	return b
}

func NewVertexBuffer2(shader *Shader, data bufferData) *VertexBuffer {
	b := &VertexBuffer{
		data: data,
	}

	mainthread.Call(func() {
		b.vao = gl.GenVertexArrays()
		b.vbo = gl.GenBuffers()
		b.ebo = gl.GenBuffers()

		gl.BindVertexArray(b.vao)

		gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, sof*data.numVerts*data.stride, nil, gl.DYNAMIC_DRAW)

		indexSize := 4 // uint32 // TODO - make this modifiable?
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexSize*len(b.data.indices), nil, gl.DYNAMIC_DRAW)

		for i := range b.data.buffers {
			switch subBuffer := b.data.buffers[i].(type) {
			case *SubBuffer[float32]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.attr.Name)
				size := int(subBuffer.attr.Size())
				gl.VertexAttribPointer(loc, size, gl.FLOAT, false, size*sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			case *SubBuffer[glVec2]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.attr.Name)
				size := int(subBuffer.attr.Size())
				gl.VertexAttribPointer(loc, size, gl.FLOAT, false, size*sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			case *SubBuffer[glVec3]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.attr.Name)
				size := int(subBuffer.attr.Size())
				gl.VertexAttribPointer(loc, size, gl.FLOAT, false, size*sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			case *SubBuffer[glVec4]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.attr.Name)
				size := int(subBuffer.attr.Size())
				// TODO!!! - gl.VertexAttribPointerWithOffset: https://github.com/go-gl/gl/pull/135/files#diff-b335630551682c19a781afebcf4d07bf978fb1f8ac04c6bf87428ed5106870f5R67
				gl.VertexAttribPointer(loc, size, gl.FLOAT, false, size*sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			default:
				panic("Unknown!")
			}
		}
	})

	b.Clear() // TODO - fix

	runtime.SetFinalizer(b, (*VertexBuffer).delete)

	return b
}

func NewVertexBuffer(shader *Shader, numVerts, numIndices int) *VertexBuffer {
	data := NewSubBuffers(shader, numVerts, numIndices)
	return NewVertexBuffer2(shader, data)
}

func (v *VertexBuffer) delete() {
	if v.deleted {
		return
	}
	v.deleted = true

	mainthread.CallNonBlock(func() {
		gl.DeleteVertexArrays(v.vao)
		gl.DeleteBuffers(v.vbo)
	})
}

func (v *VertexBuffer) deallocCPUBuffers() {
	v.data = bufferData{}
}

func (v *VertexBuffer) Clear() {
	for i := range v.data.buffers {
		v.data.buffers[i].Clear()
	}
	v.data.indices = v.data.indices[:0]
	v.numIndicesToDraw = 0
	v.numVerts = 0
	// v.materialSet = false
	v.bufferedToGPU = false
}

func (v *VertexBuffer) Reserve(indices []uint32, numVerts int, dests []interface{}) bool {
	// // If material is set and it doesn't match the reserved material
	// if v.materialSet && v.state != state {
	// 	// fmt.Println("VertexBuffer.Reserve - Material Doesn't match")
	// 	return false
	// }
	if len(v.data.indices)+len(indices) > cap(v.data.indices) {
		// fmt.Println("VertexBuffer.Reserve - Not enough index capacity")
		return false
	}
	if v.data.buffers[0].Len()+numVerts > v.data.buffers[0].Cap() {
		// fmt.Println("VertexBuffer.Reserve - Not enough vertex capacity")
		return false
	}

	v.bufferedToGPU = false

	// v.materialSet = true
	// v.state = state

	// currentElement := v.buffers[0].VertexCount()
	currentElement := v.numVerts
	for i := range indices {
		v.data.indices = append(v.data.indices, currentElement+indices[i])
	}
	v.numVerts += uint32(numVerts)
	v.numIndicesToDraw = len(v.data.indices)

	for i := range v.data.buffers {
		switch subBuffer := v.data.buffers[i].(type) {
		case *SubBuffer[float32]:
			d := dests[i].(*[]float32)
			*d = subBuffer.Reserve(numVerts)
		case *SubBuffer[glVec2]:
			d := dests[i].(*[]glVec2)
			*d = subBuffer.Reserve(numVerts)
		case *SubBuffer[glVec3]:
			d := dests[i].(*[]glVec3)
			*d = subBuffer.Reserve(numVerts)
		case *SubBuffer[glVec4]:
			d := dests[i].(*[]glVec4)
			*d = subBuffer.Reserve(numVerts)
		default:
			panic("Unknown!")
		}
	}
	return true
}

// Buffers the entire vertex buffer
// This function returns with the Element array buffer bound, so it is ready to be drawn
func (v *VertexBuffer) mainthreadBufferData() {
	gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)
	offset := 0
	var buf []byte
	for i := range v.data.buffers {
		buf = v.data.buffers[i].Buffer()
		gl.BufferSubDataByte(gl.ARRAY_BUFFER, offset, buf)
		offset += v.data.buffers[i].Offset()
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
	gl.BufferSubDataUint32(gl.ELEMENT_ARRAY_BUFFER, 0, v.data.indices)

	if v.deallocAfterBuffer {
		v.deallocCPUBuffers()
	}
	v.bufferedToGPU = true
}

func (v *VertexBuffer) mainthreadDraw() {
	gl.BindVertexArray(v.vao)

	if !v.bufferedToGPU {
		v.mainthreadBufferData()
		gl.DrawElements(gl.TRIANGLES, v.numIndicesToDraw, gl.UNSIGNED_INT, 0)
	} else {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
		gl.DrawElements(gl.TRIANGLES, v.numIndicesToDraw, gl.UNSIGNED_INT, 0)
	}
}

func (v *VertexBuffer) Draw() {
	if v.numIndicesToDraw <= 0 {
		return
	}

	state.drawVertBuffer(v)
}

// BufferPool
// TODO - Idea Improvements: You'd be able to calculate in the pass how many draws with the same material you'd be doing. Based on that you could have really well sized buffers. Also in here you could have different VertexBuffer sizes and order them as needed into a final draw slice
type BufferPool struct {
	shader            *Shader
	triangleBatchSize int
	triangleCount     int
	buffers           []*VertexBuffer
	currentIndex      int
	nextClean         int // Tracks the next clean vertex buffer (ie clean = buffers that haven't been written reserved on in this case
}

func NewBufferPool(shader *Shader, triangleBatchSize int) *BufferPool {
	return &BufferPool{
		shader:            shader,
		triangleBatchSize: triangleBatchSize,
		triangleCount:     0,
		buffers:           make([]*VertexBuffer, 0),
		currentIndex:      0,
		nextClean:         0,
	}
}

func (b *BufferPool) Clear() {
	for i := range b.buffers {
		b.buffers[i].Clear()
	}
	b.triangleCount = 0
	b.currentIndex = 0
	b.nextClean = 0
}

// Updates the buffer pool, so that the next reserve call will return a brand new vertex buffer
func (b *BufferPool) gotoNextClean() {
	b.currentIndex = b.nextClean
}

// Returns the vertexbuffer that we reserved to
func (b *BufferPool) Reserve(indices []uint32, numVerts int, dests []interface{}) *VertexBuffer {
	for i := b.currentIndex; i < len(b.buffers); i++ {
		success := b.buffers[i].Reserve(indices, numVerts, dests)
		if success {
			b.triangleCount += len(indices) / 3
			b.currentIndex = i
			b.nextClean = b.currentIndex + 1
			return b.buffers[i]
		}
	}

	// Since we are making a new buffer, ensure we have at least enough for what we are trying to reserve. Else use the default triangleBatchSize
	// TODO: For indexBatchSize: 3x is a bit hacky. assumes each triangle has 3 indices
	vertBatchSize := max(numVerts, b.triangleBatchSize)
	indexBatchSize := max(len(indices), 3*b.triangleBatchSize)

	newBuff := NewVertexBuffer(b.shader, vertBatchSize, indexBatchSize)
	success := newBuff.Reserve(indices, numVerts, dests)
	if !success {
		panic(fmt.Sprintf("Failed to reserve on freshly created buffer:\nReserve: %v, %v, %v\nOn: %v %v",
			len(indices), numVerts, len(dests), vertBatchSize, indexBatchSize,
		))
	}
	b.buffers = append(b.buffers, newBuff)
	b.triangleCount += len(indices) / 3

	b.currentIndex = len(b.buffers) - 1
	b.nextClean = b.currentIndex + 1

	return newBuff
}

// func (b *BufferPool) Draw() {
// 	lastMaterial := Material(nil)
// 	for i := range b.buffers {
// 		// fmt.Println(i, len(b.buffers[i].indices), b.buffers[i].buffers[0].Len(), b.buffers[i].buffers[0].Cap())
// 		if lastMaterial != b.buffers[i].material {
// 			lastMaterial = b.buffers[i].material
// 			if lastMaterial != nil {
// 				// fmt.Println("Binding New Material", lastMaterial)
// 				lastMaterial.Bind()
// 			}
// 		}
// 		b.buffers[i].Draw()
// 	}
// }

// func openglDraw(buffers []*VertexBuffer) {
// 	lastMaterial := Material(nil)
// 	for i := range buffers {
// 		// fmt.Println(i, len(b.buffers[i].indices), b.buffers[i].buffers[0].Len(), b.buffers[i].buffers[0].Cap())
// 		if lastMaterial != buffers[i].material {
// 			lastMaterial = buffers[i].material
// 			if lastMaterial != nil {
// 				// fmt.Println("Binding New Material", lastMaterial)
// 				lastMaterial.Bind()
// 			}
// 		}
// 		buffers[i].Draw()
// 	}
// }
