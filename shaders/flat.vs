#version 330 core

out vec3 FragPos;
out vec3 Normal;
out float Alpha;

layout (location = 0) in vec3 aPos;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
uniform float alpha;

void main() {
    Alpha = alpha;
    gl_Position = projection * view * model * vec4(aPos, 1.0);
}
