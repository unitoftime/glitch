#version 300 es

layout (location = 0) in vec3 position;
layout (location = 1) in vec4 color;
/* layout (location = 1) in vec3 normal; */
layout (location = 2) in vec2 texture;

/* out vec3 Normal; */
/* out vec3 FragPos; */
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
   TexCoord = texture;
   /* FragPos = vec3(model * vec4(position, 1.0)); */
   /* Normal = mat3(transpose(inverse(model))) * normal; // DIdn't really understand this */
   /* FragPosLightSpace = shadowMatrix * vec4(FragPos, 1.0); */
   /* gl_Position = projection * view * model * vec4(position, 1.0f); */
   /* FragPos = vec3(model * vec4(position, 1.0)); */

   gl_Position = projection * view * model * vec4(position, 1.0f);
}
