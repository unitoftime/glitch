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

// This is essentially the camera zoom level
// E.G. If you were zoomed in 4x, you'd pass in 1.0/4.0 (ie 1.0 texture texel maps to 4.0 screen pixels)
uniform float texelsPerPixel;

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

  // --------------------------------------------------------------------------------
  // Attempt 2: https://colececil.io/blog/2017/scaling-pixel-art-without-destroying-it/
  // --------------------------------------------------------------------------------
  // Note: This looks less blurry
  /* float texelsPerPixel = 100.0; // TODO - Not sure where to get this from */
  /* float texelsPerPixel = 1.0 / 8.0; */
  /* float texelsPerPixel = 1.0 / 4.0; // (1920.0 / 4.0))/ 1920.0; // TODO: pass in with uniform */


  /* float texelsPerPixel = 1.0 / 8.0; */
  /* float texelsPerPixel = 1.0 / view[0][0]; */
  /* float texelsPerPixel = view[0][0]; */

  vec2 textureSize2d = vec2(textureSize(texture1, 0));
  vec2 pixel = TexCoord * textureSize2d.xy;

  vec2 locationWithinTexel = fract(pixel);
  vec2 interpolationAmount = clamp(locationWithinTexel / texelsPerPixel, 0.0, 0.5) + clamp((locationWithinTexel - 1.0) / texelsPerPixel + 0.5, 0.0, 0.5);
  /* vec2 interpolationAmount = clamp(locationWithinTexel * texelsPerPixel, 0.0, 0.5) + clamp((locationWithinTexel - 1.0) * texelsPerPixel + 0.5, 0.0, 0.5); */
  vec2 finalTextureCoords = (floor(pixel) + interpolationAmount) / textureSize2d.xy;

  // output
  vec4 color = texture(texture1, finalTextureCoords);

  // --------------------------------------------------------------------------------
  // https://www.shadertoy.com/view/MlB3D3
  // --------------------------------------------------------------------------------
  /* vec2 textureSize2d = vec2(textureSize(texture1, 0)); */
  /* vec2 pix = (TexCoord * textureSize2d.xy) * texelsPerPixel; */

  /* pix = floor(pix) + min(fract(pix) / fwidth(pix), 1.0) - 0.5; */
  /* /\* pix = floor(pix) + smoothstep(0.0, 1.0, fract(pix) / fwidth(pix)) - 0.5; // Sharper *\/ */

  /* // sample and return */
  /* vec4 color = texture(texture1, pix / textureSize2d.xy); */

  if (color.a == 0.0) {
    discard;
  }

  FragColor = ourColor * color;
}
