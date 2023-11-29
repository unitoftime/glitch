package glitch

import (

	// "math"
	"cmp"
	"slices"

	"github.com/unitoftime/glitch/internal/gl"
	// "github.com/unitoftime/glitch/internal/mainthread"
)

// TODO - This whole file needs to be rewritten. I'm thinking:
// 1. Single renderpass wrapper that lets met toggle all configs, shaders, meshes, etc
// 2. Things that are changed infrequently: SetBlah()
// 3. Things that change frequently: include in draw command

type BatchTarget interface {
	Add(GeometryFiller, Mat4, RGBA, Material, bool)
}

type Target interface {
	// TODO - Should this be differentiated from being a source Vs a target binding. For example, I'm using this now to bind the target that we draw to. But If I want to have another function on frambuffers to use them as image texture inputs, what would that API be called?
	Bind()
}

// TODO: I kindof want to implement one or the other
type GeometryFiller interface {
	GetBuffer() *VertexBuffer // Returns a prebuild VertexBuffer

	// NumVerts() int // Returns the number of verts to reserve
	// Indices() []uint32 // Retursn the indices to reserve

	// TODO: I think you can simplify all of the draw options into one struct and pass it by pointer
	Fill(*RenderPass, glMat4, RGBA, BufferState) *VertexBuffer
}

// https://realtimecollisiondetection.net/blog/?p=86
// Sort by:
// - Front-to-back vs Back-to-front (single bit)
// - Depth bits
// - Material / Uniforms / Textures
// - Sort by: x, y, z, depth?
// I also feel like I'm finding myself mirroring this: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#DrawImageOptions
type drawCommand struct {
	// command uint64
	// fillFunc func(VertexFormat, []interface{})
	// mesh *Mesh
	filler GeometryFiller
	matrix Mat4
	mask RGBA
	state BufferState
}

func SortDrawCommands(buf []drawCommand, sortMode SoftwareSortMode) {
	if sortMode == SoftwareSortNone { return } // Skip if sorting disabled

	if sortMode == SoftwareSortX {
		slices.SortFunc(buf, func(a, b drawCommand) int {
			return -cmp.Compare(a.matrix[i4_3_0], b.matrix[i4_3_0]) // sort by x
		})
	} else if sortMode == SoftwareSortY {
		slices.SortFunc(buf, func(a, b drawCommand) int {
			return -cmp.Compare(a.matrix[i4_3_1], b.matrix[i4_3_1]) // sort by y
		})
	} else if sortMode == SoftwareSortZ {
		slices.SortFunc(buf, func(a, b drawCommand) int {
			return -cmp.Compare(a.matrix[i4_3_2], b.matrix[i4_3_2]) // sort by z
		})
	}//  else if sortMode == SoftwareSortCommand {
	// 	slices.SortFunc(buf, func(a, b drawCommand) int {
	// 		return -cmp.Compare(a.command, b.command) // sort by command
	// 	})
	// }
}


type cmdList struct{
	Opaque []drawCommand
	Translucent []drawCommand
}

func (c *cmdList) Add(translucent bool, cmd drawCommand) {
	if translucent {
		c.Translucent = append(c.Translucent, cmd)
	} else {
		c.Opaque = append(c.Opaque, cmd)
	}
}

func (c *cmdList) SortTranslucent(sortMode SoftwareSortMode) {
	SortDrawCommands(c.Translucent, sortMode)
}

func (c *cmdList) SortOpaque(sortMode SoftwareSortMode) {
	SortDrawCommands(c.Opaque, sortMode)
}

func (c *cmdList) Clear() {
	c.Opaque = c.Opaque[:0]
	c.Translucent = c.Translucent[:0]
}

type meshBuffer struct {
	buffer *VertexBuffer
}
func newMeshBuffer(shader *Shader, mesh *Mesh) meshBuffer {
	if len(mesh.indices) % 3 != 0 {
		panic("Mesh indices must have 3 indices per triangle!")
	}
	numVerts := len(mesh.positions)
	numTris := len(mesh.indices) / 3
	meshBuf := meshBuffer{
		buffer: NewVertexBuffer(shader, numVerts, numTris),
	}
	return meshBuf
}

