#version 330 core

layout (location = 0) out vec4 FragColor;

in vec2 TexCoords;

uniform sampler3D tex;

void main() {
    // vec3 color = texture(tex, vec3(TexCoords, 0.0));
    // FragColor = vec4(color, 1.0);

    vec4 color = texture(tex, vec3(floor(TexCoords.x), floor(TexCoords.y), 0));
    // vec4 color = texture(tex, vec3(3, 2, 0));
    FragColor = vec4(color.xyz, 1.0);
}
