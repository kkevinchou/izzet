// Package physics provides a small rigid-body simulation for spheres and cubes.
//
// Bodies are addressed by BodyID handles. Use World.CreateSphere,
// World.CreateCube, or World.CreateBox to add bodies, World.Step or
// World.Simulate to advance the simulation, and World.Position, World.Rotation,
// or World.Transform to read the current body state.
package physics
