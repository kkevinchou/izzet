#version 330 core

out vec3 FragPos;
out vec3 Normal;
out float Alpha;

layout (location = 0) in vec3 aPos;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

out VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    vec2 TexCoord;
} vs_out;

void main() {
    gl_Position = projection * view * model * vec4(aPos, 1.0);
    FragPos = (model * vec4(aPos, 1.0)).xyz;
}
