#version 330 core

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;

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
    vec4 totalPos = vec4(aPos, 1);
    vec4 totalNormal = vec4(aNormal, 1.0);

    // NOTE - just transposing the inverse of the rotation matrix didn't look right for some static geometry
    vs_out.Normal = vec3(transpose(inverse(model)) * totalNormal);
    vs_out.FragPos = vec3(model * totalPos);
    vs_out.View = view;

    vs_out.FragPosLightSpace = lightSpaceMatrix * (model * totalPos);
    gl_Position = (projection * (view * (model * totalPos)));
}