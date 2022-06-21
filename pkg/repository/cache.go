// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//counterfeiter:generate . Logger
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
}

//counterfeiter:generate . RepoCache
type RepoCache interface {
	Set(submitted, persisted *unstructured.Unstructured)
	UnchangedSinceCached(submitted *unstructured.Unstructured, existingObj *unstructured.Unstructured) *unstructured.Unstructured
	UnchangedSinceCachedFromList(local *unstructured.Unstructured, remote []*unstructured.Unstructured) *unstructured.Unstructured
}

// fixme: get loggers from contexts (per request)
func NewCache(l Logger) RepoCache {
	return &cache{
		logger:         l,
		submittedCache: make(map[string]unstructured.Unstructured),
		persistedCache: make(map[string]unstructured.Unstructured),
	}
}

type cache struct {
	logger         Logger
	submittedCache map[string]unstructured.Unstructured
	persistedCache map[string]unstructured.Unstructured
}

func (c *cache) Set(submitted, persisted *unstructured.Unstructured) {
	key := getKey(submitted)
	c.submittedCache[key] = *submitted
	c.persistedCache[key] = *persisted
}

func (c *cache) UnchangedSinceCachedFromList(submitted *unstructured.Unstructured, existingList []*unstructured.Unstructured) *unstructured.Unstructured {
	key := getKey(submitted)
	c.logger.Info("checking for changes since cached", "key", key)
	if !c.isSubmittedCacheHit(submitted, key) {
		return nil
	}

	persistedCached := c.getPersistedCached(key)

	for _, existing := range existingList {
		if c.isPersistedCacheHit(key, existing, persistedCached) {
			return existing
		} else {
			continue
		}
	}

	c.logger.Info("miss: no matching existing object on apiserver", "key", key)
	return nil
}

func (c *cache) UnchangedSinceCached(submitted *unstructured.Unstructured, existingObj *unstructured.Unstructured) *unstructured.Unstructured {
	key := getKey(submitted)
	c.logger.Info("checking for changes since cached", "key", key)
	if !c.isSubmittedCacheHit(submitted, key) {
		return nil
	}

	persistedCached := c.getPersistedCached(key)

	if c.isPersistedCacheHit(key, existingObj, persistedCached) {
		return existingObj
	} else {
		return nil
	}
}

func (c *cache) isSubmittedCacheHit(submitted *unstructured.Unstructured, key string) bool {
	submittedCached, submittedFoundInCache := c.submittedCache[key]
	submittedUnchanged := submittedFoundInCache && reflect.DeepEqual(submittedCached, *submitted)

	if submittedUnchanged {
		c.logger.Info("no changes since last submission, checking existing objects on apiserver", "key", key)
		return true
	} else {
		if submittedFoundInCache {
			c.logger.Info("miss: submitted object in cache is different from submitted object", "key", key)
		} else {
			c.logger.Info("miss: object not in cache", "key", key)
		}
		return false
	}
}

func (c *cache) isPersistedCacheHit(key string, existingObj *unstructured.Unstructured, persistedCached *unstructured.Unstructured) bool {
	c.logger.Info("considering object", "key", key, "existingName", existingObj.GetName())
	existingSpec, ok := existingObj.Object["spec"]
	if !ok {
		c.logger.Info("object on apiserver has no spec", "key", key)
		return false
	}

	persistedCachedSpec, ok := persistedCached.Object["spec"]
	if !ok {
		c.logger.Info("persisted object in cache has no spec", "key", key)
		return false
	}

	sameSame := reflect.DeepEqual(existingSpec, persistedCachedSpec)
	if sameSame {
		c.logger.Info("hit: persisted object in cache matches spec on apiserver", "key", key)
		return true
	} else {
		c.logger.Info("miss: persisted object in cache DOES NOT match spec on apiserver", "key", key)
		return false
	}
}

func getKey(obj *unstructured.Unstructured) string {
	// todo: probably should hash object for key
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	var name string
	if obj.GetName() == "" {
		name = obj.GetGenerateName()
	} else {
		name = obj.GetName()
	}
	ns := obj.GetNamespace()
	return fmt.Sprintf("%s:%s:%s", ns, kind, name)
}

func (c *cache) getPersistedCached(key string) *unstructured.Unstructured {
	persisted := c.persistedCache[key]
	return &persisted
}
