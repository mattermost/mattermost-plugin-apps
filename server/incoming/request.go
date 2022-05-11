package incoming

import (
	"context"
	"time"

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
	ctx            context.Context
	mm             *pluginapi.Client
	config         config.Service
	Log            utils.Logger
	sessionService SessionService

	requestID string
	pluginID  string

	toApp                 *apps.App
	fromApp               *apps.App
	actingUserID          string
	actingUserAccessToken string
}

func NewRequest(config config.Service, log utils.Logger, session SessionService) *Request {
	// TODO <>/<>: is the incoming Mattermost request ID available, and should it be used?
	requestID := model.NewId()
	return &Request{
		ctx:            context.Background(),
		mm:             config.MattermostAPI(),
		config:         config,
		Log:            log.With("request_id", requestID),
		sessionService: session,
		requestID:      requestID,
	}
}

// Clone creates a shallow copy of request, allowing clones to apply per-request changes.
func (r *Request) Clone() *Request {
	clone := *r
	return &clone
}

func (r *Request) WithCtx(ctx context.Context) *Request {
	r = r.Clone()
	r.ctx = ctx
	return r
}

func (r *Request) WithTimeout(timeout time.Duration, cancelFunc *context.CancelFunc) *Request {
	if timeout == 0 {
		return r
	}
	r = r.Clone()
	ctx, f := context.WithTimeout(r.Ctx(), timeout)
	r.ctx = ctx
	if cancelFunc != nil {
		*cancelFunc = f
	}
	return r
}

func (r *Request) ToApp(app *apps.App) *Request {
	if app == nil {
		return r
	}
	r.Log = r.Log.With("app_id", app.AppID)
	r.toApp = app
	return r
}

func (r *Request) FromApp(app *apps.App) *Request {
	if app == nil {
		return r
	}
	r.Log = r.Log.With("from_app_id", app.AppID)
	r.fromApp = app
	return r
}

func (r *Request) FromPlugin(pluginID string) *Request {
	r.Log = r.Log.With("from_plugin_id", pluginID)
	r.pluginID = pluginID
	return r
}

func (r *Request) WithActingUser(id, token string) *Request {
	r = r.Clone()
	r.actingUserID = id
	r.actingUserAccessToken = token
	r.Log = r.Log.With("acting_user_id", id)
	return r
}

func (r *Request) WithActingUserFromContext(cc apps.Context) *Request {
	id := ""
	if cc.ActingUser != nil {
		id = cc.ActingUser.Id
	}
	return r.WithActingUser(id, cc.ActingUserAccessToken)
}

func (r *Request) Ctx() context.Context {
	return r.ctx
}

func (r *Request) To() *apps.App {
	return r.toApp
}

func (r *Request) SourceApp() *apps.App {
	return r.fromApp
}

func (r *Request) ActingUserID() string {
	return r.actingUserID
}

func (r *Request) UserAccessToken() (string, error) {
	if r.actingUserAccessToken != "" {
		return r.actingUserAccessToken, nil
	}
	if r.toApp == nil {
		return "", errors.New("missing destination app ID in request")
	}

	session, err := r.sessionService.GetOrCreate(r, r.toApp.AppID, r.actingUserID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get session")
	}
	r.actingUserAccessToken = session.Token

	return r.actingUserAccessToken, nil
}

func (r *Request) GetMMClient() (mmclient.Client, error) {
	token, err := r.UserAccessToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to use the current user's token to access Mattermost")
	}
	return mmclient.NewHTTPClient(r.config.Get(), token), nil
}

func (r *Request) Config() config.Service {
	return r.config
}