type drawCall struct {
	buffer *VertexBuffer
	model Mat4
}

// This is essentially a generalized 2D render pass
type RenderPass struct {
	shader *Shader
	texture *Texture
	uniforms map[string]any
	buffer *BufferPool
	commands []cmdList
	currentLayer int8 // TODO - layering code relies on the fact that this is a uint8, when you change, double check every usage of layers.

	blendMode BlendMode

	DepthTest bool // If set true, enable hardware depth testing. This changes how software sorting works. currently If you change this mid-pass you might get weird behavior.
	SoftwareSort SoftwareSortMode

	drawCalls []drawCall
}

type SoftwareSortMode uint8
const (
	SoftwareSortNone SoftwareSortMode = iota
	SoftwareSortX // Sort based on the X position
	SoftwareSortY // Sort based on the Y position
	SoftwareSortZ // Sort based on the Z position
	SoftwareSortCommand // Sort by the computed drawCommand.command
)

// const DefaultLayer uint8 = 127/2
// const DefaultLayer int8 = 0

func NewRenderPass(shader *Shader) *RenderPass {
	defaultBatchSize := 1024 * 8 // 10000 // TODO - arbitrary

	r := &RenderPass{
		shader: shader,
		texture: nil,
		uniforms: make(map[string]any),
		buffer: NewBufferPool(shader, defaultBatchSize),
		commands: make([]cmdList, 256), // TODO - hardcoding from sizeof(uint8)
		drawCalls: make([]drawCall, 0),
		blendMode: BlendModeNormal,
	}
	return r
}


func (r *RenderPass) Clear() {
	// Clear stuff
	r.buffer.Clear()
	for l := range r.commands {
		r.commands[l].Clear()
	}

	// TODO: I'm not 100% sure if this is needed, but we may need to clear the vbo pointers that exist in draw calls
	for i := range r.drawCalls {
		r.drawCalls[i] = drawCall{}
	}
	r.drawCalls = r.drawCalls[:0]
}

// TODO - I think I could use a linked list of layers and just use an int here
func (r *RenderPass) SetLayer(layer int8) {
	r.currentLayer = layer
}
func (r RenderPass) Layer() int8 {
	return r.currentLayer
}

// Sets the blend mode for this render pass
func (r *RenderPass) SetBlendMode(bm BlendMode) {
	r.blendMode = bm
}

func (r *RenderPass) Batch() {
	r.SortInSoftware()

	// TODO: This isn't an efficient order for fill rate. You should reverse the order (but make an initial batch pass where you draw translucent geometry in the right order)
	// for l := range r.commands { // Draw front to back
	for l := len(r.commands)-1; l >= 0; l-- { // Reverse order so that layer 0 is drawn last
		for _, c := range r.commands[l].Opaque {
			r.applyDrawCommand(c)
		}
		for _, c := range r.commands[l].Translucent {
			r.applyDrawCommand(c)
		}
	}
}

func (r *RenderPass) applyDrawCommand(c drawCommand) {
	if c.filler == nil { return } // Skip nil meshes

	// TODO: if command is a buffered mesh then just draw that
	if c.filler.GetBuffer() != nil {
		// Because we are about to use a custom meshBuffer, we need to make sure the next time we do an autobatch we use a new VertexBuffer, So we call this function to foward the autobuffer to the next clean buffer
		r.buffer.gotoNextClean()
		r.drawCalls = append(r.drawCalls, drawCall{c.filler.GetBuffer(), c.matrix})
		return
	}

	// Else we are auto-batching the mesh because the mesh is small
	vertexBuffer := c.filler.Fill(r, c.matrix.gl(), c.mask, c.state)

	// numVerts := c.filler.NumVerts()
	// indices := c.filler.Indices()
	// vertexBuffer := r.buffer.Reserve(c.state, indices, numVerts, r.shader.tmpBuffers)
	// r.batchToBuffers(c, r.shader.tmpBuffers)

	// If the last buffer to draw isn't the currently used vertexBuffer, then we need to add it to the list
	if len(r.drawCalls) <= 0 || r.drawCalls[len(r.drawCalls) - 1].buffer != vertexBuffer {
		// Append the draw call to our list. Because we've already pre-applied the model matrix, we use Mat4Ident here
		r.drawCalls = append(r.drawCalls, drawCall{vertexBuffer, Mat4Ident})
	}
}

