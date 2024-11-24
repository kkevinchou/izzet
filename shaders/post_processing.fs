#version 330 core
out vec4 FragColor;

in vec2 TexCoords;

uniform int kuwahara;
uniform sampler2D image;

const float A = 2.51;
const float B = 0.03;
const float C = 2.43;
const float D = 0.59;
const float E = 0.14;

const float MAX_FLOAT = 100000000;
const int MAX_KERNEL_SIZE = 7;

// Narokowicz ACES tone mapping function
vec3 acesToneMapping(vec3 color)
{
    color = (color * (A * color + B)) / (color * (C * color + D) + E);
    return clamp(color, 0.0, 1.0);
}

void calculatePixelColors(ivec2 texel, out vec3 pixelColors[MAX_KERNEL_SIZE*MAX_KERNEL_SIZE], out vec3 meanColor, vec2 texelSize) {
    vec3 sumPixelColors = vec3(0);

    int idx = 0;
    for (int i = texel.x; i < texel.x + MAX_KERNEL_SIZE; ++i) {
        for (int j = texel.y; j < texel.y + MAX_KERNEL_SIZE; ++j) {
            vec2 localTexel = clamp(
                vec2(TexCoords.x + (i * texelSize.x), TexCoords.y + (j * texelSize.y)),
                vec2(0.0), vec2(1.0)
            );
            vec3 pixelColor = texture(image, localTexel).rgb;

            pixelColors[idx++] = pixelColor;
            sumPixelColors += pixelColor;
        }
    }

    meanColor = sumPixelColors / float(MAX_KERNEL_SIZE*MAX_KERNEL_SIZE);
    return;
}

float calculateStandardDeviation(vec3 centerPixelColor, vec3 pixelColors[MAX_KERNEL_SIZE*MAX_KERNEL_SIZE]) {
    // std calculation
    float squaredDiff[MAX_KERNEL_SIZE*MAX_KERNEL_SIZE];
    float avgSquaredDiff = 0;

    for (int i = 0; i < (MAX_KERNEL_SIZE*MAX_KERNEL_SIZE); ++i) {
        squaredDiff[i] = pow(length(pixelColors[i] - centerPixelColor), 2);
        avgSquaredDiff += squaredDiff[i];
    }

    avgSquaredDiff /= (MAX_KERNEL_SIZE*MAX_KERNEL_SIZE);
    return avgSquaredDiff;
}

void main()
{
    vec3 color = texture(image, TexCoords).rgb;
    vec2 texelSize = 1.0 / textureSize(image, 0);

    if (kuwahara == 1) {
        vec3 centerPixelColor = color;
        int xQuadrant[2] = int[2](-MAX_KERNEL_SIZE+1, 0);
        int yQuadrant[2] = int[2](-MAX_KERNEL_SIZE+1, 0);

        float minStandardDeviation = MAX_FLOAT;
        vec3 minStandardDeviationColor;

        vec3 pixelColors[MAX_KERNEL_SIZE*MAX_KERNEL_SIZE];
        vec3 meanColor = vec3(0);

        for (int i = 0; i < 2; ++i) {
            for (int j = 0; j < 2; ++j) {
                ivec2 texel = ivec2(xQuadrant[i], yQuadrant[j]);
                calculatePixelColors(texel, pixelColors, meanColor, texelSize);
                float standardDeviation = calculateStandardDeviation(centerPixelColor, pixelColors);
                if (standardDeviation < minStandardDeviation) {
                    minStandardDeviation = standardDeviation;
                    minStandardDeviationColor = meanColor;
                }
            }
        }
        color = minStandardDeviationColor;
    }

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
