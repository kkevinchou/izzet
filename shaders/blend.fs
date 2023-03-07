#version 330 core
out vec4 FragColor;

in vec2 TexCoords;

uniform sampler2D texture0;
uniform sampler2D texture1;


void main()
{
    vec3 color = texture(texture0, TexCoords).rgb + texture(texture1, TexCoords).rgb;
    // color = color / (color + vec3(1.0));
    FragColor = vec4(color, 1.0);
}