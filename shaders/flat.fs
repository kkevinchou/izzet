
#version 330 core
out vec4 FragColor;

in vec3 FragPos;  
in vec3 Normal;  
in float Alpha;

uniform vec3 color;

void main()
{
    vec3 objectColor = color;
    FragColor = vec4(objectColor, Alpha);
}