// TODO - Mat?
func (r *RenderPass) Draw(target Target) {
	r.Batch()

	// Bind render target
	target.Bind()

	// mainthread.Call(r.mainthreadDepthTest)
	state.enableDepthTest(gl.LEQUAL)// TODO - rehook for depthtest flags

	r.shader.Bind()
	for k,v := range r.uniforms {
		ok := r.shader.SetUniform(k, v)
		if !ok {
			panic("Error setting uniform - todo decrease this to log")
		}
	}

	openglDraw(r.shader, r.drawCalls)
}

func (r *RenderPass) SetTexture(slot int, texture *Texture) {
	// TODO - use correct texture slot
	r.texture = texture
}

// TODO: Maybe do this to prevent allocations from the `any` cast
// func SetUniform[T any](r *RenderPass, name, val T) {
// }

func (r *RenderPass) SetUniform(name string, value any) {
	r.uniforms[name] = value
}

// Option 1: I was thinking that I could add in the Z component on top of the Y component at the very end. but only use the early Y component for the sorting.
// Option 2: I could also just offset the geometry when I create the sprite (or after). Then simply use the transforms like normal. I'd just have to offset the sprite by the height, and then not add the height to the Y transformation
// Option 3: I can batch together these sprites into a single thing that is then rendered
func (r *RenderPass) Add(filler GeometryFiller, mat Mat4, mask RGBA, material Material, translucent bool) {
	if mask.A != 0 && mask.A != 1 {
		translucent = true
	}

	if r.DepthTest {
		// If we are doing depth testing, then use the r.CurrentLayer field to determine the depth (normalizing from (0 to 1). Notably the standard ortho cam is (-1, 1) which this range fits into but is easier to normalize to // TODO - make that depth range tweakable?
		// TODO - hardcoded because layer is a uint8. You probably want to make layer an int and then just set depth based on that
		// depth := 1 - (float32(r.currentLayer) / float32(math.MaxUint8))
		// // fmt.Println("Old/New: ", mat[i4_3_2], depth)
		// mat[i4_3_2] = depth // Set Z translation to the depth

		// fmt.Println("Apply: ", mat.Apply(Vec3{0, 0, 0}))

		// mat[i4_3_2] = mat[i4_3_1] // Set Z translation to the y point

		// Add the layer to the depth
		mat[i4_3_2] -= float64(r.currentLayer)
		// fmt.Println("Depth: ", mat[i4_3_2])

		r.commands[r.currentLayer].Add(translucent, drawCommand{
			filler, mat, mask, BufferState{material, r.blendMode},
		})
	} else {
		r.commands[r.currentLayer].Add(translucent, drawCommand{
			filler, mat, mask, BufferState{material, r.blendMode},
		})
	}
}

func (r *RenderPass) SortInSoftware() {
	if r.DepthTest {
		// TODO - do special sort function for depth test code:
		// 1. Fully Opaque or fully transparent groups of meshes: Don't sort inside that group
		// 2. Partially transparent groups of meshes: sort inside that group
		// 3. Take into account blendMode

		// Sort translucent buffer
		for l := range r.commands {
			r.commands[l].SortTranslucent(r.SoftwareSort)
		}

		return
	}

	for l := range r.commands {
		r.commands[l].SortOpaque(r.SoftwareSort)
	}
}

