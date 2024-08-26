#version 430 core

out vec4 FragColor;

in vec2 TexCoords;

uniform sampler3D tex;
uniform sampler2D tex1;

void main() {
    vec4 color1 = texture(tex, vec3(TexCoords.xy, 0));
    vec4 color2 = texture(tex1, TexCoords);

    vec4 color = color1;

    FragColor = vec4(color.xyz, 1.0);
}
