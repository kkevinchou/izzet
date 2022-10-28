#version 330 core
out vec4 FragColor;

in VS_OUT {
    vec3 FragPos;
    vec3 Normal;
    vec4 FragPosLightSpace;
    mat4 View;
    vec2 TexCoord;
} fs_in;

uniform vec3 viewPos;
uniform sampler2D modelTexture;
uniform sampler2D shadowMap;
uniform float shadowDistance;
uniform vec3 directionalLightDir;

// pbr materials
uniform int hasPBRMaterial;
uniform int hasPBRBaseColorTexture;
uniform vec4 pbrBaseColorFactor;

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
    float bias = max(0.0000000001 * (1.0 - dot(normal, lightDir)), 0.000000001);
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

void main()
{
    vec3 lightColor = vec3(1.0, 1.0, 1.0);

    // ambient
    float ambientStrength = 0.5;
    vec3 ambient = ambientStrength * lightColor;
    vec3 specular = vec3(0, 0, 0);
        
    // diffuse 
    vec3 normal = normalize(fs_in.Normal);
    vec3 lightDir = normalize(directionalLightDir);
    float diff = max(dot(normal, -lightDir), 0.0);
    vec3 diffuse = diff * lightColor;

    vec3 color;
    if (hasPBRMaterial == 1) {
        color = pbrBaseColorFactor.xyz;
        if (hasPBRBaseColorTexture == 1) {
            color = texture(modelTexture, fs_in.TexCoord).xyz;
        }
    } else {
        color = texture(modelTexture, fs_in.TexCoord).xyz;
    }
    
    // calculate shadow
    float shadow = ShadowCalculation(fs_in.FragPosLightSpace, normal, lightDir);
    vec3 lighting = (ambient + (1.0 - shadow) * (diffuse + specular)) * color; 

    FragColor = vec4(lighting, 1);
}