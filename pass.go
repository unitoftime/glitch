package glitch

import (
	"cmp"
	"slices"

	"github.com/unitoftime/glitch/internal/gl"
)

// Sorting
// 1. FBO Target
// 2. Shader program
// 3. Geometry

// TODO - This whole file needs to be rewritten. I'm thinking:
// 1. Single renderpass wrapper that lets met toggle all configs, shaders, meshes, etc
// 2. Things that are changed infrequently: SetBlah()
// 3. Things that change frequently: include in draw command

// TODO: I kindof want to implement one or the other
type GeometryFiller interface {
	GetBuffer() *VertexBuffer // Returns a prebuild VertexBuffer

	// TODO: I think you can simplify all of the draw options into one struct and pass it by pointer
	Fill(*RenderPass, glMat4, RGBA, BufferState) *VertexBuffer
}

type BatchTarget interface {
	// Add(GeometryFiller, Mat4, RGBA, Material, bool)
	Add(*Mesh, glMat4, RGBA, Material, bool)
}

type Target interface {
	// TODO - Should this be differentiated from being a source Vs a target binding. For example, I'm using this now to bind the target that we draw to. But If I want to have another function on frambuffers to use them as image texture inputs, what would that API be called?
	Bind()
}

// https://realtimecollisiondetection.net/blog/?p=86
// Sort by:
// - Front-to-back vs Back-to-front (single bit)
// - Depth bits
// - Material / Uniforms / Textures
// - Sort by: x, y, z, depth?
// I also feel like I'm finding myself mirroring this: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#DrawImageOptions
type drawCommand struct {
	// mesh *Mesh
	filler GeometryFiller
	matrix glMat4
	mask RGBA
	state BufferState
}

func SortDrawCommands(buf []drawCommand, sortMode SoftwareSortMode) {
	if sortMode == SoftwareSortNone { return } // Skip if sorting disabled

	if sortMode == SoftwareSortX {
		slices.SortStableFunc(buf, func(a, b drawCommand) int {
			return -cmp.Compare(a.matrix[i4_3_0], b.matrix[i4_3_0]) // sort by x
		})
	} else if sortMode == SoftwareSortY {
		slices.SortStableFunc(buf, func(a, b drawCommand) int {
			return -cmp.Compare(a.matrix[i4_3_1], b.matrix[i4_3_1]) // sort by y
		})
	} else if sortMode == SoftwareSortZ {
		slices.SortStableFunc(buf, func(a, b drawCommand) int {
			return -cmp.Compare(a.matrix[i4_3_2], b.matrix[i4_3_2]) // sort by z
		})
	}//  else if sortMode == SoftwareSortCommand {
	// 	slices.SortStableFunc(buf, func(a, b drawCommand) int {
	// 		return -cmp.Compare(a.command, b.command) // sort by command
	// 	})
	// }
}


type cmdList struct{
	Opaque []drawCommand
	Translucent []drawCommand
}

