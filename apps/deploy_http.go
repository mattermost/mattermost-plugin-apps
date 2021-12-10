// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/dgrijalva/jwt-go"

	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

// HTTP contains metadata for an app that is already, deployed externally
// and us accessed over HTTP. The JSON name `http` must match the type.
type HTTP struct {
	// All call and static paths are relative to the RootURL.
	RootURL string `json:"root_url,omitempty"`

	// UseJWT instructs the proxy to authenticate outgoing requests with a JWT.
	UseJWT bool `json:"use_jwt,omitempty"`
}

func (h *HTTP) Validate() error {
	if h == nil {
		return nil
	}
	if h.RootURL == "" {
		return utils.NewInvalidError("root_url must be set for HTTP apps")
	}
	err := httputils.IsValidURL(h.RootURL)
	if err != nil {
		return utils.NewInvalidError("invalid root_url: %q: %v", h.RootURL, err)
	}
	return nil
}

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
}
