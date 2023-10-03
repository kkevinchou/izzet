package other

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

// NOT REALLY A GOOD PACKAGE, FIGURE OUT WHERE TO PUT THIS STUFF LATER

func CreateEntitiesFromScene(document *modelspec.Document) []*entities.Entity {
	// modelConfig := &model.ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}
	var result []*entities.Entity

	for _, scene := range document.Scenes {
		for _, node := range scene.Nodes {
			entity := entities.InstantiateEntity(node.Name)
			entity.Node = parseNode(node, true, mgl32.Ident4(), document.Name)
			entities.SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
			entities.SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
			entities.SetScale(entity, utils.Vec3F32ToF64(node.Scale))

			result = append(result, entity)
		}
	}

	return result
}

// func parseNode(node *modelspec.Node, parentTransform mgl32.Mat4, ignoreTransform bool) []RenderData {
func parseNode(node *modelspec.Node, ignoreTransform bool, parentTransform mgl32.Mat4, namespace string) entities.Node {
	transform := node.Transform
	if ignoreTransform {
		transform = mgl32.Ident4()
	}
	transform = parentTransform.Mul4(transform)

	eNode := entities.Node{
		MeshID:     *node.MeshID,
		Transform:  transform,
		MeshHandle: modellibrary.NewHandle(namespace, *node.MeshID),
	}

	var children []entities.Node
	for _, childNode := range node.Children {
		children = append(children, parseNode(childNode, false, transform, namespace))
	}

	eNode.Children = children
	return eNode
}
