#version 100
// Spec: https://registry.khronos.org/OpenGL/specs/es/2.0/GLSL_ES_Specification_1.00.pdf

// Required for webgl - TODO - is this needed for version 100?
precision highp float;

varying vec4 ourColor;
varying vec2 TexCoord;

//vec4 FragColor;

//texture samplers
uniform sampler2D texture1;

void main()
{
  vec4 tex = texture2D(texture1, TexCoord);
  if (tex.a == 0.0) {
    discard;
  }
  // linearly interpolate between both textures (80% container, 20% awesomeface)
  //FragColor = mix(texture(texture1, TexCoord), texture(texture2, TexCoord), 0.2);
  gl_FragColor = ourColor * tex;
  //  FragColor = vec4(ourColor, 1.0) * texture(texture1, TexCoord);
  //  FragColor = vec4(ourColor, 1.0);
}
