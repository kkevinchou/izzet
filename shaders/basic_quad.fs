#version 330 core
layout (location = 0) out vec4 FragColor;
layout (location = 1) out vec4 PickingColor;

in vec2 TexCoords;
uniform sampler2D basictexture;
uniform vec3 pickingColor;

uniform int doColorOverride;
uniform vec3 colorOverride;
uniform float colorOverrideIntensity;

void main() {
    vec4 t = texture(basictexture, TexCoords);
    if (t.a == 0) {
        discard;
    }

    if (doColorOverride == 1) {
        FragColor = vec4(colorOverride * colorOverrideIntensity, 1);
    } else {
        FragColor = t;
    }
    PickingColor = vec4(pickingColor, 1);
}