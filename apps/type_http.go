// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
)

type HTTP struct {
	// All call and static paths are relative to the RootURL.
	RootURL string `json:"root_url,omitempty"`
}

func (h *HTTP) IsValid() error {
	if h == nil {
		return nil
	}
	if h.RootURL == "" {
		return utils.NewInvalidError(errors.New("root_url must be set for HTTP apps"))
	}
	err := utils.IsValidHTTPURL(h.RootURL)
	if err != nil {
		return utils.NewInvalidError(errors.Wrapf(err, "invalid root_url: %q", h.RootURL))
	}
	return nil
}

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
}
