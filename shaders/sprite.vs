#version 300 es

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec4 aColor;
layout (location = 2) in vec2 aTexCoord;

out vec4 ourColor;
out vec2 TexCoord;

uniform mat4 projection;
uniform mat4 view;
//uniform mat4 transform;

void main()
{
	gl_Position = projection * view * vec4(aPos, 1.0);
//	gl_Position = projection * transform * vec4(aPos, 1.0);
//	gl_Position = vec4(aPos, 1.0);
	ourColor = aColor;
	TexCoord = vec2(aTexCoord.x, aTexCoord.y);
}