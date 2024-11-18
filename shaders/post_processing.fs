#version 330 core
out vec4 FragColor;

in vec2 TexCoords;

uniform int kuwahara;
uniform int doPostProcessing;
uniform sampler2D image;

const float A = 2.51;
const float B = 0.03;
const float C = 2.43;
const float D = 0.59;
const float E = 0.14;

const float MAX_FLOAT = 100;

// Narokowicz ACES tone mapping function
vec3 acesToneMapping(vec3 color)
{
    color = (color * (A * color + B)) / (color * (C * color + D) + E);
    return clamp(color, 0.0, 1.0);
}

void calculatePixelColors(vec2 texel, out vec3 pixelColors[9], out vec3 meanColor, vec2 texelSize) {
    int xOffsets[9] = int[9](-4, -3, -2, -1, 0, 1, 2, 3, 4);
    int yOffsets[9] = int[9](-4, -3, -2, -1, 0, 1, 2, 3, 4);
    vec3 sumPixelColors;

    for (int i = 0; i < 9; ++i) {
        int xOffset = xOffsets[i];
        for (int j = 0; j < 9; ++j) {
            int yOffset = yOffsets[j];

            vec2 localTexel = vec2(texel.x + (xOffset * texelSize.x), texel.y + (yOffset * texelSize.y));
            vec3 pixelColor = texture(image, localTexel).rgb;

            pixelColors[(9 * i) + j] = pixelColor;
            sumPixelColors += pixelColor;
        }
    }

    meanColor = sumPixelColors / (9.0 * 9.0);
    return;
}

float calculateStandardDeviation(vec3 pixelColors[9], vec3 meanColor) {
    // std calculation
    float squaredDiff[9]; 
    float avgSquaredDiff;

    for (int i = 0; i < 9; ++i) {
        squaredDiff[i] = pow(length(pixelColors[i] - meanColor), 2);
        avgSquaredDiff += squaredDiff[i];
    }

    avgSquaredDiff /= 9.0;
    return avgSquaredDiff;
}

void main()
{
    vec3 color = texture(image, TexCoords).rgb;
    vec2 texelSize = 1.0 / textureSize(image, 0);

    if (kuwahara == 1) {
        // int xQuadrant[2] = int[2](-1, 1);
        // int yQuadrant[2] = int[2](-1, 1);
        // int xQuadrant[2] = int[2](-4, 4);
        // int yQuadrant[2] = int[2](-4, 4);
        int xQuadrant[2] = int[2](0, 0);
        int yQuadrant[2] = int[2](0, 0);

        float minStandardDeviation = MAX_FLOAT;
        vec3 minStandardDeviationColor;

        for (int i = 0; i < 2; ++i) {
            for (int j = 0; j < 2; ++j) {
                vec2 texel = vec2(
                    TexCoords.x + (xQuadrant[i] * texelSize.x),
                    TexCoords.y + (yQuadrant[j] * texelSize.y)
                );
                vec3 pixelColors[9];
                vec3 meanColor;

                calculatePixelColors(texel, pixelColors, meanColor, texelSize);
                float standardDeviation = calculateStandardDeviation(pixelColors, meanColor);
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
