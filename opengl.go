package glitch

import (
	"fmt"
	"reflect"
	"unsafe"
	"runtime"

	"github.com/unitoftime/gl"
)

const sof int = 4 // SizeOf(Float)

type VertexBuffer struct {
	vao, vbo, ebo gl.Buffer

	materialSet bool
	material Material
	format VertexFormat
	stride int

	buffers []ISubBuffer
	indices []uint32
	bufferedToGPU bool
}

// TODO - rename
type ISubBuffer interface {
	Clear()
	VertexCount() uint32
	Buffer() []byte
	Offset() int
	Len() int
	Cap() int
	// BufferSubData(int)
}

type SupportedSubBuffers interface {
	glVec4 | glVec3 | glVec2 | float32
}

type SubBuffer[T SupportedSubBuffers] struct {
	attr Attr
	// name string
	// attrSize AttribSize
	maxVerts int
	offset int
	vertexCount int
	buffer []T
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

func (b *SubBuffer[T]) VertexCount() uint32 {
	return uint32(b.vertexCount)
}

func (b *SubBuffer[T]) Buffer() []byte {
	// TODO - maybe I could put this in the struct so its not allocated every time?
	// https://github.com/golang/go/issues/45380
	var tt T
	t := interface{}(tt)

	buff := b.buffer
	switch s := t.(type) {
	case float32:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&buff))
		h.Len *= 1 * 4
		h.Cap *= 1 * 4
		// fmt.Println("float32", h.Len, h.Cap)
		return *(*[]byte)(unsafe.Pointer(h))
	case glVec2:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&buff))
		h.Len *= 2 * 4
		h.Cap *= 2 * 4
		// fmt.Println("Vec2", h.Len, h.Cap)
		return *(*[]byte)(unsafe.Pointer(h))
	case glVec3:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&buff))
		h.Len *= 3 * 4
		h.Cap *= 3 * 4
		// fmt.Println("Vec3", h.Len, h.Cap)
		return *(*[]byte)(unsafe.Pointer(h))
	case glVec4:
		h := (*reflect.SliceHeader)(unsafe.Pointer(&buff))
		h.Len *= 4 * 4
		h.Cap *= 4 * 4
		// fmt.Println("Vec3", h.Len, h.Cap)
		return *(*[]byte)(unsafe.Pointer(h))
	default:
		panic(fmt.Sprintf("Error: %T", s))
	}
	return nil
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

func ReserveSubBuffer[T SupportedSubBuffers](b *SubBuffer[T], count int) []T {
	start := len(b.buffer)
	b.buffer = b.buffer[:len(b.buffer)+count]
	end := len(b.buffer)
	b.vertexCount += count

	return b.buffer[start:end]
}

