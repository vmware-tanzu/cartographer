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
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const submittedCachePrefix = "submitted"
const persistedCachePrefix = "persisted"
const CacheExpiryDuration = 1 * time.Hour

//counterfeiter:generate . ExpiringCache
type ExpiringCache interface {
	Get(key interface{}) (val interface{}, ok bool)
	Set(key interface{}, val interface{}, ttl time.Duration)
}

//counterfeiter:generate . RepoCache
type RepoCache interface {
	Set(submitted, persisted *unstructured.Unstructured)
	UnchangedSinceCached(local, remote *unstructured.Unstructured) bool
	Refresh(submitted *unstructured.Unstructured)
}

func NewCache(c ExpiringCache) RepoCache {
	return &cache{
		ec: c,
	}
}

type cache struct {
	ec ExpiringCache
}

func (c *cache) Set(submitted, persisted *unstructured.Unstructured) {
	submittedKey := getKey(submitted, submittedCachePrefix)
	persistedKey := getKey(submitted, persistedCachePrefix)
	c.ec.Set(submittedKey, *submitted, CacheExpiryDuration)
	c.ec.Set(persistedKey, *persisted, CacheExpiryDuration)
}

func (c *cache) Refresh(submitted *unstructured.Unstructured) {
	submittedKey := getKey(submitted, submittedCachePrefix)
	persistedKey := getKey(submitted, persistedCachePrefix)
	if submittedCached, ok := c.ec.Get(submittedKey); ok {
		if persistedCached, ok := c.ec.Get(persistedKey); ok {
			c.ec.Set(submittedKey, submittedCached, CacheExpiryDuration)
			c.ec.Set(persistedKey, persistedCached, CacheExpiryDuration)
		}
	}
}

func (c *cache) UnchangedSinceCached(submitted, existing *unstructured.Unstructured) bool {
	submittedKey := getKey(submitted, submittedCachePrefix)
	persistedKey := getKey(submitted, persistedCachePrefix)
	submittedCached, ok := c.ec.Get(submittedKey)
	submittedUnchanged := ok && reflect.DeepEqual(submittedCached, *submitted)

	persistedCached := c.getPersistedCached(persistedKey)

	if !submittedUnchanged {
		return false
	}

	existingSpec, ok := existing.Object["spec"]
	if !ok {
		return false
	}

	persistedCachedSpec, ok := persistedCached.Object["spec"]
	if !ok {
		return false
	}

	return reflect.DeepEqual(existingSpec, persistedCachedSpec)
}

func getKey(obj *unstructured.Unstructured, prefix string) string {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	name := obj.GetName()
	ns := obj.GetNamespace()
	return fmt.Sprintf("%s:%s:%s:%s", prefix, ns, kind, name)
}

func (c *cache) getPersistedCached(persistedKey string) *unstructured.Unstructured {
	var persistedCached unstructured.Unstructured

	persistedCachedUntyped, ok := c.ec.Get(persistedKey)
	if !ok {
		persistedCachedUntyped = unstructured.Unstructured{}
	}

	persistedCached, ok = persistedCachedUntyped.(unstructured.Unstructured)
	if !ok {
		persistedCached = unstructured.Unstructured{}
	}

	return &persistedCached
}
