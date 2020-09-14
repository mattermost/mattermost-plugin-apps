// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package cloudapps

import "github.com/mattermost/mattermost-plugin-cloudapps/server/utils"

// <><> TODO remove mock, implement for real
func (r *registry) GetApp(appID AppID) (*App, error) {
	app, found := r.apps[appID]
	if !found {
		return nil, utils.ErrNotFound
	}
	return app, nil
}
