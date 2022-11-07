#version 330 core
out vec4 FragColor;

in vec3 FragPos;

uniform vec4 pickingColor;

void main()
{
    FragColor = pickingColor;
}
