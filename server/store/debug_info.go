// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type KVDebugAppInfo struct {
	AppCount       int
	AppByNamespace map[string]int
	AppByUserID    map[string]int
	UserCount      int
	TokenCount     int
}

type KVDebugInfo struct {
	AppCount          int
	ManifestsCount    int
	OAuth2StateCount  int
	SubscriptionCount int
	Total             int
	Apps              map[apps.AppID]*KVDebugAppInfo
}

func (i KVDebugInfo) forAppID(appID apps.AppID) *KVDebugAppInfo {
	appInfo, ok := i.Apps[appID]
	if ok {
		return appInfo
	}
	appInfo = &KVDebugAppInfo{
		AppByNamespace: map[string]int{},
		AppByUserID:    map[string]int{},
	}
	i.Apps[appID] = appInfo
	return appInfo
}

func (s *Service) GetDebugKVInfo() (*KVDebugInfo, error) {
	info := KVDebugInfo{
		Apps: map[apps.AppID]*KVDebugAppInfo{},
	}
	mm := s.conf.MattermostAPI()
	for i := 0; ; i++ {
		keys, err := mm.KV.ListKeys(i, ListKeysPerPage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}

		for _, key := range keys {
			info.Total++
			isHashKey := false
			if len(key) == hashKeyLength {
				gns, appID, userID, ns, _, _ := ParseHashkey(key)
				appInfo := info.forAppID(appID)
				isHashKey = true
				switch gns {
				case KVAppPrefix:
					appInfo.AppCount++
					appInfo.AppByNamespace[ns]++
					appInfo.AppByUserID[userID]++

				case KVUserPrefix:
					appInfo.UserCount++

				default:
					isHashKey = false
				}
			}
			if isHashKey {
				continue
			}

			switch {
			case strings.HasPrefix(key, KVSubPrefix):
				info.SubscriptionCount++

			case strings.HasPrefix(key, KVTokenPrefix):
				appID, _, err := parseKey(key)
				if err != nil {
					continue
				}
				info.forAppID(appID).TokenCount++

			case strings.HasPrefix(key, KVOAuth2StatePrefix):
				info.OAuth2StateCount++

			case strings.HasPrefix(key, KVAppPrefix):
				info.AppCount++

			case strings.HasPrefix(key, KVLocalManifestPrefix):
				info.ManifestsCount++
			}
		}
		if len(keys) < ListKeysPerPage {
			break
		}
	}
	return &info, nil
}
