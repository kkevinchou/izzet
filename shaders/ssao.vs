#version 330 core

layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord0;
layout (location = 3) in vec2 aTexCoord1;
layout (location = 4) in ivec3 jointIndices;
layout (location = 5) in vec3 jointWeights;


out vec2 TexCoords;

void main()
{
    TexCoords = aTexCoord0;
    gl_Position = vec4(aPos, 1.0);
}