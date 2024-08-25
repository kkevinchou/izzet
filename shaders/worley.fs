#version 330 core

layout (location = 0) out vec4 FragColor;

in vec3 FragPos;  

void main() {
    FragColor = vec4(0.2, 0.8, 0.2, 1);
}
