#version 330 core
const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in int aDistance;
layout (location = 3) in int aRegionID;

out VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    flat int Distance;
    flat int RegionID;
} vs_out;

uniform mat4 model;
uniform mat4 modelRotationMatrix;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 lightSpaceMatrix;
uniform int isAnimated;
uniform int colorTextureCoordIndex;

const vec3 errorColor = vec3(255.0 / 255, 28.0 / 255, 217.0 / 121.0);

void main() {
    vec4 totalPos = vec4(0.0);
	vec4 totalNormal = vec4(0.0);

    totalPos = vec4(aPos, 1);
    totalNormal = vec4(aNormal, 1.0);

    // NOTE - just transposing the inverse of the rotation matrix didn't look right for some static geometry
    vs_out.Normal = vec3(transpose(inverse(model)) * totalNormal);
    vs_out.FragPos = vec3(model * totalPos);
    vs_out.View = view;

    vs_out.FragPosLightSpace = lightSpaceMatrix * (model * totalPos);
    vs_out.Distance = aDistance;
    vs_out.RegionID = aRegionID;

    gl_Position = (projection * (view * (model * totalPos)));
}