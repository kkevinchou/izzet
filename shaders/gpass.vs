#version 330 core
const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord0;
layout (location = 3) in vec2 aTexCoord1;
layout (location = 4) in ivec3 jointIndices;
layout (location = 5) in vec3 jointWeights;

out vec3 FragPos;
out vec2 TexCoords;
out vec3 Normal;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

uniform int isAnimated;
uniform mat4 jointTransforms[MAX_JOINTS];

void main()
{
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

    vec4 viewPos = view * model * totalPos;
    FragPos = viewPos.xyz; 
    TexCoords = aTexCoord0;
    
    mat4 normalMatrix = transpose(inverse(view * model));
    Normal = vec3(normalMatrix * totalNormal);
    
    gl_Position = projection * viewPos;
}