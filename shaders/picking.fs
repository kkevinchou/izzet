#version 330 core
out vec4 FragColor;

in vec3 FragPos;

uniform vec3 pickingColor;

void main()
{
    FragColor = vec4(pickingColor, 1);
}
