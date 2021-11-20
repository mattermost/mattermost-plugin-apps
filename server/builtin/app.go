package builtin

import (
	"bytes"
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
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
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
	pDebugBindings      = "/debug-bindings"
	pDebugClean         = "/debug-clean"
	pDebugKVInfo        = "/debug/kv/info"
	pDebugKVCreate      = "/debug/kv/create"
	pDebugKVEdit        = "/debug/kv/edit"
	pDebugKVEditModal   = "/debug/kv/edit-modal"
	pDebugKVClean       = "/debug/kv/clean"
	pDebugKVList        = "/debug/kv/list"
	pInfo               = "/info"
	pList               = "/list"
	pUninstall          = "/uninstall"
	pEnable             = "/enable"
	pDisable            = "/disable"
	pInstallHTTP        = "/install-http"
	pInstallListed      = "/install-listed"
	pInstallConsent     = "/install-consent"
	pInstallConsentForm = "/install-consent/form"
)

const (
	pLookupAppID     = "/q/app_id"
	pLookupNamespace = "/q/namespace"
)

type handler func(apps.CallRequest) apps.CallResponse

type builtinApp struct {
	conf        config.Service
	proxy       proxy.Service
	appservices appservices.Service
	httpOut     httpout.Service
	router      map[string]handler
}

var _ upstream.Upstream = (*builtinApp)(nil)

func NewBuiltinApp(conf config.Service, proxy proxy.Service, appservices appservices.Service, httpOut httpout.Service) *builtinApp {
	a := &builtinApp{
		conf:        conf,
		proxy:       proxy,
		appservices: appservices,
		httpOut:     httpOut,
	}

	a.router = map[string]handler{
		// App's own bindings.
		appspath.Bindings: a.bindings,

		// Actions available to all users.
		pInfo: a.info,

		// Actions that require sysadmin.
		pDebugBindings:      requireAdmin(a.debugBindings),
		pDebugClean:         requireAdmin(a.debugClean),
		pDebugKVClean:       requireAdmin(a.debugKVClean),
		pDebugKVCreate:      requireAdmin(a.debugKVCreate),
		pDebugKVEdit:        requireAdmin(a.debugKVEdit),
		pDebugKVEditModal:   requireAdmin(a.debugKVEditModal),
		pDebugKVInfo:        requireAdmin(a.debugKVInfo),
		pDebugKVList:        requireAdmin(a.debugKVList),
		pDisable:            requireAdmin(a.disable),
		pEnable:             requireAdmin(a.enable),
		pInstallConsent:     requireAdmin(a.installConsent),
		pInstallConsentForm: requireAdmin(a.installConsentForm),
		pInstallListed:      requireAdmin(a.installListed),
		pInstallHTTP:        requireAdmin(a.installHTTP),
		pList:               requireAdmin(a.list),
		pUninstall:          requireAdmin(a.uninstall),

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
		GrantedPermissions: apps.Permissions{
			apps.PermissionActAsUser,
		},
	}
}

func (a *builtinApp) Roundtrip(_ apps.App, creq apps.CallRequest, async bool) (out io.ReadCloser, err error) {
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
	}(a.conf.Logger())

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
	return readcloser(h(creq))
}

func (a *builtinApp) GetStatic(_ apps.App, path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.NewNotFoundError("static support is not implemented")
}

func requireAdmin(h handler) handler {
	return func(creq apps.CallRequest) apps.CallResponse {
		if creq.Context.ActingUser == nil || creq.Context.ActingUser.Id != creq.Context.ActingUserID {
			return apps.NewErrorResponse(utils.NewInvalidError(
				"no or invalid ActingUser in the context, please make sure Expand.ActingUser is set"))
		}
		if !creq.Context.ActingUser.IsSystemAdmin() {
			return apps.NewErrorResponse(utils.NewUnauthorizedError(
				"user %s (%s) is not a sysadmin", creq.Context.ActingUser.GetDisplayName(model.ShowUsername), creq.Context.ActingUserID))
		}
		return h(creq)
	}
}

func (a *builtinApp) newLocalizer(creq apps.CallRequest) *i18n.Localizer {
	return a.conf.I18N().GetUserLocalizer(creq.Context.ActingUserID)
}
