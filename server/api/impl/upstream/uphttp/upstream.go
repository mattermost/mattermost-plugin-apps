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
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type Upstream struct {
	rootURL   string
	appSecret string
}

func NewUpstream(app *api.App) *Upstream {
	return &Upstream{app.Manifest.HTTPRootURL, app.Secret}
}

func (u *Upstream) InvokeNotification(n *api.Notification) error {
	// TODO
	resp, err := u.post("", u.rootURL+"/notify/"+string(n.Subject), n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (u *Upstream) InvokeCall(call *api.Call) *api.CallResponse {
	resp, err := u.post(call.Context.ActingUserID, u.rootURL+call.URL, call)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	defer resp.Body.Close()

	cr := api.CallResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	return &cr
}

func (u *Upstream) GetBindings(call *api.Call) ([]*api.Binding, error) {
	resp, err := u.post(call.Context.ActingUserID, u.rootURL+call.URL, call)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return upstream.DecodeBindingsResponse(resp.Body)
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

	// fmt.Printf("<><> post: %s\n", url)
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
