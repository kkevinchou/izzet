#version 410 core

layout (triangles, invocations = 4) in;
layout (triangle_strip, max_vertices = 3) out;

const int MAX_CASCADES = 16;

uniform int cascadeCount;
uniform mat4 lightSpaceMatrixArray[MAX_CASCADES];

in VS_OUT {
    vec4 WorldPos;
} gs_in[];

void main()
{
    if (gl_InvocationID >= cascadeCount) {
        return;
    }

    for (int i = 0; i < 3; i++) {
        gl_Position = lightSpaceMatrixArray[gl_InvocationID] * gs_in[i].WorldPos;
        gl_Layer = gl_InvocationID;
        EmitVertex();
    }
    EndPrimitive();
}
