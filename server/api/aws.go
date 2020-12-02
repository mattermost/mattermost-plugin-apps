// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

type AWS interface {
	InstallApp(releaseURL string) error
	InvokeLambda(functionName string, request interface{}) ([]byte, error)
}
