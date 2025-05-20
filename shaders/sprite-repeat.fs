#version 300 es

// Required for webgl
#ifdef GL_ES
precision highp float;
#endif

out vec4 FragColor;

in vec4 ourColor;
in vec2 TexCoord;

uniform float iTime;
uniform vec2 zoom;
uniform vec4 repeatRect; // Note: This is in UV space

//texture samplers
uniform sampler2D texture1;

void main()
{
  // Repeat
  vec2 rMin = vec2(repeatRect.x, repeatRect.y);
  vec2 rMax = vec2(repeatRect.z, repeatRect.w);
  vec2 rSize = rMax - rMin;
  vec2 rCoord = rMin + mod(TexCoord * zoom, rSize);

  vec4 tex = texture(texture1, rCoord);
  if (tex.a == 0.0) {
    discard;
  }

  FragColor = ourColor * tex;
}
