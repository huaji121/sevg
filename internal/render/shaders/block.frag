#version 450
layout(location = 0) in vec2 vUV;
layout(location = 1) in float vLight;
layout(location = 2) in float vDist;
layout(location = 3) in float vSky;
layout(location = 4) in float vBlock;
layout(location = 5) in vec3 vWorld;

layout(set = 2, binding = 0) uniform sampler2D uTex;
layout(set = 2, binding = 1) uniform sampler2DArray uDynTex;

layout(set = 3, binding = 0) uniform UBO {
    mat4 uViewProj;
    vec4 uFogColor;
    vec4 uFogParams;
    vec4 uCam;
    vec4 uSun;
    vec4 uDyn;
} u;

layout(location = 0) out vec4 oColor;

void main() {
    vec4 tex = texture(uTex, vUV);
    if (tex.a < 0.035) discard;
    // Keep unlit caves readable instead of clipping to pure black.
    vec3 light = vec3(0.13, 0.14, 0.17) + u.uSun.rgb * vSky + vec3(1.0, 0.72, 0.42) * vBlock;
    // Player-held torch: a sampled 3D light volume, so all chunks update in
    // the same frame instead of being remeshed one at a time.
    float dyn = 0.0;
    if (u.uDyn.w > 0.5) {
        vec3 lp = (vWorld - u.uDyn.xyz) / u.uDyn.w;
        if (all(greaterThanEqual(lp, vec3(0.0))) && all(lessThanEqual(lp, vec3(1.0)))) {
            float layer = clamp(lp.z, 0.0, 1.0) * (u.uDyn.w - 1.0);
            float l0 = floor(layer);
            float f = fract(layer);
            float d0 = texture(uDynTex, vec3(lp.xy, l0)).r;
            float d1 = texture(uDynTex, vec3(lp.xy, min(l0 + 1.0, u.uDyn.w - 1.0))).r;
            dyn = mix(d0, d1, f);
        }
    }
    light += vec3(1.0, 0.72, 0.42) * dyn;
    vec3 col = tex.rgb * vLight * light;
    float fog = clamp((vDist - u.uFogParams.x) / max(u.uFogParams.y - u.uFogParams.x, 0.001), 0.0, 1.0);
    col = mix(col, u.uFogColor.rgb, fog * 0.92);
    oColor = vec4(col, tex.a);
}