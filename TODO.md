# TODO

## Lighting
* distance fog
* distance blur
* screen Space Ambient Occlusion
* create directional light from menu
* light emitter

## Water
* realistic water rendering like sea of thieves

## Geometry
* experiment with using half edges

## Other
* parent scale transforms should not affect the translation of its children
* refactor MeshSpecification to instead refer to primitives. VAOs are at the primitive level (one material/texture, one set of positions) rather than at the mesh level
* pretty sure either the depth texture (from the light view) or shadow mapping logic is broken
* toggleable spatial partition
* handle processing normal maps
    * we currently only support normals from primitives
    ```
    {
        "attributes": {
        "POSITION": 50,
        "TEXCOORD_0": 51,
        "NORMAL": 52,
        "TANGENT": 53
        },
        "mode": 4,
        "material": 0,
        "indices": 0
    }
    ```
    * but they can be stored in textures in the material definition as well
    ```
    {
      "pbrMetallicRoughness": {
        "baseColorFactor": [
          0.5879999995231628,
          0.5879999995231628,
          0.5879999995231628,
          1
        ],
        "baseColorTexture": {
          "index": 52
        },
        "metallicRoughnessTexture": {
          "index": 49
        }
      },
      "normalTexture": {
        "index": 53
      }
    }
    ```
# Recently Completed