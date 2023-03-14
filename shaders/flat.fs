
#version 330 core
layout (location = 0) out vec4 FragColor;
layout (location = 1) out vec4 PickingColor;

in vec3 FragPos;  
in vec3 Normal;  

uniform vec3 color;
uniform float intensity;
uniform vec3 pickingColor;

void main()
{
    FragColor = vec4(color * intensity, 1);
    PickingColor = vec4(pickingColor, 1);
}
