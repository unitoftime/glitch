#version 300 es

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

struct PointLight {
   vec3 position;
   float constant;
   float linear;
   float quadratic;
   vec3 ambient;
   vec3 diffuse;
   vec3 specular;
};

struct SpotLight {
    vec3 position;
    vec3 direction;
    float cutOff;
    float outerCutOff;
    float constant;
    float linear;
    float quadratic;
    vec3 ambient;
    vec3 diffuse;
    vec3 specular;
};

out vec4 FragColor;
in vec3 FragPos;
in vec3 Normal;
in vec2 TexCoord;
in vec4 FragPosLightSpace;

uniform vec3 viewPos;
uniform Material material;
uniform DirLight dirLight;
#define NR_POINT_LIGHTS 1
uniform PointLight pointLights[NR_POINT_LIGHTS];
uniform sampler2D tex;
uniform sampler2D shadowMap;

vec3 CalcDirLight(DirLight light, vec3 normal, vec3 viewDir);
vec3 CalcPointLight(PointLight light, vec3 normal, vec3 fragPos, vec3 viewDir);
vec3 CalcSpotLight(SpotLight light, vec3 normal, vec3 fragPos, vec3 viewDir);

void main()
{
//   FragColor = vec4(0.0, 1.0, 0.0, 1.0);
//   FragColor = vec4(ourColor, 1.0);
//   FragColor = texture(tex, TexCoord) * vec4(ourColor, 1.0);
//   FragColor = texture(tex, TexCoord);
//    // ambient
//    vec3 ambient = light.ambient * material.ambient;
//    // diffuse
//    vec3 norm = normalize(Normal);
//    vec3 lightDir = normalize(light.position - FragPos);
//    float diff = max(dot(norm, lightDir), 0.0);
//    vec3 diffuse = light.diffuse * (diff * material.diffuse);
//    // specular
//    vec3 viewDir = normalize(viewPos - FragPos);
//    vec3 reflectDir = reflect(-lightDir, norm);
//    float spec = pow(max(dot(viewDir, reflectDir), 0.0), material.shininess);
//    vec3 specular = light.specular * (spec * material.specular);
// //   vec3 result = texture(tex, TexCoord).xyz;
//    vec3 result = ambient + diffuse + specular;
   // Calculate a few properties
   vec3 norm = normalize(Normal);
   vec3 viewDir = normalize(viewPos - FragPos);
   // Calculate Directional Lighting
   vec3 result = CalcDirLight(dirLight, norm, viewDir);
   // Calculate Point Lighting
   //for(int i = 0; i < NR_POINT_LIGHTS; i++)
   //   result += normalize(CalcPointLight(pointLights[i], norm, FragPos, viewDir));
   // Calculate Spot Lighting
//   result += CalcSpotLight(spotLight, norm, FragPos, viewDir);
   vec3 color = result * vec3(0.2, 0.2, 0.2);
   FragColor = vec4(color, 1.0);
// Linearize Depth Buffer:
   // float near = 0.1;
   // float far  = 100.0;
   // float depth = gl_FragCoord.z;
   // float z = depth * 2.0 - 1.0; // back to NDC
   // depth = (2.0 * near * far) / (far + near - z * (far - near));
   // depth = depth / far; // divide by far for demonstration
   // FragColor += vec4(vec3(depth), 1.0);
// Display Depth Buffer:   FragColor += vec4(vec3(gl_FragCoord.z), 1.0);
   //FragColor += vec4(0.3, 0.3, 0.3, 0.0);
   //FragColor += texture(tex, TexCoord);
//   FragColor *= texture(tex, TexCoord);
//   FragColor = vec4(FragColor.rgb, 1.0);
}
float ShadowCalculation(vec4 fragPosLightSpace)
{
    // perform perspective divide
    vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
    // transform to [0,1] range
    projCoords = projCoords * 0.5 + 0.5;
    // get closest depth value from light's perspective (using [0,1] range fragPosLight as coords)
    float closestDepth = texture(shadowMap, projCoords.xy).r;
    // get depth of current fragment from light's perspective
    float currentDepth = projCoords.z;
    // check whether current frag pos is in shadow
    float shadow = currentDepth > closestDepth  ? 1.0 : 0.0;
    return shadow;
}
vec3 CalcDirLight(DirLight light, vec3 normal, vec3 viewDir)
{
   vec3 lightDir = normalize(-light.direction);
//   vec3 lightDir = normalize(light.direction - FragPos);
   // diffuse shading
   float diff = max(dot(normal, lightDir), 0.0);
   // specular shading
   vec3 reflectDir = reflect(-lightDir, normal);
   float spec = pow(max(dot(viewDir, reflectDir), 0.0), material.shininess);
   // combine results
// For Textures:   vec3 ambient  = light.ambient  * vec3(texture(material.diffuse, TexCoords));
// For Textures:   vec3 diffuse  = light.diffuse  * diff * vec3(texture(material.diffuse, TexCoords));
// For SpecMaps:   vec3 specular = light.specular * spec * vec3(texture(material.specular, TexCoords));
   vec3 ambient = light.ambient * material.ambient;
   vec3 diffuse = light.diffuse * (diff * material.diffuse);
   vec3 specular = light.specular * (spec * material.specular);
//old way:   return (ambient + diffuse + specular);
// with shadows:
   float shadow = ShadowCalculation(FragPosLightSpace);
   return (ambient + (1.0 - shadow) * (diffuse + specular));
}
// calculates the color when using a point light.
vec3 CalcPointLight(PointLight light, vec3 normal, vec3 fragPos, vec3 viewDir)
{
    vec3 lightDir = normalize(light.position - fragPos);
    // diffuse shading
    float diff = max(dot(normal, lightDir), 0.0);
    // specular shading
    vec3 reflectDir = reflect(-lightDir, normal);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), material.shininess);
    // attenuation
    float distance = length(light.position - fragPos);
    float attenuation = 1.0 / (light.constant + light.linear * distance + light.quadratic * (distance * distance));
    // combine results