func (c *cmdList) Add(translucent bool, cmd drawCommand) *drawCommand {
	if translucent {
		c.Translucent = append(c.Translucent, cmd)
		return &c.Translucent[len(c.Translucent)-1]
	} else {
		c.Opaque = append(c.Opaque, cmd)
		return &c.Opaque[len(c.Opaque)-1]
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

// type meshBuffer struct {
// 	buffer *VertexBuffer
// }
// func newMeshBuffer(shader *Shader, mesh *Mesh) meshBuffer {
// 	if len(mesh.indices) % 3 != 0 {
// 		panic("Mesh indices must have 3 indices per triangle!")
// 	}
// 	numVerts := len(mesh.positions)
// 	numTris := len(mesh.indices) / 3
// 	meshBuf := meshBuffer{
// 		buffer: NewVertexBuffer(shader, numVerts, numTris),
// 	}
// 	return meshBuf
// }

type drawCall struct {
	buffer *VertexBuffer
	model glMat4
}

// This is essentially a generalized 2D render pass
type RenderPass struct {
	shader *Shader
	texture *Texture
	uniforms map[string]any
	uniformMat4 map[string]glMat4
	buffer *BufferPool
	commands []cmdList
	currentLayer int8 // TODO - layering code relies on the fact that this is a uint8, when you change, double check every usage of layers.

	blendMode BlendMode

	DepthTest bool // If set true, enable hardware depth testing. This changes how software sorting works. currently If you change this mid-pass you might get weird behavior.
	SoftwareSort SoftwareSortMode

	DepthBump bool
	depthBump float32

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
		uniformMat4: make(map[string]glMat4),
		buffer: NewBufferPool(shader, defaultBatchSize),
		commands: make([]cmdList, 256), // TODO - hardcoding from sizeof(uint8)
		drawCalls: make([]drawCall, 0),
		blendMode: BlendModeNormal,
	}
	return r
}

func (r *RenderPass) Clear() {
	r.depthBump = 0

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

// TODO: This shouldnt be global on the pass, but rather local on material
// Sets the blend mode for this render pass
func (r *RenderPass) SetBlendMode(bm BlendMode) {
	r.blendMode = bm
}

func (r *RenderPass) Batch() {
	r.SortInSoftware()

	// TODO: This isn't an efficient order for fill rate. You should reverse the order (but make an initial batch pass where you draw translucent geometry in the right order)
	// for l := len(r.commands)-1; l >= 0; l-- { // Reverse order so that layer 0 is drawn last
	// 	for i := range r.commands[l].Opaque {
	// 		r.applyDrawCommand(r.commands[l].Opaque[i])
	// 	}
	// 	for i := range r.commands[l].Translucent {
	// 		r.applyDrawCommand(r.commands[l].Translucent[i])
	// 	}
	// }

	// Opaque goes front to back (0 to 255)
	for l := range r.commands {
		for i := range r.commands[l].Opaque {
			// fmt.Println("- Opaque: (layer, x, z)", l, r.commands[l].Opaque[i].matrix[i4_3_1], r.commands[l].Opaque[i].matrix[i4_3_2])
			r.applyDrawCommand(r.commands[l].Opaque[i])
		}
	}

	// Translucent goes from back to front (255 to 0)
	for l := len(r.commands)-1; l >= 0; l-- { // Reverse order so that layer 0 is drawn last
		for i := range r.commands[l].Translucent {
			// fmt.Println("- Transl: (layer, x, z)", l, r.commands[l].Translucent[i].matrix[i4_3_1], r.commands[l].Translucent[i].matrix[i4_3_2])
			r.applyDrawCommand(r.commands[l].Translucent[i])
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
	vertexBuffer := c.filler.Fill(r, c.matrix, c.mask, c.state)

	// If the last buffer to draw isn't the currently used vertexBuffer, then we need to add it to the list
	if len(r.drawCalls) <= 0 || r.drawCalls[len(r.drawCalls) - 1].buffer != vertexBuffer {
		// Append the draw call to our list. Because we've already pre-applied the model matrix, we use Mat4Ident here
		r.drawCalls = append(r.drawCalls, drawCall{vertexBuffer, glMat4Ident})
	}
}

// TODO - Mat?
func (r *RenderPass) Draw(target Target) {
	r.Batch()

	// Bind render target
	target.Bind()

	state.enableDepthTest(r.DepthTest)
	if r.DepthTest {
		// state.setDepthFunc(gl.LEQUAL)
		state.setDepthFunc(gl.LESS)
	}

	r.shader.Bind()
	for k, v := range r.uniformMat4 {
		ok := r.shader.SetUniformMat4(k, v)
		if !ok {
			panic("Error setting uniform - todo decrease this to log")
		}
	}

	for k, v := range r.uniforms {
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

func (r *RenderPass) SetCamera2D(camera *CameraOrtho) {
	r.uniformMat4["projection"] = camera.Projection.gl()
	r.uniformMat4["view"] = camera.View.gl()
}

func (r *RenderPass) SetUniform(name string, value any) {
	r.uniforms[name] = value
}

// Option 1: I was thinking that I could add in the Z component on top of the Y component at the very end. but only use the early Y component for the sorting.
// Option 2: I could also just offset the geometry when I create the sprite (or after). Then simply use the transforms like normal. I'd just have to offset the sprite by the height, and then not add the height to the Y transformation
// Option 3: I can batch together these sprites into a single thing that is then rendered
func (r *RenderPass) Add(filler *Mesh, mat glMat4, mask RGBA, material Material, translucent bool) {
	if mask.A == 0 { return } // discard b/c its completely transparent

	if mask.A != 1 {
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
		if r.DepthBump {
			r.depthBump -= 0.00001 // TODO: Very very arbitrary
		}
		mat[i4_3_2] -= float32(r.currentLayer) + r.depthBump
		// fmt.Println("Depth: ", mat[i4_3_2])
	}

	r.commands[r.currentLayer].Add(translucent, drawCommand{
		filler, mat, mask, BufferState{material, r.blendMode},
	})
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


// Not thread safe
var lastState BufferState
// Not thread safe
func openglDraw(shader *Shader, draws []drawCall) {
	lastState = BufferState{}
	for i := range draws {
		buffer := draws[i].buffer
		// fmt.Println(i, len(b.buffers[i].indices), b.buffers[i].buffers[0].Len(), b.buffers[i].buffers[0].Cap())
		// TODO: Push this inward so concept like last state and whatever is managed by vertbuffer
		if lastState != buffer.state {
			lastState = buffer.state
			lastState.Bind()
		}

		ok := shader.SetUniformMat4("model", (draws[i].model))
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
	buffer.deallocAfterBuffer = true

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

	// TODO: Could use copy funcs if you want to restrict buffer types
	// for bufIdx, attr := range pass.shader.attrFmt {
	// 	switch attr.Swizzle {
	// 	case PositionXYZ:
	// 		buffer.buffers[bufIdx].SetData(mesh.positions)
	// 	case ColorRGBA:
	// 		buffer.buffers[bufIdx].SetData(mesh.colors)
	// 	case TexCoordXY:
	// 		buffer.buffers[bufIdx].SetData(mesh.texCoords)
	// 	default:
	// 		panic("unsupported")
	// 	}
	// }

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
