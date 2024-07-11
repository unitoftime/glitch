#version 300 es

// Adapted from: https://www.redblobgames.com/x/2404-distance-field-effects/distance-field-effects.js

// Required for webgl
#ifdef GL_ES
precision highp float;
#endif

out vec4 FragColor;

in vec4 ourColor;
in vec2 TexCoord;

uniform sampler2D texture1;

//#extension GL_OES_standard_derivatives : enable
//precision mediump float;

/* varying vec2 v_texcoord; */
/* uniform sampler2D u_mtsdf_font; */

/* uniform vec2 u_unit_range; */
/* uniform float u_rounded_fonts; */
/* uniform float u_rounded_outlines; */
uniform float u_threshold;
/* uniform float u_out_bias; */
/* uniform float u_outline_width_absolute; */
uniform float u_outline_width_relative;
uniform float u_outline_blur;
/* uniform float u_gradient; */
/* uniform float u_gamma; */
uniform vec4 u_outline_color;

/* sample code from https://github.com/Chlumsky/msdfgen */
float median(float r, float g, float b) {
  return max(min(r, g), min(max(r, g), b));
}

float screenPxRange() {
  float distanceRange = 10.0;
  vec2 texSize = vec2(textureSize(texture1, 0));
  vec2 u_unit_range = vec2(distanceRange / texSize.x, distanceRange / texSize.y);
  vec2 screenTexSize =  vec2(1.0) / fwidth(TexCoord);
  return max(0.5 * dot(u_unit_range, screenTexSize), 1.0);
}

void main() {
  float u_rounded_fonts = 0.0;
  float u_rounded_outlines = 0.0;
  /* float u_threshold = 0.5; */
  float u_out_bias = 0.25;
  float u_outline_width_absolute = 0.3;
  /* float u_outline_width_relative = 0.05; */
  /* float u_outline_blur = 0.0; */
  float u_gradient = 0.0;
  float u_gamma = 1.0;

  /* vec4 inner_color = vec4(1, 1, 1, 1); */
  /* vec4 outer_color = vec4(0, 0, 0, 1); */
  vec4 inner_color = ourColor;
  vec4 outer_color = u_outline_color * ourColor;

  // distances are stored with 1.0 meaning "inside" and 0.0 meaning "outside"
  vec4 distances = texture2D(texture1, TexCoord);
  float d_msdf = median(distances.r, distances.g, distances.b);
  float d_sdf = distances.a; // mtsdf format only
  d_msdf = min(d_msdf, d_sdf + 0.1);  // HACK: to fix glitch in msdf near edges

  // blend between sharp and rounded corners
  float d_inner = mix(d_msdf, d_sdf, u_rounded_fonts);
  float d_outer = mix(d_msdf, d_sdf, u_rounded_outlines);

  // typically 0.5 is the threshold, >0.5 inside <0.5 outside
  float inverted_threshold = 1.0 - u_threshold; // because I want the ui to be +larger -smaller
  float width = screenPxRange();
  float inner = width * (d_inner - inverted_threshold) + 0.5 + u_out_bias;
  float outer = width * (d_outer - inverted_threshold + u_outline_width_relative) + 0.5 + u_out_bias + u_outline_width_absolute;

  float inner_opacity = clamp(inner, 0.0, 1.0);
  float outer_opacity = clamp(outer, 0.0, 1.0);

  if (u_outline_blur > 0.0) {
    // NOTE: the smoothstep fails when the two edges are the same, and I wish it
    // would act like a step function instead of failing.
    // NOTE: I'm using d_sdf here because I want the shadows to be rounded
    // even when outlines are sharp. But I don't yet have implemented a way
    // to see the sharp outline with a rounded shadow.
    float blur_start = u_outline_width_relative + u_outline_width_absolute / width;
    outer_color.a = smoothstep(blur_start,
                               blur_start * (1.0 - u_outline_blur),
                               inverted_threshold - d_sdf - u_out_bias / width);
  }

  // apply some lighting (hard coded angle)
  if (u_gradient > 0.0) {
     // NOTE: this is not a no-op so it changes the rendering even when
     // u_gradient is 0.0. So I use an if() instead. But ideally I'd
     // make this do nothing when u_gradient is 0.0.
     vec2 normal = normalize(vec3(dFdx(d_inner), dFdy(d_inner), 0.01)).xy;
     float light = 0.5 * (1.0 + dot(normal, normalize(vec2(-0.3, -0.5))));
     inner_color = mix(inner_color, vec4(light, light, light, 1),
                       smoothstep(u_gradient + inverted_threshold, inverted_threshold, d_inner));
  }

  // unlike in the 2403 experiments, I do know the color is light
  // and the shadow is dark so I can implement gamma correction
  inner_opacity = pow(inner_opacity, 1.0 / u_gamma);

  vec4 color = (inner_color * inner_opacity) + (outer_color * (outer_opacity - inner_opacity));

  if (color.a == 0.0) {
    discard;
  }

  FragColor = color;
}

/* #version 300 es */

