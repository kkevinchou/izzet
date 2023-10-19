#version 330 core
out vec4 FragColor;

uniform vec3 resolution;
uniform vec4 color;

void main() {
    // place the origin of our local coordinate system to be the middle of the rendered area
    vec2 uv = gl_FragCoord.xy / vec2(1024, 1024) * 2 - 1;

    float dist = length(uv);

    if (dist > 1 || dist < 0.88) {
        FragColor = vec4(0, 0, 0, 0);
    } else {
        FragColor = vec4(vec3(color), color.a);
    }
}