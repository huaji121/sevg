#version 450
layout(location = 0) in vec2 aPos;
layout(location = 1) in vec2 aUV;
layout(location = 2) in vec4 aColor;

layout(set = 1, binding = 0) uniform UBO {
    vec4 uScreen;
} u;

layout(location = 0) out vec2 vUV;
layout(location = 1) out vec4 vColor;

void main() {
    gl_Position = vec4(aPos.x / u.uScreen.x * 2.0 - 1.0, 1.0 - aPos.y / u.uScreen.y * 2.0, 0.0, 1.0);
    vUV = aUV;
    vColor = aColor;
}
