#version 450
layout(location = 0) in vec2 vUV;
layout(location = 1) in float vLight;
layout(location = 2) in float vDist;
layout(location = 3) in float vSky;
layout(location = 4) in float vBlock;

layout(set = 2, binding = 0) uniform sampler2D uTex;

layout(set = 3, binding = 0) uniform UBO {
    mat4 uViewProj;
    vec4 uFogColor;
    vec4 uFogParams;
    vec4 uCam;
    vec4 uSun;
} u;

layout(location = 0) out vec4 oColor;

void main() {
    vec4 tex = texture(uTex, vUV);
    if (tex.a < 0.035) discard;
    // Keep unlit caves readable instead of clipping to pure black.
    vec3 light = vec3(0.13, 0.14, 0.17) + u.uSun.rgb * vSky + vec3(1.0, 0.72, 0.42) * vBlock;
    vec3 col = tex.rgb * vLight * light;
    float fog = clamp((vDist - u.uFogParams.x) / max(u.uFogParams.y - u.uFogParams.x, 0.001), 0.0, 1.0);
    col = mix(col, u.uFogColor.rgb, fog * 0.92);
    oColor = vec4(col, tex.a);
}