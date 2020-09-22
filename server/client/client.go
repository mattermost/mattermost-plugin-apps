// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package client

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattermost/mattermost-plugin-apps/server/appmodel"
)

const (
	jwtTTL = time.Minute
)

type Client interface {
	SendNotification(ss appmodel.SubscriptionSubject, msg interface{})
	InstallComplete()
}

type client struct {
	token string
	app   *appmodel.App
}

func New(userID string, app *appmodel.App) (Client, error) {
	token, err := createJWT(userID, app.Secret)
	if err != nil {
		return nil, err
	}

	return &client{
		token: token,
		app:   app,
	}, nil
}

func createJWT(userID, secret string) (string, error) {
	var err error
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := at.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (c *client) DoPost(url string, body interface{}) {
	piper, pipew := io.Pipe()
	go func() {
		defer pipew.Close()
		encodeErr := json.NewEncoder(pipew).Encode(body)
		if encodeErr != nil {
			// <><> TODO log
			return
		}
	}()

	req, err := http.NewRequest(http.MethodPost, url, piper)
	if err != nil {
		// <><> TODO log
		return
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// <><> TODO log
		// TODO ticket: progressive backoff on errors
		return
	}
	resp.Body.Close()
}
