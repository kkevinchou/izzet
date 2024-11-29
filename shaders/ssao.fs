#version 330 core

layout (location = 0) out vec4 FragColor;
  
in vec2 TexCoords;

uniform sampler2D gPosition;
uniform sampler2D gNormal;
uniform sampler2D texNoise;

// parameters (you'd probably want to use them as uniforms to more easily tweak the effect)
int kernelSize = 64;

uniform float radius;
uniform float bias;

uniform vec3 samples[64];
uniform mat4 projection;

// tile noise texture over screen, based on screen dimensions divided by noise size
const vec2 noiseScale = vec2(1639.0/4.0, 1024.0/4.0); // screen = 800x600

void main() {
    vec3 fragPos   = texture(gPosition, TexCoords).xyz;
    vec3 normal    = normalize(texture(gNormal, TexCoords).rgb);
    vec3 randomVec = normalize(texture(texNoise, TexCoords * noiseScale).xyz);

    vec3 tangent   = normalize(randomVec - normal * dot(randomVec, normal));
    vec3 bitangent = cross(normal, tangent);
    mat3 TBN       = mat3(tangent, bitangent, normal);  

    float occlusion = 0.0;
    for(int i = 0; i < kernelSize; ++i) {
        // get sample position
        vec3 samplePos = TBN * samples[i]; // from tangent to view-space
        samplePos = fragPos + samplePos * radius; 

        // project sample position (to sample texture) (to get position on screen/texture)
        vec4 offset = vec4(samplePos, 1.0);
        offset = projection * offset; // from view to clip-space
        offset.xyz /= offset.w; // perspective divide
        offset.xyz = offset.xyz * 0.5 + 0.5; // transform to range 0.0 - 1.0
        
        // get sample depth
        float sampleDepth = texture(gPosition, offset.xy).z; // get depth value of kernel sample
        
        // range check & accumulate
        // float rangeCheck = smoothstep(0.0, 1.0, radius / length(fragPos - sampleDepth));
        float rangeCheck = smoothstep(0.0, 1.0, radius / abs(fragPos.z - sampleDepth));
        occlusion += (sampleDepth >= samplePos.z + bias ? 1.0 : 0.0) * rangeCheck;      
    }
    occlusion = 1.0 - (occlusion / kernelSize);
    
    FragColor = vec4(occlusion, occlusion, occlusion, 1);

    // FragColor = vec4(1 - fragPos.z, 1 - fragPos.z, 1 - fragPos.z, 1);
    // FragColor = vec4(1, 1, 1, 1);
}