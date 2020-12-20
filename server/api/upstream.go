// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"github.com/dgrijalva/jwt-go"
)

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type Upstream interface {
	GetBindings(*Call) ([]*Binding, error)
	Call(*Call) *CallResponse
	Notify(*Call) error
}

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
}
