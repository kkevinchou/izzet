#version 330 core

layout (location = 0) out vec4 FragColor;
layout (location = 1) out uint PickingColor;

// material parameters
uniform vec3  albedo;
uniform float metallic;
uniform float roughness;
uniform float ao; // ambient occlusion

// lights
const int MAX_LIGHTS = 10;

struct Light {
    // the light type
    // 0 - directional
    int type;

    // general props
    vec4 diffuse;

    // directional
    vec3 dir;

    // positional
    vec3 position;
};

uniform int lightCount;
uniform Light lights[MAX_LIGHTS];

uniform vec3 viewPos;
uniform sampler2D modelTexture;

// shadows
uniform sampler2D shadowMap;
uniform float shadowDistance;

// point light shadows
uniform samplerCube depthCubeMap;
uniform float far_plane;

uniform float ambientFactor;

// pbr materials
uniform int hasPBRMaterial;
uniform int hasPBRBaseColorTexture;
uniform vec4 pbrBaseColorFactor;
uniform float bias;

uniform uint entityID;

uniform int applyToneMapping;

const float PI = 3.14159265359;

const vec4 errorColor = vec4(255.0 / 255, 28.0 / 255, 217.0 / 121.0, 1.0);

in VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    vec2 TexCoord;
} fs_in;

const float A = 2.51;
const float B = 0.03;
const float C = 2.43;
const float D = 0.59;
const float E = 0.14;

// ACES tone mapping function
vec3 acesToneMapping(vec3 color)
{
    color = (color * (A * color + B)) / (color * (C * color + D) + E);
    return clamp(color, 0.0, 1.0);
}

const mat3 ACESInputMat = mat3
(
    0.59719, 0.35458, 0.04823,
    0.07600, 0.90834, 0.01566,
    0.02840, 0.13383, 0.83777
);

// ODT_SAT => XYZ => D60_2_D65 => sRGB
const mat3 ACESOutputMat = mat3
(
     1.60475, -0.53108, -0.07367,
    -0.10208,  1.10813, -0.00605,
    -0.00327, -0.07276,  1.07602
);

vec3 RRTAndODTFit(vec3 v)
{
    vec3 a = v * (v + 0.0245786f) - 0.000090537f;
    vec3 b = v * (0.983729f * v + 0.4329510f) + 0.238081f;
    return a / b;
}

vec3 ACESFitted(vec3 color) {
    color = ACESInputMat * color;

    // Apply RRT and ODT
    color = RRTAndODTFit(color);

    color = ACESOutputMat * color;

    // Clamp to [0, 1]
    color = clamp(color, 0, 1);

    return color;
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
    float shadow = currentDepth -  bias > closestDepth ? 1.0 : 0.0;        
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
    float bias = max(0.00025 * (1.0 - dot(normal, lightDir)), 0.00005);
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
    vec3 F0 = vec3(0.04); 
    F0 = mix(F0, in_albedo, metallic);

    // calculate per-light radiance
    vec3 H = normalize(fragToCam + fragToLight);

    // float attenuation = 1.0 / (1 + 0.01 * lightDistance + 0.001 * (lightDistance * lightDistance));
    float attenuation = 1.0 / (lightDistance * lightDistance);
    if (do_attenuation == 0) {
        attenuation = 1.0;
    }

    vec3 radiance = lightColor * attenuation; 
    
    // cook-torrance brdf
    float NDF = DistributionGGX(normal, H, roughness); // what proportion of microfacts are aligned with the bisecting vector, causing light to bounce towards the camera
    float G = GeometrySmith(normal, fragToCam, fragToLight, roughness); // how much of the microfacets are self shadowing
    vec3 F = fresnelSchlick(max(dot(H, fragToCam), 0.0), F0); // how much energy is reflected in a specular fashion
    
    vec3 numerator = NDF * G * F;
    float denominator = 4.0 * max(dot(normal, fragToCam), 0.0) * max(dot(normal, fragToLight), 0.0) + 0.0001;
    vec3 specular = numerator / denominator;  

    vec3 kD = vec3(1.0) - F;
    kD *= 1.0 - metallic;
        
    // add to outgoing radiance Lo
    float NdotL = max(dot(normal, fragToLight), 0.0);

    return (kD * in_albedo / PI + specular) * radiance * NdotL;
}

void main()
{		
    vec3 normal = normalize(fs_in.Normal);
	           
    // reflectance equation
    vec3 Lo = vec3(0.0);
    vec3 in_albedo = albedo;
    if (hasPBRBaseColorTexture == 1) {
        in_albedo = in_albedo * texture(modelTexture, fs_in.TexCoord).xyz;
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

        vec3 lightColor = vec3(light.diffuse) * light.diffuse.w; // multiply color by intensity

        // in gltf 2.0 if we have both the base color factor and base color texture defined
        // the base color factor is a linear multiple of the texture values
        // https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.html#metallic-roughness-material

        Lo += (1 - shadow) * calculateLightOut(normal, fragToCam, fragToLight, distance, lightColor, in_albedo, do_attenuation);
    }
  
    vec3 ambient = vec3(ambientFactor) * in_albedo * ao;
    vec3 color = ambient + Lo;
	
    if (applyToneMapping == 1) {
        // HDR tone mapping
        // color = color / (color + vec3(1.0));

        color = acesToneMapping(color);

        // Gamma correction
        // unclear if we actually need to do gamma correction. seems like GLTF expects us to internally
        // store textures in SRGB format which we then need to gamma correct herea.
        // PARAMETERS:
        //     gl.Enable(gl.FRAMEBUFFER_SRGB)
        //         OpenGL setting for how the fragment shader outputs colors
        //     lightColor
        //         The color of the light. i've tested with (1, 1, 1) to (20, 20, 20)
        //     gamma correction in the fragment shader
        //         I've experimented with enabling/disabling. it seems like if i gamma correct
        //         I want to disable the OpenGL setting, and if I don't, I want to enable it instead.
        color = pow(color, vec3(1.0/2.2));
    }

    FragColor = vec4(color, 1.0);
    PickingColor = entityID;
}
