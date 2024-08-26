#version 430 core

out vec4 FragColor;

in vec2 TexCoords;

uniform sampler3D tex;
uniform sampler2D tex1;

void main() {
    // vec3 color = texture(tex, vec3(TexCoords, 0.0));
    // FragColor = vec4(color, 1.0);

    // vec4 color1 = texture(tex, vec3(0.6, 0.3, 0));
    vec4 color1 = texture(tex, vec3(TexCoords.xy, 0));
    vec4 color2 = texture(tex1, TexCoords);
    // vec4 color = texture(tex, vec3(floor(TexCoords.x), floor(TexCoords.y), 0));

    vec4 color = color1;

    // vec4 color = texture(tex, vec3(0, 3, 0));
    FragColor = vec4(color.xyz, 1.0);
}
