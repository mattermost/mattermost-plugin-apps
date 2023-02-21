// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package httpin

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

// CreateTimer create or updates a new statefull timer.
//
//	Path: /api/v1/timer
//	Method: POST
//	Input: JSON {at, call, state}
//	Output: None
func (s *Service) CreateTimer(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var t apps.Timer

	err := json.NewDecoder(req.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.AppServices.CreateTimer(r, t)
	if err != nil {
		http.Error(w, err.Error(), httputils.ErrorToStatus(err))
		return
	}
}
