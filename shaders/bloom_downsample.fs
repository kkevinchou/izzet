#version 330 core

// This shader performs downsampling on a texture,
// as taken from Call Of Duty method, presented at ACM Siggraph 2014.
// This particular method was customly designed to eliminate
// "pulsating artifacts and temporal stability issues".

// Remember to add bilinear minification filter for this texture!
// Remember to use a floating-point texture format (for HDR)!
// Remember to use edge clamping for this texture!
uniform sampler2D srcTexture;
uniform int bloomThresholdEnabled;
uniform float bloomThreshold;
uniform int karis;

in vec2 texCoord;
layout (location = 0) out vec4 downsample;

vec3 PowVec3(vec3 v, float p)
{
    return vec3(pow(v.x, p), pow(v.y, p), pow(v.z, p));
}

const float invGamma = 1.0 / 2.2;
vec3 ToSRGB(vec3 v) { return PowVec3(v, invGamma); }

float RGBToLuminance(vec3 col)
{
    return dot(col, vec3(0.2126f, 0.7152f, 0.0722f));
}

float KarisAverage(vec3 col)
{
    // Formula is 1 / (1 + luma)
    float luma = RGBToLuminance(ToSRGB(col)) * 0.25f;
    return 1.0f / (1.0f + luma);
}

void main()
{
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

    vec3 v;
    v = e*0.125;
    v += (a+c+g+i)*0.03125;
    v += (b+d+f+h)*0.0625;
    v += (j+k+l+m)*0.125;

    if (karis == 1) {
        vec3 groups[5];
        groups[0] = (a+b+d+e) * (0.125f/4.0f);
        groups[1] = (b+c+e+f) * (0.125f/4.0f);
        groups[2] = (d+e+g+h) * (0.125f/4.0f);
        groups[3] = (e+f+h+i) * (0.125f/4.0f);
        groups[4] = (j+k+l+m) * (0.5f/4.0f);
        groups[0] *= KarisAverage(groups[0]);
        groups[1] *= KarisAverage(groups[1]);
        groups[2] *= KarisAverage(groups[2]);
        groups[3] *= KarisAverage(groups[3]);
        groups[4] *= KarisAverage(groups[4]);
        v = groups[0]+groups[1]+groups[2]+groups[3]+groups[4];
        v = max(v, 0.0001f);
    }

    if (bloomThresholdEnabled == 1) {
        if (RGBToLuminance(v) < bloomThreshold) {
            v /= 2;
        }
    }

    downsample = vec4(v, 1);
}