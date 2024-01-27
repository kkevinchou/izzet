uniform sampler2D Texture;

in vec2 Frag_UV;
in vec4 Frag_Color;
uniform int IsFontTexture;

out vec4 Out_Color;

void main()
{
    if (IsFontTexture == 0) {
        Out_Color = texture(Texture, Frag_UV.st).rgba;
    } else {
        Out_Color = vec4(Frag_Color.rgb, Frag_Color.a * texture(Texture, Frag_UV.st).r);
    }
    // Out_Color = vec4(Frag_Color.rgb, Frag_Color.a * texture(Texture, Frag_UV.st).r);
}
