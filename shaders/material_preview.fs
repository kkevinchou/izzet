#version 330 core

in  vec2 vNDC;
out vec4 FragColor;

uniform vec3   uAlbedo;
uniform sampler2D uAlbedoMap;
uniform bool   uUseAlbedoMap;

uniform float  uMetallic;
uniform sampler2D uMetallicMap;
uniform bool   uUseMetallicMap;

uniform float  uRoughness;
uniform sampler2D uRoughnessMap;
uniform bool   uUseRoughnessMap;

uniform float  uAO;
uniform sampler2D uAOMap;
uniform bool   uUseAOMap;

uniform vec3   uLightDir;    // normalized, view-space
uniform vec3   uLightColor;

const float PI = 3.14159265359;

// ==== PBR helper functions ====
float DistributionGGX(vec3 N, vec3 H, float roughness) {
    float a      = roughness*roughness;
    float a2     = a*a;
    float NdotH  = max(dot(N, H), 0.0);
    float denom  = (NdotH*NdotH*(a2 - 1.0) + 1.0);
    return a2 / (PI * denom * denom);
}

float GeometrySchlickGGX(float NdotV, float roughness) {
    float r = (roughness + 1.0);
    float k = (r*r) / 8.0;
    return NdotV / (NdotV*(1.0 - k) + k);
}

float GeometrySmith(vec3 N, vec3 V, vec3 L, float roughness) {
    float NdotV = max(dot(N, V), 0.0);
    float NdotL = max(dot(N, L), 0.0);
    float ggx1  = GeometrySchlickGGX(NdotV, roughness);
    float ggx2  = GeometrySchlickGGX(NdotL, roughness);
    return ggx1 * ggx2;
}

vec3 FresnelSchlick(float cosTheta, vec3 F0) {
    return F0 + (1.0 - F0) * pow(1.0 - cosTheta, 5.0);
}

// ==== Main ====
void main() {
    // 1) Sphere mask & normal
    float x = vNDC.x;
    float y = vNDC.y;
    float r2 = x*x + y*y;
    if (r2 > 1.0) discard;
    float z = sqrt(1.0 - r2);
    vec3 N = normalize(vec3(x, y, z));

    // 2) Spherical UV (for textures)
    vec2 uv = vec2(
        0.5 + atan(N.z, N.x)/(2.0*PI),
        acos(N.y)/PI
    );

    // 3) Material inputs
    vec3 albedo    = uUseAlbedoMap    ? texture(uAlbedoMap,    uv).rgb : uAlbedo;
    float metallic  = uUseMetallicMap  ? texture(uMetallicMap,  uv).r   : uMetallic;
    float roughness = uUseRoughnessMap ? texture(uRoughnessMap, uv).r   : uRoughness;
    float ao        = uUseAOMap        ? texture(uAOMap,        uv).r   : uAO;

    // 4) Geometry vectors
    vec3 V = vec3(0.0, 0.0, 1.0);         // camera looks down +Z
    vec3 L = normalize(uLightDir);       // directional light
    vec3 H = normalize(V + L);

    float NdotL = max(dot(N, L), 0.0);
    float NdotV = max(dot(N, V), 0.0);

    // 5) Fresnel reflectance at normal incidence
    vec3 F0 = mix(vec3(0.04), albedo, metallic);
    vec3 F  = FresnelSchlick(max(dot(H, V), 0.0), F0);

    // 6) Cook–Torrance BRDF
    float D = DistributionGGX(N, H, roughness);
    float G = GeometrySmith(N, V, L, roughness);
    vec3  numerator   = D * G * F;
    float denominator = 4.0 * NdotV * NdotL + 0.0001;
    vec3  specular    = numerator / denominator;

    // 7) Energy conservation: diffuse factor
    vec3 kS = F;
    vec3 kD = (1.0 - kS) * (1.0 - metallic);
    vec3 diffuse = kD * albedo / PI;

    // 8) Ambient (simple AO-modulated)
    vec3 ambient = vec3(0.03) * albedo * ao;

    // 9) Final color
    vec3 Lo = (diffuse + specular) * uLightColor * NdotL;
    vec3 color = ambient + Lo;

    // 10) Gamma correction
    color = color / (color + vec3(1.0));
    color = pow(color, vec3(1.0/2.2));

    FragColor = vec4(color, 1.0);
}