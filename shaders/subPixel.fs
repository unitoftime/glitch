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
  // --- For Pixel art games ---
  // TODO - This isn't working, I'm not sure why
  // Note: Ensure you enable Linear filtering (ie smooth) on textures
  // https://www.youtube.com/watch?v=Yu8k7a1hQuU - Pixel art rendering
  // Adapted from: https://www.shadertoy.com/view/MlB3D3
  // More reading: https://jorenjoestar.github.io/post/pixel_art_filtering/

  // Attempt 1 - This looks blurry
  /* vec2 textureSize2d = vec2(textureSize(texture1, 0)); */
  /* /\* vec2 textureSize2d = vec2(1024, 1024); *\/ */
  /* vec2 pixel = TexCoord * textureSize2d.xy; */

  /* // Calculate the scale of the view matrix (used for scaling the subpixels) */
  /* /\* vec2 scale = vec2(view[3][0]/4.0, view[3][1]/4.0); *\/ */
  /* /\* float scale = 1.0; *\/ */
  /* vec2 scale = vec2(length(vec3(view[0][0], view[0][1], view[0][2])), length(vec3(view[1][0], view[1][1], view[1][2]))); */

  /* /\* vec2 scale = vec2(5.0 * view[0][0], 5.0 * view[0][0]); *\/ */
  /* /\* vec2 scale = vec2(10.0 * view[0][0], 10.0 * view[0][0]); *\/ */


  /* /\* scale = scale * 0.5; // TODO - Magic number, this just seems to look good *\/ */

  /* // emulate point sampling */
  /* vec2 uv = floor(pixel) + 0.5; */
  /* // subpixel aa algorithm (COMMENT OUT TO COMPARE WITH POINT SAMPLING) */
  /* // TODO - This is shimmering, I'm not sure why, I think the scale is wrong */
  /* uv += 1.0 - clamp((1.0 - fract(pixel)) * scale, 0.0, 1.0); */

  /* // output */
  /* vec4 color = texture(texture1, uv / textureSize2d.xy); */

  // Attempt 2: https://colececil.io/blog/2017/scaling-pixel-art-without-destroying-it/
  // Note: This looks less blurry
  float texelsPerPixel = 100.0; // TODO - Not sure where to get this from
  vec2 textureSize2d = vec2(textureSize(texture1, 0));
  vec2 pixel = TexCoord * textureSize2d.xy;

  vec2 locationWithinTexel = fract(pixel);
  vec2 interpolationAmount = clamp(locationWithinTexel / texelsPerPixel, 0.0, 0.5) + clamp((locationWithinTexel - 1.0) / texelsPerPixel + 0.5, 0.0, 0.5);
  vec2 finalTextureCoords = (floor(pixel) + interpolationAmount) / textureSize2d.xy;

  // output
  vec4 color = texture(texture1, finalTextureCoords);

  if (color.a == 0.0) {
    discard;
  }

  FragColor = ourColor * color;
}
