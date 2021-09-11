package builtin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"runtime/debug"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
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
	fUserID         = "user"
)

const (
	pDebugBindings      = "/debug-bindings"
	pDebugClean         = "/debug-clean"
	pInfo               = "/info"
	pList               = "/list"
	pUninstall          = "/uninstall"
	pEnable             = "/enable"
	pDisable            = "/disable"
	pInstallURL         = "/install-url"
	pInstallS3          = "/install-s3"
	pInstallMarketplace = "/install-marketplace"
	pInstallConsent     = "/install-consent"
)

type builtinApp struct {
	conf    config.Service
	proxy   proxy.Service
	store   *store.Service
	httpOut httpout.Service

	router map[string]func(apps.CallRequest) apps.CallResponse
}

var _ upstream.Upstream = (*builtinApp)(nil)

func NewBuiltinApp(conf config.Service, proxy proxy.Service, store *store.Service, httpOut httpout.Service) *builtinApp {
	a := &builtinApp{
		conf:    conf,
		proxy:   proxy,
		store:   store,
		httpOut: httpOut,
		router:  map[string]func(apps.CallRequest) apps.CallResponse{},
	}

	a.router[apps.DefaultBindings.Path] = a.getBindings

	// Actions available to all users
	a.handle(pInfo, false, a.info)

	// Actions that require sysadmin
	a.handle(pDebugBindings, SysadminOnly, a.debugBindings)
	a.handle(pDebugClean, SysadminOnly, a.debugClean)
	a.handle(pList, SysadminOnly, a.list)
	a.withLookup(pDisable, SysadminOnly, a.disableSubmit, a.disableLookup)
	a.withLookup(pEnable, SysadminOnly, a.enableSubmit, a.enableLookup)
	a.withLookup(pInstallMarketplace, SysadminOnly, a.installMarketplaceSubmit, a.installMarketplaceLookup)
	a.withLookup(pInstallS3, SysadminOnly, a.installS3Submit, a.installS3Lookup)
	a.withLookup(pUninstall, SysadminOnly, a.uninstallSubmit, a.uninstallLookup)
	a.handle(pInstallURL, SysadminOnly, a.installURLSubmit)
	a.withForm(pInstallConsent, SysadminOnly, a.installConsentSubmit, a.installConsentForm)

	return a
}

func Manifest(conf config.Config) apps.Manifest {
	return apps.Manifest{
		AppID:       AppID,
		AppType:     apps.AppTypeBuiltin,
		Version:     apps.AppVersion(conf.BuildConfig.BuildHashShort),
		DisplayName: AppDisplayName,
		Description: AppDescription,
	}
}

func App(conf config.Config) apps.App {
	return apps.App{
		Manifest:    Manifest(conf),
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
			data, errr := json.Marshal(apps.CallResponse{
				Type:     apps.CallResponseTypeOK,
				Markdown: txt,
			})
			if errr != nil {
				err = errr
				return
			}
			err = nil
			out = ioutil.NopCloser(bytes.NewReader(data))
		}
	}(a.conf.Logger())

	f, ok := a.router[creq.Path]
	if !ok {
		return nil, utils.NewNotFoundError("%s is not found", creq.Path)
	}
	cresp := f(creq)
	data, err := json.Marshal(cresp)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(data)), nil
}

func (a *builtinApp) GetStatic(_ apps.App, path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.NewNotFoundError("static support is not implemented")
}

func mdResponse(format string, args ...interface{}) apps.CallResponse {
	return apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: fmt.Sprintf(format, args...),
	}
}

func formResponse(form apps.Form) apps.CallResponse {
	return apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: &form,
	}
}

func dataResponse(data interface{}) apps.CallResponse {
	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: data,
	}
}

func emptyForm(_ apps.CallRequest) apps.CallResponse {
	return formResponse(apps.Form{})
}

func submitPath(p string) string {
	return path.Join(p, "submit")
}

func formPath(p string) string {
	return path.Join(p, "form")
}

func lookupPath(p string) string {
	return path.Join(p, "lookup")
}

const SysadminOnly = true

func (a *builtinApp) handle(
	path string,
	sysadminOnly bool,
	submitf func(apps.CallRequest) apps.CallResponse,
) {
	a.router[submitPath(path)] = requireSysadmin(sysadminOnly, submitf)
}

func (a *builtinApp) withLookup(
	path string,
	sysadminOnly bool,
	submitf func(apps.CallRequest) apps.CallResponse,
	lookupf func(apps.CallRequest) ([]apps.SelectOption, error),
) {
	a.handle(path, sysadminOnly, submitf)

	if lookupf == nil {
		return
	}
	type lookupResponse struct {
		Items []apps.SelectOption `json:"items"`
	}
	a.router[lookupPath(path)] = requireSysadmin(sysadminOnly,
		func(creq apps.CallRequest) apps.CallResponse {
			opts, err := lookupf(creq)
			if err != nil {
				return apps.NewErrorCallResponse(err)
			}
			return dataResponse(lookupResponse{opts})
		})
}

func (a *builtinApp) withForm(
	path string,
	sysadminOnly bool,
	submitf func(apps.CallRequest) apps.CallResponse,
	formf func(apps.CallRequest) (*apps.Form, error),
) {
	a.handle(path, sysadminOnly, submitf)

	if formf == nil {
		return
	}
	a.router[lookupPath(path)] = requireSysadmin(sysadminOnly,
		func(creq apps.CallRequest) apps.CallResponse {
			form, err := formf(creq)
			if err != nil {
				return apps.NewErrorCallResponse(err)
			}
			return formResponse(*form)
		})
}

func requireSysadmin(require bool, handler func(apps.CallRequest) apps.CallResponse) func(apps.CallRequest) apps.CallResponse {
	if !require {
		return handler
	}
	return func(creq apps.CallRequest) apps.CallResponse {
		if creq.Context.ExpandedContext.AdminAccessToken == "" {
			return apps.NewErrorCallResponse(utils.NewUnauthorizedError("no admin token in the request"))
		}
		return handler(creq)
	}
}
