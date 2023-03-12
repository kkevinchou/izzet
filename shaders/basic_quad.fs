#version 330 core
layout (location = 0) out vec4 FragColor;
layout (location = 1) out vec4 PickingColor;

in vec2 TexCoords;
uniform sampler2D basictexture;
uniform vec3 pickingColor;

void main() {
    vec4 t = texture(basictexture, TexCoords);
    if (t.a == 0) {
        discard;
    }

    FragColor = t;
    PickingColor = vec4(pickingColor, 1);
}