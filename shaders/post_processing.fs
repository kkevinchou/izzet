#version 330 core
out vec4 FragColor;

in vec2 TexCoords;

uniform sampler2D image;

vec3 compute_color()
{
    return texture(image, TexCoords).rgb;
}

const float A = 2.51;
const float B = 0.03;
const float C = 2.43;
const float D = 0.59;
const float E = 0.14;

// Narokowicz ACES tone mapping function
vec3 acesToneMapping(vec3 color)
{
    color = (color * (A * color + B)) / (color * (C * color + D) + E);
    return clamp(color, 0.0, 1.0);
}

void main()
{
    vec3 color = compute_color();
    color = acesToneMapping(color);

    // Gamma correction
    // unclear if we actually need to do gamma correction. seems like GLTF expects us to internally
    // store textures in SRGB format which we then need to gamma correct here.
    // PARAMETERS:
    //     gl.Enable(gl.FRAMEBUFFER_SRGB)
    //         OpenGL setting for how the fragment shader outputs colors
    //     lightColor
    //         The color of the light. i've tested with (1, 1, 1) to (20, 20, 20)
    //     gamma correction in the fragment shader
    //         I've experimented with enabling/disabling. it seems like if i gamma correct
    //         I want to disable the OpenGL setting, and if I don't, I want to enable it instead.
    color = pow(color, vec3(1.0 / 2.2));
    FragColor = vec4(color, 1.0);
}