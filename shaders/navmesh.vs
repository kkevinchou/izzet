#version 330 core
const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec3 aColor;

out VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    mat4 View;
    vec2 TexCoord;
    vec4 Color;
    flat uint EntityID;
} vs_out;

uniform mat4 model;
uniform mat4 modelRotationMatrix;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 lightSpaceMatrix;
uniform int isAnimated;
uniform int colorTextureCoordIndex;
uniform uint entityID;

void main() {
    vec4 totalPos = vec4(0.0);
	vec4 totalNormal = vec4(0.0);

    totalPos = vec4(aPos, 1);
    totalNormal = vec4(aNormal, 1.0);

    // NOTE - just transposing the inverse of the rotation matrix didn't look right for some static geometry
    vs_out.Normal = vec3(transpose(inverse(model)) * totalNormal);
    vs_out.FragPos = vec3(model * totalPos);
    vs_out.View = view;

    vs_out.TexCoord = vec2(0.0);
    vs_out.Color = vec4(aColor, 0.9);
    vs_out.EntityID = entityID;

    gl_Position = (projection * (view * (model * totalPos)));
}
