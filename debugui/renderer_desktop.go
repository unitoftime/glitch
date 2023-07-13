//go:build !js
package debugui

// Note: this whole file is ripped directly from: https://github.com/inkyblackness/imgui-go-examples/tree/master/internal

import (
	_ "embed" // using embed for the shader sources
	// "fmt"

	// "github.com/inkyblackness/imgui-go-examples/internal/renderers/gl/v3.2-core/gl"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/unitoftime/glitch/internal/gl"
)

//go:embed main.vert
var unversionedVertexShader string

//go:embed main.frag
var unversionedFragmentShader string

// OpenGL3 implements a renderer based on github.com/go-gl/gl (v3.2-core).
type OpenGL3 struct {
	imguiIO imgui.IO

	glslVersion            string
	fontTexture            gl.Texture
	shaderHandle           gl.Program
	vertHandle             gl.Shader
	fragHandle             gl.Shader
	attribLocationTex      gl.Uniform
	attribLocationProjMtx  gl.Uniform
	attribLocationPosition gl.Attrib
	attribLocationUV       gl.Attrib
	attribLocationColor    gl.Attrib
	vboHandle              gl.Buffer
	elementsHandle         gl.Buffer
}

// NewOpenGL3 attempts to initialize a renderer.
// An OpenGL context has to be established before calling this function.
func NewOpenGL3(io imgui.IO) (*OpenGL3, error) {
	// TODO: Already initialized by glitch?
	// err := gl.Init()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to initialize OpenGL: %w", err)
	// }

	renderer := &OpenGL3{
		imguiIO:     io,
		glslVersion: "#version 150",
	}
	renderer.createDeviceObjects()

	io.SetBackendFlags(io.GetBackendFlags() | imgui.BackendFlagsRendererHasVtxOffset)

	return renderer, nil
}

// Dispose cleans up the resources.
func (renderer *OpenGL3) Dispose() {
	renderer.invalidateDeviceObjects()
}

