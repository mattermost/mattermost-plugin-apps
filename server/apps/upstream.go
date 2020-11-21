// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/dgrijalva/jwt-go"
)

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type Upstream interface {
	GetUpstreamBindings(*Context) ([]*Binding, error)
	GetManifest(manifestURL string) (*Manifest, error)
	PostUpstreamCall(*Call) (*CallResponse, error)
	PostUpstreamNotification(*Notification) error
}

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
}
