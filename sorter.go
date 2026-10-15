package glitch

import (
	"math"
	"slices"
)

// you were here creating the sorter
// 1. every draw command needs ALL of the data to be sorted
// 2. you need to simplify the sorting to its all command based imo
type Sorter struct {
	DepthTest    bool
	SoftwareSort SoftwareSortMode
	DepthBump    bool
	depthBump    float32
	currentLayer int8

	// States that are used for forming the draw command
	// blendMode BlendMode
	// shader *Shader
	// material Material

	// camera *CameraOrtho

	commands []cmdList
	keys     []uint64 // Scratch buffer for sorting, reused every draw (See: drawCommands)
}

func NewSorter() *Sorter {
	return &Sorter{
		commands: make([]cmdList, math.MaxUint8+1), // One cmdList per int8 layer
	}
}

// Maps an int8 layer [-128, 127] to a command list index [0, 255], preserving order
func layerIndex(layer int8) int {
	return int(layer) - math.MinInt8
}

func (s *Sorter) SetLayer(layer int8) {
	s.currentLayer = layer
}

func (s *Sorter) Layer() int8 {
	return s.currentLayer
}

func (s *Sorter) Clear() {
	s.depthBump = 0

	// Clear stuff
	for l := range s.commands {
		s.commands[l].Clear()
	}
}

// func (s *Sorter) SetShader(shader *Shader) {
// 	s.shader = shader
// }

// func (s *Sorter) SetMaterial(material Material) {
// 	s.material = material
// }

// func (s *Sorter) SetCamera(camera *CameraOrtho) {
// 	s.camera = camera
// }

func (s *Sorter) Draw(target BatchTarget) {
	if s.DepthTest {
		// TODO - do special sort function for depth test code:
		// 1. Fully Opaque or fully transparent groups of meshes: Don't sort inside that group
		// 2. Partially transparent groups of meshes: sort inside that group
		// 3. Take into account blendMode

		// Opaque goes front to back (0 to 255), unsorted
		for l := range s.commands {
			s.drawCommands(target, s.commands[l].Opaque, SoftwareSortNone)
		}

		// Translucent goes from back to front (255 to 0)
		for l := len(s.commands) - 1; l >= 0; l-- { // Reverse order so that layer 0 is drawn last
			s.drawCommands(target, s.commands[l].Translucent, s.SoftwareSort)
		}
	} else {
		for l := len(s.commands) - 1; l >= 0; l-- { // Reverse order so that layer 0 is drawn last
			s.drawCommands(target, s.commands[l].Opaque, s.SoftwareSort)
			s.drawCommands(target, s.commands[l].Translucent, s.SoftwareSort)
		}
	}

	s.Clear()
}

// Draws the commands to the target in the order given by the sort mode
// Note: Rather than sorting the commands (which are large), we sort uint64 keys that index into them.
// The high 32 bits hold the sort value and the low 32 bits hold the command index. Because the index breaks ties,
// commands with equal values keep the order they were added in (ie the same result as a stable sort)
func (s *Sorter) drawCommands(target BatchTarget, cmds []drawCommand, sortMode SoftwareSortMode) {
	matIdx, descending, sorted := sortModeKey(sortMode)
	if !sorted {
		for i := range cmds {
			s.applyDrawCommand(target, cmds[i])
		}
		return
	}

	s.keys = s.keys[:0]
	for i := range cmds {
		key := sortableFloatBits(cmds[i].matrix[matIdx])
		if descending {
			key = ^key
		}
		s.keys = append(s.keys, uint64(key)<<32|uint64(i))
	}
	slices.Sort(s.keys)

	for _, k := range s.keys {
		s.applyDrawCommand(target, cmds[uint32(k)])
	}
}

func (s *Sorter) applyDrawCommand(target BatchTarget, c drawCommand) {
	target.Add(c.filler, c.matrix, c.mask, c.material)
}

