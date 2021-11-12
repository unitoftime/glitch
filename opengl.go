package glitch

import (
	"fmt"

	"github.com/faiface/mainthread"
	"github.com/jstewart7/gl"
)

// TODO - right now we only support floats (for simplicity)
type VertexFormat []Attrib
type UniformFormat []Attrib

type Attrib struct {
	Name string
	Size AttribSize
}
type AttribSize int

const sof int = 4 // SizeOf(Float)
const (
	// TODO - others
	AttrInt AttribSize = AttribSize(1)
	AttrFloat AttribSize = AttribSize(1)
	AttrVec2 AttribSize = AttribSize(2)
	AttrVec3 AttribSize = AttribSize(3)
	AttrVec4 AttribSize = AttribSize(4)
	AttrMat2 AttribSize = AttribSize(2 * 2)
	AttrMat23 AttribSize = AttribSize(2 * 3)
	AttrMat24 AttribSize = AttribSize(2 * 4)
	AttrMat3 AttribSize = AttribSize(3 * 3)
	AttrMat32 AttribSize = AttribSize(3 * 2)
	AttrMat34 AttribSize = AttribSize(3 * 4)
	AttrMat4 AttribSize = AttribSize(4 * 4)
	AttrMat42 AttribSize = AttribSize(4 * 2)
	AttrMat43 AttribSize = AttribSize(4 * 3)
)

type VertexBuffer struct {
	vao, vbo, ebo gl.Buffer

	format VertexFormat
	stride int

	buffers []ISubBuffer
	indices []uint32
}

// TODO - rename
type ISubBuffer interface {
	Clear()
	VertexCount() uint32
	Buffer() interface{}
	Offset() int
	Len() int
	Cap() int
}

type SupportedSubBuffers interface {
	Vec3 | Vec2 | float32
}

type SubBuffer[T SupportedSubBuffers] struct {
	name string
	attrSize AttribSize
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

func (b *SubBuffer[T]) Buffer() interface{} {
	return b.buffer
}

func (b *SubBuffer[T]) Offset() int {
	return sof * int(b.attrSize) * b.maxVerts
}

func ReserveSubBuffer[T SupportedSubBuffers](b *SubBuffer[T], count int) []T {
	start := len(b.buffer)
	b.buffer = b.buffer[:len(b.buffer)+count]
	end := len(b.buffer)
	b.vertexCount += count

	return b.buffer[start:end]
}

func NewVertexBuffer(shader *Shader, numVerts, numTris int) *VertexBuffer {
	format := shader.attrFmt // TODO - cleanup this variable
	b := &VertexBuffer{
		format: format,
		buffers: make([]ISubBuffer, len(format)),
		indices: make([]uint32, 3 * numTris), // 3 indices per triangle
	}

	b.stride = 0
	offset := 0
	for i := range format {
		b.stride += (int(format[i].Size) * sof)

		if format[i].Size == AttrVec3 {
			b.buffers[i] = &SubBuffer[Vec3]{
				name: format[i].Name,
				attrSize: format[i].Size,
				maxVerts: numVerts,
				vertexCount: 0,
				offset: offset,
				buffer: make([]Vec3, numVerts),
			}
		} else if format[i].Size == AttrVec2 {
			b.buffers[i] = &SubBuffer[Vec2]{
				name: format[i].Name,
				attrSize: format[i].Size,
				maxVerts: numVerts,
				vertexCount: 0,
				offset: offset,
				buffer: make([]Vec2, numVerts),
			}
		} else if format[i].Size == AttrFloat {
			b.buffers[i] = &SubBuffer[float32]{
				name: format[i].Name,
				attrSize: format[i].Size,
				maxVerts: numVerts,
				vertexCount: 0,
				offset: offset,
				buffer: make([]float32, numVerts),
			}
		} else {
			panic(fmt.Sprintf("Unknown format: %v", format[i]))
		}

		offset += sof * int(format[i].Size) * numVerts
	}

	// vertices := make([]float32, 8 * sof * numVerts)
	fakeVertices := make([]float32, 4 * numVerts * b.stride) // 4 = sof

	mainthread.Call(func() {
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
				loc := gl.GetAttribLocation(shader.program, subBuffer.name)
				gl.VertexAttribPointer(loc, int(subBuffer.attrSize), gl.FLOAT, false, int(subBuffer.attrSize) * sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			case *SubBuffer[Vec2]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.name)
				gl.VertexAttribPointer(loc, int(subBuffer.attrSize), gl.FLOAT, false, int(subBuffer.attrSize) * sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			case *SubBuffer[Vec3]:
				loc := gl.GetAttribLocation(shader.program, subBuffer.name)
				gl.VertexAttribPointer(loc, int(subBuffer.attrSize), gl.FLOAT, false, int(subBuffer.attrSize) * sof, subBuffer.offset)
				gl.EnableVertexAttribArray(loc)
			default:
				panic("Unknown!")
			}
		}
	})

	b.Clear() // TODO - fix

	return b
}

