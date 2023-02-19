// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type TestingCachedStore[T Cloneable[T]] map[string]T

var _ CachedStore[apps.App] = (TestingCachedStore[apps.App])(nil)

func MakeTestCachedStore[T Cloneable[T]](name string, c *CachedStoreCluster, log utils.Logger) (CachedStore[T], error) {
	return TestingCachedStore[T]{}, nil
}

func (s TestingCachedStore[T]) Get(key string) *T {
	value, ok := s[key]
	if ok {
		return &value
	}
	return nil
}

func (s TestingCachedStore[T]) Put(r *incoming.Request, key string, value *T) error {
	if value != nil {
		s[key] = *value
	} else {
		delete(s, key)
	}
	return nil
}

func (s TestingCachedStore[T]) Index() CachedIndex[T] {
	out := CachedIndex[T]{}
	for key, value := range s {
		out[key] = &CachedIndexEntry[T]{
			Key:       key,
			ValueHash: valueHash(value),
			data:      value.Clone(),
		}
	}
	return out
}
