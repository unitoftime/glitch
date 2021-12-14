package main

import (
	"embed"
	"image"
	"image/draw"
	_ "image/png"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/unitoftime/glitch"
	"github.com/unitoftime/glitch/shaders"
)

//go:embed man.png
var f embed.FS
func loadImage(path string) (*image.NRGBA, error) {
	file, err := f.Open(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
	return nrgba, nil
}

func main() {
	glitch.Run(run2)
}

func run2() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch", glitch.WindowConfig{
		Vsync: true,
	})
	if err != nil { panic(err) }

	attrFmt := glitch.VertexFormat{
		glitch.Attrib{"aPos", glitch.AttrVec3},
		glitch.Attrib{"aColor", glitch.AttrVec3},
		glitch.Attrib{"aTexCoord", glitch.AttrVec2},
	}
	uniformFmt := glitch.UniformFormat{
		glitch.Attrib{"projection", glitch.AttrMat4},
		glitch.Attrib{"transform", glitch.AttrMat4},
	}
	shader, err := glitch.NewShader(shaders.SpriteVertexShader, shaders.SpriteFragmentShader, attrFmt, uniformFmt)
	if err != nil { panic(err) }

	shader.Bind()
	identMat := mgl32.Ident4()
	shader.SetUniform("transform", identMat)

	projMat := mgl32.Ortho2D(0, float32(1920), 0, float32(1080))
	shader.SetUniform("projection", projMat)

	pass := glitch.NewRenderPass(win, shader)

	manImage, err := loadImage("man.png")
	if err != nil {
		panic(err)
	}
	texture := glitch.NewTexture(160, 200, manImage.Pix)
	pass.SetTexture(0, texture)

	mesh := glitch.NewQuadMesh()

	for !win.ShouldClose() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}

		pass.Clear()
		// pass.Draw(mesh)
		mat := glitch.Mat4Ident
		mat = *mat.Scale(100).TranslateX(50).TranslateY(50)
		pass.Draw(mesh, &mat)

		glitch.Clear(glitch.RGBA{0.1, 0.2, 0.3, 1.0})
		pass.Execute()

		win.Update()
	}

	// glitch.SetTarget(win)
	// glitch.SetShader(shader)
	// {
	// 	identMat := mgl32.Mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	// 	shader.SetUniform("transform", identMat)

	// 	projMat := mgl32.Ortho2D(0, float32(1920), 0, float32(1080))
	// 	shader.SetUniform("projection", projMat)
	// }

	// manImage, err := loadImage("man.png")
	// if err != nil {
	// 	panic(err)
	// }
	// texture := glitch.NewTexture(160, 200, manImage.Pix)
	// mesh := glitch.NewQuadMesh()
	// glitch.Draw(mesh)

	// for !win.ShouldClose() {
	// 	if win.Pressed(glitch.KeyBackspace) {
	// 		win.Close()
	// 	}

	// 	glitch.Clear(glitch.RGBA{0.1, 0.2, 0.3, 1.0})

	// 	glitch.SetTexture(0, texture)

	// 	glitch.FinalizeDraw()

	// 	win.Update()
	// }
}

// func run() {
// 	win, err := glitch.NewWindow(1920, 1080, "Glitch", glitch.WindowConfig{
// 		Vsync: true,
// 	})
// 	if err != nil { panic(err) }

// 	attrFmt := glitch.VertexFormat{
// 		glitch.Attrib{"aPos", glitch.AttrVec3},
// 		glitch.Attrib{"aColor", glitch.AttrVec3},
// 		glitch.Attrib{"aTexCoord", glitch.AttrVec2},
// 	}
// 	uniformFmt := glitch.AttributeFormat{
// 		glitch.Attrib{"projection", glitch.AttrMat4},
// 		glitch.Attrib{"transform", glitch.AttrMat4},
// 	}
// 	shader, err := glitch.NewShader(vertexSource, fragmentSource, attrFmt, uniformFmt)

// 	if err != nil { panic(err) }
// 	shader.Bind()

// 	identMat := mgl32.Mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
// 	shader.SetUniform("transform", identMat)

// 	projMat := mgl32.Ortho2D(0, float32(1920), 0, float32(1080))
// 	shader.SetUniform("projection", projMat)

// 	batch := glitch.NewVertexBuffer(shader, 1000, 1000)

// 	// // w := float32(160.0)/4
// 	// // h := float32(200.0)/4
// 	// // x := float32(100)
// 	// // y := float32(100)
// 	// w := float32(160)
// 	// h := float32(200)
// 	// x := float32(50)
// 	// y := float32(50)
// 	// R := float32(1)
// 	// G := float32(1)
// 	// B := float32(1)
// 	// batch.Add([]float32{
// 	// 	// positions       // colors           // texture coords
// 	// 	x+w	,  y+h, 0.0,   R, G, B,   1.0, 0.0, // top right
// 	// 	x+w	,  y+0, 0.0,   R, G, B,   1.0, 1.0, // bottom right
// 	// 	x+0	,  y+0, 0.0,   R, G, B,   0.0, 1.0, // bottom left
// 	// 	x+0	,  y+h, 0.0,   R, G, B,   0.0, 0.0,  // top left
// 	// },
// 	// 	[]uint32{
// 	// 		0, 1, 3, // first triangle
// 	// 		1, 2, 3,  // second triangle
// 	// 	})
// 	// // fmt.Println(currentElement, x, y, w, h, len(b.vertices), len(b.indices))