// func (r *RenderPass) batchToBuffers(c drawCommand, destBuffs []interface{}) {
// 	mat32 := c.matrix.gl()

// 	// Append all mesh buffers to shader buffers
// 	for bufIdx, attr := range r.shader.attrFmt {
// 		// TODO - I'm not sure of a good way to break up this switch statement
// 		switch attr.Swizzle {
// 			// Positions
// 		case PositionXY:
// 			posBuf := *(destBuffs[bufIdx]).(*[]glVec2)
// 			if c.matrix == Mat4Ident {
// 				// If matrix is identity, don't transform anything
// 				for i := range c.mesh.positions {
// 					posBuf[i] = *(*glVec2)(c.mesh.positions[i][:2])
// 				}
// 			} else {
// 				for i := range c.mesh.positions {
// 					vec := mat32.Apply(c.mesh.positions[i])
// 					posBuf[i] = *(*glVec2)(vec[:2])
// 				}
// 			}

// 		case PositionXYZ:
// 			posBuf := *(destBuffs[bufIdx]).(*[]glVec3)
// 			if c.matrix == Mat4Ident {
// 				// If matrix is identity, don't transform anything
// 				for i := range c.mesh.positions {
// 					posBuf[i] = c.mesh.positions[i]
// 				}
// 			} else {
// 				for i := range c.mesh.positions {
// 					vec := mat32.Apply(c.mesh.positions[i])
// 					posBuf[i] = vec
// 				}
// 			}

// 			// Normals
// 			// TODO - Renormalize if batching
// 			// case NormalXY:
// 			// 	normBuf := *(destBuffs[bufIdx]).(*[]Vec2)
// 			// 	for i := range c.mesh.normals {
// 			// 		vec := c.mesh.normals[i]
// 			// 		normBuf[i] = *(*Vec2)(vec[:2])
// 			// 	}

// 		case NormalXYZ:
// 			renormalizeMat := c.matrix.Inv().Transpose().gl()
// 			normBuf := *(destBuffs[bufIdx]).(*[]glVec3)
// 			for i := range c.mesh.normals {
// 				vec := renormalizeMat.Apply(c.mesh.normals[i])
// 				normBuf[i] = vec
// 			}

// 			// Colors
// 		case ColorR:
// 			colBuf := *(destBuffs[bufIdx]).(*[]float32)
// 			for i := range c.mesh.colors {
// 				colBuf[i] = c.mesh.colors[i][0] * float32(c.mask.R)
// 			}
// 		case ColorRG:
// 			colBuf := *(destBuffs[bufIdx]).(*[]glVec2)
// 			for i := range c.mesh.colors {
// 				colBuf[i] = glVec2{
// 					c.mesh.colors[i][0] * float32(c.mask.R),
// 					c.mesh.colors[i][1] * float32(c.mask.G),
// 				}
// 			}
// 		case ColorRGB:
// 			colBuf := *(destBuffs[bufIdx]).(*[]glVec3)
// 			for i := range c.mesh.colors {
// 				colBuf[i] = glVec3{
// 					c.mesh.colors[i][0] * float32(c.mask.R),
// 					c.mesh.colors[i][1] * float32(c.mask.G),
// 					c.mesh.colors[i][2] * float32(c.mask.B),
// 				}
// 			}
// 		case ColorRGBA:
// 			colBuf := *(destBuffs[bufIdx]).(*[]glVec4)
// 			for i := range c.mesh.colors {
// 				colBuf[i] = glVec4{
// 					c.mesh.colors[i][0] * float32(c.mask.R),
// 					c.mesh.colors[i][1] * float32(c.mask.G),
// 					c.mesh.colors[i][2] * float32(c.mask.B),
// 					c.mesh.colors[i][3] * float32(c.mask.A),
// 				}
// 			}

// 		case TexCoordXY:
// 			texBuf := *(destBuffs[bufIdx]).(*[]glVec2)
// 			for i := range c.mesh.texCoords {
// 				texBuf[i] = c.mesh.texCoords[i]
// 			}
// 		default:
// 			panic("Unsupported")
// 		}
// 	}

