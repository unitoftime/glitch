package main

import (
	"embed"
	"image"
	"image/draw"
	_ "image/png"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jstewart7/glitch"
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
	glitch.Run(run)
}

func run() {
	win, err := glitch.NewWindow(1920, 1080, "Glitch", glitch.WindowConfig{
		Vsync: true,
	})
	if err != nil { panic(err) }

	shader, err := glitch.NewShader(vertexSource, fragmentSource, glitch.AttributeFormat{
		glitch.Attrib{"projection", glitch.Mat4},
		glitch.Attrib{"transform", glitch.Mat4},
	})
	if err != nil { panic(err) }
	shader.Bind()

	identMat := mgl32.Mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	shader.SetUniform("transform", identMat)

	projMat := mgl32.Ortho2D(0, float32(1920), 0, float32(1080))
	shader.SetUniform("projection", projMat)

	batch := glitch.NewVertexBuffer(1000, 1000, shader, glitch.VertexFormat{
		glitch.Attrib{"aPos", glitch.Vec3},
		glitch.Attrib{"aColor", glitch.Vec3},
		glitch.Attrib{"aTexCoord", glitch.Vec2},
	})

	// // w := float32(160.0)/4
	// // h := float32(200.0)/4
	// // x := float32(100)
	// // y := float32(100)
	// w := float32(160)
	// h := float32(200)
	// x := float32(50)
	// y := float32(50)
	// R := float32(1)
	// G := float32(1)
	// B := float32(1)
	// batch.Add([]float32{
	// 	// positions       // colors           // texture coords
	// 	x+w	,  y+h, 0.0,   R, G, B,   1.0, 0.0, // top right
	// 	x+w	,  y+0, 0.0,   R, G, B,   1.0, 1.0, // bottom right
	// 	x+0	,  y+0, 0.0,   R, G, B,   0.0, 1.0, // bottom left
	// 	x+0	,  y+h, 0.0,   R, G, B,   0.0, 0.0,  // top left
	// },
	// 	[]uint32{
	// 		0, 1, 3, // first triangle
	// 		1, 2, 3,  // second triangle
	// 	})
	// // fmt.Println(currentElement, x, y, w, h, len(b.vertices), len(b.indices))
	manImage, err := loadImage("man.png")
	if err != nil {
		panic(err)
	}
	texture := glitch.NewTexture(160, 200, manImage.Pix)

	sprite := glitch.NewSprite(texture, glitch.R(0, 0, 160, 200))

	sprite.Draw(batch, 50, 50)
	// batch.Add(sprite)

	for !win.ShouldClose() {
		if win.Pressed(glitch.KeyBackspace) {
			win.Close()
		}

		glitch.Clear(glitch.RGBA{0.1, 0.2, 0.3, 1.0})

		texture.Bind()
		batch.Bind()
		batch.Draw()

		win.Update()
	}
}

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
)
