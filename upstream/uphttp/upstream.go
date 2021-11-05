// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"bytes"
	"context"
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
	if !app.Manifest.Contains(apps.DeployHTTP) {
		return "", errors.New("failed to get root URL: no http section in manifest.json")
	}
	return app.Manifest.HTTP.RootURL, nil
}

func (u *Upstream) Roundtrip(ctx context.Context, app apps.App, creq apps.CallRequest, async bool) (io.ReadCloser, error) {
	if async {
		go func() {
			resp, _ := u.invoke(context.Background(), creq.Context.BotUserID, app, creq)
			if resp != nil {
				resp.Body.Close()
			}
		}()
		return nil, nil
	}

	resp, err := u.invoke(ctx, creq.Context.ActingUserID, app, creq) // nolint:bodyclose
	if err != nil {
		return nil, errors.Wrap(err, "failed to invoke via HTTP")
	}
	return resp.Body, nil
}

func (u *Upstream) invoke(ctx context.Context, fromMattermostUserID string, app apps.App, creq apps.CallRequest) (*http.Response, error) {
	rootURL, err := u.appRootURL(app, creq.Path)
	if err != nil {
		return nil, err
	}
	callURL, err := utils.CleanURL(rootURL + "/" + creq.Path)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(creq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, callURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// TODO: find a better way to control the use of JWT that both OpenFaaS and
	// HTTP can share. For now, hard-limit the use of JWT to the HTTP gateway
	// itself.
	if app.Manifest.Contains(apps.DeployHTTP) && app.Manifest.HTTP.UseJWT {
		jwtoken := ""
		jwtoken, err = createJWT(fromMattermostUserID, app.Secret)
		if err != nil {
			return nil, err
		}
		req.Header.Set(apps.OutgoingAuthHeader, "Bearer "+jwtoken)
	}

	// Execute the request.
	resp, err := u.httpOut.MakeClient(u.devMode).Do(req)
	switch {
	case err != nil:
		return nil, err

	case resp.StatusCode == http.StatusNotFound:
		return nil, utils.NewNotFoundError(err)

	case resp.StatusCode != http.StatusOK:
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
