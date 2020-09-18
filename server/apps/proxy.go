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
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type Proxy interface {
	OnUserHasBeenCreated(pluginContext *plugin.Context, user *model.User)
	OnUserJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
	OnUserLeftChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
	OnUserJoinedTeam(pluginContext *plugin.Context, teamMember *model.TeamMember, actor *model.User)
	OnUserLeftTeam(pluginContext *plugin.Context, teamMember *model.TeamMember, actor *model.User)
	OnChannelHasBeenCreated(pluginContext *plugin.Context, channel *model.Channel)
}

type proxy struct {
	configurator configurator.Service
	mm           *pluginapi.Client
	registry     Registry
	Subscriptions
}

var _ Proxy = (*proxy)(nil)

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

func (p *proxy) SendChangeNotification(s *Subscription, msg interface{}) {
	conf := p.configurator.GetConfig()

	client, err := p.GetNotificationClient(s.AppID)
	if err != nil {
		// <><> TODO log
		return
	}

	app, err := p.registry.GetApp(s.AppID)
	if err != nil {
		// <><> TODO log
		return
	}

	token, err := p.CreateJWT(conf.BotUserID, app.Secret)
	if err != nil {
		// <><> TODO log
		return
	}

	piper, pipew := io.Pipe()
	go func() {
		defer pipew.Close()
		encodeErr := json.NewEncoder(pipew).Encode(msg)
		if encodeErr != nil {
			// <><> TODO log
			return
		}
	}()

	req, err := http.NewRequest(http.MethodPost, path.Join(app.Manifest.RootURL, "notify", string(s.Subject)), piper)
	if err != nil {
		// <><> TODO log
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		// <><> TODO log
		// TODO ticket: progressive backoff on errors
		return
	}
	resp.Body.Close()
}

func (p *proxy) GetNotificationClient(appID AppID) (*http.Client, error) {
	// <><> TODO cache the client per app
	return nil, nil
}
