// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import "github.com/dgrijalva/jwt-go"

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
}