// 	//================================================================================
// 	// TODO The hardcoding is a bit slower. Keeping it around in case I want to do some performance analysis
// 	// Notes: Ran gophermark with 1000000 gophers.
// 	// - Hardcoded: ~ 120 to 125 ms range
// 	// - Switch Statement: ~ 125 to 130 ms range
// 	// - Switch Statement (with shader changed to use vec2s for position): ~ 122 to 127 ms range
// 	// work and append
// 	// 	posBuf := *(destBuffs[0]).(*[]Vec3)
// 	// 	for i := range c.mesh.positions {
// 	// 		vec := c.matrix.Apply(c.mesh.positions[i])
// 	// 		posBuf[i] = vec
// 	// 	}

// 	// 	colBuf := *(destBuffs[1]).(*[]Vec4)
// 	// 	for i := range c.mesh.colors {
// 	// 		colBuf[i] = Vec4{
// 	// 			c.mesh.colors[i][0] * c.mask.R,
// 	// 			c.mesh.colors[i][1] * c.mask.G,
// 	// 			c.mesh.colors[i][2] * c.mask.B,
// 	// 			c.mesh.colors[i][3] * c.mask.A,
// 	// 		}
// 	// 	}

// 	// 	texBuf := *(destBuffs[2]).(*[]Vec2)
// 	// 	for i := range c.mesh.texCoords {
// 	// 		texBuf[i] = c.mesh.texCoords[i]
// 	// 	}
// 	//================================================================================
// }

// Not thread safe
var lastState BufferState
// Not thread safe
func openglDraw(shader *Shader, draws []drawCall) {
	lastState = BufferState{}
	for i := range draws {
		buffer := draws[i].buffer
		// fmt.Println(i, len(b.buffers[i].indices), b.buffers[i].buffers[0].Len(), b.buffers[i].buffers[0].Cap())
		if lastState != buffer.state {
			lastState = buffer.state
			lastState.Bind()
		}

		ok := shader.SetUniformMat4("model", &(draws[i].model))
		if !ok {
			panic("Error setting model uniform - all shaders must have 'model' uniform")
		}

		buffer.Draw()
	}
}

//--------------------------------------------------------------------------------
func (pass *RenderPass) BufferMesh(mesh *Mesh, material Material, translucent bool) *VertexBuffer {

	bufferState := BufferState{material, pass.blendMode}

	if len(mesh.indices) % 3 != 0 {
		panic("Cmd.Mesh indices must have 3 indices per triangle!")
	}
	numVerts := len(mesh.positions)
	numTris := len(mesh.indices) / 3
	buffer := NewVertexBuffer(pass.shader, numVerts, numTris)

	success := buffer.Reserve(bufferState, mesh.indices, numVerts, pass.shader.tmpBuffers)
	if !success {
		panic("Something went wrong")
	}

	// cmd := drawCommand{
	// 	mesh, Mat4Ident, White, bufferState,
	// }
	// TODO: Translucent?
	// TODO: Depth sorting?
	batchToBuffers(pass.shader, mesh, glMat4Ident, White)

	// pass.copyToBuffer(cmd, pass.shader.tmpBuffers)

	return buffer
}

// // This is like batchToBuffer but doesn't pre-apply the model matrix of the mesh
// func (r *RenderPass) copyToBuffer(c drawCommand, destBuffs []interface{}) {
// 	// // For now I'm just going to modify the drawCommand to use Mat4Ident and then pass to batchToBuffers
// 	// c.matrix = Mat4Ident
// 	// batchToBuffers(c, destBuffs)

// 	numVerts := c.filler.NumVerts()
// 	indices := c.filler.Indices()
// 	vertexBuffer := pass.buffer.Reserve(state, indices, numVerts, pass.shader.tmpBuffers)
// 	batchToBuffers(pass.shader, m, mat, mask)
// }
