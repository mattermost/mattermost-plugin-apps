// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	GetLocation(lr *LocationRegistry, userID, channelID string) (LocationInt, error)
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

func (s *Service) GetLocation(lr *LocationRegistry, userID, channelID string) (LocationInt, error) {
	url, err := url.Parse(lr.FetchURL)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing the url")
	}
	q := url.Query()
	q.Add("userID", userID)
	q.Add("channelID", channelID)
	url.RawQuery = q.Encode()

	app, err := s.Registry.Get(lr.AppID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting app")
	}

	resp, err := s.get(app, userID, url.String())
	if err != nil {
		return nil, errors.Wrap(err, "error fetching the location")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("returned with status %s", resp.Status)
	}

	var bareLocation Location
	buf, _ := ioutil.ReadAll(resp.Body)
	decoder := json.NewDecoder(bytes.NewReader(buf))
	err = decoder.Decode(&bareLocation)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding the bare location")
	}

	var location LocationInt
	decoder = json.NewDecoder(bytes.NewReader(buf))
	switch bareLocation.GetType() {
	case LocationChannelHeaderIcon:
		var specificLocation ChannelHeaderIconLocation
		err = decoder.Decode(&specificLocation)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding channel header icon location")
		}
		location = &specificLocation
	case LocationPostMenuItem:
		var specificLocation PostMenuItemLocation
		err = decoder.Decode(&specificLocation)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding post menu item location")
		}
		location = &specificLocation
	default:
		return nil, errors.New("location not recognized")
	}
	return location, nil
}
