//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by diegen. DO NOT EDIT.

package dies

import (
	testing "dies.dev/testing"
	testingx "testing"
)

func TestClusterBlueprintTypeDie_MissingMethods(t *testingx.T) {
	die := ClusterBlueprintTypeBlank
	ignore := []string{"TypeMeta", "ObjectMeta"}
	diff := testing.DieFieldDiff(die).Delete(ignore...)
	if diff.Len() != 0 {
		t.Errorf("found missing fields for ClusterBlueprintTypeDie: %s", diff.List())
	}
}

func TestClusterBlueprintTypeSpecDie_MissingMethods(t *testingx.T) {
	die := ClusterBlueprintTypeSpecBlank
	ignore := []string{}
	diff := testing.DieFieldDiff(die).Delete(ignore...)
	if diff.Len() != 0 {
		t.Errorf("found missing fields for ClusterBlueprintTypeSpecDie: %s", diff.List())
	}
}

func TestClusterBlueprintTypeStatusDie_MissingMethods(t *testingx.T) {
	die := ClusterBlueprintTypeStatusBlank
	ignore := []string{}
	diff := testing.DieFieldDiff(die).Delete(ignore...)
	if diff.Len() != 0 {
		t.Errorf("found missing fields for ClusterBlueprintTypeStatusDie: %s", diff.List())
	}
}
