
#version 330 core
out vec4 FragColor;

uniform vec3 resolution;

void main() {
    // place the origin of our local coordinate system to be the middle of the rendered area
    vec2 uv = gl_FragCoord.xy / vec2(1024, 1024) * 2 - 1;

    float distance = 1 - length(uv);
    distance = step(0, distance);

    FragColor = vec4(0.5, 0, 0.5, distance - 0.2);

    // FragColor = vec4(0, 1, 0, 1);
}