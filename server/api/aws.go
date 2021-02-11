// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import "github.com/mattermost/mattermost-plugin-apps/apps"

type AWS interface {
	ProvisionAppFromURL(releaseURL string, shouldUpdate bool) error
	InvokeLambda(appID apps.AppID, appVersion apps.AppVersion, functionName, invocationType string, request interface{}) ([]byte, error)
}
