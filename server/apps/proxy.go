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
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type proxy struct {
	configurator configurator.Service
	mm           *pluginapi.Client
	registry     Registry
	Subscriptions
}

func NewProxy(mm *pluginapi.Client, configurator configurator.Service, subs Subscriptions) Proxy {
	return &proxy{
		configurator:  configurator,
		Subscriptions: subs,
		mm:            mm,
	}
}

func (p *proxy) CreateJWT(userID, secret string) (string, error) {
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

func (p *proxy) SendChangeNotification(sub *Subscription, msg interface{}) {
	conf := p.configurator.GetConfig()

	app, err := p.registry.GetApp(sub.AppID)
	if err != nil {
		// <><> TODO log
		return
	}

	resp, err := p.Post(app, path.Join(app.Manifest.RootURL, "notify", string(sub.Subject)), msg)
	if err != nil {
		// <><> TODO log
		return
	}
	resp.Body.Close()
}

func (p *proxy) Post(app *App, url string, msg interface{}) (*http.Response, error) {
	conf := p.configurator.GetConfig()
	client, err := p.GetNotificationClient(app.Manifest.AppID)
	if err != nil {
		return nil, err
	}

	token, err := p.CreateJWT(conf.BotUserID, app.Secret)
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
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		// TODO ticket: progressive backoff on errors
		return nil, err
	}
	if encodeErr != nil {
		// TODO ticket: progressive backoff on errors
		return nil, encodeErr
	}

	return resp, nil
}

func (p *proxy) GetNotificationClient(appID AppID) (*http.Client, error) {
	// <><> TODO cache the client per app
	return &http.Client{}, nil
}

func (p *proxy) GetWishClient(appID AppID) (*http.Client, error) {
	// <><> TODO cache the client per app
	return &http.Client{}, nil
}
