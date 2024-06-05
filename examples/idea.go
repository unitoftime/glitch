package main

/*

// Sorting
// target | shader | translucency type | material | depth | mesh

func idea() {
	// Create a window, which implements some 'target' interface (ie something that can be rendered to)
	win, err := glitch.NewWindow(1920, 1080, "Glitch", glitch.WindowConfig{
		Vsync: true,
	})
	if err != nil { panic(err) }

	// Create a shader program with specified attributes and uniforms
	attrFmt := glitch.VertexFormat{
		glitch.Attrib{"aPos", glitch.AttrVec3},
		glitch.Attrib{"aColor", glitch.AttrVec3},
		glitch.Attrib{"aTexCoord", glitch.AttrVec2},
	}
	uniformFmt := glitch.AttributeFormat{
		glitch.Attrib{"projection", glitch.AttrMat4},
		glitch.Attrib{"transform", glitch.AttrMat4},
	}
	shader, err := glitch.NewShader(vertexSource, fragmentSource, attrFmt, uniformFmt)
	if err != nil { panic(err) }

	// Load Textures
	manImage, err := loadImage("man.png")
	if err != nil {
		panic(err)
	}
	texture := glitch.NewTexture(160, 200, manImage.Pix)

	// Create Meshes
	// meshData, err := loadGltf("man.gltf")
	// mesh := glitch.NewMesh(meshData)
	mesh := glitch.NewMesh(geometry.Quad())

	// Create Text
	// Any way to combine this into quad and stuff like that? Or is it different enough?
	// fontData, err := loadText("font.ttf")
	// glitch.NewFont(fontData)

	// Draw stuff
	// Option 1 - more stateful
	glitch.SetTarget(win)
	glitch.SetShader(shader)
	glitch.SetUniform("transform", identMat)
	glitch.SetUniform("projection", identMat)

	for !win.Closed() {

		glitch.Draw(mesh, matrix) // Draws but transforms verts via matrix

		matrix := identMatrix
		glitch.DrawColorMask(mesh, matrix, color) // Draws but transforms verts via matrix and colors via color
		// What if people want to modify other vertex attributes?

		// Essentially
		{
			// Setup shader params
			glitch.SetUniform("transform", identMat)
			glitch.SetUniform("projection", identMat)
			// pass in data into that same context
			glitch.Draw(mesh, matrix)
			// 1. extra parameters let you modify a few special attributes
		}

		{
			glitch.SetShader(2dShader)
			glitch.SetUniform("myUniform", uniformValue)

			glitch.SetTarget(win)
			glitch.Draw(mesh, matrix)

			glitch.SetTarget(myFramebuffer)
			glitch.Draw(mesh, matrix)
		}

		{
			glitch.SetShader(textShader)
			glitch.SetTarget(win)
			glitch.Draw(mesh, matrix)
		}
	}

	// Option 2 - More object based? Shader being the main thing you draw against
	shader.SetTarget(win)
	shader.SetTexture(0, texture)
	shader.SetUniform("transform", mat)
	shader.SetUniform("projection", mat)
	shader.Draw(mesh, matrix)
	shader.Execute()


	// Option 3 - Pass Based
	pass := shader.NewRenderPass(win) // rendering to a target

	pass.Clear()
	pass.SetTexture(0, texture)
	pass.SetUniform("transform", mat)
	pass.SetUniform("projection", mat)
	pass.Draw(mesh, matrix)
	pass.Execute()
	pass.Execute(win, )

	pass.Clear()
	mesh.Draw(pass, matrix)

	glitch.SetTarget(win)
	glitch.SetShader(shader)
	{
		identMat := mgl32.Mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
		shader.SetUniform("transform", identMat)

		projMat := mgl32.Ortho2D(0, float32(1920), 0, float32(1080))
		shader.SetUniform("projection", projMat)
	}

	for !win.ShouldClose() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}

		glitch.Clear(glitch.RGBA{0.1, 0.2, 0.3, 1.0})

		glitch.SetTexture(0, texture)
		glitch.Draw(mesh)

		glitch.FinalizeDraw()

		win.Update() // SwapBuffers, PollEvents
	}
}
*/
