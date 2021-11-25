package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"runtime/debug"
	"strings"

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
)

const (
	pDebugBindings     = "/debug/bindings"
	pDebugClean        = "/debug/clean"
	pDebugKVInfo       = "/debug/kv/info"
	pDebugKVCreate     = "/debug/kv/create"
	pDebugKVEdit       = "/debug/kv/edit"
	pDebugKVEditModal  = "/debug/kv/edit-modal"
	pDebugKVClean      = "/debug/kv/clean"
	pDebugKVList       = "/debug/kv/list"
	pDebugSessionsList = "/debug/session/list"
	pDisable           = "/disable"
	pEnable            = "/enable"
	pInfo              = "/info"
	pInstallConsent    = "/install-consent"
	pInstallHTTP       = "/install/http"
	pInstallListed     = "/install/listed"
	pList              = "/list"
	pUninstall         = "/uninstall"
)

type handler struct {
	requireSysadmin bool
	commandBinding  func(*i18n.Localizer) apps.Binding
	lookupf         func(*incoming.Request, apps.CallRequest) ([]apps.SelectOption, error)
	submitf         func(*incoming.Request, apps.CallRequest) apps.CallResponse
	formf           func(*incoming.Request, apps.CallRequest) (*apps.Form, error)
}

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
		// Actions available to all users
		pInfo: a.info(),

		// Actions that require sysadmin
		pDebugBindings:     a.debugBindings(),
		pDebugClean:        a.debugClean(),
		pDebugKVClean:      a.debugKVClean(),
		pDebugKVCreate:     a.debugKVCreate(),
		pDebugKVEdit:       a.debugKVEdit(),
		pDebugKVEditModal:  a.debugKVEditModal(),
		pDebugKVInfo:       a.debugKVInfo(),
		pDebugKVList:       a.debugKVList(),
		pDebugSessionsList: a.debugSessionsList(),
		pDisable:           a.disable(),
		pEnable:            a.enable(),
		pInstallConsent:    a.installConsent(),
		pInstallHTTP:       a.installHTTP(),
		pInstallListed:     a.installListed(),
		pList:              a.list(),
		pUninstall:         a.uninstall(),
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
		GrantedPermissions: apps.Permissions{
			apps.PermissionActAsUser,
		},
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

	// The bindings call does not have a call type, so make it into a submit so
	// that the router can handle it.
	if creq.Path == appspath.Bindings {
		return readcloser(a.bindings(creq))
	}

	callPath, callType := path.Split(creq.Path)
	callPath = strings.TrimRight(callPath, "/")
	h, ok := a.router[callPath]
	if !ok {
		return nil, utils.NewNotFoundError(callPath)
	}

	if h.requireSysadmin {
		if creq.Context.ActingUser == nil || creq.Context.ActingUser.Id != creq.Context.ActingUserID {
			return nil, apps.NewErrorResponse(utils.NewInvalidError(
				"no or invalid ActingUser in the context, please make sure Expand.ActingUser is set"))
		}
		if !creq.Context.ActingUser.IsSystemAdmin() {
			return nil, apps.NewErrorResponse(utils.NewUnauthorizedError(
				"user %s (%s) is not a sysadmin", creq.Context.ActingUser.GetDisplayName(model.ShowUsername), creq.Context.ActingUserID))
		}
	}

	switch apps.CallType(callType) {
	case apps.CallTypeForm:
		if h.formf == nil {
			return nil, utils.ErrNotFound
		}
		form, err := h.formf(r, creq)
		if err != nil {
			return nil, err
		}
		return readcloser(apps.NewFormResponse(*form))

	case apps.CallTypeLookup:
		if h.lookupf == nil {
			return nil, utils.ErrNotFound
		}
		opts, err := h.lookupf(r, creq)
		if err != nil {
			return nil, err
		}
		return readcloser(apps.NewLookupResponse(opts))

	case apps.CallTypeSubmit:
		if h.submitf == nil {
			return nil, utils.ErrNotFound
		}
		return readcloser(h.submitf(r, creq))
	}

	return nil, utils.NewNotFoundError("%s does not handle %s", callPath, callType)
}

func (a *builtinApp) GetStatic(_ context.Context, _ apps.App, path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.NewNotFoundError("static support is not implemented")
}

func (a *builtinApp) newLocalizer(creq apps.CallRequest) *i18n.Localizer {
	return a.conf.I18N().GetUserLocalizer(creq.Context.ActingUserID)
}
