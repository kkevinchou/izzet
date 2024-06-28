# Context

## Current State

- The last item completed was creating the adjacency list on the triangulated poly meshes in polymesh.go. This is a part of BuildPolyMesh

## Things to do

- rcBuildContours
  - Holes are currently not handled, I just panic
- rcBuildPolyMesh
  - Polygons are all triangles at the moment, I still need to merge them into convex polygons
  - There's potentially some work remaining for removing edge vertices, and finding portal edges but I don't understand that part yet
- rcBuildPolyMeshDetail
  - This is step creates an actual detailed mesh for navigation. The points on the mesh will better track the geometry of the level as it will consider the height level of the polygon via sampling and also create the points in world space.
  - This step hasn't been started yet but is the last step in constructing the navigation mesh

## Edge cases remaining

- The winding order still needs figuring out. A lot of the geometry functions and their tests in polymesh_test.go assume a counter clockwise ordering. However, the Recast algorithm seems to construct contours in a clockwise order.