func (v *VertexBuffer) Bind() {
	mainthread.Call(func() {
		gl.BindVertexArray(v.vao)
	})
}

func (v *VertexBuffer) Clear() {
	// v.vertices = v.vertices[:0]
	for i := range v.buffers {
		v.buffers[i].Clear()
	}
	v.indices = v.indices[:0]
}

func (v *VertexBuffer) Reserve(indices []uint32, numVerts int, dests []interface{}) bool {
	if len(v.indices) + len(indices) > cap(v.indices) {
		return false
	}
	if v.buffers[0].Len() + numVerts > v.buffers[0].Cap() {
		return false
	}

	currentElement := v.buffers[0].VertexCount()
	for i := range indices {
		v.indices = append(v.indices, currentElement + indices[i])
	}

	for i := range v.buffers {
		switch subBuffer := v.buffers[i].(type) {
		case *SubBuffer[float32]:
			d := dests[i].(*[]float32)
			*d = ReserveSubBuffer[float32](subBuffer, numVerts)
		case *SubBuffer[Vec2]:
			d := dests[i].(*[]Vec2)
			*d = ReserveSubBuffer[Vec2](subBuffer, numVerts)
		case *SubBuffer[Vec3]:
			d := dests[i].(*[]Vec3)
			*d = ReserveSubBuffer[Vec3](subBuffer, numVerts)
		default:
			panic("Unknown!")
		}
	}
	return true
}

func (v *VertexBuffer) Draw() {
	mainthread.Call(func() {
		gl.BindVertexArray(v.vao)

		gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)
		offset := 0
		for i := range v.buffers {
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, v.buffers[i].Buffer())
			offset += v.buffers[i].Offset()
		}

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
		gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, v.indices)

		gl.DrawElements(gl.TRIANGLES, len(v.indices), gl.UNSIGNED_INT, 0)
	})
}

// BufferPool
type BufferPool struct {
	shader *Shader
	triangleBatchSize int
	triangleCount int
	buffers []*VertexBuffer
}
func NewBufferPool(shader *Shader, triangleBatchSize int) *BufferPool {
	return &BufferPool{
		shader: shader,
		triangleBatchSize: triangleBatchSize,
		triangleCount: 0,
		buffers: make([]*VertexBuffer, 0),
	}
}

func (b *BufferPool) Clear() {
	for i := range b.buffers {
		b.buffers[i].Clear()
	}
	b.triangleCount = 0
}

func (b *BufferPool) Reserve(indices []uint32, numVerts int, dests []interface{}) bool {
	for i := range b.buffers {
		success := b.buffers[i].Reserve(indices, numVerts, dests)
		if success {
			b.triangleCount += len(indices) / 3
			return true
		}
	}

	fmt.Printf("NEW BATCH: %d\n", b.triangleCount)
	newBuff := NewVertexBuffer(b.shader, b.triangleBatchSize, b.triangleBatchSize)
	success := newBuff.Reserve(indices, numVerts, dests)
	if !success {
		panic("SOMETHING WENT WRONG")
	}
	b.buffers = append(b.buffers, newBuff)
	b.triangleCount += len(indices) / 3
	return success
}

func (b *BufferPool) Draw() {
	for i := range b.buffers {
		b.buffers[i].Draw()
	}
}
