package model_test

import (
	"fmt"
	"testing"

	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/lib/model"
)

func TestFillWeightsWithNoChange(t *testing.T) {
	jointIDs := []int{0, 1, 2, 0}
	weights := []float32{0.55, 0.25, 0.20, 0}

	newJointIDs, newWeights := model.FillWeights(jointIDs, weights)
	if !IntSliceEqual(jointIDs, newJointIDs) {
		t.Errorf("expected jointIDs to match %v, %v", jointIDs, newJointIDs)
	}

	if !Float32SliceEqual(weights, newWeights) {
		t.Fatal("expected weights to match")
	}

}

func TestFillWeightsDroppingWeight(t *testing.T) {
	jointIDs := []int{0, 1, 2, 3, 4}
	weights := []float32{0.55, 0.25, 0.02, 0.15, 0.05}

	newJointIDs, newWeights := model.FillWeights(jointIDs, weights)

	expectedJointIDs := []int{0, 1, 3, 4}
	if !IntSliceEqual(expectedJointIDs, newJointIDs) {
		t.Fatal("expected jointIDs to match")
	}

	expectedWeights := []float32{0.55, 0.25, 0.15, 0.05}
	if !Float32SliceEqual(expectedWeights, newWeights) {
		t.Errorf("expected weights to match %v, %v", expectedWeights, newWeights)
	}
}

func TestFillWeightsWithAddedWeight(t *testing.T) {
	jointIDs := []int{0, 1}
	weights := []float32{0.75, 0.25}

	newJointIDs, newWeights := model.FillWeights(jointIDs, weights)

	expectedJointIDs := []int{0, 1, 0, 0}
	if !IntSliceEqual(expectedJointIDs, newJointIDs) {
		t.Fatal("expected jointIDs to match")
	}

	expectedWeights := []float32{0.75, 0.25, 0, 0}
	if !Float32SliceEqual(expectedWeights, newWeights) {
		t.Fatal("expected weights to match")
	}

	if len(newJointIDs) != settings.AnimationMaxJointWeights {
		t.Fatal("expected length of joint ids to match settings.AnimationMaxJointWeights")

	}
}
func TestNormalizeWeightsHappyPath(t *testing.T) {
	jointWeights := []model.JointWeight{
		{JointID: 0, Weight: 0.5},
		{JointID: 1, Weight: 0.5},
		{JointID: 2, Weight: 0.5},
		{JointID: 3, Weight: 0.5},
	}

	model.NormalizeWeights(jointWeights)
	var expected float32 = 0.25
	if jointWeights[0].Weight != expected {
		t.Fatal(fmt.Sprintf("joint weight should be %f but was instead: %f", expected, jointWeights[0].Weight))
	}
	expected = 0.25
	if jointWeights[1].Weight != expected {
		t.Fatal(fmt.Sprintf("joint weight should be %f but was instead: %f", expected, jointWeights[1].Weight))
	}
	expected = 0.25
	if jointWeights[2].Weight != expected {
		t.Fatal(fmt.Sprintf("joint weight should be %f but was instead: %f", expected, jointWeights[2].Weight))
	}
	expected = 0.25
	if jointWeights[3].Weight != expected {
		t.Fatal(fmt.Sprintf("joint weight should be %f but was instead: %f", expected, jointWeights[3].Weight))
	}
}

func TestNormalizeWeightsWithAdjustments(t *testing.T) {
	jointWeights := []model.JointWeight{
		model.JointWeight{JointID: 0, Weight: 0.25},
		model.JointWeight{JointID: 0, Weight: 0.25},
	}

	model.NormalizeWeights(jointWeights)
	var expected float32 = 0.5
	if jointWeights[0].Weight != expected {
		t.Fatal(fmt.Sprintf("joint weight should be %f but was instead: %f", expected, jointWeights[0].Weight))
	}
	expected = 0.5
	if jointWeights[1].Weight != expected {
		t.Fatal(fmt.Sprintf("joint weight should be %f but was instead: %f", expected, jointWeights[1].Weight))
	}
}

func IntSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func Float32SliceEqual(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
