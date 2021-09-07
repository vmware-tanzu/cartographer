// Code generated by counterfeiter. DO NOT EDIT.
package repositoryfakes

import (
	"sync"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeRepository struct {
	AssureObjectExistsOnClusterStub        func(*unstructured.Unstructured) error
	assureObjectExistsOnClusterMutex       sync.RWMutex
	assureObjectExistsOnClusterArgsForCall []struct {
		arg1 *unstructured.Unstructured
	}
	assureObjectExistsOnClusterReturns struct {
		result1 error
	}
	assureObjectExistsOnClusterReturnsOnCall map[int]struct {
		result1 error
	}
	GetClusterTemplateStub        func(v1alpha1.ClusterTemplateReference) (templates.Template, error)
	getClusterTemplateMutex       sync.RWMutex
	getClusterTemplateArgsForCall []struct {
		arg1 v1alpha1.ClusterTemplateReference
	}
	getClusterTemplateReturns struct {
		result1 templates.Template
		result2 error
	}
	getClusterTemplateReturnsOnCall map[int]struct {
		result1 templates.Template
		result2 error
	}
	GetPipelineStub        func(string, string) (*v1alpha1.Pipeline, error)
	getPipelineMutex       sync.RWMutex
	getPipelineArgsForCall []struct {
		arg1 string
		arg2 string
	}
	getPipelineReturns struct {
		result1 *v1alpha1.Pipeline
		result2 error
	}
	getPipelineReturnsOnCall map[int]struct {
		result1 *v1alpha1.Pipeline
		result2 error
	}
	GetSchemeStub        func() *runtime.Scheme
	getSchemeMutex       sync.RWMutex
	getSchemeArgsForCall []struct {
	}
	getSchemeReturns struct {
		result1 *runtime.Scheme
	}
	getSchemeReturnsOnCall map[int]struct {
		result1 *runtime.Scheme
	}
	GetSupplyChainStub        func(string) (*v1alpha1.ClusterSupplyChain, error)
	getSupplyChainMutex       sync.RWMutex
	getSupplyChainArgsForCall []struct {
		arg1 string
	}
	getSupplyChainReturns struct {
		result1 *v1alpha1.ClusterSupplyChain
		result2 error
	}
	getSupplyChainReturnsOnCall map[int]struct {
		result1 *v1alpha1.ClusterSupplyChain
		result2 error
	}
	GetSupplyChainsForWorkloadStub        func(*v1alpha1.Workload) ([]v1alpha1.ClusterSupplyChain, error)
	getSupplyChainsForWorkloadMutex       sync.RWMutex
	getSupplyChainsForWorkloadArgsForCall []struct {
		arg1 *v1alpha1.Workload
	}
	getSupplyChainsForWorkloadReturns struct {
		result1 []v1alpha1.ClusterSupplyChain
		result2 error
	}
	getSupplyChainsForWorkloadReturnsOnCall map[int]struct {
		result1 []v1alpha1.ClusterSupplyChain
		result2 error
	}
	GetTemplateStub        func(v1alpha1.TemplateReference) (templates.Template, error)
	getTemplateMutex       sync.RWMutex
	getTemplateArgsForCall []struct {
		arg1 v1alpha1.TemplateReference
	}
	getTemplateReturns struct {
		result1 templates.Template
		result2 error
	}
	getTemplateReturnsOnCall map[int]struct {
		result1 templates.Template
		result2 error
	}
	GetWorkloadStub        func(string, string) (*v1alpha1.Workload, error)
	getWorkloadMutex       sync.RWMutex
	getWorkloadArgsForCall []struct {
		arg1 string
		arg2 string
	}
	getWorkloadReturns struct {
		result1 *v1alpha1.Workload
		result2 error
	}
	getWorkloadReturnsOnCall map[int]struct {
		result1 *v1alpha1.Workload
		result2 error
	}
	StatusUpdateStub        func(client.Object) error
	statusUpdateMutex       sync.RWMutex
	statusUpdateArgsForCall []struct {
		arg1 client.Object
	}
	statusUpdateReturns struct {
		result1 error
	}
	statusUpdateReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRepository) AssureObjectExistsOnCluster(arg1 *unstructured.Unstructured) error {
	fake.assureObjectExistsOnClusterMutex.Lock()
	ret, specificReturn := fake.assureObjectExistsOnClusterReturnsOnCall[len(fake.assureObjectExistsOnClusterArgsForCall)]
	fake.assureObjectExistsOnClusterArgsForCall = append(fake.assureObjectExistsOnClusterArgsForCall, struct {
		arg1 *unstructured.Unstructured
	}{arg1})
	stub := fake.AssureObjectExistsOnClusterStub
	fakeReturns := fake.assureObjectExistsOnClusterReturns
	fake.recordInvocation("AssureObjectExistsOnCluster", []interface{}{arg1})
	fake.assureObjectExistsOnClusterMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) AssureObjectExistsOnClusterCallCount() int {
	fake.assureObjectExistsOnClusterMutex.RLock()
	defer fake.assureObjectExistsOnClusterMutex.RUnlock()
	return len(fake.assureObjectExistsOnClusterArgsForCall)
}

func (fake *FakeRepository) AssureObjectExistsOnClusterCalls(stub func(*unstructured.Unstructured) error) {
	fake.assureObjectExistsOnClusterMutex.Lock()
	defer fake.assureObjectExistsOnClusterMutex.Unlock()
	fake.AssureObjectExistsOnClusterStub = stub
}

func (fake *FakeRepository) AssureObjectExistsOnClusterArgsForCall(i int) *unstructured.Unstructured {
	fake.assureObjectExistsOnClusterMutex.RLock()
	defer fake.assureObjectExistsOnClusterMutex.RUnlock()
	argsForCall := fake.assureObjectExistsOnClusterArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) AssureObjectExistsOnClusterReturns(result1 error) {
	fake.assureObjectExistsOnClusterMutex.Lock()
	defer fake.assureObjectExistsOnClusterMutex.Unlock()
	fake.AssureObjectExistsOnClusterStub = nil
	fake.assureObjectExistsOnClusterReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) AssureObjectExistsOnClusterReturnsOnCall(i int, result1 error) {
	fake.assureObjectExistsOnClusterMutex.Lock()
	defer fake.assureObjectExistsOnClusterMutex.Unlock()
	fake.AssureObjectExistsOnClusterStub = nil
	if fake.assureObjectExistsOnClusterReturnsOnCall == nil {
		fake.assureObjectExistsOnClusterReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.assureObjectExistsOnClusterReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) GetClusterTemplate(arg1 v1alpha1.ClusterTemplateReference) (templates.Template, error) {
	fake.getClusterTemplateMutex.Lock()
	ret, specificReturn := fake.getClusterTemplateReturnsOnCall[len(fake.getClusterTemplateArgsForCall)]
	fake.getClusterTemplateArgsForCall = append(fake.getClusterTemplateArgsForCall, struct {
		arg1 v1alpha1.ClusterTemplateReference
	}{arg1})
	stub := fake.GetClusterTemplateStub
	fakeReturns := fake.getClusterTemplateReturns
	fake.recordInvocation("GetClusterTemplate", []interface{}{arg1})
	fake.getClusterTemplateMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) GetClusterTemplateCallCount() int {
	fake.getClusterTemplateMutex.RLock()
	defer fake.getClusterTemplateMutex.RUnlock()
	return len(fake.getClusterTemplateArgsForCall)
}

func (fake *FakeRepository) GetClusterTemplateCalls(stub func(v1alpha1.ClusterTemplateReference) (templates.Template, error)) {
	fake.getClusterTemplateMutex.Lock()
	defer fake.getClusterTemplateMutex.Unlock()
	fake.GetClusterTemplateStub = stub
}

func (fake *FakeRepository) GetClusterTemplateArgsForCall(i int) v1alpha1.ClusterTemplateReference {
	fake.getClusterTemplateMutex.RLock()
	defer fake.getClusterTemplateMutex.RUnlock()
	argsForCall := fake.getClusterTemplateArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) GetClusterTemplateReturns(result1 templates.Template, result2 error) {
	fake.getClusterTemplateMutex.Lock()
	defer fake.getClusterTemplateMutex.Unlock()
	fake.GetClusterTemplateStub = nil
	fake.getClusterTemplateReturns = struct {
		result1 templates.Template
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetClusterTemplateReturnsOnCall(i int, result1 templates.Template, result2 error) {
	fake.getClusterTemplateMutex.Lock()
	defer fake.getClusterTemplateMutex.Unlock()
	fake.GetClusterTemplateStub = nil
	if fake.getClusterTemplateReturnsOnCall == nil {
		fake.getClusterTemplateReturnsOnCall = make(map[int]struct {
			result1 templates.Template
			result2 error
		})
	}
	fake.getClusterTemplateReturnsOnCall[i] = struct {
		result1 templates.Template
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetPipeline(arg1 string, arg2 string) (*v1alpha1.Pipeline, error) {
	fake.getPipelineMutex.Lock()
	ret, specificReturn := fake.getPipelineReturnsOnCall[len(fake.getPipelineArgsForCall)]
	fake.getPipelineArgsForCall = append(fake.getPipelineArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.GetPipelineStub
	fakeReturns := fake.getPipelineReturns
	fake.recordInvocation("GetPipeline", []interface{}{arg1, arg2})
	fake.getPipelineMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) GetPipelineCallCount() int {
	fake.getPipelineMutex.RLock()
	defer fake.getPipelineMutex.RUnlock()
	return len(fake.getPipelineArgsForCall)
}

func (fake *FakeRepository) GetPipelineCalls(stub func(string, string) (*v1alpha1.Pipeline, error)) {
	fake.getPipelineMutex.Lock()
	defer fake.getPipelineMutex.Unlock()
	fake.GetPipelineStub = stub
}

func (fake *FakeRepository) GetPipelineArgsForCall(i int) (string, string) {
	fake.getPipelineMutex.RLock()
	defer fake.getPipelineMutex.RUnlock()
	argsForCall := fake.getPipelineArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRepository) GetPipelineReturns(result1 *v1alpha1.Pipeline, result2 error) {
	fake.getPipelineMutex.Lock()
	defer fake.getPipelineMutex.Unlock()
	fake.GetPipelineStub = nil
	fake.getPipelineReturns = struct {
		result1 *v1alpha1.Pipeline
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetPipelineReturnsOnCall(i int, result1 *v1alpha1.Pipeline, result2 error) {
	fake.getPipelineMutex.Lock()
	defer fake.getPipelineMutex.Unlock()
	fake.GetPipelineStub = nil
	if fake.getPipelineReturnsOnCall == nil {
		fake.getPipelineReturnsOnCall = make(map[int]struct {
			result1 *v1alpha1.Pipeline
			result2 error
		})
	}
	fake.getPipelineReturnsOnCall[i] = struct {
		result1 *v1alpha1.Pipeline
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetScheme() *runtime.Scheme {
	fake.getSchemeMutex.Lock()
	ret, specificReturn := fake.getSchemeReturnsOnCall[len(fake.getSchemeArgsForCall)]
	fake.getSchemeArgsForCall = append(fake.getSchemeArgsForCall, struct {
	}{})
	stub := fake.GetSchemeStub
	fakeReturns := fake.getSchemeReturns
	fake.recordInvocation("GetScheme", []interface{}{})
	fake.getSchemeMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) GetSchemeCallCount() int {
	fake.getSchemeMutex.RLock()
	defer fake.getSchemeMutex.RUnlock()
	return len(fake.getSchemeArgsForCall)
}

func (fake *FakeRepository) GetSchemeCalls(stub func() *runtime.Scheme) {
	fake.getSchemeMutex.Lock()
	defer fake.getSchemeMutex.Unlock()
	fake.GetSchemeStub = stub
}

func (fake *FakeRepository) GetSchemeReturns(result1 *runtime.Scheme) {
	fake.getSchemeMutex.Lock()
	defer fake.getSchemeMutex.Unlock()
	fake.GetSchemeStub = nil
	fake.getSchemeReturns = struct {
		result1 *runtime.Scheme
	}{result1}
}

func (fake *FakeRepository) GetSchemeReturnsOnCall(i int, result1 *runtime.Scheme) {
	fake.getSchemeMutex.Lock()
	defer fake.getSchemeMutex.Unlock()
	fake.GetSchemeStub = nil
	if fake.getSchemeReturnsOnCall == nil {
		fake.getSchemeReturnsOnCall = make(map[int]struct {
			result1 *runtime.Scheme
		})
	}
	fake.getSchemeReturnsOnCall[i] = struct {
		result1 *runtime.Scheme
	}{result1}
}

func (fake *FakeRepository) GetSupplyChain(arg1 string) (*v1alpha1.ClusterSupplyChain, error) {
	fake.getSupplyChainMutex.Lock()
	ret, specificReturn := fake.getSupplyChainReturnsOnCall[len(fake.getSupplyChainArgsForCall)]
	fake.getSupplyChainArgsForCall = append(fake.getSupplyChainArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.GetSupplyChainStub
	fakeReturns := fake.getSupplyChainReturns
	fake.recordInvocation("GetSupplyChain", []interface{}{arg1})
	fake.getSupplyChainMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) GetSupplyChainCallCount() int {
	fake.getSupplyChainMutex.RLock()
	defer fake.getSupplyChainMutex.RUnlock()
	return len(fake.getSupplyChainArgsForCall)
}

func (fake *FakeRepository) GetSupplyChainCalls(stub func(string) (*v1alpha1.ClusterSupplyChain, error)) {
	fake.getSupplyChainMutex.Lock()
	defer fake.getSupplyChainMutex.Unlock()
	fake.GetSupplyChainStub = stub
}

func (fake *FakeRepository) GetSupplyChainArgsForCall(i int) string {
	fake.getSupplyChainMutex.RLock()
	defer fake.getSupplyChainMutex.RUnlock()
	argsForCall := fake.getSupplyChainArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) GetSupplyChainReturns(result1 *v1alpha1.ClusterSupplyChain, result2 error) {
	fake.getSupplyChainMutex.Lock()
	defer fake.getSupplyChainMutex.Unlock()
	fake.GetSupplyChainStub = nil
	fake.getSupplyChainReturns = struct {
		result1 *v1alpha1.ClusterSupplyChain
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetSupplyChainReturnsOnCall(i int, result1 *v1alpha1.ClusterSupplyChain, result2 error) {
	fake.getSupplyChainMutex.Lock()
	defer fake.getSupplyChainMutex.Unlock()
	fake.GetSupplyChainStub = nil
	if fake.getSupplyChainReturnsOnCall == nil {
		fake.getSupplyChainReturnsOnCall = make(map[int]struct {
			result1 *v1alpha1.ClusterSupplyChain
			result2 error
		})
	}
	fake.getSupplyChainReturnsOnCall[i] = struct {
		result1 *v1alpha1.ClusterSupplyChain
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetSupplyChainsForWorkload(arg1 *v1alpha1.Workload) ([]v1alpha1.ClusterSupplyChain, error) {
	fake.getSupplyChainsForWorkloadMutex.Lock()
	ret, specificReturn := fake.getSupplyChainsForWorkloadReturnsOnCall[len(fake.getSupplyChainsForWorkloadArgsForCall)]
	fake.getSupplyChainsForWorkloadArgsForCall = append(fake.getSupplyChainsForWorkloadArgsForCall, struct {
		arg1 *v1alpha1.Workload
	}{arg1})
	stub := fake.GetSupplyChainsForWorkloadStub
	fakeReturns := fake.getSupplyChainsForWorkloadReturns
	fake.recordInvocation("GetSupplyChainsForWorkload", []interface{}{arg1})
	fake.getSupplyChainsForWorkloadMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) GetSupplyChainsForWorkloadCallCount() int {
	fake.getSupplyChainsForWorkloadMutex.RLock()
	defer fake.getSupplyChainsForWorkloadMutex.RUnlock()
	return len(fake.getSupplyChainsForWorkloadArgsForCall)
}

func (fake *FakeRepository) GetSupplyChainsForWorkloadCalls(stub func(*v1alpha1.Workload) ([]v1alpha1.ClusterSupplyChain, error)) {
	fake.getSupplyChainsForWorkloadMutex.Lock()
	defer fake.getSupplyChainsForWorkloadMutex.Unlock()
	fake.GetSupplyChainsForWorkloadStub = stub
}

func (fake *FakeRepository) GetSupplyChainsForWorkloadArgsForCall(i int) *v1alpha1.Workload {
	fake.getSupplyChainsForWorkloadMutex.RLock()
	defer fake.getSupplyChainsForWorkloadMutex.RUnlock()
	argsForCall := fake.getSupplyChainsForWorkloadArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) GetSupplyChainsForWorkloadReturns(result1 []v1alpha1.ClusterSupplyChain, result2 error) {
	fake.getSupplyChainsForWorkloadMutex.Lock()
	defer fake.getSupplyChainsForWorkloadMutex.Unlock()
	fake.GetSupplyChainsForWorkloadStub = nil
	fake.getSupplyChainsForWorkloadReturns = struct {
		result1 []v1alpha1.ClusterSupplyChain
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetSupplyChainsForWorkloadReturnsOnCall(i int, result1 []v1alpha1.ClusterSupplyChain, result2 error) {
	fake.getSupplyChainsForWorkloadMutex.Lock()
	defer fake.getSupplyChainsForWorkloadMutex.Unlock()
	fake.GetSupplyChainsForWorkloadStub = nil
	if fake.getSupplyChainsForWorkloadReturnsOnCall == nil {
		fake.getSupplyChainsForWorkloadReturnsOnCall = make(map[int]struct {
			result1 []v1alpha1.ClusterSupplyChain
			result2 error
		})
	}
	fake.getSupplyChainsForWorkloadReturnsOnCall[i] = struct {
		result1 []v1alpha1.ClusterSupplyChain
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetTemplate(arg1 v1alpha1.TemplateReference) (templates.Template, error) {
	fake.getTemplateMutex.Lock()
	ret, specificReturn := fake.getTemplateReturnsOnCall[len(fake.getTemplateArgsForCall)]
	fake.getTemplateArgsForCall = append(fake.getTemplateArgsForCall, struct {
		arg1 v1alpha1.TemplateReference
	}{arg1})
	stub := fake.GetTemplateStub
	fakeReturns := fake.getTemplateReturns
	fake.recordInvocation("GetTemplate", []interface{}{arg1})
	fake.getTemplateMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) GetTemplateCallCount() int {
	fake.getTemplateMutex.RLock()
	defer fake.getTemplateMutex.RUnlock()
	return len(fake.getTemplateArgsForCall)
}

func (fake *FakeRepository) GetTemplateCalls(stub func(v1alpha1.TemplateReference) (templates.Template, error)) {
	fake.getTemplateMutex.Lock()
	defer fake.getTemplateMutex.Unlock()
	fake.GetTemplateStub = stub
}

func (fake *FakeRepository) GetTemplateArgsForCall(i int) v1alpha1.TemplateReference {
	fake.getTemplateMutex.RLock()
	defer fake.getTemplateMutex.RUnlock()
	argsForCall := fake.getTemplateArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) GetTemplateReturns(result1 templates.Template, result2 error) {
	fake.getTemplateMutex.Lock()
	defer fake.getTemplateMutex.Unlock()
	fake.GetTemplateStub = nil
	fake.getTemplateReturns = struct {
		result1 templates.Template
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetTemplateReturnsOnCall(i int, result1 templates.Template, result2 error) {
	fake.getTemplateMutex.Lock()
	defer fake.getTemplateMutex.Unlock()
	fake.GetTemplateStub = nil
	if fake.getTemplateReturnsOnCall == nil {
		fake.getTemplateReturnsOnCall = make(map[int]struct {
			result1 templates.Template
			result2 error
		})
	}
	fake.getTemplateReturnsOnCall[i] = struct {
		result1 templates.Template
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetWorkload(arg1 string, arg2 string) (*v1alpha1.Workload, error) {
	fake.getWorkloadMutex.Lock()
	ret, specificReturn := fake.getWorkloadReturnsOnCall[len(fake.getWorkloadArgsForCall)]
	fake.getWorkloadArgsForCall = append(fake.getWorkloadArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.GetWorkloadStub
	fakeReturns := fake.getWorkloadReturns
	fake.recordInvocation("GetWorkload", []interface{}{arg1, arg2})
	fake.getWorkloadMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) GetWorkloadCallCount() int {
	fake.getWorkloadMutex.RLock()
	defer fake.getWorkloadMutex.RUnlock()
	return len(fake.getWorkloadArgsForCall)
}

func (fake *FakeRepository) GetWorkloadCalls(stub func(string, string) (*v1alpha1.Workload, error)) {
	fake.getWorkloadMutex.Lock()
	defer fake.getWorkloadMutex.Unlock()
	fake.GetWorkloadStub = stub
}

func (fake *FakeRepository) GetWorkloadArgsForCall(i int) (string, string) {
	fake.getWorkloadMutex.RLock()
	defer fake.getWorkloadMutex.RUnlock()
	argsForCall := fake.getWorkloadArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRepository) GetWorkloadReturns(result1 *v1alpha1.Workload, result2 error) {
	fake.getWorkloadMutex.Lock()
	defer fake.getWorkloadMutex.Unlock()
	fake.GetWorkloadStub = nil
	fake.getWorkloadReturns = struct {
		result1 *v1alpha1.Workload
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetWorkloadReturnsOnCall(i int, result1 *v1alpha1.Workload, result2 error) {
	fake.getWorkloadMutex.Lock()
	defer fake.getWorkloadMutex.Unlock()
	fake.GetWorkloadStub = nil
	if fake.getWorkloadReturnsOnCall == nil {
		fake.getWorkloadReturnsOnCall = make(map[int]struct {
			result1 *v1alpha1.Workload
			result2 error
		})
	}
	fake.getWorkloadReturnsOnCall[i] = struct {
		result1 *v1alpha1.Workload
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) StatusUpdate(arg1 client.Object) error {
	fake.statusUpdateMutex.Lock()
	ret, specificReturn := fake.statusUpdateReturnsOnCall[len(fake.statusUpdateArgsForCall)]
	fake.statusUpdateArgsForCall = append(fake.statusUpdateArgsForCall, struct {
		arg1 client.Object
	}{arg1})
	stub := fake.StatusUpdateStub
	fakeReturns := fake.statusUpdateReturns
	fake.recordInvocation("StatusUpdate", []interface{}{arg1})
	fake.statusUpdateMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) StatusUpdateCallCount() int {
	fake.statusUpdateMutex.RLock()
	defer fake.statusUpdateMutex.RUnlock()
	return len(fake.statusUpdateArgsForCall)
}

func (fake *FakeRepository) StatusUpdateCalls(stub func(client.Object) error) {
	fake.statusUpdateMutex.Lock()
	defer fake.statusUpdateMutex.Unlock()
	fake.StatusUpdateStub = stub
}

func (fake *FakeRepository) StatusUpdateArgsForCall(i int) client.Object {
	fake.statusUpdateMutex.RLock()
	defer fake.statusUpdateMutex.RUnlock()
	argsForCall := fake.statusUpdateArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) StatusUpdateReturns(result1 error) {
	fake.statusUpdateMutex.Lock()
	defer fake.statusUpdateMutex.Unlock()
	fake.StatusUpdateStub = nil
	fake.statusUpdateReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) StatusUpdateReturnsOnCall(i int, result1 error) {
	fake.statusUpdateMutex.Lock()
	defer fake.statusUpdateMutex.Unlock()
	fake.StatusUpdateStub = nil
	if fake.statusUpdateReturnsOnCall == nil {
		fake.statusUpdateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.statusUpdateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.assureObjectExistsOnClusterMutex.RLock()
	defer fake.assureObjectExistsOnClusterMutex.RUnlock()
	fake.getClusterTemplateMutex.RLock()
	defer fake.getClusterTemplateMutex.RUnlock()
	fake.getPipelineMutex.RLock()
	defer fake.getPipelineMutex.RUnlock()
	fake.getSchemeMutex.RLock()
	defer fake.getSchemeMutex.RUnlock()
	fake.getSupplyChainMutex.RLock()
	defer fake.getSupplyChainMutex.RUnlock()
	fake.getSupplyChainsForWorkloadMutex.RLock()
	defer fake.getSupplyChainsForWorkloadMutex.RUnlock()
	fake.getTemplateMutex.RLock()
	defer fake.getTemplateMutex.RUnlock()
	fake.getWorkloadMutex.RLock()
	defer fake.getWorkloadMutex.RUnlock()
	fake.statusUpdateMutex.RLock()
	defer fake.statusUpdateMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRepository) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ repository.Repository = new(FakeRepository)
