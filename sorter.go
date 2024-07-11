package glitch

import (
	"cmp"
	"slices"
)

// you were here creating the sorter
// 1. every draw command needs ALL of the data to be sorted
// 2. you need to simplify the sorting to its all command based imo
type Sorter struct {
	DepthTest bool
	SoftwareSort SoftwareSortMode
	DepthBump bool
	depthBump float32
	currentLayer int8

	// States that are used for forming the draw command
	// blendMode BlendMode
	// shader *Shader
	// material Material

	// camera *CameraOrtho

	commands []cmdList
}

func NewSorter() *Sorter {
	return &Sorter{
		commands: make([]cmdList, 256), // TODO - hardcoding from sizeof(uint8)
	}
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
	s.sort()

	if s.DepthTest {
		// Opaque goes front to back (0 to 255)
		for l := range s.commands {
			for i := range s.commands[l].Opaque {
				// fmt.Println("- Opaque: (layer, x, z)", l, s.commands[l].Opaque[i].matrix[i4_3_1], s.commands[l].Opaque[i].matrix[i4_3_2])
				s.applyDrawCommand(target, s.commands[l].Opaque[i])
			}
		}

		// Translucent goes from back to front (255 to 0)
		for l := len(s.commands)-1; l >= 0; l-- { // Reverse order so that layer 0 is drawn last
			for i := range s.commands[l].Translucent {
			// for i := len(s.commands[l].Translucent)-1; i >= 0; i-- {
				// fmt.Println("- Transl: (layer, x, z)", l, s.commands[l].Translucent[i].matrix[i4_3_1], s.commands[l].Translucent[i].matrix[i4_3_2])
				s.applyDrawCommand(target, s.commands[l].Translucent[i])
			}
		}
	} else {
		for l := len(s.commands)-1; l >= 0; l-- { // Reverse order so that layer 0 is drawn last
			for i := range s.commands[l].Opaque {
				s.applyDrawCommand(target, s.commands[l].Opaque[i])
			}
			for i := range s.commands[l].Translucent {
				s.applyDrawCommand(target, s.commands[l].Translucent[i])
			}
		}
	}

	s.Clear()
}

func (s *Sorter) applyDrawCommand(target BatchTarget, c drawCommand) {
	target.Add(c.filler, c.matrix, c.mask, c.material, false)
}

func (s *Sorter) Add(filler GeometryFiller, mat glMat4, mask RGBA, material Material, translucent bool) {
	if mask.A == 0 { return } // discard b/c its completely transparent

	if mask.A != 1 {
		translucent = true
	}

	// TODO: If depthtest if false, I want to ensure that the drawCalls are sorted as if they are translucent
	if !s.DepthTest {
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
	}

	// state := BufferState{materialGroup{s.material, material}, s.blendMode}

	// if s.camera != nil {
	// 	material.camera = s.camera
	// }
	if s.DepthTest {
		material.depth = DepthModeLess
	}

	s.commands[s.currentLayer].Add(translucent, drawCommand{
		filler, mat, mask, material,
	})
}

func (s *Sorter) sort() {
	if s.DepthTest {
		// TODO - do special sort function for depth test code:
		// 1. Fully Opaque or fully transparent groups of meshes: Don't sort inside that group
		// 2. Partially transparent groups of meshes: sort inside that group
		// 3. Take into account blendMode

		// Sort translucent buffer
		for l := range s.commands {
			s.commands[l].SortTranslucent(s.SoftwareSort)
		}

		return
	}

	for l := range s.commands {
		s.commands[l].SortTranslucent(s.SoftwareSort)
	}

	for l := range s.commands {
		s.commands[l].SortOpaque(s.SoftwareSort)
	}
}

//--------------------------------------------------------------------------------
type cmdList struct{
	Opaque []drawCommand // TODO: This is kindof more like the unsorted list
	Translucent []drawCommand // TODO: This is kindof more like the sorted list

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

type SoftwareSortMode uint8
const (
	SoftwareSortNone SoftwareSortMode = iota
	SoftwareSortX // Sort based on the X position
	SoftwareSortY // Sort based on the Y position
	SoftwareSortZ // Sort based on the Z position
	SoftwareSortCommand // Sort by the computed drawCommand.command
)

type drawCommand struct {
	filler GeometryFiller
	matrix glMat4
	mask RGBA
	material Material
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
			return cmp.Compare(a.matrix[i4_3_2], b.matrix[i4_3_2]) // sort by z
		})
	}//  else if sortMode == SoftwareSortCommand {
	// 	slices.SortStableFunc(buf, func(a, b drawCommand) int {
	// 		return -cmp.Compare(a.command, b.command) // sort by command
	// 	})
	// }
}
