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
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type Client interface {
	GetManifest(manifestURL string) (*api.Manifest, error)
	PostCall(call *api.Call) (*api.CallResponse, error)
	PostNotification(n *api.Notification) error
	GetLocations(appID api.AppID, userID, channelID string) ([]api.LocationInt, error)
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

	resp, err := c.post(app, call.Context.ActingUserID, call.FormURL, call)
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

func (c *client) GetLocations(appID api.AppID, userID, channelID string) ([]api.LocationInt, error) {
	app, err := c.store.GetApp(appID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting app")
	}

	url, err := url.Parse(app.Manifest.LocationsURL)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing the url")
	}
	q := url.Query()
	q.Add("user_id", userID)
	q.Add("channel_id", channelID)
	url.RawQuery = q.Encode()

	resp, err := c.get(app, userID, url.String())
	if err != nil {
		return nil, errors.Wrap(err, "error fetching the location")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("returned with status %s", resp.Status)
	}

	var bareLocations []map[string]interface{}
	locations := []api.LocationInt{}
	err = json.NewDecoder(resp.Body).Decode(&bareLocations)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling bare location list")
	}
	for _, bareLocation := range bareLocations {
		bareLocation["app_id"] = appID
		location, err := api.LocationFromMap(bareLocation)
		if err != nil {
			return nil, errors.Wrap(err, "error passing from map to location")
		}
		locations = append(locations, location)
	}

	return locations, nil
}
