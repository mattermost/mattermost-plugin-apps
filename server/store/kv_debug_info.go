// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
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
	Apps                   map[apps.AppID]*KVDebugAppInfo
	AppsTotal              int
	CachedStoreTotal       int
	CachedStoreCountByName map[string]int
	Debug                  int
	OAuth2StateCount       int
	Other                  int
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

func GetKVDebugInfo(r *incoming.Request) (*KVDebugInfo, error) {
	info := KVDebugInfo{
		Apps:                   map[apps.AppID]*KVDebugAppInfo{},
		CachedStoreCountByName: map[string]int{},
	}
	for i := 0; ; i++ {
		keys, err := r.API.Mattermost.KV.ListKeys(i, ListKeysPerPage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}
		if len(keys) == 0 {
			break
		}

		for _, key := range keys {
			info.Total++
			isHashKey := false
			if len(key) == hashKeyLength {
				gns, appID, userID, ns, _, _ := ParseHashkey(key)
				appInfo := info.forAppID(appID)
				isHashKey = true
				switch gns {
				case KVPrefix:
					appInfo.AppKVCount++
					appInfo.AppKVCountByNamespace[ns]++
					appInfo.AppKVCountByUserID[userID]++
					info.AppsTotal++

				case UserPrefix:
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
			case strings.HasPrefix(key, TokenPrefix):
				appID, _, err := parseSessionKey(key)
				if err != nil {
					continue
				}
				info.forAppID(appID).TokenCount++
				info.AppsTotal++

			case strings.HasPrefix(key, OAuth2StatePrefix):
				info.OAuth2StateCount++

			case strings.HasPrefix(key, CachedPrefix):
				name, _, _ := parseCachedStoreKey(key)
				if name != "" {
					info.CachedStoreCountByName[name]++
				}
				info.CachedStoreTotal++

			case key == "mmi_botid":
				info.Other++

			case strings.HasPrefix(key, DebugPrefix):
				info.Debug++
			}
		}
	}
	return &info, nil
}
