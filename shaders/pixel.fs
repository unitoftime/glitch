#version 300 es

// Required for webgl
#ifdef GL_ES
precision highp float;
#endif

out vec4 FragColor;

in vec4 ourColor;
in vec2 TexCoord;

//texture samplers
uniform sampler2D texture1;

void main()
{
  vec4 tex = texture(texture1, TexCoord);
  if (tex.a == 0.0) {
    discard;
  }
  // linearly interpolate between both textures (80% container, 20% awesomeface)
  //FragColor = mix(texture(texture1, TexCoord), texture(texture2, TexCoord), 0.2);
  FragColor = ourColor * tex;
  //  FragColor = vec4(ourColor, 1.0) * texture(texture1, TexCoord);
  //  FragColor = vec4(ourColor, 1.0);
}
