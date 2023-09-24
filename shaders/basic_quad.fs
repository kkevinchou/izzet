#version 330 core
layout (location = 0) out vec4 FragColor;
layout (location = 1) out uint PickingColor;

in vec2 TexCoords;
uniform sampler2D basictexture;
uniform uint entityID;

void main() {
    vec4 t = texture(basictexture, TexCoords);
    // if (t.a < 0.01) {
    //     discard;
    // }

    FragColor = t;
    PickingColor = entityID;
}