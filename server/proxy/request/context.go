package request

import (
	"log"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Context struct {
	mm             *pluginapi.Client
	config         config.Service
	sessionService session.Service

	Log utils.Logger

	RequestID             string
	PluginID              string
	AppID                 apps.AppID
	ActingUserID          string
	actingUserAccessToken string
	sysAdminChecked       bool
}

// Clone creates a shallow copy of context, allowing clones to apply per-request changes.
// TODO
func (c *Context) Clone() *Context {
	return &Context{
		Log:            c.Log,
		mm:             c.mm,
		sessionService: c.sessionService,
	}
}

type contextOption func(*Context)

// TODO: Are other fields needed?
func NewContextFromAppContext(cc apps.Context, mm *pluginapi.Client, config config.Service, session session.Service) *Context {
	c := NewContext(mm, config, session)

	c.AppID = cc.AppID
	c.ActingUserID = cc.ActingUserID
	c.actingUserAccessToken = cc.ActingUserAccessToken

	return c
}

func NewContext(mm *pluginapi.Client, config config.Service, session session.Service, opts ...contextOption) *Context {
	c := &Context{
		mm:             mm,
		config:         config,
		sessionService: session,
		Log:            utils.NewPluginLogger(mm),
	}

	/*
		// TODO
		for _, opt := range opts {
			opt(c)
		}
	*/

	return c
}

func (c Context) UpdateAppContext(cc apps.Context) apps.Context {
	updated := cc
	updated.ActingUserID = c.ActingUserID
	updated.ExpandedContext = apps.ExpandedContext{
		ActingUserAccessToken: c.actingUserAccessToken,
	}
	return updated
}

func (c Context) UserAccessToken() (string, error) {
	if c.actingUserAccessToken != "" {
		return c.actingUserAccessToken, nil
	}

	if c.AppID == "" {
		return "", errors.New("missing appID in context")
	}

	session, err := c.sessionService.GetOrCreate(c.AppID, c.ActingUserID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get session")
	}
	log.Printf("session: %#+v\n", session)
	log.Printf("err: %#+v\n", err)

	c.actingUserAccessToken = session.Token

	return c.actingUserAccessToken, nil
}

func (c Context) GetMMClient() (mmclient.Client, error) {
	conf := c.config.Get()

	token, err := c.UserAccessToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to use the current user's token for admin access to Mattermost")
	}

	return mmclient.NewHTTPClient(conf, token), nil
}
