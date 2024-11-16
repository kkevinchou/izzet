#version 330 core
out vec4 FragColor;

in vec2 TexCoords;

uniform sampler2D scene;
uniform sampler2D bloomBlur;
uniform float bloomIntensity = 0.04f;

vec3 compute_bloom_color()
{
    vec3 hdrColor = texture(scene, TexCoords).rgb;
    vec3 bloomColor = texture(bloomBlur, TexCoords).rgb;
    return mix(hdrColor, bloomColor, bloomIntensity); // linear interpolation
}

void main()
{
    vec3 color = compute_bloom_color();
    FragColor = vec4(color, 1.0);
}