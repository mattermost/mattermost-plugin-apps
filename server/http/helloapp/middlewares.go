package helloapp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/pkg/errors"
)

func checkAuthentication(f func(http.ResponseWriter, *http.Request, apps.JWTClaims)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		authValue := req.Header.Get(apps.OutgoingAuthHeader)
		if !strings.HasPrefix(authValue, "Bearer ") {
			httputils.WriteBadRequestError(w, errors.Errorf("missing %s: Bearer header", apps.OutgoingAuthHeader))
			return
		}

		jwtoken := strings.TrimPrefix(authValue, "Bearer ")
		claims := apps.JWTClaims{}
		_, err := jwt.ParseWithClaims(jwtoken, &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(appSecret), nil
		})
		if err != nil {
			httputils.WriteBadRequestError(w, err)
			return
		}

		f(w, req, claims)
	}
}

func extractUserAndChannelID(f func(http.ResponseWriter, *http.Request, string, string)) func(http.ResponseWriter, *http.Request, apps.JWTClaims) {
	return func(w http.ResponseWriter, req *http.Request, _ apps.JWTClaims) {
		userID := req.URL.Query().Get("user_id")
		if userID == "" {
			httputils.WriteBadRequestError(w, errors.New("missing user ID"))
			return
		}

		channelID := req.URL.Query().Get("channel_id")
		if channelID == "" {
			httputils.WriteBadRequestError(w, errors.New("missing channel ID"))
			return
		}

		f(w, req, userID, channelID)
	}
}