var numVertexBuffers int
func NewVertexBuffer(shader *Shader, numVerts, numTris int) *VertexBuffer {
	numVertexBuffers++
	// fmt.Println("NewVertexBuffer:", shader, numVerts, numTris)
	format := shader.attrFmt // TODO - cleanup this variable
	b := &VertexBuffer{
		format: format,
		buffers: make([]ISubBuffer, len(format)),
		indices: make([]uint32, 3 * numTris), // 3 indices per triangle
	}

	b.stride = 0
	offset := 0
	for i := range format {
		b.stride += (int(format[i].Size()) * sof)

		if format[i].Type == AttrVec4 {
			b.buffers[i] = &SubBuffer[glVec4]{
				attr: format[i].Attr,
				// name: format[i].Name,
				// attrSize: format[i].Size,
				maxVerts: numVerts,
				vertexCount: 0,
				offset: offset,
				buffer: make([]glVec4, numVerts),
			}
		} else if format[i].Type == AttrVec3 {
			b.buffers[i] = &SubBuffer[glVec3]{
				attr: format[i].Attr,
				// name: format[i].Name,
				// attrSize: format[i].Size,
				maxVerts: numVerts,
				vertexCount: 0,
				offset: offset,
				buffer: make([]glVec3, numVerts),
			}
		} else if format[i].Type == AttrVec2 {
			b.buffers[i] = &SubBuffer[glVec2]{
				attr: format[i].Attr,
				// name: format[i].Name,
				// attrSize: format[i].Size,
				maxVerts: numVerts,
				vertexCount: 0,
				offset: offset,
				buffer: make([]glVec2, numVerts),
			}
		} else if format[i].Type == AttrFloat {
			b.buffers[i] = &SubBuffer[float32]{
				attr: format[i].Attr,
				// name: format[i].Name,
				// attrSize: format[i].Size,
				maxVerts: numVerts,
				vertexCount: 0,
				offset: offset,
				buffer: make([]float32, numVerts),
			}
		} else {
			panic(fmt.Sprintf("Unknown format: %v", format[i]))
		}

		offset += sof * int(format[i].Size()) * numVerts
	}

	// vertices := make([]float32, 8 * sof * numVerts)
	fakeVertices := make([]float32, 4 * numVerts * b.stride) // 4 = sof

	mainthreadCall(func() {
		b.vao = gl.GenVertexArrays()
		b.vbo = gl.GenBuffers()
		b.ebo = gl.GenBuffers()

		gl.BindVertexArray(b.vao)

		componentSize := 4 // float32
		gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, componentSize * numVerts * b.stride, fakeVertices, gl.DYNAMIC_DRAW)

		indexSize := 4 // uint32 // TODO - make this modifiable?
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexSize * len(b.indices), b.indices, gl.DYNAMIC_DRAW)

		for i := range b.buffers {
			switch subBuffer := b.buffers[i].(type) {
			case *SubBuffer[float32]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.attr.Name)
				size := int(subBuffer.attr.Size())
				gl.VertexAttribPointer(loc, size, gl.FLOAT, false, size * sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			case *SubBuffer[glVec2]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.attr.Name)
				size := int(subBuffer.attr.Size())
				gl.VertexAttribPointer(loc, size, gl.FLOAT, false, size * sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			case *SubBuffer[glVec3]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.attr.Name)
				size := int(subBuffer.attr.Size())
				gl.VertexAttribPointer(loc, size, gl.FLOAT, false, size * sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			case *SubBuffer[glVec4]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.attr.Name)
				size := int(subBuffer.attr.Size())
				gl.VertexAttribPointer(loc, size, gl.FLOAT, false, size * sof, subBuffer.offset)
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

func (v *VertexBuffer) delete() {
	mainthreadCallNonBlock(func() {
		gl.DeleteVertexArrays(v.vao)
		gl.DeleteBuffers(v.vbo)
	})
}

func (v *VertexBuffer) Bind() {
	mainthreadCall(func() {
		gl.BindVertexArray(v.vao)
	})
}

func (v *VertexBuffer) Clear() {
	// v.vertices = v.vertices[:0]
	for i := range v.buffers {
		v.buffers[i].Clear()
	}
	v.indices = v.indices[:0]
	v.material = nil
	v.materialSet = false
	v.bufferedToGPU = false
}

func (v *VertexBuffer) Reserve(material Material, indices []uint32, numVerts int, dests []interface{}) bool {
	// If material is set and it doesn't match the reserved material
	if v.materialSet && v.material != material {
		// fmt.Println("VertexBuffer.Reserve - Material Doesn't match")
		return false
	}
	if len(v.indices) + len(indices) > cap(v.indices) {
		// fmt.Println("VertexBuffer.Reserve - Not enough index capacity")
		return false
	}
	if v.buffers[0].Len() + numVerts > v.buffers[0].Cap() {
		// fmt.Println("VertexBuffer.Reserve - Not enough vertex capacity")
		return false
	}

	v.bufferedToGPU = false

	// fmt.Println("Establishing Buffer", material)
	v.materialSet = true
	v.material = material

	currentElement := v.buffers[0].VertexCount()
	for i := range indices {
		v.indices = append(v.indices, currentElement + indices[i])
	}

	for i := range v.buffers {
		switch subBuffer := v.buffers[i].(type) {
		case *SubBuffer[float32]:
			d := dests[i].(*[]float32)
			*d = ReserveSubBuffer[float32](subBuffer, numVerts)
		case *SubBuffer[glVec2]:
			d := dests[i].(*[]glVec2)
			*d = ReserveSubBuffer[glVec2](subBuffer, numVerts)
		case *SubBuffer[glVec3]:
			d := dests[i].(*[]glVec3)
			*d = ReserveSubBuffer[glVec3](subBuffer, numVerts)
		case *SubBuffer[glVec4]:
			d := dests[i].(*[]glVec4)
			*d = ReserveSubBuffer[glVec4](subBuffer, numVerts)
		default:
			panic("Unknown!")
		}
	}
	return true
}

func (v *VertexBuffer) Draw() {
	if len(v.indices) <= 0 {
		return
	}

	mainthreadCall(func() {
		gl.BindVertexArray(v.vao)

		if !v.bufferedToGPU {
			gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)
			offset := 0
			for i := range v.buffers {
				gl.BufferSubData(gl.ARRAY_BUFFER, offset, v.buffers[i].Buffer())
				// byteBuff := v.buffers[i].Buffer()
				// switch t := byteBuff.(type) {
				// case []Vec2:
				// 	gl.BufferSubData(gl.ARRAY_BUFFER, offset, [][2]float32(t))
				// case []Vec3:
				// 	gl.BufferSubData(gl.ARRAY_BUFFER, offset, [][3]float32(t))
				// }
				offset += v.buffers[i].Offset()
			}

			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
			gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, v.indices)

			v.bufferedToGPU = true
			gl.DrawElements(gl.TRIANGLES, len(v.indices), gl.UNSIGNED_INT, 0)
		} else {
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
			gl.DrawElements(gl.TRIANGLES, len(v.indices), gl.UNSIGNED_INT, 0)
		}
	})
}

