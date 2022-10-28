#version 330 core
out vec4 FragColor;

// material parameters
uniform vec3  albedo;
uniform float metallic;
uniform float roughness;
uniform float ao;

// lights
uniform vec3 lightPositions[4];
uniform vec3 lightColors[4];
uniform vec3 directionalLightDir;

uniform vec3 viewPos;
uniform sampler2D modelTexture;

// shadows
uniform sampler2D shadowMap;
uniform float shadowDistance;

// pbr materials
uniform int hasPBRMaterial;
uniform int hasPBRBaseColorTexture;
uniform vec4 pbrBaseColorFactor;

const float PI = 3.14159265359;

in VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    vec2 TexCoord;
} fs_in;

float ShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir)
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

vec3 calculateLightOut(vec3 normal, vec3 fragToCam, vec3 fragToLight, float lightDistance, vec3 lightColor, vec3 in_albedo) {
    vec3 F0 = vec3(0.04); 
    F0 = mix(F0, in_albedo, metallic);

    // calculate per-light radiance
    vec3 H = normalize(fragToCam + fragToLight);
    float attenuation = 1.0 / (lightDistance * lightDistance);
    vec3 radiance     = lightColor * attenuation;        
    
    // cook-torrance brdf
    float NDF = DistributionGGX(normal, H, roughness);        
    float G   = GeometrySmith(normal, fragToCam, fragToLight, roughness);      
    vec3 F    = fresnelSchlick(max(dot(H, fragToCam), 0.0), F0);       
    
    vec3 kS = F;
    vec3 kD = vec3(1.0) - kS;
    kD *= 1.0 - metallic;	  
    
    vec3 numerator    = NDF * G * F;
    float denominator = 4.0 * max(dot(normal, fragToCam), 0.0) * max(dot(normal, fragToLight), 0.0) + 0.0001;
    vec3 specular     = numerator / denominator;  
        
    // add to outgoing radiance Lo
    float NdotL = max(dot(normal, fragToLight), 0.0);                
    return (kD * in_albedo / PI + specular) * radiance * NdotL; 
}

void main()
{		
    vec3 normal = normalize(fs_in.Normal);
	           
    // reflectance equation
    vec3 Lo = vec3(0.0);
    // for(int i = 0; i < 4; ++i) {
        // vec3 fragToLight = normalize(lightPositions[i] - fs_in.FragPos);
        vec3 fragToCam = normalize(viewPos - fs_in.FragPos);
        // float distance = length(lightPositions[i] - fs_in.FragPos);
        
        float distance = 1;
        vec3 fragToLight = -directionalLightDir;
        vec3 lightColor = vec3(5);
        float shadow = ShadowCalculation(fs_in.FragPosLightSpace, normal, fragToLight);
        // shadow = 0;

        vec3 in_albedo = albedo; 
        if (hasPBRBaseColorTexture == 1) {
            in_albedo = texture(modelTexture, fs_in.TexCoord).xyz;
        }
        Lo += (1 - shadow) * calculateLightOut(normal, fragToCam, fragToLight, distance, lightColor, in_albedo);
    // }   
  
    vec3 ambient = vec3(0.1) * in_albedo * ao;
    // ambient = vec3(0);
    vec3 color = ambient + Lo;
	
    color = color / (color + vec3(1.0));

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
   
    FragColor = vec4(color, 1.0);
}
