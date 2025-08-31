#version 330 core

layout (location = 0) out vec4 FragColor;
layout (location = 1) out uint PickingColor;

// material parameters
uniform vec3  albedo;
uniform float metallic;
uniform float roughness;
uniform vec3  translation;
uniform vec3  scale;

uniform int repeatTexture;

// lights
const int MAX_LIGHTS = 10;

struct Light {
    // the light type
    // 0 - directional
    int type;

    // general props
    vec3 diffuse;

    // directional
    vec3 dir;

    // positional
    vec3 position;

    float range;
};

uniform int lightCount;
uniform Light lights[MAX_LIGHTS];

uniform vec3 viewPos;
uniform sampler2D modelTexture;

// ambient occlusion blur
uniform sampler2D ambientOcclusion;

// shadows
uniform sampler2D shadowMap;
uniform float shadowDistance;

// depth map
uniform sampler2D cameraDepthMap;

// point light shadows
uniform samplerCube depthCubeMap;
uniform float far_plane;

uniform float ambientFactor;
uniform float specularFactor;

// pbr materials
uniform int hasPBRBaseColorTexture;

uniform float pointLightBias;
uniform float shadowMapMinBias;
uniform float shadowMapAngleBiasRate;

uniform uint entityID;

const float PI = 3.14159265359;

const vec4 errorColor = vec4(255.0 / 255, 28.0 / 255, 217.0 / 121.0, 1.0);

in VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    vec2 TexCoord;
    vec3 ColorOverride;
} fs_in;

const float A = 2.51;
const float B = 0.03;
const float C = 2.43;
const float D = 0.59;
const float E = 0.14;

uniform int fogDensity;

uniform int width;
uniform int height;
uniform float far;
uniform float near;

uniform int enableAmbientOcclusion;

// ACES tone mapping function
vec3 acesToneMapping(vec3 color)
{
    color = (color * (A * color + B)) / (color * (C * color + D) + E);
    return clamp(color, 0.0, 1.0);
}

float PointLightShadowCalculation(vec3 fragPos, vec3 lightPos)
{
    // get vector between fragment position and light position
    vec3 fragToLight = fragPos - lightPos;
    // ise the fragment to light vector to sample from the depth map    
    float closestDepth = texture(depthCubeMap, fragToLight).r;
    // it is currently in linear range between [0,1], let's re-transform it back to original depth value
    closestDepth *= far_plane;
    // now get current linear depth as the length between the fragment and light position
    float currentDepth = length(fragToLight);
    // test for shadows
    // float bias = 0.05; // we use a much larger bias since depth is now in [near_plane, far_plane] range
    float shadow = currentDepth - pointLightBias > closestDepth ? 1.0 : 0.0;
    // display closestDepth as debug (to visualize depth cubemap)
    // FragColor = vec4(vec3(closestDepth / far_plane), 1.0);    
        
    return shadow;
}

float DirectionalLightShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir)
{
    if (length(vec3(fs_in.View * vec4(fs_in.FragPos, 1))) > shadowDistance) {
        return 0;
    }

    // perform perspective divide
    vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
    // transform to [0,1] range
    projCoords = projCoords * 0.5 + 0.5;

    // get closest depth value from light's perspective (using [0,1] range fragPosLight as coords)
    // float closestDepth = texture(shadowMap, projCoords.xy).r; // QUESTION: why is it .r? is it because it's a grayscale texture?
    // get depth of current fragment from light's perspective
    float currentDepth = projCoords.z;
    // check whether current frag pos is in shadow
    // float shadow = currentDepth > closestDepth  ? 1.0 : 0.0;

    // bias term needs to be tweaked depending on geometry
    float bias = max(shadowMapAngleBiasRate * (1.0 - dot(normal, lightDir)), shadowMapMinBias);
    // bias = 0;
    
    float shadow = 0.0;
    vec2 texelSize = 1.0 / textureSize(shadowMap, 0);
    for(int x = -1; x <= 1; ++x)
    {
        for(int y = -1; y <= 1; ++y)
        {
            float pcfDepth = texture(shadowMap, projCoords.xy + vec2(x, y) * texelSize).r; 
            shadow += currentDepth - bias > pcfDepth ? 1.0 : 0.0;        
        }    
    }
    shadow /= 9.0;

    return shadow;
}

