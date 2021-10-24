package glitch

import (
	"fmt"

	"github.com/ungerik/go3d/vec3"

	"github.com/faiface/mainthread"
	"github.com/jstewart7/gl"
)

// TODO - right now we only support floats (for simplicity)
type VertexFormat []Attrib
type AttributeFormat []Attrib

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

	buffers []SubBuffer
	indices []uint32
}

type SubBuffer struct {
	name string
	attrSize AttribSize
	maxVerts int
	offset int
	vertexCount int
	buffer []float32
}

// func (b *SubBuffer) NumVerts() uint32 {
// 	return uint32(len(b.buffer) / int(b.attrSize))
// }

func NewVertexBuffer(shader *Shader, numVerts, numTris int) *VertexBuffer {
	format := shader.attrFmt // TODO - cleanup this variable
	b := &VertexBuffer{
		format: format,
		// vertices: make([]float32, 8 * sof * numVerts), // 8 * floats (sizeof float) * num triangles
		buffers: make([]SubBuffer, len(format)),
		indices: make([]uint32, 3 * numTris), // 3 indices per triangle
	}

	b.stride = 0
	offset := 0
	for i := range format {
		b.stride += (int(format[i].Size) * sof)

		b.buffers[i] = SubBuffer{
			name: format[i].Name,
			attrSize: format[i].Size,
			maxVerts: numVerts,
			vertexCount: 0,
			offset: offset,
			buffer: make([]float32, int(format[i].Size) * numVerts),
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
			loc := shader.getAttribLocation(b.buffers[i].name)
			gl.VertexAttribPointer(loc, int(b.buffers[i].attrSize), gl.FLOAT, false, int(b.buffers[i].attrSize) * sof, b.buffers[i].offset)
			gl.EnableVertexAttribArray(loc)
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
		v.buffers[i].buffer = v.buffers[i].buffer[:0]
		v.buffers[i].vertexCount = 0
	}
	v.indices = v.indices[:0]
}

// TODO - guarantee correct sizes?
func (v *VertexBuffer) Add(positions []vec3.T, colors, texCoords []float32, indices []uint32, matrix *Mat4) bool {
	// TODO - only checking indices
	if len(v.indices) + len(indices) > cap(v.indices) {
		return false
	}

	// currentElement := v.buffers[0].NumVerts()
	currentElement := uint32(v.buffers[0].vertexCount)

	{
		for i := range positions {
			// vec := matrix.MulVec3(&positions[i])
			vec := MatMul(matrix, positions[i])
			v.buffers[0].buffer = append(v.buffers[0].buffer, vec[0])
			v.buffers[0].buffer = append(v.buffers[0].buffer, vec[1])
			v.buffers[0].buffer = append(v.buffers[0].buffer, vec[2])
		}
	}
	// v.buffers[0].buffer = append(v.buffers[0].buffer, positions...)
	v.buffers[1].buffer = append(v.buffers[1].buffer, colors...)
	v.buffers[2].buffer = append(v.buffers[2].buffer, texCoords...)

	for i := range indices {
		v.indices = append(v.indices, currentElement + indices[i])
	}

	// TODO - Note: Each vec3 element in positions slice represents a vert
	v.buffers[0].vertexCount += len(positions)

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
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, v.buffers[i].buffer)
			offset += sof * int(v.buffers[i].attrSize) * v.buffers[i].maxVerts
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

func (b *BufferPool) Add(positions []vec3.T, colors, texCoords []float32, indices []uint32, matrix *Mat4) {
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

func (b *BufferPool) Draw() {
	for i := range b.buffers {
		b.buffers[i].Draw()
	}
}