/* // Required for webgl */
/* #ifdef GL_ES */
/* precision highp float; */
/* #endif */

/* out vec4 FragColor; */

/* in vec4 ourColor; */
/* in vec2 TexCoord; */

/* uniform sampler2D texture1; */
/* /\* uniform vec4 bgColor; *\/ */
/* /\* uniform vec4 fgColor; *\/ */

/* // https://github.com/Chlumsky/msdfgen?tab=readme-ov-file#using-a-multi-channel-distance-field */
/* float median(float r, float g, float b) { */
/*     return max(min(r, g), min(max(r, g), b)); */
/* } */

/* void main() { */
/*   float screenPxRange = 2.5; */
/*   float u_in_bias = 0.2; */
/*   float u_out_bias = 0.2; */
/*   float u_outline = 0.1; */

/*   vec4 inner_color = vec4(1, 1, 1, 1); */
/*   vec4 outer_color = vec4(1, 0, 0, 1); */


/*   vec4 msd = texture(texture1, TexCoord); */
/*   float sd = median(msd.r, msd.g, msd.b); */
/*   float width = screenPxRange; */
/*   float inner = width * (sd - 0.5 + u_in_bias) + 0.5 + u_out_bias; */
/*   float outer = width * (sd - 0.5 + u_in_bias + u_outline) + 0.5 + u_out_bias; */

/*   float inner_opacity = clamp(inner, 0.0, 1.0); */
/*   float outer_opacity = clamp(outer, 0.0, 1.0); */

/*   FragColor = (inner_color * inner_opacity) + (outer_color * outer_opacity); */



/*   /\* float screenPxRange = 2.5; *\/ */
/*   /\* float screenPxRange2 = 2.5 *\/ */

/*   /\* vec4 bgColor = vec4(1.0, 0.0, 0.0, 1.0); *\/ */
/*   /\* vec4 fgColor = vec4(0.0, 1.0, 0.0, 1.0); *\/ */

/*   /\* float borderWidth = 0.5; *\/ */
/*   /\* /\\* float borderEdge = 0.1; *\\/ *\/ */
/*   /\* /\\* vec4 borderColor = vec3(1.0, 0.0, 0.0); *\\/ *\/ */


/*   /\* vec3 msd = texture(texture1, TexCoord).rgb; *\/ */
/*   /\* float sd = median(msd.r, msd.g, msd.b); *\/ */
/*   /\* float screenPxDistance = screenPxRange * (sd - 0.5); *\/ */
/*   /\* float opacity = clamp(screenPxDistance + 0.5, 0.0, 1.0); *\/ */

/*   /\* float screenPxDistance2 = screenPxRange2 * (sd - borderWidth); *\/ */
/*   /\* float opacity2 = clamp(screenPxDistance2 + borderWidth, 0.0, 1.0); *\/ */

/*   /\* float finalAlpha = opacity + (1.0 - opacity) * opacity2; *\/ */
/*   /\* vec4 finalColor = mix(fgColor, bgColor, opacity / finalAlpha); *\/ */

/*   /\* FragColor = finalColor; *\/ */
/* } */

/* /\* void main() { *\/ */
/* /\*   float screenPxRange = 4.5; *\/ */
/* /\*   vec4 bgColor = vec4(0.0, 0.0, 0.0, 0.0); *\/ */
/* /\*   vec4 fgColor = vec4(0.0, 1.0, 0.0, 1.0); *\/ */


/* /\*   vec3 msd = texture(texture1, TexCoord).rgb; *\/ */
/* /\*   float sd = median(msd.r, msd.g, msd.b); *\/ */
/* /\*   float screenPxDistance = screenPxRange*(sd - 0.5); *\/ */
/* /\*   float opacity = clamp(screenPxDistance + 0.5, 0.0, 1.0); *\/ */
/* /\*   FragColor = mix(bgColor, fgColor, opacity); *\/ */
/* /\* } *\/ */

/* //RBG */
/* /\* void main() { *\/ */
/* /\*   vec4 distances = texture2D(u_mtsdf_font, v_texcoord); *\/ */
/* /\*   float d = median(distances.r, distances.g, distances.b); *\/ */
/* /\*   float width = screenPxRange(); *\/ */
/* /\*   float inner = width * (d - 0.5 + u_in_bias            ) + 0.5 + u_out_bias; *\/ */
/* /\*   float outer = width * (d - 0.5 + u_in_bias + u_outline) + 0.5 + u_out_bias; *\/ */

/* /\*   float inner_opacity = clamp(inner, 0.0, 1.0); *\/ */
/* /\*   vec4 inner_color = vec4(1, 1, 1, 1); *\/ */
/* /\*   float outer_opacity = clamp(outer, 0.0, 1.0); *\/ */
/* /\*   vec4 outer_color = vec4(0, 0, 0, 1); *\/ */

/* /\*   vec4 color = (inner_color * inner_opacity) + (outer_color * outer_opacity); *\/ */
/* /\*   gl_FragColor = color; *\/ */
/* /\* } *\/ */
