#version 300 es
// Required for webgl
#ifdef GL_ES
precision highp float;
#endif

out vec4 FragColor;

in vec4 ourColor;
in vec2 TexCoord;

// View matrix uniform
uniform mat4 view;

//texture samplers
uniform sampler2D texture1;

void main()
{
  float s = 16.0;
  vec4 color = texture(texture1, (floor(TexCoord * s)+.5)/s);
  /* vec4 color = texture(texture1, TexCoord); */

  if (color.a == 0.0) {
    discard;
  }

  FragColor = ourColor * color;
}
