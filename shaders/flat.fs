
#version 330 core
out vec4 FragColor;

in vec3 FragPos;  
in vec3 Normal;  

uniform vec3 color;
uniform float intensity;

void main()
{
    FragColor = vec4(color * intensity, 1);
}