// PreRender clears the framebuffer.
func (renderer *OpenGL3) PreRender(clearColor [3]float32) {
	gl.ClearColor(clearColor[0], clearColor[1], clearColor[2], 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

// Render translates the ImGui draw data to OpenGL3 commands.
func (renderer *OpenGL3) Render(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData) {
	// Avoid rendering when minimized, scale coordinates for retina displays (screen coordinates != framebuffer coordinates)
	displayWidth, displayHeight := displaySize[0], displaySize[1]
	fbWidth, fbHeight := framebufferSize[0], framebufferSize[1]
	if (fbWidth <= 0) || (fbHeight <= 0) {
		return
	}
	drawData.ScaleClipRects(imgui.Vec2{
		X: fbWidth / displayWidth,
		Y: fbHeight / displayHeight,
	})

	// Backup GL state
	lastActiveTexture := gl.GetInteger(gl.ACTIVE_TEXTURE)
	gl.ActiveTexture(gl.TEXTURE0)
	lastProgram := gl.GetInteger(gl.CURRENT_PROGRAM)
	lastTexture := gl.GetInteger(gl.TEXTURE_BINDING_2D)
	// lastSampler := gl.GetInteger(gl.SAMPLER_BINDING)
	lastArrayBuffer := gl.GetInteger(gl.ARRAY_BUFFER_BINDING)
	lastElementArrayBuffer := gl.GetInteger(gl.ELEMENT_ARRAY_BUFFER_BINDING)
	lastVertexArray := gl.GetInteger(gl.VERTEX_ARRAY_BINDING)
	// var lastPolygonMode [2]int32
	// gl.GetIntegerv(gl.POLYGON_MODE, lastPolygonMode)
	lastViewport := make([]int32, 4)
	gl.GetIntegerv(gl.VIEWPORT, lastViewport)
	lastScissorBox := make([]int32, 4)
	gl.GetIntegerv(gl.SCISSOR_BOX, lastScissorBox)
	lastBlendSrcRgb := gl.GetInteger(gl.BLEND_SRC_RGB)
	lastBlendDstRgb := gl.GetInteger(gl.BLEND_DST_RGB)
	lastBlendSrcAlpha := gl.GetInteger(gl.BLEND_SRC_ALPHA)
	lastBlendDstAlpha := gl.GetInteger(gl.BLEND_DST_ALPHA)
	lastBlendEquationRgb := gl.GetInteger(gl.BLEND_EQUATION_RGB)
	lastBlendEquationAlpha := gl.GetInteger(gl.BLEND_EQUATION_ALPHA)
	lastEnableBlend := gl.IsEnabled(gl.BLEND)
	lastEnableCullFace := gl.IsEnabled(gl.CULL_FACE)
	lastEnableDepthTest := gl.IsEnabled(gl.DEPTH_TEST)
	lastEnableScissorTest := gl.IsEnabled(gl.SCISSOR_TEST)

	// Setup render state: alpha-blending enabled, no face culling, no depth testing, scissor enabled, polygon fill
	gl.Enable(gl.BLEND)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.SCISSOR_TEST)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	// Setup viewport, orthographic projection matrix
	// Our visible imgui space lies from draw_data->DisplayPos (top left) to draw_data->DisplayPos+data_data->DisplaySize (bottom right).
	// DisplayMin is typically (0,0) for single viewport apps.
	gl.Viewport(0, 0, int(fbWidth), int(fbHeight))
	// orthoProjection := [4][4]float32{
	// 	{2.0 / displayWidth, 0.0, 0.0, 0.0},
	// 	{0.0, 2.0 / -displayHeight, 0.0, 0.0},
	// 	{0.0, 0.0, -1.0, 0.0},
	// 	{-1.0, 1.0, 0.0, 1.0},
	// }
	orthoProjection := []float32{
		2.0 / displayWidth, 0.0, 0.0, 0.0,
		0.0, 2.0 / -displayHeight, 0.0, 0.0,
		0.0, 0.0, -1.0, 0.0,
		-1.0, 1.0, 0.0, 1.0,
	}
	gl.UseProgram(renderer.shaderHandle)
	gl.Uniform1i(renderer.attribLocationTex, 0)
	gl.UniformMatrix4fv(renderer.attribLocationProjMtx, orthoProjection)
	// gl.BindSampler(0, 0) // Rely on combined texture/sampler state. //TODO:????

	// Recreate the VAO every time
	// (This is to easily allow multiple GL contexts. VAO are not shared among GL contexts, and
	// we don't track creation/deletion of windows so we don't have an obvious key to use to cache them.)
	vaoHandle := gl.GenVertexArrays()
	gl.BindVertexArray(vaoHandle)
	gl.BindBuffer(gl.ARRAY_BUFFER, renderer.vboHandle)
	gl.EnableVertexAttribArray(renderer.attribLocationPosition)
	gl.EnableVertexAttribArray(renderer.attribLocationUV)
	gl.EnableVertexAttribArray(renderer.attribLocationColor)
	vertexSize, vertexOffsetPos, vertexOffsetUv, vertexOffsetCol := imgui.VertexBufferLayout()
	// gl.VertexAttribPointerWithOffset(uint32(renderer.attribLocationPosition), 2, gl.FLOAT, false, int32(vertexSize), uintptr(vertexOffsetPos))
	// gl.VertexAttribPointerWithOffset(uint32(renderer.attribLocationUV), 2, gl.FLOAT, false, int32(vertexSize), uintptr(vertexOffsetUv))
	// gl.VertexAttribPointerWithOffset(uint32(renderer.attribLocationColor), 4, gl.UNSIGNED_BYTE, true, int32(vertexSize), uintptr(vertexOffsetCol))
	gl.VertexAttribPointer(renderer.attribLocationPosition, 2, gl.FLOAT, false, vertexSize, vertexOffsetPos)
	gl.VertexAttribPointer(renderer.attribLocationUV, 2, gl.FLOAT, false, vertexSize, vertexOffsetUv)
	gl.VertexAttribPointer(renderer.attribLocationColor, 4, gl.UNSIGNED_BYTE, true, vertexSize, vertexOffsetCol)
	indexSize := imgui.IndexBufferLayout()
	drawType := gl.UNSIGNED_SHORT
	const bytesPerUint32 = 4
	if indexSize == bytesPerUint32 {
		drawType = gl.UNSIGNED_INT
	}

	// Draw
	for _, list := range drawData.CommandLists() {
		vertexBuffer, vertexBufferSize := list.VertexBuffer()
		gl.BindBuffer(gl.ARRAY_BUFFER, renderer.vboHandle)
		gl.BufferDataImguiPassthrough(gl.ARRAY_BUFFER, vertexBufferSize, vertexBuffer, gl.STREAM_DRAW)

		indexBuffer, indexBufferSize := list.IndexBuffer()
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, renderer.elementsHandle)
		gl.BufferDataImguiPassthrough(gl.ELEMENT_ARRAY_BUFFER, indexBufferSize, indexBuffer, gl.STREAM_DRAW)

		for _, cmd := range list.Commands() {
			if cmd.HasUserCallback() {
				cmd.CallUserCallback(list)
			} else {
				gl.BindTexture(gl.TEXTURE_2D, gl.Texture{uint32(cmd.TextureID())})
				clipRect := cmd.ClipRect()
				gl.Scissor(int32(clipRect.X), int32(fbHeight)-int32(clipRect.W), int32(clipRect.Z-clipRect.X), int32(clipRect.W-clipRect.Y))
				// gl.DrawElementsBaseVertexWithOffset(gl.TRIANGLES, int32(cmd.ElementCount()), uint32(drawType),
				// 	uintptr(cmd.IndexOffset()*indexSize), int32(cmd.VertexOffset()))
				gl.DrawElements(gl.TRIANGLES, int(cmd.ElementCount()), gl.Enum(drawType),
					cmd.IndexOffset()*indexSize)
					// uintptr(cmd.IndexOffset()*indexSize), int32(cmd.VertexOffset()))
			}
		}
	}
	gl.DeleteVertexArrays(vaoHandle)

	// Restore modified GL state
	gl.UseProgram(gl.Program(lastProgram))
	gl.BindTexture(gl.TEXTURE_2D, gl.Texture{lastTexture.Value})
	// gl.BindSampler(0, lastSampler)
	gl.ActiveTexture(gl.Enum(lastActiveTexture.Value))
	gl.BindVertexArray(gl.Buffer(lastVertexArray))
	gl.BindBuffer(gl.ARRAY_BUFFER, gl.Buffer(lastArrayBuffer))
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, gl.Buffer(lastElementArrayBuffer))
	gl.BlendEquationSeparate(gl.Enum(lastBlendEquationRgb.Value), gl.Enum(lastBlendEquationAlpha.Value))
	gl.BlendFuncSeparate(gl.Enum(lastBlendSrcRgb.Value), gl.Enum(lastBlendDstRgb.Value), gl.Enum(lastBlendSrcAlpha.Value), gl.Enum(lastBlendDstAlpha.Value))
	if lastEnableBlend {
		gl.Enable(gl.BLEND)
	} else {
		gl.Disable(gl.BLEND)
	}
	if lastEnableCullFace {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
	if lastEnableDepthTest {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	if lastEnableScissorTest {
		gl.Enable(gl.SCISSOR_TEST)
	} else {
		gl.Disable(gl.SCISSOR_TEST)
	}
	// gl.PolygonMode(gl.FRONT_AND_BACK, uint32(lastPolygonMode[0]))
	gl.Viewport(int(lastViewport[0]), int(lastViewport[1]), int(lastViewport[2]), int(lastViewport[3]))
	gl.Scissor(lastScissorBox[0], lastScissorBox[1], lastScissorBox[2], lastScissorBox[3])
}

func (renderer *OpenGL3) createDeviceObjects() {
	// Backup GL state
	lastTexture := gl.GetInteger(gl.TEXTURE_BINDING_2D)
	lastArrayBuffer := gl.GetInteger(gl.ARRAY_BUFFER_BINDING)
	lastVertexArray := gl.GetInteger(gl.VERTEX_ARRAY_BINDING)

	vertexShader := renderer.glslVersion + "\n" + unversionedVertexShader
	fragmentShader := renderer.glslVersion + "\n" + unversionedFragmentShader

	renderer.shaderHandle = gl.CreateProgram()
	renderer.vertHandle = gl.CreateShader(gl.VERTEX_SHADER)
	renderer.fragHandle = gl.CreateShader(gl.FRAGMENT_SHADER)

	// glShaderSource := func(handle uint32, source string) {
	// 	csource, free := gl.Strs(source + "\x00")
	// 	defer free()

	// 	gl.ShaderSource(handle, 1, csource, nil)
	// }
	// glShaderSource(renderer.vertHandle, vertexShader)
	// glShaderSource(renderer.fragHandle, fragmentShader)

	gl.ShaderSource(renderer.vertHandle, vertexShader)
	gl.ShaderSource(renderer.fragHandle, fragmentShader)

	gl.CompileShader(renderer.vertHandle)
	gl.CompileShader(renderer.fragHandle)
	gl.AttachShader(renderer.shaderHandle, renderer.vertHandle)
	gl.AttachShader(renderer.shaderHandle, renderer.fragHandle)
	gl.LinkProgram(renderer.shaderHandle)

	renderer.attribLocationTex = gl.GetUniformLocation(renderer.shaderHandle, "Texture")
	renderer.attribLocationProjMtx = gl.GetUniformLocation(renderer.shaderHandle, "ProjMtx")
	renderer.attribLocationPosition = gl.GetAttribLocation(renderer.shaderHandle, "Position")
	renderer.attribLocationUV = gl.GetAttribLocation(renderer.shaderHandle, "UV")
	renderer.attribLocationColor = gl.GetAttribLocation(renderer.shaderHandle, "Color")

	renderer.vboHandle = gl.GenBuffers()
	renderer.elementsHandle = gl.GenBuffers()

	renderer.createFontsTexture()

	// Restore modified GL state
	gl.BindTexture(gl.TEXTURE_2D, gl.Texture{lastTexture.Value})
	gl.BindBuffer(gl.ARRAY_BUFFER, gl.Buffer(lastArrayBuffer))
	gl.BindVertexArray(gl.Buffer(lastVertexArray))
}

func (renderer *OpenGL3) createFontsTexture() {
	// Build texture atlas
	io := imgui.CurrentIO()
	image := io.Fonts().TextureDataAlpha8()

	// Upload texture to graphics system
	lastTexture := gl.GetInteger(gl.TEXTURE_BINDING_2D)
	// gl.GenTextures(1, &renderer.fontTexture)
	renderer.fontTexture = gl.CreateTexture()
	gl.BindTexture(gl.TEXTURE_2D, renderer.fontTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.TexImage2DFullImguiPassthrough(gl.TEXTURE_2D, 0, gl.RED, image.Width, image.Height,
		gl.RED, gl.UNSIGNED_BYTE, image.Pixels)

	// Store our identifier
	io.Fonts().SetTextureID(imgui.TextureID(renderer.fontTexture.Value))

	// Restore state
	gl.BindTexture(gl.TEXTURE_2D, gl.Texture{lastTexture.Value})
}

func (renderer *OpenGL3) invalidateDeviceObjects() {
	if !renderer.vboHandle.Valid() {
		gl.DeleteBuffers(renderer.vboHandle)
	}
	renderer.vboHandle = gl.NoBuffer
	if !renderer.elementsHandle.Valid() {
		gl.DeleteBuffers(renderer.elementsHandle)
	}
	renderer.elementsHandle = gl.NoBuffer

	if (!renderer.shaderHandle.Valid()) && (!renderer.vertHandle.Valid()) {
		gl.DetachShader(renderer.shaderHandle, renderer.vertHandle)
	}
	if !renderer.vertHandle.Valid() {
		gl.DeleteShader(renderer.vertHandle)
	}
	renderer.vertHandle = gl.NoShader

	if (!renderer.shaderHandle.Valid()) && (!renderer.fragHandle.Valid()) {
		gl.DetachShader(renderer.shaderHandle, renderer.fragHandle)
	}
	if !renderer.fragHandle.Valid() {
		gl.DeleteShader(renderer.fragHandle)
	}
	renderer.fragHandle = gl.NoShader

	if !renderer.shaderHandle.Valid() {
		gl.DeleteProgram(renderer.shaderHandle)
	}
	renderer.shaderHandle = gl.NoProgram

	if !renderer.fontTexture.Valid() {
		gl.DeleteTexture(renderer.fontTexture)
		imgui.CurrentIO().Fonts().SetTextureID(0)
		renderer.fontTexture = gl.NoTexture
	}
}
