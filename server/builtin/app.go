package builtin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"

	"github.com/mattermost/mattermost-plugin-apps/apps"
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
	fURL            = "url"
	fConsent        = "consent"
	fSecret         = "secret"
	fAppID          = "app"
	fVersion        = "version"
	fIncludePlugins = "include_plugins"
	fDeployType     = "deploy_type"
)

const (
	pDebugBindings       = "/debug-bindings"
	pDebugClean          = "/debug-clean"
	pInfo                = "/info"
	pList                = "/list"
	pUninstall           = "/uninstall"
	pUninstallLookup     = "/uninstall/lookup"
	pEnable              = "/enable"
	pEnableLookup        = "/enable/lookup"
	pDisable             = "/disable"
	pDisableLookup       = "/disable/lookup"
	pInstallHTTP         = "/install-http"
	pInstallListed       = "/install-listed"
	pInstallListedLookup = "/install-listed/lookup"
	pInstallConsent      = "/install-consent"
	pInstallConsentForm  = "/install-consent/form"
)

type handler func(apps.CallRequest) apps.CallResponse

type builtinApp struct {
	conf    config.Service
	proxy   proxy.Service
	httpOut httpout.Service
	router  map[string]handler
}

var _ upstream.Upstream = (*builtinApp)(nil)

func NewBuiltinApp(conf config.Service, proxy proxy.Service, httpOut httpout.Service) *builtinApp {
	a := &builtinApp{
		conf:    conf,
		proxy:   proxy,
		httpOut: httpOut,
	}

	a.router = map[string]handler{
		// Actions available to all users
		pInfo: a.info,

		// Actions that require sysadmin
		pDebugBindings:       requireAdmin(a.debugBindings),
		pDebugClean:          requireAdmin(a.debugClean),
		pDisable:             requireAdmin(a.disable),
		pDisableLookup:       requireAdmin(a.disableLookup),
		pEnable:              requireAdmin(a.enable),
		pEnableLookup:        requireAdmin(a.enableLookup),
		pInstallConsent:      requireAdmin(a.installConsent),
		pInstallConsentForm:  requireAdmin(a.installConsentForm),
		pInstallListed:       requireAdmin(a.installListed),
		pInstallListedLookup: requireAdmin(a.installListedLookup),
		pInstallHTTP:         requireAdmin(a.installHTTP),
		pList:                requireAdmin(a.list),
		pUninstall:           requireAdmin(a.uninstall),
		pUninstallLookup:     requireAdmin(a.uninstallLookup),
	}

	return a
}

func Manifest(conf config.Config) apps.Manifest {
	return apps.Manifest{
		AppID:       AppID,
		Version:     apps.AppVersion(conf.BuildConfig.BuildHashShort),
		DisplayName: AppDisplayName,
		Description: AppDescription,
		Bindings: &apps.Call{
			Path: "/bindings",
			Expand: &apps.Expand{
				Locale: apps.ExpandAll,
			},
		},
		Deploy: apps.Deploy{},
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
			apps.PermissionActAsAdmin,
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
		if creq.Context.AdminAccessToken == "" {
			return apps.NewErrorResponse(
				utils.NewUnauthorizedError("no admin token in the request"))
		}
		return h(creq)
	}
}
