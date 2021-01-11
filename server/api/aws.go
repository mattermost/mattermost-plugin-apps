// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

type AWS interface {
	ProvisionApp(releaseURL string) error
	InvokeLambda(appID, appVersion, functionName, invocationType string, request interface{}) ([]byte, error)
}
