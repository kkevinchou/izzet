package gltf_test

import (
	"fmt"
	"testing"

	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/assets/loaders/gltf"
)

var testFile string = "../../../../_assets/gltf/mannequin_m.gltf"

// bug hint: when a joint is defined but has no poses our
// animation loading code freaks out. i removed the joint animatiosn from the legs
// and it seems to point to the origin afterwards
// this means the original animation looked wonky probably because there was no pose info
// for the joint which our animation loading code did not understand. likely need to see
// how we handled poses where a joint does not have any poses

func TestBasic(t *testing.T) {
	d, err := gltf.ParseGLTF("", testFile, &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL})
	if err != nil {
		t.Error(err)
	}
	_ = d
}

// our current animation system only works on models with keyframes that include all joints (even if redundant).
func TestFullKeyFrames(t *testing.T) {
	m, err := gltf.ParseGLTF("", testFile, &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL})
	if err != nil {
		t.Error(err)
	}

	count := jointCount(m.RootJoint)
	for _, animation := range m.Animations {
		for i, kf := range animation.KeyFrames {
			if len(kf.Pose) != count {
				t.Error(fmt.Sprintf("animation %s on key frame %d has %d joints rather than the expected %d", animation.Name, i, len(kf.Pose), count))
			}
		}

	}
}

func jointCount(joint *modelspec.Joint) int {
	count := 1
	for _, child := range joint.Children {
		count += jointCount(child)
	}
	return count
}
