#version 410 core
const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in ivec3 jointIndices;
layout (location = 2) in vec3 jointWeights;

out VS_OUT {
    vec4 WorldPos;
} vs_out;

uniform mat4 model;
uniform mat4 jointTransforms[MAX_JOINTS];
uniform int isAnimated;

void main() {
    vec4 totalPos = vec4(0.0);

    if (isAnimated == 1) {
        for(int i = 0; i < MAX_WEIGHTS; i++){
            int jointIndex = jointIndices[i];

            mat4 jointTransform = jointTransforms[jointIndex];
            vec4 posePosition = jointTransform * vec4(aPos, 1.0);
            totalPos += posePosition * jointWeights[i];
        }

        totalPos = totalPos / totalPos.w;
    } else {
        totalPos = vec4(aPos, 1);
    }

    vs_out.WorldPos = model * totalPos;
    gl_Position = vs_out.WorldPos;
}
