#version 330 core

// This shader performs downsampling on a texture,
// as taken from Call Of Duty method, presented at ACM Siggraph 2014.
// This particular method was customly designed to eliminate
// "pulsating artifacts and temporal stability issues".

// Remember to add bilinear minification filter for this texture!
// Remember to use a floating-point texture format (for HDR)!
// Remember to use edge clamping for this texture!
uniform sampler2D srcTexture;
uniform int threshold;

in vec2 texCoord;
layout (location = 0) out vec4 downsample;

void main()
{
    // vec2 srcTexelSize = 1.0 / srcResolution;
    vec2 srcTexelSize = 1.0 / vec2(textureSize(srcTexture, 0));
    float x = srcTexelSize.x;
    float y = srcTexelSize.y;

    // Take 13 samples around current texel:
    // a - b - c
    // - j - k -
    // d - e - f
    // - l - m -
    // g - h - i
    // === ('e' is the current texel) ===

    vec3 a = texture(srcTexture, vec2(texCoord.x - 2*x, texCoord.y + 2*y)).rgb;
    vec3 b = texture(srcTexture, vec2(texCoord.x,       texCoord.y + 2*y)).rgb;
    vec3 c = texture(srcTexture, vec2(texCoord.x + 2*x, texCoord.y + 2*y)).rgb;

    vec3 d = texture(srcTexture, vec2(texCoord.x - 2*x, texCoord.y)).rgb;
    vec3 e = texture(srcTexture, vec2(texCoord.x,       texCoord.y)).rgb;
    vec3 f = texture(srcTexture, vec2(texCoord.x + 2*x, texCoord.y)).rgb;

    vec3 g = texture(srcTexture, vec2(texCoord.x - 2*x, texCoord.y - 2*y)).rgb;
    vec3 h = texture(srcTexture, vec2(texCoord.x,       texCoord.y - 2*y)).rgb;
    vec3 i = texture(srcTexture, vec2(texCoord.x + 2*x, texCoord.y - 2*y)).rgb;

    vec3 j = texture(srcTexture, vec2(texCoord.x - x, texCoord.y + y)).rgb;
    vec3 k = texture(srcTexture, vec2(texCoord.x + x, texCoord.y + y)).rgb;
    vec3 l = texture(srcTexture, vec2(texCoord.x - x, texCoord.y - y)).rgb;
    vec3 m = texture(srcTexture, vec2(texCoord.x + x, texCoord.y - y)).rgb;

    // Apply weighted distribution:
    // 0.5 + 0.125 + 0.125 + 0.125 + 0.125 = 1
    // a,b,d,e * 0.125
    // b,c,e,f * 0.125
    // d,e,g,h * 0.125
    // e,f,h,i * 0.125
    // j,k,l,m * 0.5
    // This shows 5 square areas that are being sampled. But some of them overlap,
    // so to have an energy preserving downsample we need to make some adjustments.
    // The weights are the distributed, so that the sum of j,k,l,m (e.g.)
    // contribute 0.5 to the final color output. The code below is written
    // to effectively yield this sum. We get:
    // 0.125*5 + 0.03125*4 + 0.0625*4 = 1

    vec3 v;
    v = e*0.125;
    v += (a+c+g+i)*0.03125;
    v += (b+d+f+h)*0.0625;
    v += (j+k+l+m)*0.125;

    // testing
    // b = texture(srcTexture, vec2(texCoord.x,       texCoord.y + 1*y)).rgb;
    // d = texture(srcTexture, vec2(texCoord.x - 1*x, texCoord.y)).rgb;
    // f = texture(srcTexture, vec2(texCoord.x + 1*x, texCoord.y)).rgb;
    // h = texture(srcTexture, vec2(texCoord.x,       texCoord.y - 1*y)).rgb;
    // v = (e * 0.1) + 0.1875 * (b + d + f + h);
    // v = f;

    // downsample = vec4(1, 1, 1, 1);
    // downsample = texture(srcTexture, texCoord).xyzw;
    downsample = vec4(v, 1);
    if (threshold == 1) {
        if (downsample.r + downsample.g + downsample.b < 1) {
            downsample = vec4(0, 0, 0, 1);
        }
    }
    // downsample = vec4(texture(srcTexture, vec2(texCoord.x + 1*x, texCoord.y)).rgb, 1);
}