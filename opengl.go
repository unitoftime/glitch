package glitch

import (
	// "fmt"

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
	Int AttribSize = AttribSize(1)
	Float AttribSize = AttribSize(1)
	Vec2 AttribSize = AttribSize(2)
	Vec3 AttribSize = AttribSize(3)
	Vec4 AttribSize = AttribSize(4)
	Mat2 AttribSize = AttribSize(2 * 2)
	Mat23 AttribSize = AttribSize(2 * 3)
	Mat24 AttribSize = AttribSize(2 * 4)
	Mat3 AttribSize = AttribSize(3 * 3)
	Mat32 AttribSize = AttribSize(3 * 2)
	Mat34 AttribSize = AttribSize(3 * 4)
	Mat4 AttribSize = AttribSize(4 * 4)
	Mat42 AttribSize = AttribSize(4 * 2)
	Mat43 AttribSize = AttribSize(4 * 3)
)
// const sof int = 4 // SizeOf(Float)
// const (
// 	// TODO - others
// 	Int AttribSize = AttribSize(4)
// 	Float AttribSize = AttribSize(sof)
// 	Vec2 AttribSize = AttribSize(2 * sof)
// 	Vec3 AttribSize = AttribSize(3 * sof)
// 	Vec4 AttribSize = AttribSize(4 * sof)
// 	Mat2 AttribSize = AttribSize(2 * 2 * sof)
// 	Mat23 AttribSize = AttribSize(2 * 3 * sof)
// 	Mat24 AttribSize = AttribSize(2 * 4 * sof)
// 	Mat3 AttribSize = AttribSize(3 * 3 * sof)
// 	Mat32 AttribSize = AttribSize(3 * 2 * sof)
// 	Mat34 AttribSize = AttribSize(3 * 4 * sof)
// 	Mat4 AttribSize = AttribSize(4 * 4 * sof)
// 	Mat42 AttribSize = AttribSize(4 * 2 * sof)
// 	Mat43 AttribSize = AttribSize(4 * 3 * sof)
// )

type VertexBuffer struct {
	vao, vbo, ebo gl.Buffer

	format VertexFormat
	stride int

	vertices []float32
	indices []uint32
}

func NewVertexBuffer(numVerts, numTris int, shader *Shader, format VertexFormat) *VertexBuffer {
	b := &VertexBuffer{
		format: format,
		vertices: make([]float32, 8 * sof * numVerts), // 8 * floats (sizeof float) * num triangles
		indices: make([]uint32, 3 * numTris), // 3 indices per triangle
	}

	mainthread.Call(func() {
		b.vao = gl.GenVertexArrays()
		b.vbo = gl.GenBuffers()
		b.ebo = gl.GenBuffers()

		gl.BindVertexArray(b.vao)

		componentSize := 4 // float32
		gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, componentSize * len(b.vertices), b.vertices, gl.DYNAMIC_DRAW)

		indexSize := 4 // uint32 // TODO - make this modifiable?
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexSize * len(b.indices), b.indices, gl.DYNAMIC_DRAW)

		b.stride = 0
		for i := range format {
			b.stride += (int(format[i].Size) * sof)
		}

		// fmt.Println("stride", b.stride)

		offset := 0
		for i := range format {
			loc := shader.getAttribLocation(format[i].Name)
			gl.VertexAttribPointer(loc, int(format[i].Size), gl.FLOAT, false, b.stride, offset)
			gl.EnableVertexAttribArray(loc)
			// fmt.Println("name", format[i].Name)
			// fmt.Println("size", format[i].Size)
			// fmt.Println("offset", offset)
			offset += (int(format[i].Size) * sof)
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
	v.vertices = v.vertices[:0]
	v.indices = v.indices[:0]
}

// // TODO - guarantee correct sizes?
// func (v *VertexBuffer) Add(positions, colors, texCoords []float32, indices []uint32) {
// 	currentElement := uint32(len(v.vertices) / v.stride)

// 	v.vertices = append(v.vertices, vertices...)
// 	for i := range indices {
// 		v.indices = append(v.indices, currentElement + indices[i])
// 	}
// }

// TODO - guarantee correct sizes?
func (v *VertexBuffer) Add(vertices []float32, indices []uint32) {
	currentElement := uint32(len(v.vertices) / v.stride)

	v.vertices = append(v.vertices, vertices...)
	for i := range indices {
		v.indices = append(v.indices, currentElement + indices[i])
	}
}

func (v *VertexBuffer) Draw() {
	mainthread.Call(func() {
		gl.BindVertexArray(v.vao)

		// fmt.Println(v.vertices)
		// fmt.Println(v.indices)
		gl.BindBuffer(gl.ARRAY_BUFFER, v.vbo)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, v.vertices)

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.ebo)
		gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, v.indices)

		gl.DrawElements(gl.TRIANGLES, len(v.indices), gl.UNSIGNED_INT, 0)
	})
}
