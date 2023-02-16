package glitch

import (
	"fmt"

	// "github.com/go-gl/mathgl/mgl32"

	"github.com/unitoftime/glitch/internal/gl"
)

type ShaderConfig struct {
	VertexShader, FragmentShader string
	VertexFormat VertexFormat
	UniformFormat UniformFormat
}

type Shader struct {
	program gl.Program
	uniforms map[string]Uniform
	attrFmt VertexFormat
	tmpBuffers []any
	tmpFloat32Slice []float32
	mainthreadBind func()
}

type Uniform struct {
	name string
	// attrType AttrType
	loc gl.Uniform
}

func NewShader(cfg ShaderConfig) (*Shader, error) {
	return NewShaderExt(cfg.VertexShader, cfg.FragmentShader, cfg.VertexFormat, cfg.UniformFormat)
}

func NewShaderExt(vertexSource, fragmentSource string, attrFmt VertexFormat, uniformFmt UniformFormat) (*Shader, error) {
	shader := &Shader{
		uniforms: make(map[string]Uniform),
		attrFmt: attrFmt,
		tmpFloat32Slice: make([]float32, 0),
	}
	err := mainthreadCallErr(func() error {
		var err error
		shader.program, err = createProgram(vertexSource, fragmentSource)
		if err != nil {
			return err
		}

		for _, uniform := range uniformFmt {
			loc := gl.GetUniformLocation(shader.program, uniform.Name)
			shader.uniforms[uniform.Name] = Uniform{uniform.Name, loc}
			// fmt.Println("Found uniform: ", uniform)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	shader.mainthreadBind = func() {
		gl.UseProgram(shader.program)
	}

	// Loop through and set all matrices to identity matrices
	shader.Bind()
	for _, uniform := range uniformFmt {
		// TODO handle other matrices
		if uniform.Type == AttrMat4 {
			// Setting uniform
			shader.SetUniform(uniform.Name, Mat4Ident)
		}
	}

	shader.tmpBuffers = make([]any, len(shader.attrFmt))
	for i, attr := range shader.attrFmt {
		shader.tmpBuffers[i] = attr.GetBuffer()
	}

	return shader, nil
}

func (s *Shader) Bind() {
	mainthreadCall(s.mainthreadBind)
}
// func (s *Shader) Bind() {
// 	mainthreadCall(func() {
// 		gl.UseProgram(s.program)
// 	})
// }

func createProgram(vertexSrc, fragmentSrc string) (gl.Program, error) {
	program := gl.CreateProgram()
	if !program.Valid() {
		return gl.Program{}, fmt.Errorf("Could not CreateProgram")
	}

	vertexShader, err := loadShader(gl.VERTEX_SHADER, vertexSrc)
	if err != nil {
		return gl.Program{}, err
	}
	fragmentShader, err := loadShader(gl.FRAGMENT_SHADER, fragmentSrc)
	if err != nil {
		gl.DeleteShader(vertexShader)
		return gl.Program{}, err
	}

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Flag shaders for deletion when program is unlinked.
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	if gl.GetProgrami(program, gl.LINK_STATUS) == 0 {
		defer gl.DeleteProgram(program)
		return gl.Program{}, fmt.Errorf("CreateProgram: %s", gl.GetProgramInfoLog(program))
	}
	return program, nil
}

func loadShader(shaderType gl.Enum, src string) (gl.Shader, error) {
	shader := gl.CreateShader(shaderType)
	if !shader.Valid() {
		return gl.Shader{}, fmt.Errorf("loadShader could not create shader (type %v)", shaderType)
	}
	gl.ShaderSource(shader, src)
	gl.CompileShader(shader)
	if gl.GetShaderi(shader, gl.COMPILE_STATUS) == 0 {
		defer gl.DeleteShader(shader)
		return gl.Shader{}, fmt.Errorf("loadShader: %s", gl.GetShaderInfoLog(shader))
	}
	return shader, nil
}

func (s *Shader) SetUniform(uniformName string, value interface{}) bool {
	ret := false
	mainthreadCall(func() {
		uniform, ok := s.uniforms[uniformName]
		// TODO - detecting if uniform is invalid, because Valid() checks if it is 0, which is a valid location index
		if !ok /* || !uniform.loc.Valid() */ {
			// TODO - panic or just return false? I feel like its bad if you think you're setting a uniform that doesn't exist.
			panic(fmt.Sprintf("Uniform not found! Or uniform location was invalid: %s", uniformName))
			ret = false
		}

		switch val := value.(type) {
		// case float32:
		// 	sliced := []float32{val}
		// 	gl.Uniform1fv(uniform.loc, sliced)
		// case Vec3:
		// 	vec := val.gl()
		// 	gl.Uniform3fv(uniform.loc, vec[:])
		// case Vec4:
		// 	vec := val.gl()
		// 	gl.Uniform4fv(uniform.loc, vec[:])
		case Mat4:
			s.tmpFloat32Slice = s.tmpFloat32Slice[:0]
			s.tmpFloat32Slice = val.writeToFloat32(s.tmpFloat32Slice)
			gl.UniformMatrix4fv(uniform.loc, s.tmpFloat32Slice)
			// mat := val.gl()
			// gl.UniformMatrix4fv(uniform.loc, mat[:])
		case *Mat4:
			s.tmpFloat32Slice = s.tmpFloat32Slice[:0]
			s.tmpFloat32Slice = val.writeToFloat32(s.tmpFloat32Slice)
			gl.UniformMatrix4fv(uniform.loc, s.tmpFloat32Slice)
			// mat := val.gl()
			// gl.UniformMatrix4fv(uniform.loc, mat[:])
		default:
			// fmt.Println("ERROR", uniform)
			panic(fmt.Sprintf("set uniform attr: invalid attribute type: %T", value))
		}
		ret = true
	})
	return ret
}

// Note: This was me playing around with a way to reduce the amount of memory allocations
// var tmpUniformSetter uniformSetter
// func init() {
// 	tmpUniformSetter.FUNC = func() {
// 		tmpUniformSetter.Func()
// 	}
// }

// func (s *Shader) SetUniform(uniformName string, value Mat4) bool {
// // func (s *Shader) SetUniform(uniformName string, value interface{}) bool {
// 	tmpUniformSetter.shader = s
// 	tmpUniformSetter.name = uniformName
// 	tmpUniformSetter.value = value

// 	mainthreadCall(tmpUniformSetter.FUNC)
// 	return true // TODO - wrong
// }
// type uniformSetter struct {
// 	shader *Shader
// 	name string
// 	value Mat4
// 	FUNC func()
// }

// func (u *uniformSetter) Func() {
// 	s := u.shader
// 	uniformName := u.name
// 	value := u.value

// 	uniform, ok := s.uniforms[uniformName]
// 	// TODO - detecting if uniform is invalid, because Valid() checks if it is 0, which is a valid location index
// 	if !ok /* || !uniform.loc.Valid() */ {
// 		// TODO - panic or just return false? I feel like its bad if you think you're setting a uniform that doesn't exist.
// 		panic(fmt.Sprintf("Uniform not found! Or uniform location was invalid: %s", uniformName))
// 	}

// 	s.tmpFloat32Slice = s.tmpFloat32Slice[:0]
// 	s.tmpFloat32Slice = value.writeToFloat32(s.tmpFloat32Slice)
// 	gl.UniformMatrix4fv(uniform.loc, s.tmpFloat32Slice)

// 	// switch val := value.(type) {
// 	// 	// case float32:
// 	// 	// 	sliced := []float32{val}
// 	// 	// 	gl.Uniform1fv(uniform.loc, sliced)
// 	// 	// case Vec3:
// 	// 	// 	vec := val.gl()
// 	// 	// 	gl.Uniform3fv(uniform.loc, vec[:])
// 	// 	// case Vec4:
// 	// 	// 	vec := val.gl()
// 	// 	// 	gl.Uniform4fv(uniform.loc, vec[:])
// 	// case Mat4:
// 	// 	s.tmpFloat32Slice = s.tmpFloat32Slice[:0]
// 	// 	s.tmpFloat32Slice = val.writeToFloat32(s.tmpFloat32Slice)
// 	// 	gl.UniformMatrix4fv(uniform.loc, s.tmpFloat32Slice)
// 	// 	// mat := val.gl()
// 	// 	// gl.UniformMatrix4fv(uniform.loc, mat[:])
// 	// case *Mat4:
// 	// 	s.tmpFloat32Slice = s.tmpFloat32Slice[:0]
// 	// 	s.tmpFloat32Slice = val.writeToFloat32(s.tmpFloat32Slice)
// 	// 	gl.UniformMatrix4fv(uniform.loc, s.tmpFloat32Slice)
// 	// 	// mat := val.gl()
// 	// 	// gl.UniformMatrix4fv(uniform.loc, mat[:])
// 	// default:
// 	// 	// fmt.Println("ERROR", uniform)
// 	// 	panic(fmt.Sprintf("set uniform attr: invalid attribute type: %T", value))
// 	// }
// }
