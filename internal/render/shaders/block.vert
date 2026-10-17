#version 450
layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aUV;
layout(location = 2) in float aLight;
layout(location = 3) in float aSky;
layout(location = 4) in float aBlock;

layout(set = 1, binding = 0) uniform UBO {
    mat4 uViewProj;
    vec4 uFogColor;
    vec4 uFogParams;
    vec4 uCam;
    vec4 uSun;
} u;

layout(location = 0) out vec2 vUV;
layout(location = 1) out float vLight;
layout(location = 2) out float vDist;
layout(location = 3) out float vSky;
layout(location = 4) out float vBlock;

void main() {
    gl_Position = u.uViewProj * vec4(aPos, 1.0);
    vUV = aUV;
    vLight = aLight;
    vDist = length(aPos - u.uCam.xyz);
    vSky = aSky;
    vBlock = aBlock;
}