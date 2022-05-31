#version 300 es

layout (location = 0) in vec3 positionIn;
/* layout (location = 1) in vec4 colorIn; */
layout (location = 1) in vec3 normalIn;
layout (location = 2) in vec2 texCoordIn;

out vec3 FragPos;
out vec3 Normal;
out vec2 TexCoord;
//out vec4 FragPosLightSpace;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
// uniform vec3 viewPos;

//uniform mat4 shadowMatrix;

void main()
{
   /* TexCoord = vec2(aTexCoord.x, 1.0 - aTexCoord.y); */
   /* FragPosLightSpace = shadowMatrix * vec4(FragPos, 1.0); */

   Normal = mat3(transpose(inverse(model))) * normalIn; // I didn't really understand this
   FragPos = vec3(model * vec4(positionIn, 1.0f));
   TexCoord = texCoordIn;

   gl_Position = projection * view * model * vec4(positionIn, 1.0f);
}
