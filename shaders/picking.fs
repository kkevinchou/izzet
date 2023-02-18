#version 330 core
out vec4 FragColor;

uniform vec3 pickingColor;

void main()
{
    FragColor = vec4(pickingColor, 1);
}