// For Textures:    vec3 ambient = light.ambient * vec3(texture(material.diffuse, TexCoords));
// For Textures:    vec3 diffuse = light.diffuse * diff * vec3(texture(material.diffuse, TexCoords));
// For SpecMaps:    vec3 specular = light.specular * spec * vec3(texture(material.specular, TexCoords));
   vec3 ambient = light.ambient * material.ambient;
   vec3 diffuse = light.diffuse * (diff * material.diffuse);
   vec3 specular = light.specular * (spec * material.specular);
    ambient *= attenuation;
    diffuse *= attenuation;
    specular *= attenuation;
//    return (ambient + diffuse + specular);
   vec3 res = 0.01 * (ambient + diffuse + specular);
   return res;
}
// calculates the color when using a spot light.
vec3 CalcSpotLight(SpotLight light, vec3 normal, vec3 fragPos, vec3 viewDir)
{
    vec3 lightDir = normalize(light.position - fragPos);
    // diffuse shading
    float diff = max(dot(normal, lightDir), 0.0);
    // specular shading
    vec3 reflectDir = reflect(-lightDir, normal);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), material.shininess);
    // attenuation
    float distance = length(light.position - fragPos);
    float attenuation = 1.0 / (light.constant + light.linear * distance + light.quadratic * (distance * distance));
    // spotlight intensity
    float theta = dot(lightDir, normalize(-light.direction));
    float epsilon = light.cutOff - light.outerCutOff;
    float intensity = clamp((theta - light.outerCutOff) / epsilon, 0.0, 1.0);
    // combine results
// For Textures:    vec3 ambient = light.ambient * vec3(texture(material.diffuse, TexCoords));
// For Textures:    vec3 diffuse = light.diffuse * diff * vec3(texture(material.diffuse, TexCoords));
// For SpecMaps:    vec3 specular = light.specular * spec * vec3(texture(material.specular, TexCoords));
   vec3 ambient = light.ambient * material.ambient;
   vec3 diffuse = light.diffuse * (diff * material.diffuse);
   vec3 specular = light.specular * (spec * material.specular);
    ambient *= attenuation * intensity;
    diffuse *= attenuation * intensity;
    specular *= attenuation * intensity;
    return (ambient + diffuse + specular);
}

// #version 300 es
// // Required for webgl
// #ifdef GL_ES
// precision highp float;
// #endif

// out vec4 FragColor;

// in vec4 ourColor;
// in vec2 TexCoord;

// //texture samplers
// uniform sampler2D texture1;

// void main()
// {
//   vec4 tex = texture(texture1, TexCoord);

//   // linearly interpolate between both textures (80% container, 20% awesomeface)
//   //FragColor = mix(texture(texture1, TexCoord), texture(texture2, TexCoord), 0.2);
//   FragColor = ourColor * tex;
//   //  FragColor = vec4(ourColor, 1.0) * texture(texture1, TexCoord);
//   //  FragColor = vec4(ourColor, 1.0);
// }
