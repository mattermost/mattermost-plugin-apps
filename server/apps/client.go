// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type Client interface {
	GetManifest(manifestURL string) (*api.Manifest, error)
	PostCall(*api.Call) (*api.CallResponse, error)
	PostNotification(*api.Notification) error
	GetBindings(*api.Context) ([]*api.Binding, error)
}

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
}

type client struct {
	store store.Service
}

func newClient(store store.Service) *client {
	return &client{
		store: store,
	}
}

func (c *client) PostNotification(n *api.Notification) error {
	app, err := c.store.GetApp(n.Context.AppID)
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

func (c *client) PostCall(call *api.Call) (*api.CallResponse, error) {
	app, err := c.store.GetApp(call.Context.AppID)
	if err != nil {
		return nil, err
	}

	resp, err := c.post(app, call.Context.ActingUserID, call.URL, call)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cr := api.CallResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

// post does not close resp.Body, it's the caller's responsibility
func (c *client) post(toApp *api.App, fromMattermostUserID string, url string, msg interface{}) (*http.Response, error) {
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
	req.Header.Set(OutgoingAuthHeader, "Bearer "+jwtoken)
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

func (c *client) get(toApp *api.App, fromMattermostUserID string, url string) (*http.Response, error) {
	client := c.getClient(toApp.Manifest.AppID)
	jwtoken, err := createJWT(fromMattermostUserID, toApp.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "error creating token")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}
	req.Header.Set(OutgoingAuthHeader, "Bearer "+jwtoken)

	resp, err := client.Do(req)
	if err != nil {
		// TODO ticket: progressive backoff on errors
		return nil, errors.Wrap(err, "error performing the request")
	}

	return resp, nil
}

func (c *client) getClient(appID api.AppID) *http.Client {
	return &http.Client{}
}

func createJWT(actingUserID, secret string) (string, error) {
	claims := JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
		ActingUserID: actingUserID,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (c *client) GetManifest(manifestURL string) (*api.Manifest, error) {
	var manifest api.Manifest
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

func (c *client) GetBindings(cc *api.Context) ([]*api.Binding, error) {
	app, err := c.store.GetApp(cc.AppID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get app")
	}

	resp, err := c.get(app, cc.ActingUserID, appendGetContext(app.Manifest.RootURL+constants.AppBindingsPath, cc))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bindings")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("returned with status %s", resp.Status)
	}

	out := []*api.Binding{}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling function")
	}
	return out, nil
}

func appendGetContext(inURL string, cc *api.Context) string {
	if cc == nil {
		return inURL
	}
	out, err := url.Parse(inURL)
	if err != nil {
		return inURL
	}
	q := out.Query()
	if cc.TeamID != "" {
		q.Add(constants.TeamID, cc.TeamID)
	}
	if cc.ChannelID != "" {
		q.Add(constants.ChannelID, cc.ChannelID)
	}
	if cc.ActingUserID != "" {
		q.Add(constants.ActingUserID, cc.ActingUserID)
	}
	if cc.PostID != "" {
		q.Add(constants.PostID, cc.PostID)
	}
	out.RawQuery = q.Encode()
	return out.String()
}
