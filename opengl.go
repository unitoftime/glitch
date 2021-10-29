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
}

type SubBuffer[T any] struct {
	name string
	attrSize AttribSize
	maxVerts int
	offset int
	vertexCount int
	buffer []T
}

func (b *SubBuffer[T]) Clear() {
	b.buffer = b.buffer[:0]
	b.vertexCount = 0
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

// Returns the last section appended
func (b *SubBuffer[T]) AppendAndReturnDest(vals []T) []T {
	start := len(b.buffer)
	b.buffer = append(b.buffer, vals...)
	end := len(b.buffer)
	b.vertexCount += len(vals)

	// fmt.Println(start, end, len(vals))
	return b.buffer[start:end]
}

func (b *SubBuffer[T]) Append(val T) {
	b.buffer = append(b.buffer, val)
	b.vertexCount += 1
}


// Returns a pre-sliced underlying buffer based on the reserved amount
// func (b *SubBuffer) Reserve(vertexCount int) interface{} {
// 	b.vertexCount += vertexCount

// 	b.buffer
// 	// Return buffer
// }

// // Returns the entire buffer based on the current vertexCount
// func (b *SubBuffer) Get() interface{} {
// 	// slice and return the buffer
// }

func NewVertexBuffer(shader *Shader, numVerts, numTris int) *VertexBuffer {
	format := shader.attrFmt // TODO - cleanup this variable
	b := &VertexBuffer{
		format: format,
		// vertices: make([]float32, 8 * sof * numVerts), // 8 * floats (sizeof float) * num triangles
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
//		gl.BufferData(gl.ARRAY_BUFFER, componentSize * numVerts * b.stride, nil, gl.DYNAMIC_DRAW)
		gl.BufferData(gl.ARRAY_BUFFER, componentSize * numVerts * b.stride, fakeVertices, gl.DYNAMIC_DRAW)

		indexSize := 4 // uint32 // TODO - make this modifiable?
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexSize * len(b.indices), b.indices, gl.DYNAMIC_DRAW)

		// fmt.Println("stride", b.stride)

		// offset := 0
		// for i := range format {
		// 	loc := shader.getAttribLocation(format[i].Name)
		// 	// gl.VertexAttribPointer(loc, int(format[i].Size), gl.FLOAT, false, b.stride, offset)
		// 	gl.VertexAttribPointer(loc, int(format[i].Size), gl.FLOAT, false, int(format[i].Size) * sof, offset)
		// 	gl.EnableVertexAttribArray(loc)
		// 	// fmt.Println("name", format[i].Name)
		// 	// fmt.Println("size", format[i].Size)
		// 	// fmt.Println("offset", offset)
		// 	// offset += (int(format[i].Size) * sof)

		// 	offset += sof * int(b.buffers[i].attrSize) * b.buffers[i].maxVerts
		// }
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

// TODO - guarantee correct sizes?
func (v *VertexBuffer) Add(positions []Vec3, colors []Vec3, texCoords []Vec2, indices []uint32, matrix *Mat4) bool {
	// TODO - only checking indices
	if len(v.indices) + len(indices) > cap(v.indices) {
		return false
	}

	// currentElement := v.buffers[0].NumVerts()
	// currentElement := uint32(v.buffers[0].vertexCount)
	currentElement := v.buffers[0].VertexCount()

	{
		// for i := range positions {
		// 	// vec := matrix.MulVec3(&positions[i])
		// 	vec := MatMul(matrix, positions[i])
		// 	v.buffers[0].buffer = append(v.buffers[0].buffer, vec[0])
		// 	v.buffers[0].buffer = append(v.buffers[0].buffer, vec[1])
		// 	v.buffers[0].buffer = append(v.buffers[0].buffer, vec[2])
		// }

		// TODO - speed comparison?
		// Alter then write
		// posBuffer := v.buffers[0].(*SubBuffer[Vec3])
		// for i := range positions {
		// 	vec := MatMul(matrix, positions[i])
		// 	posBuffer.Append(vec)
		// }

		// Write then alter
		posBuffer := v.buffers[0].(*SubBuffer[Vec3])
		dest := posBuffer.AppendAndReturnDest(positions)
		for i := range dest {
			dest[i] = MatMul(matrix, dest[i])
		}
		// fmt.Println(dest)
	}

	// v.buffers[1].buffer = append(v.buffers[1].buffer, colors...)
	// v.buffers[2].buffer = append(v.buffers[2].buffer, texCoords...)

	colBuffer := v.buffers[1].(*SubBuffer[Vec3])
	colBuffer.AppendAndReturnDest(colors)
	texBuffer := v.buffers[2].(*SubBuffer[Vec2])
	texBuffer.AppendAndReturnDest(texCoords)

	for i := range indices {
		v.indices = append(v.indices, currentElement + indices[i])
	}

	// TODO - Note: Each vec3 element in positions slice represents a vert
	// v.buffers[0].vertexCount += len(positions)

	return true
}

// // TODO - guarantee correct sizes?
// func (v *VertexBuffer) Add(vertices []float32, indices []uint32) {
// 	currentElement := uint32(len(v.vertices) / v.stride)

// 	v.vertices = append(v.vertices, vertices...)
// 	for i := range indices {
// 		v.indices = append(v.indices, currentElement + indices[i])
// 	}
// }

func (v *VertexBuffer) Draw() {
	mainthread.Call(func() {
		gl.BindVertexArray(v.vao)

		gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)
		offset := 0
		for i := range v.buffers {
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, v.buffers[i].Buffer())
			// offset += sof * int(v.buffers[i].attrSize) * v.buffers[i].maxVerts
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

func (b *BufferPool) Add(positions []Vec3, colors []Vec3, texCoords []Vec2, indices []uint32, matrix *Mat4, mask RGBA) {
	success := false
	for i := range b.buffers {
		success = b.buffers[i].Add(positions, colors, texCoords, indices, matrix)
		if success {
			break
		}
	}
	if !success {
		fmt.Printf("NEW BATCH: %d\n", b.triangleCount)
		newBuff := NewVertexBuffer(b.shader, b.triangleBatchSize, b.triangleBatchSize)
		success := newBuff.Add(positions, colors, texCoords, indices, matrix)
		if !success {
			panic("SOMETHING WENT WRONG")
		}
		b.buffers = append(b.buffers, newBuff)
	}

	b.triangleCount += len(indices) / 3
}

// Adds indices and returns sliced subbuffers
// func (b *BufferPool) Reserve(indices []uint32, vertexCount int) []SubBuffer {
// 	success := false
// 	for _, v := range b.buffers {
// 		if len(v.indices) + len(indices) <= cap(v.indices) {
// 			for i := range indices {
// 				currentElement := uint32(v.buffers[0].vertexCount) // TODO - 0 okay here? I guess it doesn't matter which vertexCount we retrieve?
// 				v.indices = append(v.indices, currentElement + indices[i])
// 			}
// 			return b.buffers[i]
// 			success = true
// 		}
// 		if success {
// 			break
// 		}
// 	}

// 	if !success {
// 		fmt.Printf("RESERVED NEW BATCH: %d\n", b.triangleCount)
// 		newBuff := NewVertexBuffer(b.shader, b.triangleBatchSize, b.triangleBatchSize)
// 		b.buffers = append(b.buffers, newBuff)
// 	}

// 	b.triangleCount += len(indices) / 3

// 	return nil
// }

func (b *BufferPool) Draw() {
	for i := range b.buffers {
		b.buffers[i].Draw()
	}
}
