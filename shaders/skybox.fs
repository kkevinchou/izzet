#version 330 core

out vec4 FragColor;

in vec3 TexCoords;

uniform vec3 skyboxTopColor;
uniform vec3 skyboxBottomColor;
uniform float skyboxMixValue;

void main()
{
    // Normalize the direction vector
    vec3 direction = normalize(TexCoords);

    // Calculate the mix factor based on the y coordinate (height)
    float t = (direction.y + 1.0) * 0.5;

    vec3 topColor = skyboxTopColor;
    vec3 bottomColor = skyboxBottomColor;

    // Interpolate between the bottom and top colors
    vec3 color = mix(bottomColor, topColor, pow(t, skyboxMixValue));

    FragColor = vec4(color, 1.0);
}