package assets

import (
	"encoding/json"
	"strings"

	"github.com/kkevinchou/izzet/internal/modelspec"
)

var (
	defaultMaterialHandle = MaterialHandle{id: "custom/default"}
	whiteMaterialHandle   = MaterialHandle{id: "custom/white"}
	defaultCubeHandle     = MeshHandle{namespace: "global", id: "cube"}
)

type AnimationHandle struct {
	id string
}

func (h AnimationHandle) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(h.id))
}

func (h *AnimationHandle) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}
	*h = AnimationHandle{id: id}
	return nil
}

type MeshHandle struct {
	namespace string
	id        string
}

func (h MeshHandle) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Namespace string
		ID        string
	}{
		Namespace: string(h.namespace),
		ID:        string(h.id),
	})
}

func (h *MeshHandle) UnmarshalJSON(data []byte) error {
	var value struct {
		Namespace string
		ID        string
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*h = MeshHandle{namespace: value.Namespace, id: value.ID}
	return nil
}

type MaterialHandle struct {
	id string
}

func newSingleEntityMeshHandle(namespace string) MeshHandle {
	return MeshHandle{namespace: namespace, id: "__merged__"}
}

func (h MaterialHandle) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID string
	}{
		ID: string(h.id),
	})
}

func (h *MaterialHandle) UnmarshalJSON(data []byte) error {
	var value struct {
		ID string
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*h = MaterialHandle{id: value.ID}
	return nil
}

func (m *AssetManager) DefaultMaterialHandle() MaterialHandle {
	return defaultMaterialHandle
}

func (m *AssetManager) WhiteMaterialHandle() MaterialHandle {
	return whiteMaterialHandle
}

func (m *AssetManager) DefaultCubeHandle() MeshHandle {
	return defaultCubeHandle
}

func (m *AssetManager) GetSingleEntityMeshHandle(namespace string) MeshHandle {
	return newSingleEntityMeshHandle(namespace)
}

func (m *AssetManager) GetDocumentMeshHandle(namespace string, id string) MeshHandle {
	return MeshHandle{namespace: namespace, id: id}
}

func (m *AssetManager) MeshHandleReferencesDocument(meshHandle MeshHandle, namespace string) bool {
	return string(meshHandle.namespace) == namespace
}

func (m *AssetManager) GetAnimationHandle(id string) AnimationHandle {
	return AnimationHandle{id: id}
}

func (m *AssetManager) IsRaptorAnimationHandle(animationHandle AnimationHandle) bool {
	return strings.Contains(string(animationHandle.id), "velociraptor")
}

// this should probably look up a document, and get the animations from there, rather than storing these locally
func (m *AssetManager) GetAnimations(animationHandle AnimationHandle) (map[string]*modelspec.AnimationSpec, map[int]*modelspec.Joint, int) {
	id := string(animationHandle.id)
	return m.Animations[id], m.Joints[id], m.RootJoints[id]
}

func (m *AssetManager) GetPrimitives(meshHandle MeshHandle) []Primitive {
	if _, ok := m.Primitives[meshHandle]; !ok {
		return nil
	}
	return m.Primitives[meshHandle]
}
