#version 330 core

out vec4 FragColor;

in vec2 TexCoords;

uniform sampler2DArray depthMap;
uniform int layer;

void main()
{
    float depthValue = texture(depthMap, vec3(TexCoords, float(layer))).r;
    FragColor = vec4(vec3(depthValue), 1.0);
}
