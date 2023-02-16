// +build !js

package gl

// Enum is equivalent to GLenum, and is normally used with one of the
// constants defined in this package.
type Enum uint32

// Attrib identifies the location of a specific attribute variable.
type Attrib struct {
	Value int
}

// Program identifies a compiled shader program.
type Program struct {
	Value uint32
}

// Shader identifies a GLSL shader.
type Shader struct {
	Value uint32
}

// Buffer identifies a GL buffer object.
type Buffer struct {
	Value uint32
}

// Framebuffer identifies a GL framebuffer.
type Framebuffer struct {
	Value uint32
}
func (o Framebuffer) Equal(o2 Framebuffer) bool {
	return o.Value == o2.Value
}


// A Renderbuffer is a GL object that holds an image in an internal format.
type Renderbuffer struct {
	Value uint32
}

// A Texture identifies a GL texture unit.
type Texture struct {
	Value uint32
}

// Uniform identifies the location of a specific uniform variable.
type Uniform struct {
	Value int32
}

type Object struct {
	Value uint32
}

func (o Object) Equal(o2 Object) bool {
	return o.Value == o2.Value
}



var NoObject = Object{0}

var NoAttrib = Attrib{0}
var NoProgram = Program{0}
var NoShader = Shader{0}
var NoBuffer = Buffer{0}
var NoFramebuffer = Framebuffer{0}
var NoRenderbuffer = Renderbuffer{0}
var NoTexture = Texture{0}
var NoUniform = Uniform{0}
