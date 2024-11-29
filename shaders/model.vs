#version 330 core
const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord0;
layout (location = 3) in vec2 aTexCoord1;
layout (location = 4) in ivec3 jointIndices;
layout (location = 5) in vec3 jointWeights;

out VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    vec2 TexCoord;
    vec3 ColorOverride;
    vec2 NDCCoord;
} vs_out;

uniform mat4 model;
uniform mat4 modelRotationMatrix;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 jointTransforms[MAX_JOINTS];
uniform mat4 lightSpaceMatrix;
uniform int isAnimated;
uniform int colorTextureCoordIndex;

const vec3 errorColor = vec3(255.0 / 255, 28.0 / 255, 217.0 / 121.0);

void main() {
    vec4 totalPos = vec4(0.0);
	vec4 totalNormal = vec4(0.0);

    if (isAnimated == 1) {
        for(int i = 0; i < MAX_WEIGHTS; i++){
            int jointIndex = jointIndices[i];

            mat4 jointTransform = jointTransforms[jointIndex];
            vec4 posePosition = jointTransform * vec4(aPos, 1.0);
            totalPos += posePosition * jointWeights[i];

            vec4 worldNormal = jointTransform * vec4(aNormal, 0.0);
            totalNormal += worldNormal * jointWeights[i];
        }

        // after animations we need to make sure we normalize w to 1;
        totalPos = totalPos / totalPos.w;
    } else {
        totalPos = vec4(aPos, 1);
        totalNormal = vec4(aNormal, 1.0);
    }

    // NOTE - just transposing the inverse of the rotation matrix didn't look right for some static geometry
    vs_out.Normal = vec3(transpose(inverse(model)) * totalNormal);
    vs_out.FragPos = vec3(model * totalPos);
    vs_out.View = view;

    vs_out.TexCoord = aTexCoord0;
    if (colorTextureCoordIndex == 0) {
        vs_out.TexCoord = aTexCoord0;
    } else if (colorTextureCoordIndex == 1) {
        vs_out.TexCoord = aTexCoord1;
    }

    vec4 clipPosition = projection * (view * (model * totalPos));
    clipPosition.xyz /= clipPosition.w; // perspective divide
    clipPosition.xyz = clipPosition.xyz * 0.5 + 0.5; // transform to range 0.0 - 1.0
    vs_out.NDCCoord = clipPosition.xy;

    vs_out.ColorOverride = errorColor;
    vs_out.FragPosLightSpace = lightSpaceMatrix * (model * totalPos);
    gl_Position = (projection * (view * (model * totalPos)));
}