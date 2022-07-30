// Code generated by counterfeiter. DO NOT EDIT.
package satokenfakes

import (
	"sync"

	"github.com/vmware-tanzu/cartographer/pkg/satoken"
)

type FakeLogger struct {
	ErrorStub        func(error, string, ...interface{})
	errorMutex       sync.RWMutex
	errorArgsForCall []struct {
		arg1 error
		arg2 string
		arg3 []interface{}
	}
	InfoStub        func(string, ...interface{})
	infoMutex       sync.RWMutex
	infoArgsForCall []struct {
		arg1 string
		arg2 []interface{}
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeLogger) Error(arg1 error, arg2 string, arg3 ...interface{}) {
	fake.errorMutex.Lock()
	fake.errorArgsForCall = append(fake.errorArgsForCall, struct {
		arg1 error
		arg2 string
		arg3 []interface{}
	}{arg1, arg2, arg3})
	stub := fake.ErrorStub
	fake.recordInvocation("Error", []interface{}{arg1, arg2, arg3})
	fake.errorMutex.Unlock()
	if stub != nil {
		fake.ErrorStub(arg1, arg2, arg3...)
	}
}

func (fake *FakeLogger) ErrorCallCount() int {
	fake.errorMutex.RLock()
	defer fake.errorMutex.RUnlock()
	return len(fake.errorArgsForCall)
}

func (fake *FakeLogger) ErrorCalls(stub func(error, string, ...interface{})) {
	fake.errorMutex.Lock()
	defer fake.errorMutex.Unlock()
	fake.ErrorStub = stub
}

func (fake *FakeLogger) ErrorArgsForCall(i int) (error, string, []interface{}) {
	fake.errorMutex.RLock()
	defer fake.errorMutex.RUnlock()
	argsForCall := fake.errorArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeLogger) Info(arg1 string, arg2 ...interface{}) {
	fake.infoMutex.Lock()
	fake.infoArgsForCall = append(fake.infoArgsForCall, struct {
		arg1 string
		arg2 []interface{}
	}{arg1, arg2})
	stub := fake.InfoStub
	fake.recordInvocation("Info", []interface{}{arg1, arg2})
	fake.infoMutex.Unlock()
	if stub != nil {
		fake.InfoStub(arg1, arg2...)
	}
}

func (fake *FakeLogger) InfoCallCount() int {
	fake.infoMutex.RLock()
	defer fake.infoMutex.RUnlock()
	return len(fake.infoArgsForCall)
}

func (fake *FakeLogger) InfoCalls(stub func(string, ...interface{})) {
	fake.infoMutex.Lock()
	defer fake.infoMutex.Unlock()
	fake.InfoStub = stub
}

func (fake *FakeLogger) InfoArgsForCall(i int) (string, []interface{}) {
	fake.infoMutex.RLock()
	defer fake.infoMutex.RUnlock()
	argsForCall := fake.infoArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeLogger) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.errorMutex.RLock()
	defer fake.errorMutex.RUnlock()
	fake.infoMutex.RLock()
	defer fake.infoMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeLogger) recordInvocation(key string, args []interface{}) {
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

var _ satoken.Logger = new(FakeLogger)
