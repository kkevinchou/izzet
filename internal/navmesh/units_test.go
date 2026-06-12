package navmesh

import "testing"

func TestWorldUnitSettingsToVoxels(t *testing.T) {
	if got, want := WorldRadiusToVoxels(0.4, 0.1), 4; got != want {
		t.Fatalf("radius voxels: got %d, want %d", got, want)
	}

	if got, want := WorldHeightToVoxels(1.8, 0.1), 18; got != want {
		t.Fatalf("height voxels: got %d, want %d", got, want)
	}

	if got, want := WorldClimbToVoxels(0.3, 0.1), 3; got != want {
		t.Fatalf("climb voxels: got %d, want %d", got, want)
	}
}

func TestWorldUnitSettingsRoundForConservativeBake(t *testing.T) {
	if got, want := WorldRadiusToVoxels(0.41, 0.1), 5; got != want {
		t.Fatalf("radius voxels: got %d, want %d", got, want)
	}

	if got, want := WorldHeightToVoxels(1.81, 0.1), 19; got != want {
		t.Fatalf("height voxels: got %d, want %d", got, want)
	}

	if got, want := WorldClimbToVoxels(0.39, 0.1), 3; got != want {
		t.Fatalf("climb voxels: got %d, want %d", got, want)
	}
}
