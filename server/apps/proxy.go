// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/appmodel"
	"github.com/mattermost/mattermost-plugin-apps/server/client"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type Proxy interface {
	OnUserJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
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

func (p *proxy) SendChangeNotification(s *appmodel.Subscription, msg interface{}) {
	conf := p.configurator.GetConfig()

	app, err := p.registry.GetApp(s.AppID)
	if err != nil {
		// <><> TODO log
		return
	}

	c, err := client.New(conf.BotUserID, app)
	if err != nil {
		// <><> TODO log
		return
	}

	c.SendNotification(appmodel.SubjectUserJoinedChannel, msg)
}
