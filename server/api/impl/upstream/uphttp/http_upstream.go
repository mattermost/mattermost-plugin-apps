// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type Upstream struct {
	rootURL   string
	appSecret string
}

func NewUpstream(app *api.App) *Upstream {
	return &Upstream{app.Manifest.RootURL, app.Secret}
}

// post does not close resp.Body, it's the caller's responsibility
func (u *Upstream) post(fromMattermostUserID string, url string, msg interface{}) (*http.Response, error) {
	client := u.getClient()
	jwtoken, err := createJWT(fromMattermostUserID, u.appSecret)
	if err != nil {
		return nil, err
	}

	piper, pipew := io.Pipe()
	go func() {
		encodeErr := json.NewEncoder(pipew).Encode(msg)
		if encodeErr != nil {
			_ = pipew.CloseWithError(encodeErr)
		}
		pipew.Close()
	}()

	req, err := http.NewRequest(http.MethodPost, url, piper)
	if err != nil {
		return nil, err
	}
	req.Header.Set(api.OutgoingAuthHeader, "Bearer "+jwtoken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, httputils.DecodeJSONError(resp.Body)
	}

	return resp, nil
}

// func (u *Upstream) get(fromMattermostUserID string, url string) (*http.Response, error) {
// 	client := u.getClient()
// 	jwtoken, err := createJWT(fromMattermostUserID, u.appSecret)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "error creating token")
// 	}

// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "error creating request")
// 	}
// 	req.Header.Set(api.OutgoingAuthHeader, "Bearer "+jwtoken)

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		// TODO ticket: progressive backoff on errors
// 		return nil, errors.Wrap(err, "error performing the request")
// 	}

// 	return resp, nil
// }

func (u *Upstream) getClient() *http.Client {
	return &http.Client{}
}

func createJWT(actingUserID, secret string) (string, error) {
	claims := api.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
		ActingUserID: actingUserID,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// func appendGetContext(inURL string, cc *api.Context) string {
// 	if cc == nil {
// 		return inURL
// 	}
// 	out, err := url.Parse(inURL)
// 	if err != nil {
// 		return inURL
// 	}
// 	q := out.Query()
// 	if cc.TeamID != "" {
// 		q.Add(api.PropTeamID, cc.TeamID)
// 	}
// 	if cc.ChannelID != "" {
// 		q.Add(api.PropChannelID, cc.ChannelID)
// 	}
// 	if cc.ActingUserID != "" {
// 		q.Add(api.PropActingUserID, cc.ActingUserID)
// 	}
// 	if cc.PostID != "" {
// 		q.Add(api.PropPostID, cc.PostID)
// 	}
// 	out.RawQuery = q.Encode()
// 	return out.String()
// }
