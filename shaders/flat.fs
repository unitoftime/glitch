#version 300 es

// Required for webgl
#ifdef GL_ES
precision highp float;
#endif

/* struct Material { */
/*    vec3 ambient; */
/*    vec3 diffuse; */
/*    vec3 specular; */
/*    float shininess; */
/* }; */

/* struct DirLight { */
/*    vec3 direction; */
/*    vec3 ambient; */
/*    vec3 diffuse; */
/*    vec3 specular; */
/* }; */

out vec4 FragColor;

in vec3 FragPos;
/* in vec3 Normal; */
in vec2 TexCoord;
/* in vec4 FragPosLightSpace; */

/* uniform vec3 viewPos; */

/* uniform Material material; */

/* uniform DirLight dirLight; */

uniform sampler2D tex;

void main()
{
  FragColor = texture(tex, TexCoord);
  /* FragColor = vec4(1.0, 0.0, 0.0, 1.0); */
}
