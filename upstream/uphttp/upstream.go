// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type Upstream struct {
	httpOut    httpout.Service
	appRootURL func(_ apps.App, path string) (string, error)
	devMode    bool
}

var _ upstream.Upstream = (*Upstream)(nil)

func NewUpstream(httpOut httpout.Service, devMode bool, appRootURL func(apps.App, string) (string, error)) *Upstream {
	if appRootURL == nil {
		appRootURL = AppRootURL
	}
	return &Upstream{
		httpOut:    httpOut,
		appRootURL: appRootURL,
		devMode:    devMode,
	}
}

func AppRootURL(app apps.App, _ string) (string, error) {
	if !app.Manifest.SupportsDeploy(apps.DeployHTTP) {
		return "", errors.New("failed to get root URL: no http section in manifest.json")
	}
	return app.Manifest.HTTP.RootURL, nil
}

func (u *Upstream) Roundtrip(app apps.App, creq apps.CallRequest, async bool) (io.ReadCloser, error) {
	if async {
		go func() {
			resp, _ := u.invoke(creq.Context.BotUserID, app, creq)
			if resp != nil {
				resp.Body.Close()
			}
		}()
		return nil, nil
	}

	resp, err := u.invoke(creq.Context.ActingUserID, app, creq) // nolint:bodyclose
	if err != nil {
		return nil, errors.Wrap(err, "failed to invoke via HTTP")
	}
	return resp.Body, nil
}

func (u *Upstream) invoke(fromMattermostUserID string, app apps.App, creq apps.CallRequest) (*http.Response, error) {
	rootURL, err := u.appRootURL(app, creq.Path)
	if err != nil {
		return nil, err
	}
	callURL, err := utils.CleanURL(rootURL + "/" + creq.Path)
	if err != nil {
		return nil, err
	}

	client := u.httpOut.MakeClient(u.devMode)
	jwtoken, err := createJWT(fromMattermostUserID, app.Secret)
	if err != nil {
		return nil, err
	}

	piper, pipew := io.Pipe()
	go func() {
		encodeErr := json.NewEncoder(pipew).Encode(creq)
		if encodeErr != nil {
			_ = pipew.CloseWithError(encodeErr)
		}
		pipew.Close()
	}()

	req, err := http.NewRequest(http.MethodPost, callURL, piper)
	if err != nil {
		return nil, err
	}
	req.Header.Set(apps.OutgoingAuthHeader, "Bearer "+jwtoken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		bb, _ := httputils.ReadAndClose(resp.Body)
		return nil, errors.New(string(bb))
	}

	return resp, nil
}

func createJWT(actingUserID, secret string) (string, error) {
	claims := apps.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
		ActingUserID: actingUserID,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
