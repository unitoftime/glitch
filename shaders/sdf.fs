#version 300 es

// Required for webgl
#ifdef GL_ES
precision highp float;
#endif

out vec4 FragColor;

in vec4 ourColor;
in vec2 TexCoord;

uniform sampler2D texture1;

/* uniform vec4 bgColor; */
/* uniform vec4 fgColor; */

void main() {
  float width = 0.4;
  float edge = 0.1;

  float borderWidth = 0.5;
  float borderEdge = 0.1;

  vec3 bgColor = vec3(0, 0, 0);
  vec3 fgColor = vec3(1, 1, 1);

  float sd = 1.0 - texture(texture1, TexCoord).r;
  float alpha = 1.0 - smoothstep(width, width + edge, sd);
  float borderAlpha = 1.0 - smoothstep(borderWidth, borderWidth + borderEdge, sd);

  float overallAlpha = alpha + (1.0 - alpha) * borderAlpha;
  vec3 mixColor = mix(bgColor, fgColor, alpha / overallAlpha);

  FragColor = vec4(mixColor, overallAlpha);
}