float DistributionGGX(vec3 N, vec3 H, float roughness) {
    float a      = roughness*roughness;
    float a2     = a*a;
    float NdotH  = max(dot(N, H), 0.0);
    float NdotH2 = NdotH*NdotH;
	
    float num   = a2;
    float denom = (NdotH2 * (a2 - 1.0) + 1.0);
    denom = PI * denom * denom;
	
    return num / denom;
}

float GeometrySchlickGGX(float NdotV, float roughness) {
    float r = (roughness + 1.0);
    float k = (r*r) / 8.0;

    float num   = NdotV;
    float denom = NdotV * (1.0 - k) + k;
	
    return num / denom;
}

float GeometrySmith(vec3 N, vec3 V, vec3 L, float roughness) {
    float NdotV = max(dot(N, V), 0.0);
    float NdotL = max(dot(N, L), 0.0);
    float ggx2  = GeometrySchlickGGX(NdotV, roughness);
    float ggx1  = GeometrySchlickGGX(NdotL, roughness);
	
    return ggx1 * ggx2;
}

vec3 fresnelSchlick(float cosTheta, vec3 F0) {
    return F0 + (1.0 - F0) * pow(clamp(1.0 - cosTheta, 0.0, 1.0), 5.0);
}  

vec3 calculateLightOut(vec3 normal, vec3 fragToCam, vec3 fragToLight, float lightDistance, vec3 lightColor, vec3 in_albedo, int do_attenuation) {
    // vec3 F0 = vec3(0.04);
    // F0 = mix(F0, in_albedo, metallic);

    float ior = 1.5;
    vec3 specularColorFactor = vec3(1.0);

    // Base dielectric F0 from IOR:
    float f0_dielectric_scalar = pow((ior - 1.0) / (ior + 1.0), 2.0); // ~0.04 when ior=1.5

    // KHR_materials_specular scales & tints the dielectric F0:
    vec3 F0_dielectric = f0_dielectric_scalar * specularFactor * specularColorFactor;

    // Final F0 mixes between dielectric and conductor (metal) paths:
    vec3 F0 = mix(F0_dielectric, in_albedo, metallic);

    vec3 H = normalize(fragToCam + fragToLight);

    float NdotL = max(dot(normal, fragToLight), 0.00001);
    float NdotV = max(dot(normal, fragToCam), 0.00001);

    // float attenuation = 1.0 / (1 + 0.01 * lightDistance + 0.001 * (lightDistance * lightDistance));
    float attenuation = 1.0 / (lightDistance * lightDistance);
    if (do_attenuation == 0) {
        attenuation = 1.0;
    }

    // incoming light energy
    vec3 radiance = lightColor * attenuation; 
    
    // cook-torrance brdf - used for specular calculation
    float D = DistributionGGX(normal, H, roughness); // what proportion of microfacts are aligned with the bisecting vector, causing light to bounce towards the camera
    float G = GeometrySmith(normal, fragToCam, fragToLight, roughness); // how much of the microfacets are self shadowing
    vec3 F = fresnelSchlick(max(dot(H, fragToCam), 0.0), F0); // how much energy is reflected in a specular fashion
    
    vec3 spec_numerator = D * G * F;
    float spec_denominator = 4.0 * NdotV * NdotL;
    vec3 specular = spec_numerator / spec_denominator;  

    // kS + kD sum to 1 to conserve energy
    vec3 kS = F;
    vec3 kD = vec3(1.0) - kS;

    // only non-metals (or partial metals) have diffuse lighting
    kD *= 1.0 - metallic;
        
    vec3 diffuse = kD * in_albedo / PI;

    return (diffuse + specular) * radiance * NdotL;
}

// OpenGL does not use linear scaling of depth values. Close objects have very noticeable affect
// on depth values while objects further away quickly approach 1.0
float depthValueToLinearDistance(float depth) {
    float ndc = depth * 2.0 - 1.0;
    float linearDepth = (2.0 * near * far) / (far + near - ndc * (far - near));
    return linearDepth;
}

// float linearFog(float dist) {
//     return 1 - (fogEnd - dist) / (fogEnd - fogStart);
// }

// float exponentialFog(float dist, float density) {
//     return 1 - pow(2, -dist * density);
// }

float exponentialSquaredFog(float dist, float density) {
    return 1 - pow(2, -pow(dist * density, 2));
}

