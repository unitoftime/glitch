#version 300 es

// Required for webgl
#ifdef GL_ES
precision highp float;
#endif

out vec4 FragColor;

in vec4 ourColor;
in vec2 TexCoord;

/* // View matrix uniform */
/* uniform mat4 view; */

//texture samplers
uniform sampler2D texture1;

float normpdf(float x, float sigma) {
  return 0.39894*exp(-0.5*x*x/(sigma*sigma))/sigma;
}

void main()
{
  /* /\* float s = 16.0; *\/ */
  /* /\* float sHalf = 8.0; *\/ */
  /* /\* vec4 color = texture(texture1, (floor(TexCoord * s)+sHalf)/s); *\/ */
  /* /\* vec4 color = texture(texture1, (floor(TexCoord * s)+.5)/s); *\/ */
  /* vec4 color = texture(texture1, TexCoord); */

  /* /\* // Reduce color range *\/ */
  /* /\* float range = 3.0; *\/ */
  /* /\* color = floor(color * range) / range; *\/ */

  /* if (color.a == 0.0) { */
  /*   discard; */
  /* } */

  /* FragColor = ourColor * color; */

  vec2 resolution = vec2(2048.0, 2048.0);
  vec3 c = texture(texture1, TexCoord.xy / resolution.xy).rgb;

  //declare stuff
  const int mSize = 11;
  const int kSize = (mSize-1)/2;
  float kernel[mSize];
  vec3 final_colour = vec3(0.0);

  //create the 1-D kernel
  float sigma = 7.0;
  float Z = 0.0;
  for (int j = 0; j <= kSize; ++j) {
    kernel[kSize+j] = kernel[kSize-j] = normpdf(float(j), sigma);
  }

  //get the normalization factor (as the gaussian has been clamped)
  for (int j = 0; j < mSize; ++j) {
    Z += kernel[j];
  }

  //read out the texels
  for (int i=-kSize; i <= kSize; ++i) {
    for (int j=-kSize; j <= kSize; ++j) {
      final_colour += kernel[kSize+j]*kernel[kSize+i]*texture(texture1, (TexCoord.xy+vec2(float(i),float(j))) / resolution.xy).rgb;
    }
  }

  FragColor = ourColor * vec4(final_colour/(Z*Z), 1.0);
}
