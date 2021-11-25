package incoming

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
	GetOrCreate(r *Request, appID apps.AppID, userID string) (*model.Session, error)
}

type Request struct {
	ctx context.Context

	mm             *pluginapi.Client
	config         config.Service
	Log            utils.Logger
	sessionService SessionService

	requestID string
	pluginID  string

	appID                 apps.AppID
	actingUserID          string
	actingUserAccessToken string
}

type RequestOption func(*Request)

func WithCtx(ctx context.Context) RequestOption {
	return func(r *Request) {
		if ctx == nil {
			panic("nil context")
		}

		r.ctx = ctx
	}
}

func WithAppContext(cc apps.Context) RequestOption {
	return func(r *Request) {
		r.SetActingUserID(cc.ActingUserID)
		r.actingUserAccessToken = cc.ActingUserAccessToken
	}
}

func WithAppID(appID apps.AppID) RequestOption {
	return func(r *Request) {
		r.SetAppID(appID)
	}
}

func NewRequest(mm *pluginapi.Client, config config.Service, session SessionService, opts ...RequestOption) *Request {
	r := &Request{
		ctx:            context.Background(),
		mm:             mm,
		config:         config,
		Log:            config.Logger(),
		sessionService: session,
		requestID:      model.NewId(),
	}

	r.Log = r.Log.With(
		"request_id", r.requestID,
	)

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Clone creates a shallow copy of request, allowing clones to apply per-request changes.
func (r *Request) Clone() *Request {
	return &Request{
		ctx:                   r.ctx,
		mm:                    r.mm,
		config:                r.config,
		Log:                   r.Log,
		sessionService:        r.sessionService,
		requestID:             r.requestID,
		pluginID:              r.pluginID,
		appID:                 r.appID,
		actingUserID:          r.actingUserID,
		actingUserAccessToken: r.actingUserAccessToken,
	}
}

func (r *Request) UpdateAppContext(cc apps.Context) apps.Context {
	updated := cc
	updated.ActingUserID = r.ActingUserID()
	updated.ExpandedContext = apps.ExpandedContext{
		ActingUserAccessToken: r.actingUserAccessToken,
	}
	return updated
}

func (r *Request) Ctx() context.Context {
	return r.ctx
}

func (r *Request) MattermostAPI() *pluginapi.Client {
	return r.mm
}

func (r *Request) Config() config.Service {
	return r.config
}

func (r *Request) PluginID() string {
	return r.pluginID
}

func (r *Request) AppID() apps.AppID {
	return r.appID
}

func (r *Request) SetAppID(appID apps.AppID) {
	r.Log = r.Log.With("app_id", appID)

	r.appID = appID
}

func (r *Request) ActingUserID() string {
	return r.actingUserID
}

func (r *Request) SetActingUserID(userID string) {
	r.Log = r.Log.With("user_id", userID)

	r.actingUserID = userID
}

func (r *Request) UserAccessToken() (string, error) {
	if r.actingUserAccessToken != "" {
		return r.actingUserAccessToken, nil
	}

	appID := r.AppID()
	if r.AppID() == "" {
		return "", errors.New("missing appID in context")
	}

	session, err := r.sessionService.GetOrCreate(r, appID, r.ActingUserID())
	if err != nil {
		return "", errors.Wrap(err, "failed to get session")
	}

	r.actingUserAccessToken = session.Token

	return r.actingUserAccessToken, nil
}

func (r *Request) GetMMClient() (mmclient.Client, error) {
	conf := r.config.Get()

	token, err := r.UserAccessToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to use the current user's token for access to Mattermost")
	}

	return mmclient.NewHTTPClient(conf, token), nil
}
