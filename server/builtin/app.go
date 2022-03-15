package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	AppID          = "apps"
	AppDisplayName = "Mattermost Apps plugin"
	AppDescription = "Install and manage Mattermost Apps"
)

const (
	fAction         = "action"
	fAppID          = "app"
	fBase64         = "base64"
	fBase64Key      = "base64_key"
	fConsent        = "consent"
	fCurrentValue   = "current_value"
	fDeployType     = "deploy_type"
	fID             = "id"
	fIncludePlugins = "include_plugins"
	fNamespace      = "namespace"
	fNewValue       = "new_value"
	fSecret         = "secret"
	fURL            = "url"
	fSessionID      = "session_id"
)

const (
	pDebugBindings        = "/debug/bindings"
	pDebugClean           = "/debug/clean"
	pDebugKVClean         = "/debug/kv/clean"
	pDebugKVCreate        = "/debug/kv/create"
	pDebugKVEdit          = "/debug/kv/edit"
	pDebugKVEditModal     = "/debug/kv/edit-modal"
	pDebugKVInfo          = "/debug/kv/info"
	pDebugKVList          = "/debug/kv/list"
	pDebugSessionsList    = "/debug/session/list"
	pDebugSessionsView    = "/debug/session/view"
	pDebugSessionsRevoke  = "/debug/session/delete"
	pDebugOAuthConfigView = "/debug/oauth/config/view"
	pEnable               = "/enable"
	pDisable              = "/disable"
	pInfo                 = "/info"
	pInstallConsent       = "/install-consent"
	pInstallConsentSource = "/install-consent/form"
	pInstallHTTP          = "/install-http"
	pInstallListed        = "/install-listed"
	pList                 = "/list"
	pUninstall            = "/uninstall"
)

const (
	pLookupAppID     = "/q/app_id"
	pLookupNamespace = "/q/namespace"
)

/*
type handler struct {
	requireSysadmin bool
	commandBinding  func(*i18n.Localizer) apps.Binding
	lookupf         func(*incoming.Request, apps.CallRequest) ([]apps.SelectOption, error)
	submitf         func(*incoming.Request, apps.CallRequest) apps.CallResponse
	formf           func(*incoming.Request, apps.CallRequest) (*apps.Form, error)
}
*/
type handler func(*incoming.Request, apps.CallRequest) apps.CallResponse

type builtinApp struct {
	conf           config.Service
	proxy          proxy.Service
	appservices    appservices.Service
	httpOut        httpout.Service
	sessionService session.Service
	router         map[string]handler
}

var _ upstream.Upstream = (*builtinApp)(nil)

func NewBuiltinApp(conf config.Service, proxy proxy.Service, appservices appservices.Service, httpOut httpout.Service, sessionService session.Service) *builtinApp {
	a := &builtinApp{
		conf:           conf,
		proxy:          proxy,
		appservices:    appservices,
		httpOut:        httpOut,
		sessionService: sessionService,
	}

	a.router = map[string]handler{
		// App's own bindings.
		appspath.Bindings: a.bindings,

		// Commands available to all users.
		pInfo: a.info,

		// Commands that require sysadmin.
		pDebugBindings:        requireAdmin(a.debugBindings),
		pDebugClean:           requireAdmin(a.debugClean),
		pDebugKVClean:         requireAdmin(a.debugKVClean),
		pDebugKVCreate:        requireAdmin(a.debugKVCreate),
		pDebugKVEdit:          requireAdmin(a.debugKVEdit),
		pDebugKVInfo:          requireAdmin(a.debugKVInfo),
		pDebugKVList:          requireAdmin(a.debugKVList),
		pDebugSessionsList:    requireAdmin(a.debugSessionsList),
		pDebugSessionsRevoke:  requireAdmin(a.debugSessionsRevoke),
		pDebugSessionsView:    requireAdmin(a.debugSessionsView),
		pDebugOAuthConfigView: requireAdmin(a.debugOAuthConfigView),
		pEnable:               requireAdmin(a.enable),
		pDisable:              requireAdmin(a.disable),
		pInstallListed:        requireAdmin(a.installListed),
		pInstallHTTP:          requireAdmin(a.installHTTP),
		pList:                 requireAdmin(a.list),
		pUninstall:            requireAdmin(a.uninstall),

		// Modals.
		pDebugKVEditModal:     requireAdmin(a.debugKVEditModal),
		pInstallConsent:       requireAdmin(a.installConsent),
		pInstallConsentSource: requireAdmin(a.installConsentForm),

		// Lookups.
		pLookupAppID:     requireAdmin(a.lookupAppID),
		pLookupNamespace: requireAdmin(a.lookupNamespace),
	}

	return a
}

