#version 330 core

in vec3 FragPos;
in vec2 TexCoord;
out vec4 FragColor;

uniform sampler2D skyboxTexture;

void main()
{
    FragColor = texture(skyboxTexture, TexCoord);
}