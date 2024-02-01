package app

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/input"
)

var zeroVec mgl64.Vec3

func IsZeroVec(v mgl64.Vec3) bool {
	return v == zeroVec
}

func GetControlVector(keyboardInput input.KeyboardInput) mgl64.Vec3 {
	var controlVector mgl64.Vec3
	if _, ok := keyboardInput[input.KeyboardKeyW]; ok {
		controlVector[2]++
	}
	if _, ok := keyboardInput[input.KeyboardKeyS]; ok {
		controlVector[2]--
	}
	if _, ok := keyboardInput[input.KeyboardKeyA]; ok {
		controlVector[0]--
	}
	if _, ok := keyboardInput[input.KeyboardKeyD]; ok {
		controlVector[0]++
	}
	if _, ok := keyboardInput[input.KeyboardKeyLShift]; ok {
		controlVector[1]--
	}
	if _, ok := keyboardInput[input.KeyboardKeySpace]; ok {
		controlVector[1]++
	}

	return controlVector
}