// 	manImage, err := loadImage("man.png")
// 	if err != nil {
// 		panic(err)
// 	}
// 	texture := glitch.NewTexture(160, 200, manImage.Pix)

// 	sprite := glitch.NewSprite(texture, glitch.R(0, 0, 160, 200))

// 	sprite.Draw(batch, 50, 50)
// 	// batch.Add(sprite)

// 	for !win.ShouldClose() {
// 		if win.Pressed(glitch.KeyBackspace) {
// 			win.Close()
// 		}

// 		glitch.Clear(glitch.RGBA{0.1, 0.2, 0.3, 1.0})

// 		texture.Bind(0)
// 		batch.Bind()
// 		batch.Draw()

// 		win.Update()
// 	}


// 	// // Interface draft 1:
// 	// // Load a program
// 	// // Load a bunch of uniform inputs
// 	// // Load a bunch of geometry inputs
// 	// // Execute
// 	// loader := glitch.NewLoader("./")
// 	// mesh, err := glitch.NewModel(loader.Model("model.gltf"))
// 	// texture := glitch.NewTexture(loader.Texture("man.png"))
// 	// shader := glitch.NewShader(vSrc, fSrc, fmt)

// 	// for {
// 	// 	// glitch.BeginDrawing() // Why? to capture previous gpu states?

// 	// 	glitch.SetTarget(win)
// 	// 	win.Clear(rgba)
// 	// 	glitch.SetShader(shader) // Creates a shader context, then when you draw things you have to draw to a certain vertex specification. So then all the models/meshes/geoms that get drawn, we pull data out in that format! And puts them into vertexbuffers based on some batching rules
// 	// 	glitch.SetCamera(camera)
// 	// 	glitch.SetMaterial(material) // Sets a group of uniforms
// 	// 	mesh.Draw(matrix) // draw based on shader layout?

// 	// 	win.Update()
// 	// 	// glitch.EndDrawing()
// 	// }

// // 	// Interface draft 2:
// // 	// Load a program
// // 	// Load a bunch of uniform inputs
// // 	// Load a bunch of geometry inputs
// // 	// Execute
// // 	loader := glitch.NewLoader("./")
// // 	mesh, err := glitch.NewModel(loader.Model("model.gltf"))
// // 	texture := glitch.NewTexture(loader.Texture("man.png")) // TODO - some way to do mipmaps
// // 	shader := glitch.NewShader(vSrc, fSrc, fmt)

// // 	for {
// // 		// glitch.BeginDrawing() // Why? to capture previous gpu states?

// // 		glitch.SetTarget(win)
// // 		win.Clear(rgba)
// // 		glitch.SetShader(shader) // Creates a shader context, then when you draw things you have to draw to a certain vertex specification. So then all the models/meshes/geoms that get drawn, we pull data out in that format! And puts them into vertexbuffers based on some batching rules
// // 		glitch.SetTexture(0, texture)

// // 		glitch.SetCamera(camera) // More like set uniform?

// // 		glitch.SetMaterial(material) // Sets a group of uniforms
// // 		mesh.Draw(matrix) // draw based on shader layout? Matrix to transform positions, color matrix to transform colors, do I need anything else?

// // /* https://learnopengl.com/Advanced-OpenGL/Blending
// // Blending: Draw in this order
// // 1. Draw all opaque objects first.
// // 2. Sort all the transparent objects.
// // 3. Draw all the transparent objects in sorted order.
// // OR DO: https://learnopengl.com/Guest-Articles/2020/OIT/Introduction
// // */

// // 		win.Update()
// // 		// glitch.EndDrawing()
// // 	}
// }

// const (
// 	vertexSource = `#version 300 es

// layout (location = 0) in vec3 aPos;
// layout (location = 1) in vec3 aColor;
// layout (location = 2) in vec2 aTexCoord;

// out vec3 ourColor;
// out vec2 TexCoord;

// uniform mat4 projection;
// uniform mat4 transform;

// void main()
// {
// 	gl_Position = projection * transform * vec4(aPos, 1.0);
// //	gl_Position = vec4(aPos, 1.0);
// 	ourColor = aColor;
// 	TexCoord = vec2(aTexCoord.x, aTexCoord.y);
// }
// `
// 	fragmentSource = `#version 300 es
// // Required for webgl
// #ifdef GL_ES
// precision highp float;
// #endif

// out vec4 FragColor;

// in vec3 ourColor;
// in vec2 TexCoord;

// //texture samplers
// uniform sampler2D texture1;

// void main()
// {
// 	// linearly interpolate between both textures (80% container, 20% awesomeface)
// 	//FragColor = mix(texture(texture1, TexCoord), texture(texture2, TexCoord), 0.2);
//   FragColor = vec4(ourColor, 1.0) * texture(texture1, TexCoord);
// //  FragColor = vec4(ourColor, 1.0);
// }
// `
// )
