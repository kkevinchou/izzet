#version 330 core
const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 4;

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;
layout (location = 3) in ivec3 jointIndices;
layout (location = 4) in vec3 jointWeights;

out VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    vec2 TexCoord;
} vs_out;

uniform mat4 model;
uniform mat4 modelRotationMatrix;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 jointTransforms[MAX_JOINTS];
uniform mat4 lightSpaceMatrix;

void main() {
    vec4 totalPos = vec4(0.0);
	vec4 totalNormal = vec4(0.0);

    // note: the total position post transformation does not necessarily have W == 1
    // i.e.
    // a = totalPos
    // b = vec4(totalPos.xyz, 1)
    // a does not equal b here.

	for(int i = 0; i < MAX_WEIGHTS; i++){
		int jointIndex = jointIndices[i];

		mat4 jointTransform = jointTransforms[jointIndex];
		vec4 posePosition = jointTransform * vec4(aPos, 1.0);
		totalPos += posePosition * jointWeights[i];

		vec4 worldNormal = jointTransform * vec4(aNormal, 0.0);
		totalNormal += worldNormal * jointWeights[i];
	}
    vs_out.Normal = vec3(transpose(inverse(modelRotationMatrix)) * totalNormal);

    vs_out.FragPos = vec3(model * totalPos);
    vs_out.FragPosLightSpace = lightSpaceMatrix * (model * totalPos);

    vs_out.View = view;
	vs_out.TexCoord = aTexCoord;

    gl_Position = (projection * (view * (model * totalPos)));
}