// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type client struct {
	s *service
}

func (s *service) newClient() *client {
	return &client{
		s: s,
	}
}

func (c *client) PostNotification(n *apps.Notification) error {
	app, err := c.s.GetApp(n.Context.AppID)
	if err != nil {
		return err
	}

	resp, err := c.post(app, "", app.Manifest.RootURL+"/notify/"+string(n.Subject), n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) PostCall(call *apps.Call) (*apps.CallResponse, error) {
	app, err := c.s.GetApp(call.Context.AppID)
	if err != nil {
		return nil, err
	}

	resp, err := c.post(app, call.Context.ActingUserID, call.URL, call)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cr := apps.CallResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

// post does not close resp.Body, it's the caller's responsibility
func (c *client) post(toApp *apps.App, fromMattermostUserID string, url string, msg interface{}) (*http.Response, error) {
	client := c.getClient(toApp.Manifest.AppID)
	jwtoken, err := createJWT(fromMattermostUserID, toApp.Secret)
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
	req.Header.Set(apps.OutgoingAuthHeader, "Bearer "+jwtoken)
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

func (c *client) get(toApp *apps.App, fromMattermostUserID string, url string) (*http.Response, error) {
	client := c.getClient(toApp.Manifest.AppID)
	jwtoken, err := createJWT(fromMattermostUserID, toApp.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "error creating token")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}
	req.Header.Set(apps.OutgoingAuthHeader, "Bearer "+jwtoken)

	resp, err := client.Do(req)
	if err != nil {
		// TODO ticket: progressive backoff on errors
		return nil, errors.Wrap(err, "error performing the request")
	}

	return resp, nil
}

func (c *client) getClient(appID apps.AppID) *http.Client {
	return &http.Client{}
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

func (c *client) GetManifest(manifestURL string) (*apps.Manifest, error) {
	var manifest apps.Manifest
	resp, err := http.Get(manifestURL) // nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (c *client) GetBindings(cc *apps.Context) ([]*apps.Binding, error) {
	app, err := c.s.GetApp(cc.AppID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get app")
	}

	resp, err := c.get(app, cc.ActingUserID, appendGetContext(app.Manifest.RootURL+apps.AppBindingsPath, cc))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bindings")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("returned with status %s", resp.Status)
	}

	out := []*apps.Binding{}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling function")
	}
	return out, nil
}

func appendGetContext(inURL string, cc *apps.Context) string {
	if cc == nil {
		return inURL
	}
	out, err := url.Parse(inURL)
	if err != nil {
		return inURL
	}
	q := out.Query()
	if cc.TeamID != "" {
		q.Add(apps.PropTeamID, cc.TeamID)
	}
	if cc.ChannelID != "" {
		q.Add(apps.PropChannelID, cc.ChannelID)
	}
	if cc.ActingUserID != "" {
		q.Add(apps.PropActingUserID, cc.ActingUserID)
	}
	if cc.PostID != "" {
		q.Add(apps.PropPostID, cc.PostID)
	}
	out.RawQuery = q.Encode()
	return out.String()
}
