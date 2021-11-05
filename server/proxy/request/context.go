package request

import (
	"context"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SessionService interface {
	GetOrCreate(appID apps.AppID, userID string) (*model.Session, error)
}

type Context struct {
	mm             *pluginapi.Client
	config         config.Service
	Log            utils.Logger
	sessionService SessionService

	requestID string
	pluginID  string

	appID                 apps.AppID
	actingUserID          string
	actingUserAccessToken string
	sysAdminChecked       bool

	Ctx context.Context
}

type ContextOption func(*Context)

func WithAppContext(cc apps.Context) ContextOption {
	return func(c *Context) {
		c.SetActingUserID(cc.ActingUserID)
		c.actingUserAccessToken = cc.ActingUserAccessToken
	}
}

func WithAppID(appID apps.AppID) ContextOption {
	return func(c *Context) {
		c.SetAppID(appID)
	}
}

func WithCtx(ctx context.Context) ContextOption {
	return func(c *Context) {
		c.Ctx = ctx
	}
}

func NewContext(mm *pluginapi.Client, config config.Service, session SessionService, opts ...ContextOption) *Context {
	c := &Context{
		mm:             mm,
		config:         config,
		Log:            config.Logger(),
		sessionService: session,
		Ctx:            context.Background(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Clone creates a shallow copy of context, allowing clones to apply per-request changes.
func (c *Context) Clone() *Context {
	return &Context{
		mm:                    c.mm,
		config:                c.config,
		Log:                   c.Log,
		sessionService:        c.sessionService,
		requestID:             c.requestID,
		pluginID:              c.pluginID,
		appID:                 c.appID,
		actingUserID:          c.actingUserID,
		actingUserAccessToken: c.actingUserAccessToken,
		sysAdminChecked:       c.sysAdminChecked,
		Ctx:                   c.Ctx,
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

func (c *Context) MattermostAPI() *pluginapi.Client {
	return c.mm
}

func (c *Context) Config() config.Service {
	return c.config
}

func (c *Context) PluginID() string {
	return c.pluginID
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
