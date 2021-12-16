// Code generated by counterfeiter. DO NOT EDIT.
package repositoryfakes

import (
	"sync"

	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type FakeRepoCache struct {
	SetStub        func(*unstructured.Unstructured, *unstructured.Unstructured)
	setMutex       sync.RWMutex
	setArgsForCall []struct {
		arg1 *unstructured.Unstructured
		arg2 *unstructured.Unstructured
	}
	UnchangedSinceCachedStub        func(*unstructured.Unstructured, *unstructured.Unstructured) *unstructured.Unstructured
	unchangedSinceCachedMutex       sync.RWMutex
	unchangedSinceCachedArgsForCall []struct {
		arg1 *unstructured.Unstructured
		arg2 *unstructured.Unstructured
	}
	unchangedSinceCachedReturns struct {
		result1 *unstructured.Unstructured
	}
	unchangedSinceCachedReturnsOnCall map[int]struct {
		result1 *unstructured.Unstructured
	}
	UnchangedSinceCachedFromListStub        func(*unstructured.Unstructured, []*unstructured.Unstructured) *unstructured.Unstructured
	unchangedSinceCachedFromListMutex       sync.RWMutex
	unchangedSinceCachedFromListArgsForCall []struct {
		arg1 *unstructured.Unstructured
		arg2 []*unstructured.Unstructured
	}
	unchangedSinceCachedFromListReturns struct {
		result1 *unstructured.Unstructured
	}
	unchangedSinceCachedFromListReturnsOnCall map[int]struct {
		result1 *unstructured.Unstructured
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRepoCache) Set(arg1 *unstructured.Unstructured, arg2 *unstructured.Unstructured) {
	fake.setMutex.Lock()
	fake.setArgsForCall = append(fake.setArgsForCall, struct {
		arg1 *unstructured.Unstructured
		arg2 *unstructured.Unstructured
	}{arg1, arg2})
	stub := fake.SetStub
	fake.recordInvocation("Set", []interface{}{arg1, arg2})
	fake.setMutex.Unlock()
	if stub != nil {
		fake.SetStub(arg1, arg2)
	}
}

func (fake *FakeRepoCache) SetCallCount() int {
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	return len(fake.setArgsForCall)
}

func (fake *FakeRepoCache) SetCalls(stub func(*unstructured.Unstructured, *unstructured.Unstructured)) {
	fake.setMutex.Lock()
	defer fake.setMutex.Unlock()
	fake.SetStub = stub
}

func (fake *FakeRepoCache) SetArgsForCall(i int) (*unstructured.Unstructured, *unstructured.Unstructured) {
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	argsForCall := fake.setArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRepoCache) UnchangedSinceCached(arg1 *unstructured.Unstructured, arg2 *unstructured.Unstructured) *unstructured.Unstructured {
	fake.unchangedSinceCachedMutex.Lock()
	ret, specificReturn := fake.unchangedSinceCachedReturnsOnCall[len(fake.unchangedSinceCachedArgsForCall)]
	fake.unchangedSinceCachedArgsForCall = append(fake.unchangedSinceCachedArgsForCall, struct {
		arg1 *unstructured.Unstructured
		arg2 *unstructured.Unstructured
	}{arg1, arg2})
	stub := fake.UnchangedSinceCachedStub
	fakeReturns := fake.unchangedSinceCachedReturns
	fake.recordInvocation("UnchangedSinceCached", []interface{}{arg1, arg2})
	fake.unchangedSinceCachedMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepoCache) UnchangedSinceCachedCallCount() int {
	fake.unchangedSinceCachedMutex.RLock()
	defer fake.unchangedSinceCachedMutex.RUnlock()
	return len(fake.unchangedSinceCachedArgsForCall)
}

func (fake *FakeRepoCache) UnchangedSinceCachedCalls(stub func(*unstructured.Unstructured, *unstructured.Unstructured) *unstructured.Unstructured) {
	fake.unchangedSinceCachedMutex.Lock()
	defer fake.unchangedSinceCachedMutex.Unlock()
	fake.UnchangedSinceCachedStub = stub
}

func (fake *FakeRepoCache) UnchangedSinceCachedArgsForCall(i int) (*unstructured.Unstructured, *unstructured.Unstructured) {
	fake.unchangedSinceCachedMutex.RLock()
	defer fake.unchangedSinceCachedMutex.RUnlock()
	argsForCall := fake.unchangedSinceCachedArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRepoCache) UnchangedSinceCachedReturns(result1 *unstructured.Unstructured) {
	fake.unchangedSinceCachedMutex.Lock()
	defer fake.unchangedSinceCachedMutex.Unlock()
	fake.UnchangedSinceCachedStub = nil
	fake.unchangedSinceCachedReturns = struct {
		result1 *unstructured.Unstructured
	}{result1}
}

func (fake *FakeRepoCache) UnchangedSinceCachedReturnsOnCall(i int, result1 *unstructured.Unstructured) {
	fake.unchangedSinceCachedMutex.Lock()
	defer fake.unchangedSinceCachedMutex.Unlock()
	fake.UnchangedSinceCachedStub = nil
	if fake.unchangedSinceCachedReturnsOnCall == nil {
		fake.unchangedSinceCachedReturnsOnCall = make(map[int]struct {
			result1 *unstructured.Unstructured
		})
	}
	fake.unchangedSinceCachedReturnsOnCall[i] = struct {
		result1 *unstructured.Unstructured
	}{result1}
}

func (fake *FakeRepoCache) UnchangedSinceCachedFromList(arg1 *unstructured.Unstructured, arg2 []*unstructured.Unstructured) *unstructured.Unstructured {
	var arg2Copy []*unstructured.Unstructured
	if arg2 != nil {
		arg2Copy = make([]*unstructured.Unstructured, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.unchangedSinceCachedFromListMutex.Lock()
	ret, specificReturn := fake.unchangedSinceCachedFromListReturnsOnCall[len(fake.unchangedSinceCachedFromListArgsForCall)]
	fake.unchangedSinceCachedFromListArgsForCall = append(fake.unchangedSinceCachedFromListArgsForCall, struct {
		arg1 *unstructured.Unstructured
		arg2 []*unstructured.Unstructured
	}{arg1, arg2Copy})
	stub := fake.UnchangedSinceCachedFromListStub
	fakeReturns := fake.unchangedSinceCachedFromListReturns
	fake.recordInvocation("UnchangedSinceCachedFromList", []interface{}{arg1, arg2Copy})
	fake.unchangedSinceCachedFromListMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepoCache) UnchangedSinceCachedFromListCallCount() int {
	fake.unchangedSinceCachedFromListMutex.RLock()
	defer fake.unchangedSinceCachedFromListMutex.RUnlock()
	return len(fake.unchangedSinceCachedFromListArgsForCall)
}

func (fake *FakeRepoCache) UnchangedSinceCachedFromListCalls(stub func(*unstructured.Unstructured, []*unstructured.Unstructured) *unstructured.Unstructured) {
	fake.unchangedSinceCachedFromListMutex.Lock()
	defer fake.unchangedSinceCachedFromListMutex.Unlock()
	fake.UnchangedSinceCachedFromListStub = stub
}

func (fake *FakeRepoCache) UnchangedSinceCachedFromListArgsForCall(i int) (*unstructured.Unstructured, []*unstructured.Unstructured) {
	fake.unchangedSinceCachedFromListMutex.RLock()
	defer fake.unchangedSinceCachedFromListMutex.RUnlock()
	argsForCall := fake.unchangedSinceCachedFromListArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRepoCache) UnchangedSinceCachedFromListReturns(result1 *unstructured.Unstructured) {
	fake.unchangedSinceCachedFromListMutex.Lock()
	defer fake.unchangedSinceCachedFromListMutex.Unlock()
	fake.UnchangedSinceCachedFromListStub = nil
	fake.unchangedSinceCachedFromListReturns = struct {
		result1 *unstructured.Unstructured
	}{result1}
}

func (fake *FakeRepoCache) UnchangedSinceCachedFromListReturnsOnCall(i int, result1 *unstructured.Unstructured) {
	fake.unchangedSinceCachedFromListMutex.Lock()
	defer fake.unchangedSinceCachedFromListMutex.Unlock()
	fake.UnchangedSinceCachedFromListStub = nil
	if fake.unchangedSinceCachedFromListReturnsOnCall == nil {
		fake.unchangedSinceCachedFromListReturnsOnCall = make(map[int]struct {
			result1 *unstructured.Unstructured
		})
	}
	fake.unchangedSinceCachedFromListReturnsOnCall[i] = struct {
		result1 *unstructured.Unstructured
	}{result1}
}

func (fake *FakeRepoCache) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	fake.unchangedSinceCachedMutex.RLock()
	defer fake.unchangedSinceCachedMutex.RUnlock()
	fake.unchangedSinceCachedFromListMutex.RLock()
	defer fake.unchangedSinceCachedFromListMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRepoCache) recordInvocation(key string, args []interface{}) {
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

var _ repository.RepoCache = new(FakeRepoCache)
