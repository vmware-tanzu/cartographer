package testing

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type BlueprintParams interface {
	GetBlueprintParams() ([]v1alpha1.BlueprintParam, error)
}

type BlueprintParamsObject struct {
	BlueprintParams []v1alpha1.BlueprintParam
}

func (p *BlueprintParamsObject) GetBlueprintParams() ([]v1alpha1.BlueprintParam, error) {
	return p.BlueprintParams, nil
}

type BlueprintParamsFile struct {
	Path string
}

func (p *BlueprintParamsFile) GetBlueprintParams() ([]v1alpha1.BlueprintParam, error) {
	paramsFile, err := os.ReadFile(p.Path)
	if err != nil {
		return nil, fmt.Errorf("could not read blueprintParamsFile: %w", err)
	}

	var paramsData []v1alpha1.BlueprintParam

	err = yaml.Unmarshal(paramsFile, &paramsData)
	if err != nil {
		return nil, fmt.Errorf("unmarshall params: %w", err)
	}

	return paramsData, nil // TODO: document
}
