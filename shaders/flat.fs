#version 300 es

// Required for webgl
#ifdef GL_ES
precision highp float;
#endif

struct Material {
   vec3 ambient;
   vec3 diffuse;
   vec3 specular;
   float shininess;
};

struct DirLight {
   vec3 direction;
   vec3 ambient;
   vec3 diffuse;
   vec3 specular;
};

out vec4 FragColor;

in vec3 FragPos;
in vec3 Normal;
in vec2 TexCoord;
/* in vec4 FragPosLightSpace; */

uniform vec3 viewPos;
uniform Material material;
uniform DirLight dirLight;

uniform sampler2D tex;

void main()
{
    // ambient
    vec3 ambient = dirLight.ambient * material.ambient;

    // diffuse
    vec3 norm = normalize(Normal);
    /* vec3 lightDir = normalize(light.position - FragPos); */ /* Calculate from position */
    vec3 lightDir = normalize(dirLight.direction);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = dirLight.diffuse * (diff * material.diffuse);

    // specular
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), material.shininess);
    vec3 specular = dirLight.specular * (spec * material.specular);

    vec3 result = ambient + diffuse + specular;
    FragColor = vec4(result, 1.0);

  /* TODO re-enable texture */
  /* FragColor = texture(tex, TexCoord); */
}
