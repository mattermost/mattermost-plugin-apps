// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"io"
	"net/http"
	"path"

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

	u := path.Join(app.RootURL, "notify", string(s.Subject))
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
