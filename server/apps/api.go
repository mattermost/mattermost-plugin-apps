// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

type API interface {
	Call(Call) (*CallResponse, error)
	InstallApp(InInstallApp) (*OutInstallApp, error)
	GetWidgets(userID, channelID string) ([]LocationInt, error)
}
