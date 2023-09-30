#version 330 core

const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in ivec3 jointIndices;
layout (location = 2) in vec3 jointWeights;

uniform mat4 model;

// animation
uniform mat4 jointTransforms[MAX_JOINTS];
uniform int isAnimated;

void main()
{
    vec4 totalPos = vec4(0.0);

    if (isAnimated == 1) {
        for(int i = 0; i < MAX_WEIGHTS; i++){
            int jointIndex = jointIndices[i];

            mat4 jointTransform = jointTransforms[jointIndex];
            vec4 posePosition = jointTransform * vec4(aPos, 1.0);
            totalPos += posePosition * jointWeights[i];
        }

        // after animations we need to make sure we normalize w to 1;
        totalPos = totalPos / totalPos.w;
    } else {
        totalPos = vec4(aPos, 1);
    }

    gl_Position = model * totalPos;
}