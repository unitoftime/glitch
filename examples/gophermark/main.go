package main

import (
	"fmt"
	"embed"
	"image"
	"image/draw"
	_ "image/png"
	"time"
	"math/rand"

	"github.com/ungerik/go3d/vec3"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jstewart7/glitch"
	"github.com/jstewart7/glitch/shaders"
)

//go:embed gopher.png
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
	uniformFmt := glitch.AttributeFormat{
		glitch.Attrib{"projection", glitch.AttrMat4},
		glitch.Attrib{"transform", glitch.AttrMat4},
	}
	shader, err := glitch.NewShader(shaders.SpriteVertexShader, shaders.SpriteFragmentShader, attrFmt, uniformFmt)
	// shader, err := glitch.NewShader(vertexSource, fragmentSource, attrFmt, uniformFmt)
	if err != nil { panic(err) }

	shader.Bind()
	identMat := mgl32.Ident4()
	shader.SetUniform("transform", identMat)

	projMat := mgl32.Ortho2D(0, float32(1920), 0, float32(1080))
	shader.SetUniform("projection", projMat)

	pass := glitch.NewRenderPass(win, shader)

	manImage, err := loadImage("gopher.png")
	if err != nil {
		panic(err)
	}
	texture := glitch.NewTexture(160, 200, manImage.Pix)
	pass.SetTexture(0, texture)

	mesh := glitch.NewQuadMesh()

	length := 10000
	man := make([]Man, length)
	for i := range man {
		man[i] = NewMan()
	}

	w := float32(160.0)/4
	h := float32(200.0)/4
	manSize := &vec3.T{w, h, 1}

	start := time.Now()
	for !win.ShouldClose() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}
		start = time.Now()
		for i := range man {
			man[i].position[0] += man[i].velocity[0]
			man[i].position[1] += man[i].velocity[1]

			if man[i].position[0] <= 0 || (man[i].position[0]+w) >= float32(1920) {
				man[i].velocity[0] = -man[i].velocity[0]
			}
			if man[i].position[1] <= 0 || (man[i].position[1]+h) >= float32(1080) {
				man[i].velocity[1] = -man[i].velocity[1]
			}
		}

		fmt.Println("Clear")
		pass.Clear()
		for i := range man {
			mat := glitch.Mat4Ident().ScaleVec3(manSize).TranslateX(man[i].position[0]).TranslateY(man[i].position[1])
			pass.Draw(mesh, mat)
		}

		glitch.Clear(glitch.RGBA{0.1, 0.2, 0.3, 1.0})

		fmt.Println("Execute")
		pass.Execute()

		fmt.Println("Update")
		win.Update()

		fmt.Println("Clock")
		dt := time.Since(start)
		fmt.Println(dt.Seconds() * 1000)
	}
}

type Man struct {
	position, velocity mgl32.Vec2
	R, G, B float32
}
func NewMan() Man {
	vScale := 5.0
	return Man{
		// position: mgl32.Vec2{100, 100},
		// position: mgl32.Vec2{float32(float64(width/2) * rand.Float64()),
		// 	float32(float64(height/2) * rand.Float64())},
		position: mgl32.Vec2{1920/2, 1080/2},
		velocity: mgl32.Vec2{float32(2*vScale * (rand.Float64()-0.5)),
			float32(2*vScale * (rand.Float64()-0.5))},
		R: rand.Float32(),
		G: rand.Float32(),
		B: rand.Float32(),
	}
}

/*
const (
	vertexSource = `
#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aColor;
layout (location = 2) in vec2 aTexCoord;

out vec3 ourColor;
out vec2 TexCoord;

uniform mat4 projection;
uniform mat4 transform;

void main()
{
	gl_Position = projection * transform * vec4(aPos, 1.0);
//	gl_Position = vec4(aPos, 1.0);
	ourColor = aColor;
	TexCoord = vec2(aTexCoord.x, aTexCoord.y);
}
`
	fragmentSource = `
#version 330 core
out vec4 FragColor;

in vec3 ourColor;
in vec2 TexCoord;

//texture samplers
uniform sampler2D texture1;

void main()
{
	// linearly interpolate between both textures (80% container, 20% awesomeface)
	//FragColor = mix(texture(texture1, TexCoord), texture(texture2, TexCoord), 0.2);
  FragColor = vec4(ourColor, 1.0) * texture(texture1, TexCoord);
//  FragColor = vec4(ourColor, 1.0);
}
`
)*/
