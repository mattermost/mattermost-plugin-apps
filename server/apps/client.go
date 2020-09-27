// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Client interface {
	PostWish(Call) (*CallResponse, error)
	PostChangeNotification(Subscription, interface{})
	GetManifest(manifestURL string) (*Manifest, error)
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
