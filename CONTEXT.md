# Context

## GLTF Animation exporting
When exporting to GLTF format, the blender exporter has a default on setting called "Optimize Animation Size". Which has a description of:
"Reduces exported filesize by removing duplicate keyframes. Can cause problems with stepped animation". Our current animation player can
only handle animated models that have keyframes for all joints (and probably for all frames as well). So, to fix my broken animations i
had to uncheck that setting. With it checked, the exported animations would have the first keyframe having transforms for all joints, while
subsequent keyframes would only have transforms for a portion of the joints. Presumably because it dropped duplicate keyframes. The GLTF spec
probably describes how to properly handle these scenarios (the vscode model renderer handles this properly). So this will be something for me
to implement in the GLTF loader or animation player. My exporting steps are to select nothing and export with "Optimize Animation Size" disabled.
It looks like this feature was added in for newer versions of blender, but when I was doing development this option was not available. Each
animation should be stashed in the same NLA track (i think)

## Collision Resolution reaching max on an entity (10)
Seems like when we are resolving collisions on slopes that cause jitter we hit the collision resolution max somehow

## GLTF animation bugs
Currently we don't handle models with multiple roots properly - we are assuming there is only one root. This is problematic because some models
can have multiple roots - e.g. models using IK

## GLTF models from mixamo
Attempting to scale models from mixamo by scaling the root armature and attempting to apply it to the model (make the scale <1, 1, 1>) will break
animations. There isn't a good way of resolvintg this at the moment. What this means in practice is that we can scale the model down, but we can't
apply the scale. This is fine for the engine because we parse the root transform from the model and we apply it to the model matrix before rendering.
Something to note here is that the vertex positions are properly scaled down even though we don't apply the scaling. i.e. if we construct a capsule
collider from the model based on its vertices, it is the correctly sized collider and matches up to the animated version of the model (which includes
the root transform matrix)