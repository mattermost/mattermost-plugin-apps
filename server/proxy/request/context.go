package request

import (
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
	appID                 apps.AppID
	actingUserID          string
	actingUserAccessToken string
	sysAdminChecked       bool
}

type ContextOption func(*Context)

func WithAppContext(cc apps.Context) ContextOption {
	return func(c *Context) {
		c.SetActingUserID(cc.ActingUserID)
		c.actingUserAccessToken = cc.ActingUserAccessToken
	}
}

func NewContext(mm *pluginapi.Client, config config.Service, session session.Service, opts ...ContextOption) *Context {
	c := &Context{
		mm:             mm,
		config:         config,
		sessionService: session,
		Log:            utils.NewPluginLogger(mm),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Clone creates a shallow copy of context, allowing clones to apply per-request changes.
func (c *Context) Clone() *Context {
	return &Context{
		Log:            c.Log,
		mm:             c.mm,
		sessionService: c.sessionService,
	}
}

func (c *Context) UpdateAppContext(cc apps.Context) apps.Context {
	updated := cc
	updated.ActingUserID = c.ActingUserID()
	updated.ExpandedContext = apps.ExpandedContext{
		ActingUserAccessToken: c.actingUserAccessToken,
	}
	return updated
}

func (c *Context) AppID() apps.AppID {
	return c.appID
}

func (c *Context) SetAppID(appID apps.AppID) {
	c.Log = c.Log.With("app_id", appID)

	c.appID = appID
}

func (c *Context) ActingUserID() string {
	return c.actingUserID
}

func (c *Context) SetActingUserID(userID string) {
	c.Log = c.Log.With("user_id", userID)

	c.actingUserID = userID
}

func (c *Context) UserAccessToken() (string, error) {
	if c.actingUserAccessToken != "" {
		return c.actingUserAccessToken, nil
	}

	appID := c.AppID()
	if c.AppID() == "" {
		return "", errors.New("missing appID in context")
	}

	session, err := c.sessionService.GetOrCreate(appID, c.ActingUserID())
	if err != nil {
		return "", errors.Wrap(err, "failed to get session")
	}

	c.actingUserAccessToken = session.Token

	return c.actingUserAccessToken, nil
}

func (c *Context) GetMMClient() (mmclient.Client, error) {
	conf := c.config.Get()

	token, err := c.UserAccessToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to use the current user's token for admin access to Mattermost")
	}

	return mmclient.NewHTTPClient(conf, token), nil
}
