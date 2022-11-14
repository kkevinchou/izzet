#version 330 core
out vec4 FragColor;

in vec2 TexCoords;
uniform sampler2D basictexture;


void main() {
    vec4 t = texture(basictexture, TexCoords);
    if (t.a == 0) {
        discard;
    }

    FragColor = t;
}