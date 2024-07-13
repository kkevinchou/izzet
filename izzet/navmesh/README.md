# Context

## Current State

- The last item completed was creating the adjacency list on the triangulated poly meshes in polymesh.go. This is a part of BuildPolyMesh

## Things to do

- rcBuildRegions
  - need to implement region merging
- rcBuildContours
  - Holes are currently not handled, I just panic
- rcBuildPolyMesh
  - some border specific logic
    - remove vertex
    - find portals
- rcBuildPolyMeshDetail
  - todo
    - implement seedArrayWithPolyCenter
  - steps
    - for each polygon
      - create a height patch
      - getHeightData
        - fill the height patch with height data from the compact height field
      - builPolyDetail
        - calculate the minimum extent
        - for each edge
          - create samples along the edge
            - the number of samples is based on sampleDist
          - simplify the samples
            - iterate through the samples
            - if the max deviation is larger than the accepted error, add the sample with the largest deviation to the list of points. otherwise, move on to the next edge
          - add the vertices to the hull
        - if the polygon minimum extent is small (sliver or small triangle), do not try to add internal points
          - triangulate the hull
          - for each triangle, set a flag to mark it if one of its edges lies on the hull
        - triangulate the hull
        - sample the polygon in a grid arrangement
        - add vertices that have the most error, repeat until error threshold is met, or all samples were added
        - create a delauney triangulation based on the added vertex
          - TODO: do an incremental add instead of full rebuiild
    - move vertices to world space
    - store vertices in the detailed poly mesh
    - store triangles in the detailed poly mesh
      - seems like it doesn't store the polygons but triangles only??
