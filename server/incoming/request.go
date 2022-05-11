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
	requestID      string

	dest apps.AppID

	sourcePluginID        string
	sourceAppID           apps.AppID
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

func (r *Request) Ctx() context.Context {
	return r.ctx
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

func (r *Request) Destination() apps.AppID {
	return r.dest
}

func (r *Request) WithDestination(appID apps.AppID) *Request {
	r.Log = r.Log.With("destination", appID)
	r.dest = appID
	return r
}

func (r *Request) SourceAppID() apps.AppID {
	return r.sourceAppID
}

func (r *Request) WithSourceAppID(appID apps.AppID) *Request {
	r.Log = r.Log.With("source_app_id", appID)
	r.sourceAppID = appID
	return r
}

func (r *Request) SourcePluginID() string {
	return r.sourcePluginID
}

func (r *Request) WithSourcePluginID(pluginID string) *Request {
	r.Log = r.Log.With("source_plugin_id", pluginID)
	r.sourcePluginID = pluginID
	return r
}

func (r *Request) ActingUserID() string {
	return r.actingUserID
}

func (r *Request) WithActingUserID(id string) *Request {
	r = r.Clone()
	r.actingUserID = id
	r.actingUserAccessToken = ""
	r.Log = r.Log.With("from_user_id", id)
	return r
}

func (r *Request) ActingUserAccessTokenForDestination() (string, error) {
	if r.dest == "" {
		return "", errors.New("missing destination app ID in request")
	}
	if r.actingUserAccessToken != "" {
		return r.actingUserAccessToken, nil
	}

	session, err := r.sessionService.GetOrCreate(r, r.dest, r.actingUserID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get session")
	}
	r.actingUserAccessToken = session.Token
	return r.actingUserAccessToken, nil
}

func (r *Request) WithPrevContext(cc apps.Context) *Request {
	id := ""
	if cc.ActingUser != nil {
		id = cc.ActingUser.Id
	}
	r = r.WithActingUserID(id)
	r.actingUserAccessToken = cc.ActingUserAccessToken
	return r
}

func (r *Request) GetMMClient() (mmclient.Client, error) {
	token, err := r.ActingUserAccessTokenForDestination()
	if err != nil {
		return nil, errors.Wrap(err, "failed to use the current user's token to access Mattermost")
	}
	return mmclient.NewHTTPClient(r.config.Get(), token), nil
}

func (r *Request) Config() config.Service {
	return r.config
}