void main()
{		
    vec3 normal = normalize(fs_in.Normal);
    float alpha = 1.0;
	           
    // reflectance equation
    vec3 Lo = vec3(0.0);
    vec3 in_albedo = albedo;
    if (hasPBRBaseColorTexture == 1) {
        vec4 texture_value = texture(modelTexture, fs_in.TexCoord);

        if (repeatTexture == 1) {
            vec2 uv;
            vec3 worldPos = fs_in.FragPos;
            vec3 absNormal = abs(normal);

            float textureFactor = 4;

            if (absNormal.z >= absNormal.x && absNormal.z >= absNormal.y) {
                // Front/Back (Z)
                uv = worldPos.xy / textureFactor - translation.xy / textureFactor;
                if (normal.z < 0.0) {
                    uv += vec2(-0.5 * scale.x, -0.5 * scale.y) / textureFactor;
                    uv.x = 1.0 - uv.x;
                } else {
                    uv += vec2(0.5 * scale.x, -0.5 * scale.y) / textureFactor;
                }
            } else if (absNormal.x >= absNormal.y && absNormal.x >= absNormal.z) {
                // Left/Right (X)
                uv = worldPos.zy / textureFactor - translation.zy / textureFactor;
                if (normal.x > 0.0) {
                    uv += vec2(-0.5 * scale.z, -0.5 * scale.y) / textureFactor;
                    uv.x = 1.0 - uv.x;
                } else {
                    uv += vec2(0.5 * scale.z, -0.5 * scale.y) / textureFactor;
                }
            } else {
                // Top/Bottom (Y)
                uv = worldPos.xz / textureFactor - translation.xz / textureFactor;
                if (normal.y > 0.0) {
                    uv += vec2(0.5 * scale.x, 0.5 * scale.z) / textureFactor;
                    uv.y = 1.0 - uv.y;
                } else {
                    uv += vec2(0.5 * scale.x, -0.5 * scale.z) / textureFactor;
                }
            }

            texture_value = texture(modelTexture, uv);
        }

        in_albedo = in_albedo * texture_value.xyz;
        alpha = texture_value.a;
    }

    // failsafe for when we pass in too many lights, i hope you like hot pink
    if (lightCount > MAX_LIGHTS) {
        FragColor = errorColor;
        return;
    }

    bool firstPointLight = true;
    for(int i = 0; i < lightCount; ++i) {
        Light light = lights[i];
        vec3 fragToCam = normalize(viewPos - fs_in.FragPos);
        float distance = length(light.position - fs_in.FragPos);
        
        vec3 fragToLight;
        float shadow = 0;
        int do_attenuation = 1;

        if (light.type == 0) {
            // directional light case
            fragToLight = -normalize(light.dir);
            shadow = DirectionalLightShadowCalculation(fs_in.FragPosLightSpace, normal, fragToLight);
            do_attenuation = 0;
        } else if (light.type == 1) {
            if (distance > light.range) {
                continue;
            }
            fragToLight = normalize(light.position - fs_in.FragPos);
            // we only support shadows for the first point light for now
            if (firstPointLight) {
                shadow = PointLightShadowCalculation(fs_in.FragPos, light.position);
                firstPointLight = false;
            }
        } else {
            FragColor = errorColor;
            return;
        }

        // in gltf 2.0 if we have both the base color factor and base color texture defined
        // the base color factor is a linear multiple of the texture values
        // https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.html#metallic-roughness-material

        Lo += (1 - shadow) * calculateLightOut(normal, fragToCam, fragToLight, distance, light.diffuse, in_albedo, do_attenuation);
    }
  
    vec3 ambient = vec3(ambientFactor) * in_albedo;

    if (enableAmbientOcclusion == 1) {
        ambient = ambient * texture(ambientOcclusion, (gl_FragCoord.xy / vec2(width, height))).r;
    }

    vec3 color = ambient + Lo;
	
    FragColor = vec4(color, 1.0);
    // FragColor = vec4(color, alpha);

    if (fogDensity > 0) {
        vec2 textureCoords = gl_FragCoord.xy / vec2(width, height);
        float depth = texture(cameraDepthMap, textureCoords).r;
        float dist = depthValueToLinearDistance(depth);

        float fogFactor = exponentialSquaredFog(dist, float(fogDensity) / 50000);
        fogFactor = clamp(fogFactor, 0.0, 1.0);

        // alpha blending doesn't work quite right with orgrimmar.
        // using the actual alpha seems to mess up loading the level, so we hard code it to 1.

        // FragColor = vec4(mix(color, vec3(1,1,1), fogFactor), alpha);
        FragColor = vec4(mix(color, vec3(1,1,1), fogFactor), 1.0);
    }

    PickingColor = entityID;
}