func Manifest(conf config.Config) apps.Manifest {
	return apps.Manifest{
		AppID:       AppID,
		Version:     apps.AppVersion(conf.PluginManifest.Version),
		DisplayName: AppDisplayName,
		Description: AppDescription,
		Deploy:      apps.Deploy{},
		Bindings: &apps.Call{
			Path: appspath.Bindings,
			Expand: &apps.Expand{
				ActingUser: apps.ExpandSummary,
				Locale:     apps.ExpandAll,
			},
		},
	}
}

func App(conf config.Config) apps.App {
	return apps.App{
		Manifest:    Manifest(conf),
		DeployType:  apps.DeployBuiltin,
		BotUserID:   conf.BotUserID,
		BotUsername: config.BotUsername,
		GrantedLocations: apps.Locations{
			apps.LocationCommand,
		},
		GrantedPermissions: apps.Permissions{},
	}
}

func (a *builtinApp) Roundtrip(ctx context.Context, _ apps.App, creq apps.CallRequest, async bool) (out io.ReadCloser, err error) {
	r := incoming.NewRequest(a.conf.MattermostAPI(), a.conf, utils.NewPluginLogger(a.conf.MattermostAPI()), a.sessionService, incoming.WithCtx(ctx), incoming.WithAppContext(creq.Context))

	defer func(log utils.Logger) {
		if x := recover(); x != nil {
			stack := string(debug.Stack())
			txt := "Call `" + creq.Path + "` panic-ed."
			log = log.With(
				"path", creq.Path,
				"values", creq.Values,
				"error", x,
				"stack", stack,
			)
			if creq.RawCommand != "" {
				txt = "Command `" + creq.RawCommand + "` panic-ed."
				log.Errorw("Recovered from a panic in a command", "command", creq.RawCommand)
			} else {
				log.Errorw("Recovered from a panic in a Call")
			}

			if a.conf.Get().DeveloperMode {
				txt += "\n"
				txt += fmt.Sprintf("Error: **%v**.\n", x)
				txt += fmt.Sprintf("Stack:\n%v", utils.CodeBlock(stack))
			} else {
				txt += "Please check the server logs for more details."
			}
			out = nil
			data, errr := json.Marshal(apps.NewTextResponse(txt))
			if errr != nil {
				err = errr
				return
			}
			err = nil
			out = ioutil.NopCloser(bytes.NewReader(data))
		}
	}(r.Log)

	readcloser := func(cresp apps.CallResponse) (io.ReadCloser, error) {
		data, err := json.Marshal(cresp)
		if err != nil {
			return nil, err
		}
		return ioutil.NopCloser(bytes.NewReader(data)), nil
	}

	h, ok := a.router[creq.Path]
	if !ok {
		return nil, utils.NewNotFoundError(creq.Path)
	}
	return readcloser(h(r, creq))
}

func (a *builtinApp) GetStatic(_ context.Context, _ apps.App, path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.NewNotFoundError("static support is not implemented")
}

func requireAdmin(h handler) handler {
	return func(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
		if creq.Context.ActingUser == nil || creq.Context.ActingUser.Id != creq.Context.ActingUserID {
			return apps.NewErrorResponse(utils.NewInvalidError(
				"no or invalid ActingUser in the context, please make sure Expand.ActingUser is set"))
		}
		if !creq.Context.ActingUser.IsSystemAdmin() {
			return apps.NewErrorResponse(utils.NewUnauthorizedError(
				"user %s (%s) is not a sysadmin", creq.Context.ActingUser.GetDisplayName(model.ShowUsername), creq.Context.ActingUserID))
		}
		return h(r, creq)
	}
}

func (a *builtinApp) newLocalizer(creq apps.CallRequest) *i18n.Localizer {
	if creq.Context.Locale != "" {
		return i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
	}

	return a.conf.I18N().GetUserLocalizer(creq.Context.ActingUserID)
}