// BufferPool
// TODO - Idea Improvements: You'd be able to calculate in the pass how many draws with the same material you'd be doing. Based on that you could have really well sized buffers. Also in here you could have different VertexBuffer sizes and order them as needed into a final draw slice
type BufferPool struct {
	shader *Shader
	triangleBatchSize int
	triangleCount int
	buffers []*VertexBuffer
	currentIndex int
	nextClean int // Tracks the next clean vertex buffer (ie clean = buffers that haven't been written reserved on in this case
}
func NewBufferPool(shader *Shader, triangleBatchSize int) *BufferPool {
	return &BufferPool{
		shader: shader,
		triangleBatchSize: triangleBatchSize,
		triangleCount: 0,
		buffers: make([]*VertexBuffer, 0),
		currentIndex: 0,
		nextClean: 0,
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
func (b *BufferPool) Reserve(material Material, indices []uint32, numVerts int, dests []interface{}) *VertexBuffer {
	for i := b.currentIndex; i < len(b.buffers); i++ {
		success := b.buffers[i].Reserve(material, indices, numVerts, dests)
		if success {
			b.triangleCount += len(indices) / 3
			b.currentIndex = i
			b.nextClean = b.currentIndex + 1
			return b.buffers[i]
		}
	}

	// fmt.Printf("NEW BATCH: %d - index: %d\n", b.triangleCount, b.currentIndex)
	newBuff := NewVertexBuffer(b.shader, b.triangleBatchSize, b.triangleBatchSize)
	success := newBuff.Reserve(material, indices, numVerts, dests)
	if !success {
		panic("SOMETHING WENT WRONG")
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

func openglDraw(buffers []*VertexBuffer) {
	lastMaterial := Material(nil)
	for i := range buffers {
		// fmt.Println(i, len(b.buffers[i].indices), b.buffers[i].buffers[0].Len(), b.buffers[i].buffers[0].Cap())
		if lastMaterial != buffers[i].material {
			lastMaterial = buffers[i].material
			if lastMaterial != nil {
				// fmt.Println("Binding New Material", lastMaterial)
				lastMaterial.Bind()
			}
		}
		buffers[i].Draw()
	}
}
