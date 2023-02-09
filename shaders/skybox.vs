#version 330 core

out vec2 FragPos;

layout (location = 0) in vec3 aPos;

void main() {
    FragPos = aPos.xy;
    gl_Position = vec4(aPos, 1.0);
}