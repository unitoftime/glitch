#version 100
// Spec: https://registry.khronos.org/OpenGL/specs/es/2.0/GLSL_ES_Specification_1.00.pdf

attribute vec2 positionIn;
attribute vec4 colorIn;
attribute vec2 texCoordIn;

varying vec4 ourColor;
varying vec2 TexCoord;

// uniform mat4 model;
uniform mat4 projection;
uniform mat4 view;
//uniform mat4 transform;

void main()
{
  gl_Position = projection * view * vec4(positionIn, 0.0, 1.0);
  //	gl_Position = projection * transform * vec4(aPos, 1.0);
  //	gl_Position = vec4(aPos, 1.0);
  ourColor = colorIn;

  TexCoord = vec2(texCoordIn.x, texCoordIn.y);
}
