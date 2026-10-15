package glitch

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/unitoftime/glitch/internal/gl"
	"github.com/unitoftime/glitch/internal/mainthread"
	"github.com/unitoftime/glitch/shaders"
)

const sof int = 4 // SizeOf(Float)
const soi int = 4 // SizeOf(Index) // TODO - make this modifiable?

// TODO - rename
type ISubBuffer interface {
	Clear()
	Buffer() []byte
	VertexSize() int // The number of bytes per vertex
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
	if len(b.buffer) == 0 {
		return nil
	}
	b.byteBuffer = unsafe.Slice((*byte)(unsafe.Pointer(&b.buffer[0])), len(b.buffer)*b.sliceScale)
	return b.byteBuffer
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

func (b *SubBuffer[T]) VertexSize() int {
	return b.sliceScale
}

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

// A VertexBuffer is either drawn whole (eg a buffered mesh), or used as a streaming chunk where batched geometry is appended
// throughout the frame and drawn in ranges. Each range only uploads the data appended since the last upload
type VertexBuffer struct {
	vao, vbo, ebo    gl.Buffer
	vboSize, eboSize int // The byte sizes of the GPU buffers

	data bufferData

	numVerts           uint32 // The number of vertices we currently have buffered
	numIndices         int    // The number of indices we currently have buffered
	uploadedVerts      int    // The number of vertices that have been written to the GPU
	uploadedIndices    int    // The number of indices that have been written to the GPU
	drawnVerts         int    // Streaming: The number of vertices that have already been drawn as part of a range
	drawnIndices       int    // Streaming: The number of indices that have already been drawn as part of a range
	needsOrphan        bool   // If true, the GPU storage may still be read by last frame's draws, so it must be orphaned before the next upload
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
		data:    data,
		vboSize: data.numVerts * data.stride, // Note: stride is already in bytes
		eboSize: soi * len(data.indices),
	}

	mainthread.Call(func() {
		b.vao = gl.GenVertexArrays()
		b.vbo = gl.GenBuffers()
		b.ebo = gl.GenBuffers()

		gl.BindVertexArray(b.vao)

		gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, b.vboSize, nil, gl.DYNAMIC_DRAW)

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, b.eboSize, nil, gl.DYNAMIC_DRAW)

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
		gl.DeleteBuffers(v.ebo)
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
	v.numIndices = 0
	v.numVerts = 0
	// v.materialSet = false

	// If the GPU storage was used, then last frame's draws may still be reading it
	v.needsOrphan = v.needsOrphan || v.uploadedIndices > 0
	v.uploadedVerts = 0
	v.uploadedIndices = 0
	v.drawnVerts = 0
	v.drawnIndices = 0
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

	// v.materialSet = true
	// v.state = state

	// currentElement := v.buffers[0].VertexCount()
	currentElement := v.numVerts
	for i := range indices {
		v.data.indices = append(v.data.indices, currentElement+indices[i])
	}
	v.numVerts += uint32(numVerts)
	v.numIndices = len(v.data.indices)

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

// Uploads the vertices and indices that were reserved since the last upload
// This function returns with the Element array buffer bound, so it is ready to be drawn
func (v *VertexBuffer) mainthreadUpload() {
	if v.needsOrphan {
		v.needsOrphan = false

		// Orphan the GPU storage, so that the driver gives us fresh memory instead of stalling until the GPU finishes reading last frame's data
		gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, v.vboSize, nil, gl.DYNAMIC_DRAW)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, v.eboSize, nil, gl.DYNAMIC_DRAW)
	}

	numVerts := int(v.numVerts)
	if v.uploadedVerts < numVerts {
		// Each attribute has its own region in the vbo, so upload the new tail of each region
		gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)
		regionOffset := 0
		for i := range v.data.buffers {
			subBuffer := v.data.buffers[i]
			tailOffset := v.uploadedVerts * subBuffer.VertexSize()
			gl.BufferSubDataByte(gl.ARRAY_BUFFER, regionOffset+tailOffset, subBuffer.Buffer()[tailOffset:])
			regionOffset += subBuffer.Offset()
		}
		v.uploadedVerts = numVerts
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
	if v.uploadedIndices < v.numIndices {
		gl.BufferSubDataUint32(gl.ELEMENT_ARRAY_BUFFER, v.uploadedIndices*soi, v.data.indices[v.uploadedIndices:])
		v.uploadedIndices = v.numIndices
	}

	if v.deallocAfterBuffer {
		v.deallocAfterBuffer = false
		v.deallocCPUBuffers()
	}
}

// Draws the indices in the range [start, start+count)
func (v *VertexBuffer) mainthreadDraw(start, count int) {
	gl.BindVertexArray(v.vao)
	v.mainthreadUpload()
	gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, start*soi)
}

// Draws the entire vertex buffer
func (v *VertexBuffer) Draw() {
	v.drawRange(0, v.numIndices)
}

// Draws the indices in the range [start, start+count)
func (v *VertexBuffer) drawRange(start, count int) {
	if count <= 0 {
		return
	}

	state.drawVertBuffer(v, start, count)
}

// Streaming: Returns the range of indices (and the number of vertices) that were added since the last call, and marks them as drawn
func (v *VertexBuffer) takePendingRange() (start, count, verts int) {
	start = v.drawnIndices
	count = v.numIndices - v.drawnIndices
	verts = int(v.numVerts) - v.drawnVerts

	v.drawnIndices = v.numIndices
	v.drawnVerts = int(v.numVerts)
	return start, count, verts
}

// BufferPool
// TODO - Idea Improvements: You'd be able to calculate in the pass how many draws with the same material you'd be doing. Based on that you could have really well sized buffers. Also in here you could have different VertexBuffer sizes and order them as needed into a final draw slice
type BufferPool struct {
	shader            *Shader
	triangleBatchSize int
	triangleCount     int
	buffers           []*VertexBuffer
	currentIndex      int // The buffer we are currently streaming into. Buffers before this one are full for this frame
}

func NewBufferPool(shader *Shader, triangleBatchSize int) *BufferPool {
	return &BufferPool{
		shader:            shader,
		triangleBatchSize: triangleBatchSize,
		triangleCount:     0,
		buffers:           make([]*VertexBuffer, 0),
		currentIndex:      0,
	}
}

func (b *BufferPool) Clear() {
	for i := range b.buffers {
		b.buffers[i].Clear()
	}
	b.triangleCount = 0
	b.currentIndex = 0
}

// Returns the vertexbuffer that we reserved to
// Note: Geometry is appended to the current buffer until it is full, so callers draw it in ranges (See: VertexBuffer.takePendingRange)
func (b *BufferPool) Reserve(indices []uint32, numVerts int, dests []interface{}) *VertexBuffer {
	for i := b.currentIndex; i < len(b.buffers); i++ {
		success := b.buffers[i].Reserve(indices, numVerts, dests)
		if success {
			b.triangleCount += len(indices) / 3
			b.currentIndex = i
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
