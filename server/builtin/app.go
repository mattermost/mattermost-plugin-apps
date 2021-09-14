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
	"strings"

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
	fDeployType     = "deploy_type"
	fUserID         = "user"
)

const (
	pDebugBindings  = "/debug-bindings"
	pDebugClean     = "/debug-clean"
	pInfo           = "/info"
	pList           = "/list"
	pUninstall      = "/uninstall"
	pEnable         = "/enable"
	pDisable        = "/disable"
	pInstallURL     = "/install-url"
	pInstallListed  = "/install-listed"
	pInstallConsent = "/install-consent"
)

type handler struct {
	requireSysadmin bool
	commandBinding  func() apps.Binding
	lookupf         func(apps.CallRequest) ([]apps.SelectOption, error)
	submitf         func(apps.CallRequest) apps.CallResponse
	formf           func(apps.CallRequest) (*apps.Form, error)
}

type builtinApp struct {
	conf    config.Service
	proxy   proxy.Service
	store   *store.Service
	httpOut httpout.Service
	router  map[string]handler
}

var _ upstream.Upstream = (*builtinApp)(nil)

func NewBuiltinApp(conf config.Service, proxy proxy.Service, store *store.Service, httpOut httpout.Service) *builtinApp {
	a := &builtinApp{
		conf:    conf,
		proxy:   proxy,
		store:   store,
		httpOut: httpOut,
	}

	a.router = map[string]handler{
		// Actions available to all users
		pInfo: a.info(),

		// Actions that require sysadmin
		pDebugBindings:  a.debugBindings(),
		pDebugClean:     a.debugClean(),
		pDisable:        a.disable(),
		pEnable:         a.enable(),
		pInstallConsent: a.installConsent(),
		pInstallListed:  a.installListed(),
		pInstallURL:     a.installURL(),
		pList:           a.list(),
		pUninstall:      a.uninstall(),
	}

	return a
}

func Manifest(conf config.Config) apps.Manifest {
	return apps.Manifest{
		AppID:       AppID,
		Version:     apps.AppVersion(conf.BuildConfig.BuildHashShort),
		DisplayName: AppDisplayName,
		Description: AppDescription,
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

	readcloser := func(cresp apps.CallResponse) (io.ReadCloser, error) {
		data, err := json.Marshal(cresp)
		if err != nil {
			return nil, err
		}
		return ioutil.NopCloser(bytes.NewReader(data)), nil
	}

	// The bindings call does not have a call type, so make it into a submit so
	// that the router can handle it.
	if creq.Path == apps.DefaultBindings.Path {
		return readcloser(a.bindings(creq))
	}

	callPath, callType := path.Split(creq.Path)
	callPath = strings.TrimRight(callPath, "/")
	h, ok := a.router[callPath]
	if !ok {
		return nil, utils.NewNotFoundError(callPath)
	}
	if h.requireSysadmin && creq.Context.AdminAccessToken == "" {
		return nil, apps.NewErrorCallResponse(utils.NewUnauthorizedError("no admin token in the request"))
	}

	switch apps.CallType(callType) {
	case apps.CallTypeForm:
		if h.formf == nil {
			return nil, utils.ErrNotFound
		}
		form, err := h.formf(creq)
		if err != nil {
			return nil, err
		}
		return readcloser(formResponse(*form))

	case apps.CallTypeLookup:
		if h.lookupf == nil {
			return nil, utils.ErrNotFound
		}
		opts, err := h.lookupf(creq)
		if err != nil {
			return nil, err
		}
		return readcloser(dataResponse(struct {
			Items []apps.SelectOption `json:"items"`
		}{opts}))

	case apps.CallTypeSubmit:
		if h.submitf == nil {
			return nil, utils.ErrNotFound
		}
		return readcloser(h.submitf(creq))
	}

	return nil, utils.NewNotFoundError("%s does not handle %s", callPath, callType)
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
