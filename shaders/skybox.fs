#version 330 core

in vec2 FragPos;
out vec4 FragColor;
float iTime = 7;

vec3 calc(float x, vec3 a, vec3 b, vec3 c, vec3 d)
{
    // sin(1/x) suggested by Phillip Trudeau
    return (b - d) * sin(1. / (vec3(x) / c + 2. / radians(180.) - a)) + d;
}

void main()
{
    vec4 fragColor = FragColor;
    vec2 fragCoord = FragPos;

    vec2 uv = (FragPos + vec2(1, 1)) / 2;

    vec3 p_dark[4] = vec3[4](
        vec3(0.3720705374951474, 0.3037080684557225, 0.26548632969565816),
        vec3(0.446163834012046, 0.39405890487346595, 0.425676737673072),
        vec3(0.16514907579431481, 0.40461292460006665, 0.8799446225003938),
        vec3(-7.057075230154481e-17, -0.08647963850488945, -0.269042973306185)
        // vec3(-7.057075230154481e-17, -0.08647963850488945, 0)
    );

    vec3 p_bright[4] = vec3[4](
        vec3( 0.38976745480184677, 0.31560358280318124,  0.27932656874),
        vec3( 1.2874522895367628,  1.0100154283349794,   0.862325457544),
        vec3( 0.12605043174959588, 0.23134451619328716,  0.526179948359),
        vec3(-0.0929868539256387, -0.07334463258550537, -0.192877259333)
        // vec3(-0.0929868539256387, -0.07334463258550537, 0)
    );

    float x = .3 + .7 * sin(uv.x * radians(60.) + (iTime - 4.) * radians(30.));

    vec3 a = mix(p_dark[0], p_bright[0], x);
    vec3 b = mix(p_dark[1], p_bright[1], x);
    vec3 c = mix(p_dark[2], p_bright[2], x);
    vec3 d = mix(p_dark[3], p_bright[3], x);

    vec3 col = calc(uv.y, a, b, c, d);
    fragColor = vec4(col, 1.);
    FragColor = fragColor;
}