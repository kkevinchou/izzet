#version 330 core
const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;

out VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    vec2 TexCoord;
} vs_out;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 lightSpaceMatrix;

void main() {
    vs_out.FragPos = vec3(model * vec4(aPos, 1.0));
    // TODO: the normal matrix is expensive to calculate and should be passed in as a uniform
    vs_out.Normal = transpose(inverse(mat3(model))) * aNormal;
    vs_out.FragPosLightSpace = lightSpaceMatrix * vec4(vs_out.FragPos, 1.0);
    vs_out.View = view;
	vs_out.TexCoord = aTexCoord;

    gl_Position = (projection * (view * (model * vec4(aPos, 1.0))));
}