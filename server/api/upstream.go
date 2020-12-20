// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"io"

	"github.com/dgrijalva/jwt-go"
)

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type Upstream interface {
	Roundtrip(call *Call) (io.ReadCloser, error)
	OneWay(call *Call) error
}

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
}
