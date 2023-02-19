package incoming

import (
	"context"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/sessionutils"
)

type SessionService interface {
	GetOrCreate(r *Request, userID string) (*model.Session, error)
}

var ErrNotABot = errors.New("not a bot")
var ErrIsABot = errors.New("is a bot")

type Request struct {
	RequestID string
	Origin    string

	Ctx     context.Context
	API     config.API // redundant with Config.API(), but convenient
	Config  config.Service
	Session SessionService
	Log     utils.Logger

	actingUser            *model.User
	actingUserAccessToken string
	actingUserID          string
	destination           apps.AppID
	sessionID             string
	sourceAppID           apps.AppID
	sourcePluginID        string
}

func NewRequest(config config.Service, session SessionService, origin, id string) *Request {
	return &Request{
		Ctx:       context.Background(),
		API:       config.API(),
		Config:    config,
		Log:       config.NewBaseLogger().With("request_id", id, "origin", origin),
		Session:   session,
		RequestID: id,
	}
}

// Clone creates a shallow copy of request, allowing clones to apply per-request changes.
func (r *Request) Clone() *Request {
	clone := *r
	return &clone
}

func (r *Request) WithCtx(ctx context.Context) *Request {
	r = r.Clone()
	r.Ctx = ctx
	return r
}

func (r *Request) WithDestination(appID apps.AppID) *Request {
	r = r.Clone()
	r.Log = r.Log.With("destination", appID)
	r.destination = appID
	return r
}

func (r *Request) WithSourceAppID(appID apps.AppID) *Request {
	r = r.Clone()
	r.Log = r.Log.With("source_app_id", appID)
	r.sourceAppID = appID
	return r
}

func (r *Request) WithSessionID(sessionID string) *Request {
	r = r.Clone()
	r.sessionID = sessionID
	return r
}

func (r *Request) WithSourcePluginID(pluginID string) *Request {
	r = r.Clone()
	if pluginID != "" {
		r.Log = r.Log.With("source_plugin_id", pluginID)
	}
	r.sourcePluginID = pluginID
	return r
}

func (r *Request) WithActingUserID(id string) *Request {
	r = r.Clone()
	r.actingUserID = id
	r.actingUser = nil
	r.actingUserAccessToken = ""
	r.Log = r.Log.With("from_user_id", id)
	return r
}

func (r *Request) WithPrevContext(cc apps.Context) *Request {
	id := ""
	if cc.ActingUser != nil {
		id = cc.ActingUser.Id
	}
	return r.WithActingUserID(id)
}

func (r *Request) SourceAppID() apps.AppID {
	return r.sourceAppID
}

func (r *Request) Destination() apps.AppID {
	return r.destination
}

func (r *Request) SourcePluginID() string {
	return r.sourcePluginID
}

func (r *Request) ActingUserID() string {
	return r.actingUserID
}

func (r *Request) GetActingUser() (*model.User, error) {
	if r.actingUser != nil {
		return r.actingUser, nil
	}
	if r.actingUserID == "" {
		return nil, utils.ErrInvalid
	}
	return r.Config.API().Mattermost.User.Get(r.actingUserID)
}

func (r *Request) RequireActingUser() error {
	if r.ActingUserID() == "" {
		return utils.NewUnauthorizedError("user ID is required")
	}
	return nil
}

func (r *Request) RequireActingUserIsNotBot() error {
	if err := r.RequireActingUser(); err != nil {
		return err
	}
	mmuser, err := r.GetActingUser()
	if err != nil {
		return err
	}
	if mmuser.IsBot {
		return utils.NewUnauthorizedError(errors.Wrapf(ErrIsABot, "@%s (%s)", mmuser.Username, mmuser.GetDisplayName(model.ShowNicknameFullName)))
	}
	return nil
}

func (r *Request) RequireActingUserIsBot() error {
	if err := r.RequireActingUser(); err != nil {
		return err
	}
	mmuser, err := r.GetActingUser()
	if err != nil {
		return err
	}
	if !mmuser.IsBot {
		return utils.NewUnauthorizedError(errors.Wrapf(ErrNotABot, "@%s (%s)", mmuser.Username, mmuser.GetDisplayName(model.ShowNicknameFullName)))
	}
	return nil
}

func (r *Request) RequireUserPermission(p *model.Permission) func() error {
	return func() error {
		if !r.Config.API().Mattermost.User.HasPermissionTo(r.ActingUserID(), p) {
			return utils.NewUnauthorizedError("access to this operation is limited to users with permission: %s", p.Id)
		}
		return nil
	}
}

func (r *Request) RequireSysadminOrPlugin() error {
	if err := r.RequireUserPermission(model.PermissionManageSystem)(); err == nil {
		return nil
	}
	if r.SourcePluginID() != "" {
		return nil
	}
	return utils.NewUnauthorizedError("access to this operation is limited to system administrators, or plugins")
}

func (r *Request) RequireSourceApp() error {
	if r.sourceAppID != "" {
		return nil
	}
	if r.sessionID == "" {
		return utils.NewUnauthorizedError("access to this operation is limited to Mattermost Apps")
	}
	s, err := r.API.Mattermost.Session.Get(r.sessionID)
	if err != nil {
		return utils.NewUnauthorizedError(err)
	}
	appID := sessionutils.GetAppID(s)
	if appID == "" {
		return utils.NewUnauthorizedError("not an app session")
	}
	r.sourceAppID = appID
	return nil
}

func (r *Request) Check(ff ...func() error) error {
	for _, f := range ff {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}
