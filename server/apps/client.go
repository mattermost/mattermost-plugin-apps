// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type Client interface {
	PostWish(Call) (*CallResponse, error)
	PostChangeNotification(Subscription, interface{})
	GetManifest(manifestURL string) (*Manifest, error)
	GetLocationsFromApp(appID AppID, userID, channelID string) ([]LocationInt, error)
}

const OutgoingAuthHeader = "Mattermost-App-Authorization"

func (s *Service) PostChangeNotification(sub Subscription, msg interface{}) {
	app, err := s.Registry.Get(sub.AppID)
	if err != nil {
		// <><> TODO log
		return
	}

	resp, err := s.post(app, "", path.Join(app.Manifest.RootURL, "notify", string(sub.Subject)), msg)
	if err != nil {
		// <><> TODO log
		return
	}
	defer resp.Body.Close()
}

func (s *Service) PostWish(call Call) (*CallResponse, error) {
	app, err := s.Registry.Get(call.Data.Context.AppID)
	if err != nil {
		return nil, err
	}
	resp, err := s.post(app, call.Data.Context.ActingUserID, call.Wish.URL, call.Data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cr := CallResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func (s *Service) post(toApp *App, fromMattermostUserID string, url string, msg interface{}) (*http.Response, error) {
	client, err := s.getAppHTTPClient(toApp.Manifest.AppID)
	if err != nil {
		return nil, err
	}
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
		// TODO ticket: progressive backoff on errors
		return nil, err
	}

	return resp, nil
}

func (s *Service) get(toApp *App, fromMattermostUserID string, url string) (*http.Response, error) {
	client, err := s.getAppHTTPClient(toApp.Manifest.AppID)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the client")
	}
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

func (s *Service) getAppHTTPClient(appID AppID) (*http.Client, error) {
	// <><> TODO cache the client per app
	return &http.Client{}, nil
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

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
}

func (s *Service) GetManifest(manifestURL string) (*Manifest, error) {
	var manifest Manifest
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

func (s *Service) GetLocationsFromApp(appID AppID, userID, channelID string) ([]LocationInt, error) {
	app, err := s.Registry.Get(appID)
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

	resp, err := s.get(app, userID, url.String())
	if err != nil {
		return nil, errors.Wrap(err, "error fetching the location")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("returned with status %s", resp.Status)
	}

	var bareLocations []map[string]interface{}
	locations := []LocationInt{}
	err = json.NewDecoder(resp.Body).Decode(&bareLocations)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling bare location list")
	}
	for _, bareLocation := range bareLocations {
		location, err := LocationFromMap(bareLocation)
		if err != nil {
			return nil, errors.Wrap(err, "error passing from map to location")
		}
		locations = append(locations, location)
	}

	return locations, nil
}
