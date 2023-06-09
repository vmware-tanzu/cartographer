package testing

import (
	"context"
	"fmt"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/controllers"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

type SupplyChain interface {
	stamp(ctx context.Context, workload *v1alpha1.Workload, apiTemplate ValidatableTemplate, template templates.Reader) (*unstructured.Unstructured, error)
	//getSupplyChain(*v1alpha1.Workload) (*v1alpha1.ClusterSupplyChain, error)
	//getTargetResourceName() string
	//getOutputs() (realizer.Outputs, error)
}

// TTSupplyChainFileSet
// Paths is a list of either paths to a supply chain
// or a directory containing supply chain files
type TTSupplyChainFileSet struct {
	Paths     []string
	YttValues Values
	YttFiles  []string
}

func (s *TTSupplyChainFileSet) getSupplyChain(workload *v1alpha1.Workload) (*v1alpha1.ClusterSupplyChain, error) {
	var noLog *NoLog

	allSupplyChains, err := s.readAllPaths()
	if err != nil {
		return nil, fmt.Errorf("read all paths, %w", err)
	}

	selectedSupplyChains, err := repository.GetSelectedSupplyChain(allSupplyChains, workload, logr.New(noLog))
	if err != nil {
		return nil, fmt.Errorf("get selected supply chain, %w", err)
	}

	if len(selectedSupplyChains) == 0 {
		return nil, fmt.Errorf("no supply chain [%s/%s] found where full selector is satisfied by labels: %v",
			workload.Namespace, workload.Name, workload.Labels)
	}

	if len(selectedSupplyChains) > 1 {
		return nil, fmt.Errorf("more than one supply chain selected for workload [%s/%s]: %+v",
			workload.Namespace, workload.Name, controllers.GetSupplyChainNames(selectedSupplyChains))
	}

	return selectedSupplyChains[0], nil
}

func (s *TTSupplyChainFileSet) readAllPaths() ([]*v1alpha1.ClusterSupplyChain, error) {
	var supplyChains []*v1alpha1.ClusterSupplyChain

	for _, path := range s.Paths {
		file, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("could not get fileinfo for path: %w", err)
		}

		if file.IsDir() {
			additionalSupplyChains, err := s.readSupplyChainDir(path)
			if err != nil {
				return nil, fmt.Errorf("read supply chain directory: %w", err)
			}

			supplyChains = append(supplyChains, additionalSupplyChains...)
		} else {
			supplyChain, err := s.readSupplyChainFile(path)
			if err != nil {
				return nil, fmt.Errorf("could not read supply chain file: %w", err)
			}

			supplyChains = append(supplyChains, supplyChain)
		}
	}

	return supplyChains, nil
}

// readSupplyChainDir is not recursive and will not walk a nested directory
func (s *TTSupplyChainFileSet) readSupplyChainDir(path string) ([]*v1alpha1.ClusterSupplyChain, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("os read directory: %w", err)
	}

	var supplyChains []*v1alpha1.ClusterSupplyChain

	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())
		supplyChain, err := s.readSupplyChainFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("read supply chain file: %s, %w", fullPath, err)
		}

		supplyChains = append(supplyChains, supplyChain)
	}

	return supplyChains, nil
}

func (s *TTSupplyChainFileSet) readSupplyChainFile(path string) (*v1alpha1.ClusterSupplyChain, error) {
	var (
		supplyChainFilepath string
		err                 error
	)

	if len(s.YttValues) != 0 || len(s.YttFiles) != 0 {
		err := ensureYTTAvailable(context.TODO())

		if err != nil {
			return nil, fmt.Errorf("ensure ytt available: %w", err)
		}

		supplyChainFilepath, err = s.preprocessYtt(context.TODO(), path)
		if err != nil {
			return nil, fmt.Errorf("failed to preprocess ytt: %w", err)
		}
		defer os.RemoveAll(supplyChainFilepath)
	} else {
		supplyChainFilepath = path
	}

	supplyChain := &v1alpha1.ClusterSupplyChain{}

	supplyChainData, err := os.ReadFile(supplyChainFilepath)
	if err != nil {
		return nil, fmt.Errorf("could not read supplyChain file: %w", err)
	}

	if err = yaml.Unmarshal(supplyChainData, supplyChain); err != nil {
		return nil, fmt.Errorf("unmarshall template: %w", err)
	}

	return supplyChain, nil
}

func (s *TTSupplyChainFileSet) preprocessYtt(ctx context.Context, supplyChainFilepath string) (string, error) {
	yt := YTT()
	yt.Values(s.YttValues)
	yt.F(supplyChainFilepath)
	for _, yttfile := range s.YttFiles {
		yt.F(yttfile)
	}
	f, err := yt.ToTempFile(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file by ytt: %w", err)
	}

	return f.Name(), nil
}

type NoLog struct{}

func (n *NoLog) Init(_ logr.RuntimeInfo)                   {}
func (n *NoLog) Enabled(_ int) bool                        { return true }
func (n *NoLog) Info(_ int, _ string, _ ...interface{})    {}
func (n *NoLog) Error(_ error, _ string, _ ...interface{}) {}
func (n *NoLog) WithValues(_ ...interface{}) logr.LogSink  { return n }
func (n *NoLog) WithName(name string) logr.LogSink         { return n }

func (s *TTSupplyChainFileSet) stamp(ctx context.Context, workload *v1alpha1.Workload, _ ValidatableTemplate, template templates.Reader) (*unstructured.Unstructured, error) {
	supplyChain, err := s.getSupplyChain(workload)
	if err != nil {
		return nil, fmt.Errorf("get supplychain: %w", err)
	}

	resource, err := getTargetResource(realizer.MakeSupplychainOwnerResources(supplyChain), s.getTargetResourceName())
	if err != nil {
		return nil, fmt.Errorf("get target resource: %w", err)
	}

	templatingContext := realizer.NewContextGenerator(workload, workload.Spec.Params, supplyChain.Spec.Params)

	resourceLabeler := controllers.BuildWorkloadResourceLabeler(workload, supplyChain)
	labels := resourceLabeler(*resource, template)

	outputs, err := s.getOutputs()

	stamper := templates.StamperBuilder(workload, templatingContext.Generate(template, *resource, outputs, labels), labels)
	actualStampedObject, err := stamper.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		return nil, fmt.Errorf("could not stamp: %w", err)
	}

	return actualStampedObject, nil
}

func getTargetResource(resources []realizer.OwnerResource, targetResourceName string) (*realizer.OwnerResource, error) {
	for _, resource := range resources {
		if resource.Name == targetResourceName {
			return &resource, nil
		}
	}

	return nil, fmt.Errorf("did not find a supply chain resource with target name: %s", targetResourceName)
}

func (s *TTSupplyChainFileSet) getTargetResourceName() string {
	panic("not implemented")
}
func (s *TTSupplyChainFileSet) getOutputs() (realizer.Outputs, error) {
	panic("not implemented")
}
