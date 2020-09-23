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

const AuthHeader = "Mattermost-App-Authorization"

func (s *Service) PostChangeNotification(appID AppID, sub *Subscription, msg interface{}) {
	app, err := s.Registry.Get(appID)
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

func (s *Service) PostWish(appID AppID, fromMattermostUserID string, w *Wish, data *CallData) (*CallResponse, error) {
	app, err := s.Registry.Get(appID)
	if err != nil {
		return nil, err
	}
	resp, err := s.post(app, fromMattermostUserID, w.URL, data)
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

	var encodeErr error
	piper, pipew := io.Pipe()
	go func() {
		defer pipew.Close()
		encodeErr = json.NewEncoder(pipew).Encode(msg)
	}()
	req, err := http.NewRequest(http.MethodPost, url, piper)
	if err != nil {
		return nil, err
	}
	req.Header.Set(AuthHeader, "Bearer "+jwtoken)

	resp, err := client.Do(req)
	if err != nil {
		// TODO ticket: progressive backoff on errors
		return nil, err
	}
	if encodeErr != nil {
		return nil, encodeErr
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
	resp, err := http.Get(manifestURL)
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
