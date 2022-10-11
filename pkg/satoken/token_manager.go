// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// wokeignore:rule=master K8s uses this term a lot
// This file is a modified version of https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/token/token_manager.go

package satoken

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	maxTTL    = 2 * time.Hour
	gcPeriod  = time.Hour * 24
	maxJitter = 10 * time.Second
)

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate k8s.io/client-go/kubernetes.Interface
//counterfeiter:generate k8s.io/client-go/kubernetes/typed/core/v1.CoreV1Interface
//counterfeiter:generate k8s.io/client-go/kubernetes/typed/core/v1.ServiceAccountInterface

//counterfeiter:generate . Logger
type Logger interface {
	Error(err error, msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
}

//counterfeiter:generate . TokenManager
type TokenManager interface {
	GetServiceAccountToken(serviceAccount *corev1.ServiceAccount) (string, error)
	Cleanup()
}

// NewManager returns a new token manager, primed with the tokenCache.
// A nil token cache will cause a new one to be automatically created.
func NewManager(cl clientset.Interface, log Logger, tokenCache map[string]*authenticationv1.TokenRequest) *Manager {
	if tokenCache == nil {
		tokenCache = make(map[string]*authenticationv1.TokenRequest)
	}
	m := &Manager{
		cl:    cl,
		cache: tokenCache,
		log:   log,
	}
	go wait.Forever(m.Cleanup, gcPeriod)
	return m
}

type Manager struct {
	cl clientset.Interface

	// cacheMutex guards the cache
	cacheMutex sync.RWMutex
	cache      map[string]*authenticationv1.TokenRequest

	log Logger
}

// GetServiceAccountToken gets a service account token from cache or
// from the TokenRequest API. This process is as follows:
// * Check the cache for the current token request.
// * If the token exists and does not require a refresh, return the current token.
// * Attempt to refresh the token.
// * If the token is refreshed successfully, save it in the cache and return the token.
// * If refresh fails and the old token is still valid, log an error and return the old token.
// * If refresh fails and the old token is no longer valid, return an error
func (m *Manager) GetServiceAccountToken(serviceAccount *corev1.ServiceAccount) (string, error) {
	key := fmt.Sprintf("%q/%q/%q", serviceAccount.Name, serviceAccount.Namespace, serviceAccount.UID)
	ctr, ok := m.get(key)

	if ok && !m.requiresRefresh(ctr) {
		return ctr.Status.Token, nil
	}

	expiration := int64(maxTTL.Seconds())
	tr := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: &expiration,
		}}
	tr, err := m.getToken(serviceAccount.Name, serviceAccount.Namespace, tr)

	if err != nil {
		switch {
		case !ok:
			return "", fmt.Errorf("fetch token: %v", err)
		case m.expired(ctr):
			return "", fmt.Errorf("token %s expired and refresh failed: %v", key, err)
		default:
			m.log.Error(err, "update token", "cacheKey", key)
			return ctr.Status.Token, nil
		}
	}

	m.set(key, tr)
	return tr.Status.Token, nil
}

func (m *Manager) Cleanup() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	for k, tr := range m.cache {
		if m.expired(tr) {
			delete(m.cache, k)
		}
	}
}

func (m *Manager) getToken(name, namespace string, tr *authenticationv1.TokenRequest) (*authenticationv1.TokenRequest, error) {
	return m.cl.CoreV1().ServiceAccounts(namespace).CreateToken(context.TODO(), name, tr, metav1.CreateOptions{})
}

func (m *Manager) get(key string) (*authenticationv1.TokenRequest, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	ctr, ok := m.cache[key]
	return ctr, ok
}

func (m *Manager) set(key string, tr *authenticationv1.TokenRequest) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.cache[key] = tr
}

func (m *Manager) expired(t *authenticationv1.TokenRequest) bool {
	return time.Now().After(t.Status.ExpirationTimestamp.Time)
}

// requiresRefresh returns true if the token is older than half of it's maxTTL
func (m *Manager) requiresRefresh(tr *authenticationv1.TokenRequest) bool {
	if tr.Spec.ExpirationSeconds == nil {
		cpy := tr.DeepCopy()
		cpy.Status.Token = ""
		m.log.Info("Expiration seconds was nil for token request", "tokenRequest", cpy)
		return false
	}
	now := time.Now()
	exp := tr.Status.ExpirationTimestamp.Time
	iat := exp.Add(-1 * time.Duration(*tr.Spec.ExpirationSeconds) * time.Second)

	jitter := time.Duration(rand.Float64()*maxJitter.Seconds()) * time.Second
	if now.After(iat.Add(maxTTL - jitter)) {
		return true
	}

	// Require a refresh if within 50% of the TTL plus a jitter from the expiration time.
	return now.After(exp.Add(-1*time.Duration((*tr.Spec.ExpirationSeconds*50)/100)*time.Second - jitter))
}
