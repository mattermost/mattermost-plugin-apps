// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package cloudapps

import (
	"encoding/json"
	"io"
	"net/http"
	"path"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/configurator"
)

type Proxy interface {
	OnUserJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
}

type proxy struct {
	configurator.Configurator
	mm       *pluginapi.Client
	registry Registry
	Subscriptions
}

var _ Proxy = (*proxy)(nil)

func NewProxy(mm *pluginapi.Client, configurator configurator.Configurator, subs Subscriptions) Proxy {
	return &proxy{
		Configurator:  configurator,
		Subscriptions: subs,
		mm:            mm,
	}
}

func (p *proxy) SendChangeNotification(s *Subscription, msg interface{}) {
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

	piper, pipew := io.Pipe()
	go func() {
		defer pipew.Close()
		encodeErr := json.NewEncoder(pipew).Encode(msg)
		if encodeErr != nil {
			// <><> TODO log
			return
		}
	}()

	u := path.Join(app.RootURL, "notify", string(SubjectUserJoinedChannel))
	// <><> TODO ticket: progressive backoff on errors
	resp, err := client.Post(u, "application/json", piper)
	if err != nil {
		// <><> TODO log
		return
	}
	defer resp.Body.Close()
}

func (p *proxy) GetNotificationClient(appID AppID) (*http.Client, error) {
	// <><> TODO cache the client per app
	return nil, nil
}
