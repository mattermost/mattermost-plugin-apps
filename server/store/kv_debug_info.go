// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type KVDebugAppInfo struct {
	AppKVCount            int
	AppKVCountByNamespace map[string]int
	AppKVCountByUserID    map[string]int
	TokenCount            int
	UserCount             int
}

func (i KVDebugAppInfo) Total() int {
	return i.AppKVCount + i.UserCount + i.TokenCount
}

type KVDebugInfo struct {
	InstalledAppCount      int
	Apps                   map[apps.AppID]*KVDebugAppInfo
	AppsTotal              int
	ManifestCount          int
	OAuth2StateCount       int
	Other                  int
	SubscriptionCount      int
	CachedStoreCount       int
	CachedStoreCountByName map[string]int
	Total                  int
}

func (i KVDebugInfo) forAppID(appID apps.AppID) *KVDebugAppInfo {
	appInfo, ok := i.Apps[appID]
	if ok {
		return appInfo
	}
	appInfo = &KVDebugAppInfo{
		AppKVCountByNamespace: map[string]int{},
		AppKVCountByUserID:    map[string]int{},
	}
	i.Apps[appID] = appInfo
	return appInfo
}

func (s *Service) GetDebugKVInfo(log utils.Logger) (*KVDebugInfo, error) {
	info := KVDebugInfo{
		Apps:                   map[apps.AppID]*KVDebugAppInfo{},
		CachedStoreCountByName: map[string]int{},
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
					appInfo.AppKVCount++
					appInfo.AppKVCountByNamespace[ns]++
					appInfo.AppKVCountByUserID[userID]++
					info.AppsTotal++

				case KVUserPrefix:
					appInfo.UserCount++
					info.AppsTotal++

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
				appID, _, err := parseTokenKey(key)
				if err != nil {
					continue
				}
				info.forAppID(appID).TokenCount++
				info.AppsTotal++

			case strings.HasPrefix(key, KVOAuth2StatePrefix):
				info.OAuth2StateCount++

			case strings.HasPrefix(key, KVInstalledAppPrefix):
				info.InstalledAppCount++

			case strings.HasPrefix(key, KVLocalManifestPrefix):
				info.ManifestCount++

			case strings.HasPrefix(key, KVCachedPrefix):
				name, _, _ := parseCachedStoreKey(key)
				if name != "" {
					info.CachedStoreCountByName[name]++
				}
				info.CachedStoreCount++

			case key == "mmi_botid":
				info.Other++
			}
		}
		if len(keys) < ListKeysPerPage {
			break
		}
	}
	return &info, nil
}
