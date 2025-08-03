#version 330 core

const vec2 QUAD[6] = vec2[](
    vec2(-1.0, -1.0),
    vec2( 1.0, -1.0),
    vec2(-1.0,  1.0),
    vec2(-1.0,  1.0),
    vec2( 1.0, -1.0),
    vec2( 1.0,  1.0)
);

out vec2 vNDC;

void main() {
    vec2 ndc = QUAD[gl_VertexID];
    vNDC = ndc;
    gl_Position = vec4(ndc, 0.0, 1.0);
}