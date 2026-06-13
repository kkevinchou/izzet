package gltf

import (
	"fmt"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

type parseContext struct {
	floatScalarAccessors map[int][]float32
}

func newParseContext() *parseContext {
	return &parseContext{
		floatScalarAccessors: map[int][]float32{},
	}
}

func (ctx *parseContext) readFloatScalarAccessor(document *gltf.Document, accessorIndex int) ([]float32, error) {
	if values, ok := ctx.floatScalarAccessors[accessorIndex]; ok {
		return values, nil
	}

	accessor := document.Accessors[accessorIndex]
	if accessor.ComponentType != gltf.ComponentFloat {
		return nil, fmt.Errorf("unexpected component type %v", accessor.ComponentType)
	}
	if accessor.Type != gltf.AccessorScalar {
		return nil, fmt.Errorf("unexpected accessor type %v", accessor.Type)
	}

	input, err := modeler.ReadAccessor(document, accessor, nil)
	if err != nil {
		return nil, fmt.Errorf("read accessor %d: %w", accessorIndex, err)
	}

	values, ok := input.([]float32)
	if !ok {
		return nil, fmt.Errorf("unexpected accessor read type %T", input)
	}

	ctx.floatScalarAccessors[accessorIndex] = values
	return values, nil
}
