#version 330 core

layout (location = 0) out vec4 FragColor;
layout (location = 1) out uint PickingColor;

in vec3 FragPos;  
in vec3 Normal;  

uniform vec3 color;
uniform float intensity;
uniform uint entityID;

void main()
{
    // FragColor = vec4(color * intensity, 1);
    FragColor = vec4(normalize(FragPos), 1);
    PickingColor = entityID;
}
