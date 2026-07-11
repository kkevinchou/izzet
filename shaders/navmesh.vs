#version 330 core
const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec3 aColor;
layout (location = 3) in vec3 aInstancePosition;
layout (location = 4) in vec3 aInstanceColor;

out VS_OUT {
    vec3 FragPos;
    vec3 ObjectPos;
    vec3 Normal;
    vec3 ObjectNormal;
    mat4 View;
    vec2 TexCoord;
    vec4 Color;
    flat uint EntityID;
    flat vec3 Scale;
    flat uint RepeatTexture;
} vs_out;

uniform mat4 model;
uniform mat4 modelRotationMatrix;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 lightSpaceMatrix;
uniform int isAnimated;
uniform int colorTextureCoordIndex;
uniform uint entityID;
uniform int useInstancing;
uniform vec3 instanceScale;

void main() {
    vec4 totalPos = vec4(0.0);
	vec4 totalNormal = vec4(0.0);
    vec3 color = aColor;

    if (useInstancing == 1) {
        totalPos = vec4(aInstancePosition + aPos * instanceScale, 1);
        color = aInstanceColor;
    } else {
        totalPos = vec4(aPos, 1);
    }
    totalNormal = vec4(aNormal, 1.0);

    // NOTE - just transposing the inverse of the rotation matrix didn't look right for some static geometry
    vs_out.Normal = vec3(transpose(inverse(model)) * totalNormal);
    vs_out.ObjectNormal = totalNormal.xyz;
    vs_out.FragPos = vec3(model * totalPos);
    vs_out.ObjectPos = totalPos.xyz;
    vs_out.View = view;

    vs_out.TexCoord = vec2(0.0);
    vs_out.Color = vec4(color, 0.9);
    vs_out.EntityID = entityID;
    vs_out.Scale = vec3(1.0);
    vs_out.RepeatTexture = uint(0);

    gl_Position = (projection * (view * (model * totalPos)));
}
