#version 330

out vec4 FragColor;
in vec2 TexCoords;

uniform sampler2D maintexture;

void main() {
    FragColor = vec4(texture(maintexture, TexCoords));
}