func (s *Sorter) Add(filler GeometryFiller, mat glMat4, mask RGBA, material Material) {
	if mask.A == 0 {
		return
	} // discard b/c its completely transparent

	// TODO: Slightly hacky but basically if the sprite expects to blend then we want to sort it as if it translucent
	translucent := false
	if mask.A != 1 {
		translucent = true
		if material.blend == BlendModeNone {
			// TODO: Hacky, correct the blend mode if its translucent
			material.SetBlendMode(BlendModeNormal)
		}
	}

	if material.blend != BlendModeNone {
		translucent = true
	}

	if s.DepthTest {
		// If we are doing depth testing, then use the s.CurrentLayer field to determine the depth (normalizing from (0 to 1). Notably the standard ortho cam is (-1, 1) which this range fits into but is easier to normalize to // TODO - make that depth range tweakable?
		// TODO - hardcoded because layer is a uint8. You probably want to make layer an int and then just set depth based on that
		// depth := 1 - (float32(s.currentLayer) / float32(math.MaxUint8))
		// // fmt.Println("Old/New: ", mat[i4_3_2], depth)
		// mat[i4_3_2] = depth // Set Z translation to the depth

		// fmt.Println("Apply: ", mat.Apply(Vec3{0, 0, 0}))

		// mat[i4_3_2] = mat[i4_3_1] // Set Z translation to the y point

		// Add the layer to the depth
		if s.DepthBump {
			s.depthBump -= 0.00001 // TODO: Very very arbitrary
		}
		mat[i4_3_2] -= float32(s.currentLayer) + s.depthBump
		// fmt.Println("Depth: ", mat[i4_3_2])

		// TODO: Hardcoded depthmode sort type
		material.depth = DepthModeLess
	} else {
		// TODO: If depthtest if false, I want to ensure that the drawCalls are sorted as if they are translucent
		translucent = true
		mat[i4_3_2] -= float32(s.currentLayer)
	}

	// state := BufferState{materialGroup{s.material, material}, s.blendMode}
	// // if s.camera != nil {
	// // 	material.camera = s.camera
	// // }
	// if s.DepthTest {
	// 	material.depth = DepthModeLess
	// }

	s.commands[layerIndex(s.currentLayer)].Add(translucent, drawCommand{
		filler, mat, mask, material,
	})
}

// --------------------------------------------------------------------------------
type cmdList struct {
	Opaque      []drawCommand // TODO: This is kindof more like the unsorted list
	Translucent []drawCommand // TODO: This is kindof more like the sorted list

}

func (c *cmdList) Add(translucent bool, cmd drawCommand) {
	if translucent {
		c.Translucent = append(c.Translucent, cmd)
		// return &c.Translucent[len(c.Translucent)-1]
	} else {
		c.Opaque = append(c.Opaque, cmd)
		// return &c.Opaque[len(c.Opaque)-1]
	}
}

func (c *cmdList) Clear() {
	c.Opaque = c.Opaque[:0]
	c.Translucent = c.Translucent[:0]
}

type SoftwareSortMode uint8

const (
	SoftwareSortNone      SoftwareSortMode = iota
	SoftwareSortX                          // Sort based on the X position
	SoftwareSortY                          // Sort based on the Y position
	SoftwareSortZ                          // Sort based on the Z position
	SoftwareSortZNegative                  // Opposite order to SoftwareSortZ
	SoftwareSortCommand                    // Sort by the computed drawCommand.command
)

type drawCommand struct {
	filler   GeometryFiller
	matrix   glMat4
	mask     RGBA
	material Material
}

// Returns the matrix index to sort on, and whether to sort it descending. Returns false if the mode doesn't sort
func sortModeKey(sortMode SoftwareSortMode) (matIdx int, descending bool, sorted bool) {
	switch sortMode {
	case SoftwareSortX:
		return i4_3_0, true, true // sort by x
	case SoftwareSortY:
		return i4_3_1, true, true // sort by y
	case SoftwareSortZ:
		return i4_3_2, false, true // sort by z
	case SoftwareSortZNegative:
		return i4_3_2, true, true // sort by z
	}
	// TODO: SoftwareSortCommand: Sort by the computed drawCommand.command
	return 0, false, false
}

// Encodes a float into a uint32 whose unsigned order matches the order of cmp.Compare (NaN first, -0 equal to +0)
func sortableFloatBits(f float32) uint32 {
	if f != f {
		return 0 // NaN sorts before everything. Note: No other float encodes to 0
	}
	if f == 0 {
		f = 0 // Treat -0 the same as +0
	}

	bits := math.Float32bits(f)
	if bits&(1<<31) != 0 {
		return ^bits // Negative: Flip all bits, so that larger magnitudes sort first
	}
	return bits | (1 << 31) // Positive: Set the sign bit, so that positives sort after negatives
}
