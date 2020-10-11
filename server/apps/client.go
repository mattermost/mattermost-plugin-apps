// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

type Client interface {
	PostWish(*Call) (*CallResponse, error)
	PostNotification(*NotificationRequest) error
	GetManifest(manifestURL string) (*store.Manifest, error)
}

const OutgoingAuthHeader = "Mattermost-App-Authorization"

func (s *service) PostNotification(n *NotificationRequest) error {
	app, err := s.Store.GetApp(n.Context.AppID)
	if err != nil {
		return err
	}
	fmt.Printf("<><> PostChangeNotification: %+v\n", app)

	resp, err := s.post(app, "", path.Join(app.Manifest.RootURL, "notify", string(n.Subject)), n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *service) PostWish(call *Call) (*CallResponse, error) {
	app, err := s.Store.GetApp(call.Request.Context.AppID)
	if err != nil {
		return nil, err
	}
	resp, err := s.post(app, call.Request.Context.ActingUserID, call.Wish.URL, call.Request)
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

// post does not close resp.Body, it's the caller's responsibility
func (s *service) post(toApp *store.App, fromMattermostUserID string, url string, msg interface{}) (*http.Response, error) {
	client, err := s.getAppHTTPClient(toApp.Manifest.AppID)
	if err != nil {
		return nil, err
	}
	jwtoken, err := createJWT(fromMattermostUserID, toApp.Secret)
	if err != nil {
		return nil, err
	}

	bb, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Printf("<><> POSTED:\n\n%s\n\n", string(bb))

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

func (s *service) getAppHTTPClient(appID store.AppID) (*http.Client, error) {
	// TODO cache the client, manage the connections
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

func (s *service) GetManifest(manifestURL string) (*store.Manifest, error) {
	var manifest store.Manifest
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
