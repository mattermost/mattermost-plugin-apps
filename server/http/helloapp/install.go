package helloapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) handleInstall(w http.ResponseWriter, req *http.Request) {
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
		return []byte(AppSecret), nil
	})
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	data := apps.CallData{}
	err = json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	// The freshly created bot token is largely useless, so we need the acting
	// user (sysadmin) to OAuth2 connect first. This can be done after OAuth2
	// (OAuther) is fully integrated.

	// TODO Install: create channel, subscribe, etc.

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.ResponseTypeOK,
			Markdown: "Installed! <><>",
			Data:     map[string]interface{}{"status": "ok"},
		})
}
